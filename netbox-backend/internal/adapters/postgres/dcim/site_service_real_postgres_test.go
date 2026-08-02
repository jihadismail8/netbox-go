package dcim

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"netbox-go/internal/adapters/postgres/bootstrap"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	identityrow "netbox-go/internal/adapters/postgres/identity"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestSiteServiceConcurrentDuplicateCreatePostgres(t *testing.T) {
	db, principal := newSiteConcurrencyPostgres(t)
	service, err := applicationdcim.NewSiteService(
		NewSiteRepository(db),
		postgresTransaction.NewUnitOfWork(db),
		postgreschangelog.NewRecorder(db),
		authz.AllowAll{},
		postgresConcurrencyClock{now: shared.NewTimestamp(time.Now().UTC())},
	)
	require.NoError(t, err)

	type result struct {
		site *domaindcim.Site
		err  error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	var workers sync.WaitGroup
	for index := range 2 {
		index := index
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			site, createErr := service.CreateSite(
				t.Context(), principal, applicationdcim.CreateSiteCommand{
					Name: applicationdcim.FieldValue("Concurrent Site"),
					Slug: applicationdcim.FieldValue("concurrent-site"),
					Description: applicationdcim.FieldValue(
						fmt.Sprintf("writer-%d", index),
					),
				},
			)
			results <- result{site: site, err: createErr}
		}()
	}
	close(start)
	workers.Wait()
	close(results)

	successes, conflicts := 0, 0
	for createResult := range results {
		if createResult.err == nil {
			require.NotNil(t, createResult.site)
			successes++
			continue
		}
		require.True(t, shared.HasReason(createResult.err, shared.ErrorReasonConflict), createResult.err)
		conflicts++
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, conflicts)

	var sites int64
	require.NoError(t, db.Model(&dcimrow.SiteRow{}).Count(&sites).Error)
	require.Equal(t, int64(1), sites)
	var changes int64
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).
		Where("kind = ? AND action = ?", domaindcim.SiteObjectType, "create").
		Count(&changes).Error)
	require.Equal(t, int64(1), changes, "the rejected duplicate must not leave a change row")
}

func newSiteConcurrencyPostgres(t *testing.T) (*gorm.DB, identity.Principal) {
	t.Helper()
	dsn := os.Getenv("NETBOX_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("NETBOX_TEST_POSTGRES_DSN is not set")
	}
	base, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	require.NoError(t, err)
	baseSQL, err := base.DB()
	require.NoError(t, err)
	schema := fmt.Sprintf("dcim_site_concurrency_%d", time.Now().UnixNano())
	require.NoError(t, base.Exec(`CREATE SCHEMA "`+schema+`"`).Error)

	db, err := gorm.Open(
		postgres.Open(dcimPostgresDSNWithSearchPath(t, dsn, schema)),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(8)
	sqlDB.SetMaxIdleConns(8)
	t.Cleanup(func() {
		_ = sqlDB.Close()
		_ = base.Exec(`DROP SCHEMA IF EXISTS "` + schema + `" CASCADE`).Error
		_ = baseSQL.Close()
	})

	entries := []bootstrap.Entry{{Name: "go_identity_users", Model: &identityrow.UserRow{}}}
	changeDependencies := []string{"go_identity_users"}
	for _, descriptor := range dcimrow.Descriptors() {
		entries = append(entries, bootstrap.Entry{
			Name: descriptor.Name, Model: descriptor.Model,
			Dependencies: descriptor.Dependencies,
		})
		changeDependencies = append(changeDependencies, descriptor.Name)
	}
	entries = append(entries, bootstrap.Entry{
		Name: "go_object_changes", Model: &postgreschangelog.ChangeRow{},
		Dependencies: changeDependencies,
	})
	registry, err := bootstrap.NewRegistry(entries...)
	require.NoError(t, err)
	_, err = bootstrap.Run(t.Context(), db, registry)
	require.NoError(t, err)
	now := time.Now().UTC()
	require.NoError(t, db.Create(&identityrow.UserRow{
		ID: 1, Username: "postgres-test", Email: "postgres-test@invalid.example",
		PasswordHash: "not-used-by-tests", IsStaff: true, IsSuperuser: true,
		IsActive: true, Permissions: []byte(`[]`), Created: now, Updated: now,
	}).Error)
	return db, identity.Principal{ID: 1, Username: "postgres-test", IsSuperuser: true}
}

func dcimPostgresDSNWithSearchPath(t *testing.T, dsn, schema string) string {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err == nil && (parsed.Scheme == "postgres" || parsed.Scheme == "postgresql") {
		query := parsed.Query()
		query.Set("search_path", schema)
		parsed.RawQuery = query.Encode()
		return parsed.String()
	}
	if strings.Contains(dsn, "=") {
		return strings.TrimSpace(dsn) + " search_path=" + schema
	}
	t.Fatalf("NETBOX_TEST_POSTGRES_DSN must be a postgres URL or keyword DSN")
	return ""
}

type postgresConcurrencyClock struct{ now shared.Timestamp }

func (clock postgresConcurrencyClock) Now() shared.Timestamp { return clock.now }
