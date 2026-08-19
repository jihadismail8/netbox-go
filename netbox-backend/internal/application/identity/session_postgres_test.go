package identity_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	identitypostgres "netbox-go/internal/adapters/postgres/identity"
	application "netbox-go/internal/application/identity"
	"netbox-go/internal/domain/shared"
)

const sessionPostgresCSRFDomain = "netbox-go/browser-csrf/v1"

var errSessionPostgresInjected = errors.New("injected session transaction failure")

type sessionPostgresResult struct {
	value string
	err   error
}

type sessionPostgresDMLTrace struct {
	creates atomic.Int64
	deletes atomic.Int64
	updates atomic.Int64
}

type sessionPostgresFault uint8

const (
	sessionPostgresFailAfterCreate sessionPostgresFault = iota + 1
	sessionPostgresFailAfterDelete
	sessionPostgresFailAfterUpdate
)

type sessionPostgresFaultStore struct {
	application.Store
	fault sessionPostgresFault
	trace *sessionPostgresDMLTrace
}

func (store *sessionPostgresFaultStore) Transaction(ctx context.Context, fn func(application.Store) error) error {
	return store.Store.Transaction(ctx, func(tx application.Store) error {
		return fn(&sessionPostgresFaultTransaction{
			Store: tx,
			fault: store.fault,
			trace: store.trace,
		})
	})
}

type sessionPostgresFaultTransaction struct {
	application.Store
	fault sessionPostgresFault
	trace *sessionPostgresDMLTrace
}

func (tx *sessionPostgresFaultTransaction) CreateSession(ctx context.Context, record application.SessionRecord) error {
	if err := tx.Store.CreateSession(ctx, record); err != nil {
		return err
	}
	tx.trace.creates.Add(1)
	if tx.fault == sessionPostgresFailAfterCreate {
		return errSessionPostgresInjected
	}
	return nil
}

func (tx *sessionPostgresFaultTransaction) DeleteSession(ctx context.Context, sessionHash []byte) error {
	if err := tx.Store.DeleteSession(ctx, sessionHash); err != nil {
		return err
	}
	tx.trace.deletes.Add(1)
	if tx.fault == sessionPostgresFailAfterDelete {
		return errSessionPostgresInjected
	}
	return nil
}

func (tx *sessionPostgresFaultTransaction) UpdateSessionCSRF(ctx context.Context, sessionHash, csrfHash []byte) error {
	if err := tx.Store.UpdateSessionCSRF(ctx, sessionHash, csrfHash); err != nil {
		return err
	}
	tx.trace.updates.Add(1)
	if tx.fault == sessionPostgresFailAfterUpdate {
		return errSessionPostgresInjected
	}
	return nil
}

type sessionPostgresHoldCommitStore struct {
	application.Store
	ready        chan<- struct{}
	release      <-chan struct{}
	transactions atomic.Int64
}

func (store *sessionPostgresHoldCommitStore) Transaction(ctx context.Context, fn func(application.Store) error) error {
	store.transactions.Add(1)
	return store.Store.Transaction(ctx, func(tx application.Store) error {
		if err := fn(tx); err != nil {
			return err
		}
		store.ready <- struct{}{}
		select {
		case <-store.release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}

type sessionPostgresCountingStore struct {
	application.Store
	transactions atomic.Int64
	entered      chan<- struct{}
}

func (store *sessionPostgresCountingStore) Transaction(ctx context.Context, fn func(application.Store) error) error {
	store.transactions.Add(1)
	if store.entered != nil {
		select {
		case store.entered <- struct{}{}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return store.Store.Transaction(ctx, fn)
}

type sessionPostgresLoginTransactionBarrierStore struct {
	application.Store
	verified     chan<- struct{}
	release      <-chan struct{}
	transactions atomic.Int64
}

func (store *sessionPostgresLoginTransactionBarrierStore) Transaction(ctx context.Context, fn func(application.Store) error) error {
	store.transactions.Add(1)
	store.verified <- struct{}{}
	select {
	case <-store.release:
		return store.Store.Transaction(ctx, fn)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestPostgresBrowserSessionRollback(t *testing.T) {
	db := newCredentialPostgres(t)
	clock := &testClock{now: time.Date(2026, 8, 17, 8, 0, 0, 0, time.UTC)}
	store := identitypostgres.NewStore(db)
	service := application.NewService(store, clock)

	_, err := service.BootstrapAdministrator(
		t.Context(), "session-rollback-admin", "", "Session-Rollback-2026!",
	)
	require.NoError(t, err)
	prior, err := service.Login(t.Context(), "session-rollback-admin", "Session-Rollback-2026!")
	require.NoError(t, err)

	loginTrace := &sessionPostgresDMLTrace{}
	failedLoginService := application.NewService(&sessionPostgresFaultStore{
		Store: store,
		fault: sessionPostgresFailAfterCreate,
		trace: loginTrace,
	}, clock)
	failedSession, err := failedLoginService.LoginReplacing(
		t.Context(), "session-rollback-admin", "Session-Rollback-2026!", prior.Secret,
	)
	require.Error(t, err)
	require.Equal(t, shared.ErrorReasonInternal, shared.ReasonOf(err))
	require.ErrorIs(t, err, errSessionPostgresInjected)
	if failedSession.Secret != "" || failedSession.CSRFToken != "" ||
		failedSession.User.ID != 0 || !failedSession.Expires.IsZero() {
		t.Fatal("failed replacing login returned session material")
	}
	require.Equal(t, int64(1), loginTrace.deletes.Load())
	require.Equal(t, int64(1), loginTrace.creates.Load())
	require.Equal(t, int64(1), sessionPostgresSessionCount(t, db))
	_, err = service.AuthenticateSession(t.Context(), prior.Secret)
	require.NoError(t, err)

	logoutTrace := &sessionPostgresDMLTrace{}
	failedLogoutService := application.NewService(&sessionPostgresFaultStore{
		Store: store,
		fault: sessionPostgresFailAfterDelete,
		trace: logoutTrace,
	}, clock)
	err = failedLogoutService.Logout(t.Context(), prior.Secret, prior.CSRFToken)
	require.Error(t, err)
	require.Equal(t, shared.ErrorReasonInternal, shared.ReasonOf(err))
	require.ErrorIs(t, err, errSessionPostgresInjected)
	require.Equal(t, int64(1), logoutTrace.deletes.Load())
	require.Equal(t, int64(1), sessionPostgresSessionCount(t, db))
	_, err = service.AuthenticateSession(t.Context(), prior.Secret)
	require.NoError(t, err)
	require.NoError(t, service.VerifyCSRF(t.Context(), prior.Secret, prior.CSRFToken))
}

func TestPostgresSessionCSRFRecoveryDurability(t *testing.T) {
	db := newCredentialPostgres(t)
	clock := &testClock{now: time.Date(2026, 8, 17, 8, 30, 0, 0, time.UTC)}
	store := identitypostgres.NewStore(db)
	service := application.NewService(store, clock)

	_, err := service.BootstrapAdministrator(
		t.Context(), "csrf-recovery-admin", "", "CSRF-Recovery-2026!",
	)
	require.NoError(t, err)
	installSessionPostgresUpdateAudit(t, db)

	t.Run("legacy recovery is durable and stable", func(t *testing.T) {
		session, loginErr := service.Login(t.Context(), "csrf-recovery-admin", "CSRF-Recovery-2026!")
		require.NoError(t, loginErr)
		legacyCSRF := "pre-i3-csrf-fixture-first"
		expected := sessionPostgresDerivedCSRF(session.Secret)
		require.False(t, sessionPostgresEqualText(legacyCSRF, expected))
		sessionPostgresSetCSRF(t, db, session.Secret, legacyCSRF)
		resetSessionPostgresUpdateAudit(t, db)

		recovered, recoveryErr := service.CSRFForSession(t.Context(), session.Secret)
		require.NoError(t, recoveryErr)
		require.True(t, sessionPostgresEqualText(recovered, expected))
		require.True(t, sessionPostgresStoredCSRFMatches(t, db, session.Secret, recovered))
		require.Equal(t, 1, sessionPostgresUpdateCount(t, db))
		require.NoError(t, service.VerifyCSRF(t.Context(), session.Secret, recovered))

		resetSessionPostgresUpdateAudit(t, db)
		recoveredAgain, recoveryErr := service.CSRFForSession(t.Context(), session.Secret)
		require.NoError(t, recoveryErr)
		require.True(t, sessionPostgresEqualText(recoveredAgain, expected))
		require.Equal(t, 0, sessionPostgresUpdateCount(t, db))
	})

	t.Run("post update failure rolls back and returns no value", func(t *testing.T) {
		session, loginErr := service.Login(t.Context(), "csrf-recovery-admin", "CSRF-Recovery-2026!")
		require.NoError(t, loginErr)
		legacyCSRF := "pre-i3-csrf-fixture-rollback"
		sessionPostgresSetCSRF(t, db, session.Secret, legacyCSRF)
		resetSessionPostgresUpdateAudit(t, db)
		trace := &sessionPostgresDMLTrace{}
		failedRecoveryService := application.NewService(&sessionPostgresFaultStore{
			Store: store,
			fault: sessionPostgresFailAfterUpdate,
			trace: trace,
		}, clock)

		recovered, recoveryErr := failedRecoveryService.CSRFForSession(t.Context(), session.Secret)
		require.Error(t, recoveryErr)
		require.Equal(t, shared.ErrorReasonInternal, shared.ReasonOf(recoveryErr))
		require.ErrorIs(t, recoveryErr, errSessionPostgresInjected)
		if recovered != "" {
			t.Fatal("failed CSRF recovery returned credential material")
		}
		require.Equal(t, int64(1), trace.updates.Load())
		require.Equal(t, 0, sessionPostgresUpdateCount(t, db))
		require.True(t, sessionPostgresStoredCSRFMatches(t, db, session.Secret, legacyCSRF))
		require.NoError(t, service.VerifyCSRF(t.Context(), session.Secret, legacyCSRF))
	})

	t.Run("zero row update is not found", func(t *testing.T) {
		updateErr := store.UpdateSessionCSRF(
			t.Context(),
			sessionPostgresDigest("absent-session-fixture"),
			sessionPostgresDigest("absent-csrf-fixture"),
		)
		if !errors.Is(updateErr, application.ErrNotFound) {
			t.Fatal("zero-row CSRF update did not return not found")
		}
	})

	t.Run("recovery commits before old CSRF logout", func(t *testing.T) {
		session, loginErr := service.Login(t.Context(), "csrf-recovery-admin", "CSRF-Recovery-2026!")
		require.NoError(t, loginErr)
		legacyCSRF := "pre-i3-csrf-fixture-recovery-first"
		sessionPostgresSetCSRF(t, db, session.Secret, legacyCSRF)

		recoveryReady := make(chan struct{}, 1)
		releaseRecovery := make(chan struct{})
		var releaseRecoveryOnce sync.Once
		releaseRecoveryTransaction := func() {
			releaseRecoveryOnce.Do(func() { close(releaseRecovery) })
		}
		defer releaseRecoveryTransaction()
		holdingStore := &sessionPostgresHoldCommitStore{
			Store:   store,
			ready:   recoveryReady,
			release: releaseRecovery,
		}
		recoveryService := application.NewService(holdingStore, clock)
		recoveryResult := make(chan sessionPostgresResult, 1)
		go func() {
			value, recoveryErr := recoveryService.CSRFForSession(t.Context(), session.Secret)
			recoveryResult <- sessionPostgresResult{value: value, err: recoveryErr}
		}()
		sessionPostgresReceive(t, recoveryReady)

		pinnedLogoutStore, logoutBackendPID := sessionPostgresPinnedStore(t, db)
		logoutEntered := make(chan struct{}, 1)
		logoutStore := &sessionPostgresCountingStore{
			Store:   pinnedLogoutStore,
			entered: logoutEntered,
		}
		logoutService := application.NewService(logoutStore, clock)
		logoutResult := make(chan error, 1)
		go func() {
			logoutResult <- logoutService.Logout(t.Context(), session.Secret, legacyCSRF)
		}()
		sessionPostgresReceive(t, logoutEntered)
		sessionPostgresAwaitAdvisoryWait(t, db, logoutBackendPID)
		releaseRecoveryTransaction()

		recoveryOutcome := sessionPostgresReceive(t, recoveryResult)
		require.NoError(t, recoveryOutcome.err)
		require.True(t, sessionPostgresEqualText(
			recoveryOutcome.value,
			sessionPostgresDerivedCSRF(session.Secret),
		))
		logoutErr := sessionPostgresReceive(t, logoutResult)
		require.Equal(t, shared.ErrorReasonForbidden, shared.ReasonOf(logoutErr))
		require.Equal(t, int64(1), holdingStore.transactions.Load())
		require.Equal(t, int64(1), logoutStore.transactions.Load())
		require.True(t, sessionPostgresSessionExists(t, db, session.Secret))
		_, authenticateErr := service.AuthenticateSession(t.Context(), session.Secret)
		require.NoError(t, authenticateErr)
	})

	t.Run("logout commits before recovery", func(t *testing.T) {
		session, loginErr := service.Login(t.Context(), "csrf-recovery-admin", "CSRF-Recovery-2026!")
		require.NoError(t, loginErr)

		logoutReady := make(chan struct{}, 1)
		releaseLogout := make(chan struct{})
		var releaseLogoutOnce sync.Once
		releaseLogoutTransaction := func() {
			releaseLogoutOnce.Do(func() { close(releaseLogout) })
		}
		defer releaseLogoutTransaction()
		holdingStore := &sessionPostgresHoldCommitStore{
			Store:   store,
			ready:   logoutReady,
			release: releaseLogout,
		}
		logoutService := application.NewService(holdingStore, clock)
		logoutResult := make(chan error, 1)
		go func() {
			logoutResult <- logoutService.Logout(
				t.Context(), session.Secret, session.CSRFToken,
			)
		}()
		sessionPostgresReceive(t, logoutReady)

		pinnedRecoveryStore, recoveryBackendPID := sessionPostgresPinnedStore(t, db)
		recoveryEntered := make(chan struct{}, 1)
		recoveryStore := &sessionPostgresCountingStore{
			Store:   pinnedRecoveryStore,
			entered: recoveryEntered,
		}
		recoveryService := application.NewService(recoveryStore, clock)
		recoveryResult := make(chan sessionPostgresResult, 1)
		go func() {
			value, recoveryErr := recoveryService.CSRFForSession(t.Context(), session.Secret)
			recoveryResult <- sessionPostgresResult{value: value, err: recoveryErr}
		}()
		sessionPostgresReceive(t, recoveryEntered)
		sessionPostgresAwaitAdvisoryWait(t, db, recoveryBackendPID)
		releaseLogoutTransaction()

		require.NoError(t, sessionPostgresReceive(t, logoutResult))
		recoveryOutcome := sessionPostgresReceive(t, recoveryResult)
		if recoveryOutcome.value != "" {
			t.Fatal("unknown-session CSRF recovery returned credential material")
		}
		require.True(t, sessionPostgresHasFailure(
			recoveryOutcome.err,
			application.SessionCredentialFailureUnknown,
		))
		require.Equal(t, int64(1), holdingStore.transactions.Load())
		require.Equal(t, int64(1), recoveryStore.transactions.Load())
		require.False(t, sessionPostgresSessionExists(t, db, session.Secret))
	})
}

func TestPostgresLoginRevalidatesAgainstConcurrentPasswordReset(t *testing.T) {
	db := newCredentialPostgres(t)
	clock := &testClock{now: time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)}
	store := identitypostgres.NewStore(db)
	service := application.NewService(store, clock)

	_, err := service.BootstrapAdministrator(
		t.Context(), "login-reset-admin", "", "Login-Before-Reset-2026!",
	)
	require.NoError(t, err)

	passwordVerified := make(chan struct{}, 1)
	releaseLogin := make(chan struct{})
	barrierStore := &sessionPostgresLoginTransactionBarrierStore{
		Store:    store,
		verified: passwordVerified,
		release:  releaseLogin,
	}
	loginService := application.NewService(barrierStore, clock)
	loginResult := make(chan struct {
		sessionMaterial bool
		err             error
	}, 1)
	go func() {
		session, loginErr := loginService.Login(
			t.Context(), "login-reset-admin", "Login-Before-Reset-2026!",
		)
		loginResult <- struct {
			sessionMaterial bool
			err             error
		}{
			sessionMaterial: session.Secret != "" || session.CSRFToken != "" ||
				session.User.ID != 0 || !session.Expires.IsZero(),
			err: loginErr,
		}
	}()
	sessionPostgresReceive(t, passwordVerified)

	require.NoError(t, service.ResetAdministratorPassword(
		t.Context(), "login-reset-admin", "Login-After-Reset-2026!",
	))
	close(releaseLogin)
	outcome := sessionPostgresReceive(t, loginResult)
	if outcome.sessionMaterial {
		t.Fatal("stale-password login returned session material")
	}
	require.Equal(t, shared.ErrorReasonUnauthenticated, shared.ReasonOf(outcome.err))
	require.Equal(t, int64(1), barrierStore.transactions.Load())
	require.Equal(t, int64(0), sessionPostgresSessionCount(t, db))

	_, err = service.AuthenticatePassword(
		t.Context(), "login-reset-admin", "Login-Before-Reset-2026!",
	)
	require.Equal(t, shared.ErrorReasonUnauthenticated, shared.ReasonOf(err))
	_, err = service.AuthenticatePassword(
		t.Context(), "login-reset-admin", "Login-After-Reset-2026!",
	)
	require.NoError(t, err)
}

func TestPostgresConcurrentLogoutHasSingleRevocation(t *testing.T) {
	db := newCredentialPostgres(t)
	clock := &testClock{now: time.Date(2026, 8, 17, 9, 30, 0, 0, time.UTC)}
	store := identitypostgres.NewStore(db)
	service := application.NewService(store, clock)

	_, err := service.BootstrapAdministrator(
		t.Context(), "concurrent-logout-admin", "", "Concurrent-Logout-2026!",
	)
	require.NoError(t, err)
	session, err := service.Login(
		t.Context(), "concurrent-logout-admin", "Concurrent-Logout-2026!",
	)
	require.NoError(t, err)
	installSessionPostgresDeleteAudit(t, db)

	ready := make(chan struct{}, 2)
	start := make(chan struct{})
	var startOnce sync.Once
	startWorkers := func() {
		startOnce.Do(func() { close(start) })
	}
	defer startWorkers()
	results := make(chan error, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			ready <- struct{}{}
			<-start
			results <- service.Logout(t.Context(), session.Secret, session.CSRFToken)
		}()
	}
	sessionPostgresReceive(t, ready)
	sessionPostgresReceive(t, ready)
	startWorkers()
	workersDone := make(chan struct{})
	go func() {
		workers.Wait()
		close(workersDone)
	}()
	sessionPostgresReceive(t, workersDone)
	close(results)

	succeeded := 0
	unknown := 0
	for logoutErr := range results {
		switch {
		case logoutErr == nil:
			succeeded++
		case sessionPostgresHasFailure(logoutErr, application.SessionCredentialFailureUnknown):
			unknown++
		default:
			t.Fatal("concurrent logout returned an unexpected classified outcome")
		}
	}
	require.Equal(t, 1, succeeded)
	require.Equal(t, 1, unknown)
	require.Equal(t, int64(0), sessionPostgresSessionCount(t, db))
	require.Equal(t, 1, sessionPostgresDeleteCount(t, db))
}

func installSessionPostgresUpdateAudit(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec("CREATE TABLE session_csrf_audit (updates integer NOT NULL)").Error)
	require.NoError(t, db.Exec("INSERT INTO session_csrf_audit (updates) VALUES (0)").Error)
	require.NoError(t, db.Exec(`
CREATE FUNCTION count_session_csrf_update() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    UPDATE session_csrf_audit SET updates = updates + 1;
    RETURN NEW;
END;
$$`).Error)
	require.NoError(t, db.Exec(`
CREATE TRIGGER count_session_csrf_update
AFTER UPDATE OF csrf_hash ON go_identity_sessions
FOR EACH ROW EXECUTE FUNCTION count_session_csrf_update()`).Error)
}

func resetSessionPostgresUpdateAudit(t *testing.T, db *gorm.DB) {
	t.Helper()
	result := db.Exec("UPDATE session_csrf_audit SET updates = 0")
	require.NoError(t, result.Error)
	require.Equal(t, int64(1), result.RowsAffected)
}

func sessionPostgresUpdateCount(t *testing.T, db *gorm.DB) int {
	t.Helper()
	var updates int
	require.NoError(t, db.Raw("SELECT updates FROM session_csrf_audit").Scan(&updates).Error)
	return updates
}

func installSessionPostgresDeleteAudit(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec("CREATE TABLE session_delete_audit (deletes integer NOT NULL)").Error)
	require.NoError(t, db.Exec("INSERT INTO session_delete_audit (deletes) VALUES (0)").Error)
	require.NoError(t, db.Exec(`
CREATE FUNCTION count_session_delete() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    UPDATE session_delete_audit SET deletes = deletes + 1;
    RETURN OLD;
END;
$$`).Error)
	require.NoError(t, db.Exec(`
CREATE TRIGGER count_session_delete
AFTER DELETE ON go_identity_sessions
FOR EACH ROW EXECUTE FUNCTION count_session_delete()`).Error)
}

func sessionPostgresDeleteCount(t *testing.T, db *gorm.DB) int {
	t.Helper()
	var deletes int
	require.NoError(t, db.Raw("SELECT deletes FROM session_delete_audit").Scan(&deletes).Error)
	return deletes
}

func sessionPostgresSetCSRF(t *testing.T, db *gorm.DB, sessionSecret, csrf string) {
	t.Helper()
	result := db.Model(&identitypostgres.SessionRow{}).
		Where("secret_hash = ?", sessionPostgresDigest(sessionSecret)).
		Update("csrf_hash", sessionPostgresDigest(csrf))
	if result.Error != nil {
		t.Fatal("could not prepare stored CSRF fixture")
	}
	require.Equal(t, int64(1), result.RowsAffected)
}

func sessionPostgresStoredCSRFMatches(t *testing.T, db *gorm.DB, sessionSecret, csrf string) bool {
	t.Helper()
	var row identitypostgres.SessionRow
	if err := db.Where(
		"secret_hash = ?", sessionPostgresDigest(sessionSecret),
	).First(&row).Error; err != nil {
		t.Fatal("could not inspect stored CSRF fixture")
	}
	return subtle.ConstantTimeCompare(row.CSRFHash, sessionPostgresDigest(csrf)) == 1
}

func sessionPostgresSessionCount(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&identitypostgres.SessionRow{}).Count(&count).Error)
	return count
}

func sessionPostgresSessionExists(t *testing.T, db *gorm.DB, sessionSecret string) bool {
	t.Helper()
	var count int64
	if err := db.Model(&identitypostgres.SessionRow{}).
		Where("secret_hash = ?", sessionPostgresDigest(sessionSecret)).
		Count(&count).Error; err != nil {
		t.Fatal("could not inspect session fixture")
	}
	return count == 1
}

func sessionPostgresPinnedStore(t *testing.T, db *gorm.DB) (*identitypostgres.Store, int) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal("could not prepare PostgreSQL contention observer")
	}
	connection, err := sqlDB.Conn(t.Context())
	if err != nil {
		t.Fatal("could not prepare PostgreSQL contention observer")
	}
	t.Cleanup(func() {
		if connection.Close() != nil {
			t.Error("could not close PostgreSQL contention observer")
		}
	})
	pinnedDB, err := gorm.Open(
		gormpostgres.New(gormpostgres.Config{Conn: connection}),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		t.Fatal("could not prepare PostgreSQL contention observer")
	}
	var backendPID int
	if err := pinnedDB.Raw("SELECT pg_backend_pid()").Scan(&backendPID).Error; err != nil {
		t.Fatal("could not prepare PostgreSQL contention observer")
	}
	return identitypostgres.NewStore(pinnedDB), backendPID
}

func sessionPostgresAwaitAdvisoryWait(t *testing.T, db *gorm.DB, backendPID int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	const waitingQuery = `
SELECT EXISTS (
    SELECT 1
    FROM pg_locks AS waiting
    JOIN pg_locks AS holding
      ON holding.locktype = waiting.locktype
     AND holding.database IS NOT DISTINCT FROM waiting.database
     AND holding.classid IS NOT DISTINCT FROM waiting.classid
     AND holding.objid IS NOT DISTINCT FROM waiting.objid
     AND holding.objsubid IS NOT DISTINCT FROM waiting.objsubid
    WHERE waiting.pid = ?
      AND waiting.locktype = 'advisory'
      AND NOT waiting.granted
      AND holding.granted
      AND holding.pid <> waiting.pid
)`
	for {
		var waiting bool
		if err := db.WithContext(ctx).Raw(waitingQuery, backendPID).Scan(&waiting).Error; err != nil {
			t.Fatal("could not observe PostgreSQL advisory-lock contention")
		}
		if waiting {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("PostgreSQL transaction did not enter the advisory-lock wait")
		default:
			runtime.Gosched()
		}
	}
}

func sessionPostgresDerivedCSRF(sessionSecret string) string {
	mac := hmac.New(sha256.New, []byte(sessionSecret))
	_, _ = mac.Write([]byte(sessionPostgresCSRFDomain))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func sessionPostgresDigest(value string) []byte {
	digest := sha256.Sum256([]byte(value))
	return digest[:]
}

func sessionPostgresEqualText(left, right string) bool {
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

func sessionPostgresHasFailure(err error, kind application.SessionCredentialFailureKind) bool {
	var failure *application.SessionCredentialFailure
	return errors.As(err, &failure) && failure.Kind == kind
}

func sessionPostgresReceive[T any](t *testing.T, values <-chan T) T {
	t.Helper()
	// Ordering comes only from explicit channels and the PostgreSQL advisory
	// transaction lock. This timer merely bounds a broken barrier.
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	select {
	case value := <-values:
		return value
	case <-timer.C:
		t.Fatal("deterministic session test barrier did not complete")
		var zero T
		return zero
	}
}
