package identity_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	postgresbootstrap "netbox-go/internal/adapters/postgres/bootstrap"
	identitypostgres "netbox-go/internal/adapters/postgres/identity"
	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

type coordinatedCredentialStore struct {
	application.Store
	ready   chan<- struct{}
	release <-chan struct{}
}

func (store *coordinatedCredentialStore) TokenByHash(ctx context.Context, hash []byte) (application.TokenRecord, domain.User, error) {
	record, user, err := store.Store.TokenByHash(ctx, hash)
	if err != nil {
		return application.TokenRecord{}, domain.User{}, err
	}
	store.ready <- struct{}{}
	<-store.release
	return record, user, nil
}

func TestPostgresTokenCredentialDurability(t *testing.T) {
	db := newCredentialPostgres(t)
	now := time.Date(2026, 8, 17, 6, 30, 0, 0, time.UTC)
	clock := &testClock{now: now}
	store := identitypostgres.NewStore(db)
	service := application.NewService(store, clock)

	userRow := identitypostgres.UserRow{
		Username: "credential-postgres-user", PasswordHash: "not-used-by-this-test",
		IsActive: true, Permissions: []byte(`[]`), Created: now, Updated: now,
	}
	require.NoError(t, db.Create(&userRow).Error)
	principal := domain.Principal{ID: userRow.ID, Username: userRow.Username}
	created, err := service.CreateToken(t.Context(), principal, application.CreateTokenInput{
		WriteEnabled: true,
	})
	require.NoError(t, err)
	installCredentialTokenAudit(t, db)

	t.Run("unknown key performs no durable write", func(t *testing.T) {
		resetCredentialTokenAudit(t, db)
		before := credentialTokenRow(t, db, created.Token.ID)
		require.Nil(t, before.LastUsed)

		user, authenticateErr := service.AuthenticateToken(t.Context(), "unknown-test-material", "192.0.2.1", false)

		require.Equal(t, domain.User{}, user)
		require.Equal(t, shared.ErrorReasonUnauthenticated, shared.ReasonOf(authenticateErr))
		after := credentialTokenRow(t, db, created.Token.ID)
		require.Nil(t, after.LastUsed)
		require.Equal(t, before.RevokedAt, after.RevokedAt)
		require.Equal(t, 0, credentialTokenUpdateCount(t, db))
	})

	t.Run("orphaned owner is internal without durable touch", func(t *testing.T) {
		orphaned, createErr := service.CreateToken(t.Context(), principal, application.CreateTokenInput{
			WriteEnabled: true,
		})
		require.NoError(t, createErr)
		missingUserID := userRow.ID + 1_000_000
		var matchingUsers int64
		require.NoError(t, db.Model(&identitypostgres.UserRow{}).
			Where("id = ?", missingUserID).
			Count(&matchingUsers).Error)
		require.Zero(t, matchingUsers)

		orphanTx := db.Begin()
		require.NoError(t, orphanTx.Error)
		rollbackPending := true
		t.Cleanup(func() {
			if rollbackPending {
				require.NoError(t, orphanTx.Rollback().Error)
			}
		})
		require.NoError(t, orphanTx.Exec(`
DO $fixture$
DECLARE
    owner_fk name;
BEGIN
    SELECT conname INTO STRICT owner_fk
    FROM pg_constraint
    WHERE conrelid = 'go_identity_tokens'::regclass
      AND confrelid = 'go_identity_users'::regclass
      AND contype = 'f';
    EXECUTE format('ALTER TABLE go_identity_tokens DROP CONSTRAINT %I', owner_fk);
END
$fixture$;`).Error)
		require.NoError(t, orphanTx.Model(&identitypostgres.TokenRow{}).
			Where("id = ?", orphaned.Token.ID).
			Update("user_id", missingUserID).Error)
		resetCredentialTokenAudit(t, orphanTx)
		before := credentialTokenRow(t, orphanTx, orphaned.Token.ID)
		require.Nil(t, before.LastUsed)
		orphanService := application.NewService(
			identitypostgres.NewStore(orphanTx),
			&testClock{now: now},
		)

		user, authenticateErr := orphanService.AuthenticateToken(
			t.Context(), orphaned.Secret, "192.0.2.1", false,
		)

		require.Equal(t, domain.User{}, user)
		require.Equal(t, shared.ErrorReasonInternal, shared.ReasonOf(authenticateErr))
		require.ErrorIs(t, authenticateErr, gorm.ErrRecordNotFound)
		after := credentialTokenRow(t, orphanTx, orphaned.Token.ID)
		require.Nil(t, after.LastUsed)
		require.Equal(t, before.UserID, after.UserID)
		require.Equal(t, 0, credentialTokenUpdateCount(t, orphanTx))
		require.NoError(t, orphanTx.Rollback().Error)
		rollbackPending = false
		restored := credentialTokenRow(t, db, orphaned.Token.ID)
		require.Equal(t, userRow.ID, restored.UserID)
		require.Nil(t, restored.LastUsed)
	})

	t.Run("recognized key initializes last used", func(t *testing.T) {
		resetCredentialTokenAudit(t, db)
		user, authenticateErr := service.AuthenticateToken(t.Context(), created.Secret, "192.0.2.1", false)

		require.NoError(t, authenticateErr)
		require.Equal(t, userRow.ID, user.ID)
		require.Equal(t, now, credentialLastUsed(t, db, created.Token.ID))
		require.Equal(t, 1, credentialTokenUpdateCount(t, db))
	})

	t.Run("application boundary is strict", func(t *testing.T) {
		resetCredentialTokenAudit(t, db)
		clock.now = now.Add(time.Minute)
		_, authenticateErr := service.AuthenticateToken(t.Context(), created.Secret, "192.0.2.1", false)
		require.NoError(t, authenticateErr)
		require.Equal(t, now, credentialLastUsed(t, db, created.Token.ID))
		require.Equal(t, 0, credentialTokenUpdateCount(t, db))

		clock.now = now.Add(time.Minute + time.Second)
		_, authenticateErr = service.AuthenticateToken(t.Context(), created.Secret, "192.0.2.1", false)
		require.NoError(t, authenticateErr)
		require.Equal(t, clock.now, credentialLastUsed(t, db, created.Token.ID))
		require.Equal(t, 1, credentialTokenUpdateCount(t, db))
	})

	t.Run("store boundary is strict", func(t *testing.T) {
		require.NoError(t, db.Model(&identitypostgres.TokenRow{}).
			Where("id = ?", created.Token.ID).
			Update("last_used", now).Error)
		resetCredentialTokenAudit(t, db)

		require.NoError(t, store.TouchToken(t.Context(), created.Token.ID, now.Add(time.Minute)))
		require.Equal(t, now, credentialLastUsed(t, db, created.Token.ID))
		require.Equal(t, 0, credentialTokenUpdateCount(t, db))

		require.NoError(t, store.TouchToken(t.Context(), created.Token.ID, now.Add(time.Minute+time.Second)))
		require.Equal(t, now.Add(time.Minute+time.Second), credentialLastUsed(t, db, created.Token.ID))
		require.Equal(t, 1, credentialTokenUpdateCount(t, db))
	})

	t.Run("store boundary preserves postgres timestamp precision", func(t *testing.T) {
		stored := now.Add(123456 * time.Microsecond)
		require.NoError(t, db.Model(&identitypostgres.TokenRow{}).
			Where("id = ?", created.Token.ID).
			Update("last_used", stored).Error)
		resetCredentialTokenAudit(t, db)

		require.NoError(t, store.TouchToken(t.Context(), created.Token.ID, stored.Add(time.Minute)))
		require.Equal(t, stored, credentialLastUsed(t, db, created.Token.ID))
		require.Equal(t, 0, credentialTokenUpdateCount(t, db))

		eligible := stored.Add(time.Minute + time.Microsecond)
		require.NoError(t, store.TouchToken(t.Context(), created.Token.ID, eligible))
		require.Equal(t, eligible, credentialLastUsed(t, db, created.Token.ID))
		require.Equal(t, 1, credentialTokenUpdateCount(t, db))
	})

	t.Run("concurrent stale callers produce one durable update", func(t *testing.T) {
		require.NoError(t, db.Model(&identitypostgres.TokenRow{}).
			Where("id = ?", created.Token.ID).
			Update("last_used", now).Error)
		resetCredentialTokenAudit(t, db)

		eligible := now.Add(time.Minute + time.Second)
		ready := make(chan struct{}, 2)
		release := make(chan struct{})
		results := make(chan error, 2)
		coordinatedStore := &coordinatedCredentialStore{
			Store: store, ready: ready, release: release,
		}
		coordinatedService := application.NewService(
			coordinatedStore,
			&testClock{now: eligible},
		)
		var workers sync.WaitGroup
		for range 2 {
			workers.Add(1)
			go func() {
				defer workers.Done()
				_, authenticateErr := coordinatedService.AuthenticateToken(
					t.Context(), created.Secret, "192.0.2.1", false,
				)
				results <- authenticateErr
			}()
		}
		for range 2 {
			<-ready
		}
		close(release)
		workers.Wait()
		close(results)
		for touchErr := range results {
			require.NoError(t, touchErr)
		}

		require.Equal(t, 1, credentialTokenUpdateCount(t, db))
		require.Equal(t, eligible, credentialLastUsed(t, db, created.Token.ID))
	})

	t.Run("revoked key performs no durable touch", func(t *testing.T) {
		require.NoError(t, db.Model(&identitypostgres.TokenRow{}).
			Where("id = ?", created.Token.ID).
			Update("last_used", now).Error)
		clock.now = now.Add(2 * time.Hour)
		require.NoError(t, service.RevokeToken(t.Context(), principal, created.Token.ID))
		before := credentialTokenRow(t, db, created.Token.ID)
		require.NotNil(t, before.RevokedAt)
		require.NotNil(t, before.LastUsed)
		resetCredentialTokenAudit(t, db)

		clock.now = clock.now.Add(2 * time.Minute)
		user, authenticateErr := service.AuthenticateToken(t.Context(), created.Secret, "192.0.2.1", false)

		require.Equal(t, domain.User{}, user)
		require.Equal(t, shared.ErrorReasonUnauthenticated, shared.ReasonOf(authenticateErr))
		after := credentialTokenRow(t, db, created.Token.ID)
		require.NotNil(t, after.LastUsed)
		require.Equal(t, before.LastUsed.UTC(), after.LastUsed.UTC())
		require.Equal(t, before.RevokedAt.UTC(), after.RevokedAt.UTC())

		require.NoError(t, store.TouchToken(t.Context(), created.Token.ID, clock.now.Add(time.Minute)))
		afterDirectTouch := credentialTokenRow(t, db, created.Token.ID)
		require.Equal(t, before.LastUsed.UTC(), afterDirectTouch.LastUsed.UTC())
		require.Equal(t, before.RevokedAt.UTC(), afterDirectTouch.RevokedAt.UTC())
		require.Equal(t, 0, credentialTokenUpdateCount(t, db))
	})

	t.Run("touch failure is fail closed", func(t *testing.T) {
		clock.now = now
		second, createErr := service.CreateToken(t.Context(), principal, application.CreateTokenInput{
			WriteEnabled: true,
		})
		require.NoError(t, createErr)
		require.NoError(t, db.Model(&identitypostgres.TokenRow{}).
			Where("id = ?", second.Token.ID).
			Update("last_used", now).Error)
		resetCredentialTokenAudit(t, db)

		readOnly := db.Begin()
		require.NoError(t, readOnly.Error)
		t.Cleanup(func() { _ = readOnly.Rollback().Error })
		require.NoError(t, readOnly.Exec("SET TRANSACTION READ ONLY").Error)
		readOnlyService := application.NewService(identitypostgres.NewStore(readOnly), &testClock{
			now: now.Add(2 * time.Minute),
		})

		user, authenticateErr := readOnlyService.AuthenticateToken(t.Context(), second.Secret, "192.0.2.1", false)

		require.Equal(t, domain.User{}, user)
		require.Equal(t, shared.ErrorReasonInternal, shared.ReasonOf(authenticateErr))
		require.NoError(t, readOnly.Rollback().Error)
		require.Equal(t, now, credentialLastUsed(t, db, second.Token.ID))
		require.Equal(t, 0, credentialTokenUpdateCount(t, db))
	})
}

func installCredentialTokenAudit(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec("CREATE TABLE credential_token_audit (updates integer NOT NULL)").Error)
	require.NoError(t, db.Exec("INSERT INTO credential_token_audit (updates) VALUES (0)").Error)
	require.NoError(t, db.Exec(`
CREATE FUNCTION count_credential_token_update() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    UPDATE credential_token_audit SET updates = updates + 1;
    RETURN NEW;
END;
$$`).Error)
	require.NoError(t, db.Exec(`
CREATE TRIGGER count_credential_token_update
AFTER UPDATE ON go_identity_tokens
FOR EACH ROW EXECUTE FUNCTION count_credential_token_update()`).Error)
}

func resetCredentialTokenAudit(t *testing.T, db *gorm.DB) {
	t.Helper()
	result := db.Exec("UPDATE credential_token_audit SET updates = 0")
	require.NoError(t, result.Error)
	require.Equal(t, int64(1), result.RowsAffected)
}

func credentialTokenUpdateCount(t *testing.T, db *gorm.DB) int {
	t.Helper()
	var updates int
	require.NoError(t, db.Raw("SELECT updates FROM credential_token_audit").Scan(&updates).Error)
	return updates
}

func newCredentialPostgres(t *testing.T) *gorm.DB {
	t.Helper()
	baseDSN := strings.TrimSpace(os.Getenv("NETBOX_TEST_POSTGRES_DSN"))
	if baseDSN == "" {
		t.Skip("NETBOX_TEST_POSTGRES_DSN is not set")
	}

	base, err := gorm.Open(
		postgres.Open(baseDSN),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	require.NoError(t, err)
	baseSQL, err := base.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, baseSQL.Close()) })
	schema := fmt.Sprintf("identity_credential_%d", time.Now().UnixNano())
	require.NoError(t, base.Exec(`CREATE SCHEMA "`+schema+`"`).Error)
	t.Cleanup(func() {
		require.NoError(t, base.Exec(`DROP SCHEMA "`+schema+`" CASCADE`).Error)
	})

	db, err := gorm.Open(
		postgres.Open(identityTestDSN(t, baseDSN, schema)),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(8)
	sqlDB.SetMaxIdleConns(8)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

	registry, err := postgresbootstrap.NewRegistry(
		postgresbootstrap.Entry{Name: "go_identity_users", Model: &identitypostgres.UserRow{}},
		postgresbootstrap.Entry{Name: "go_identity_groups", Model: &identitypostgres.GroupRow{}},
		postgresbootstrap.Entry{Name: "go_identity_permission_grants", Model: &identitypostgres.PermissionGrantRow{}},
		postgresbootstrap.Entry{Name: "go_identity_group_memberships", Model: &identitypostgres.GroupMembershipRow{}, Dependencies: []string{"go_identity_users", "go_identity_groups"}},
		postgresbootstrap.Entry{Name: "go_identity_user_permission_grants", Model: &identitypostgres.UserPermissionGrantRow{}, Dependencies: []string{"go_identity_users", "go_identity_permission_grants"}},
		postgresbootstrap.Entry{Name: "go_identity_group_permission_grants", Model: &identitypostgres.GroupPermissionGrantRow{}, Dependencies: []string{"go_identity_groups", "go_identity_permission_grants"}},
		postgresbootstrap.Entry{Name: "go_identity_tokens", Model: &identitypostgres.TokenRow{}, Dependencies: []string{"go_identity_users"}},
		postgresbootstrap.Entry{Name: "go_identity_sessions", Model: &identitypostgres.SessionRow{}, Dependencies: []string{"go_identity_users"}},
	)
	require.NoError(t, err)
	result, err := postgresbootstrap.Run(t.Context(), db, registry)
	require.NoError(t, err)
	require.Len(t, result.Created, 8)
	require.Empty(t, result.Existing)
	return db
}

func credentialTokenRow(t *testing.T, db *gorm.DB, id int64) identitypostgres.TokenRow {
	t.Helper()
	var row identitypostgres.TokenRow
	require.NoError(t, db.First(&row, id).Error)
	return row
}

func credentialLastUsed(t *testing.T, db *gorm.DB, id int64) time.Time {
	t.Helper()
	row := credentialTokenRow(t, db, id)
	require.NotNil(t, row.LastUsed)
	return row.LastUsed.UTC()
}
