package identity_test

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	identitypostgres "netbox-go/internal/adapters/postgres/identity"
	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

var errPasswordChangePostgresInjected = errors.New("injected password-change transaction failure")

type passwordChangePostgresFault uint8

const (
	passwordChangePostgresNoFault passwordChangePostgresFault = iota
	passwordChangePostgresFailBeforeDelete
	passwordChangePostgresFailAfterDelete
)

type passwordChangePostgresTrace struct {
	updateObserved atomic.Bool
	deleteAttempts atomic.Int64
	deletes        atomic.Int64
	createAttempts atomic.Int64
	creates        atomic.Int64
}

type passwordChangePostgresFaultStore struct {
	application.Store
	fault               passwordChangePostgresFault
	trace               *passwordChangePostgresTrace
	expectedNewPassword string
	expectedUpdated     time.Time
}

func (store *passwordChangePostgresFaultStore) Transaction(
	ctx context.Context,
	apply func(application.Store) error,
) error {
	return store.Store.Transaction(ctx, func(tx application.Store) error {
		return apply(&passwordChangePostgresFaultTransaction{
			Store:               tx,
			fault:               store.fault,
			trace:               store.trace,
			expectedNewPassword: store.expectedNewPassword,
			expectedUpdated:     store.expectedUpdated,
		})
	})
}

type passwordChangePostgresFaultTransaction struct {
	application.Store
	fault               passwordChangePostgresFault
	trace               *passwordChangePostgresTrace
	expectedNewPassword string
	expectedUpdated     time.Time
}

func (tx *passwordChangePostgresFaultTransaction) DeleteSessionsForUser(
	ctx context.Context,
	userID int64,
) error {
	tx.trace.deleteAttempts.Add(1)
	user, hash, err := tx.Store.UserByID(ctx, userID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(tx.expectedNewPassword)) == nil &&
		user.Updated.Equal(tx.expectedUpdated) {
		tx.trace.updateObserved.Store(true)
	}
	if tx.fault == passwordChangePostgresFailBeforeDelete {
		return errPasswordChangePostgresInjected
	}
	if err := tx.Store.DeleteSessionsForUser(ctx, userID); err != nil {
		return err
	}
	tx.trace.deletes.Add(1)
	if tx.fault == passwordChangePostgresFailAfterDelete {
		return errPasswordChangePostgresInjected
	}
	return nil
}

func (tx *passwordChangePostgresFaultTransaction) CreateSession(
	ctx context.Context,
	record application.SessionRecord,
) error {
	tx.trace.createAttempts.Add(1)
	if err := tx.Store.CreateSession(ctx, record); err != nil {
		return err
	}
	tx.trace.creates.Add(1)
	return nil
}

type passwordChangePostgresFixture struct {
	db          *gorm.DB
	store       *identitypostgres.Store
	service     *application.Service
	clock       *testClock
	user        domain.User
	sessions    []domain.BrowserSession
	token       application.CreatedToken
	oldPassword string
	newPassword string
}

type passwordChangePostgresSnapshot struct {
	user       domain.User
	hash       string
	sessions   []application.SessionRecord
	token      identitypostgres.TokenRow
	tokenFound bool
}

type passwordChangePostgresOutcome struct {
	result application.ChangePasswordResult
	err    error
}

type passwordChangePostgresLoginOutcome struct {
	session domain.BrowserSession
	err     error
}

func TestPostgresPasswordChangeRotationRollback(t *testing.T) {
	t.Run("failure after password update", func(t *testing.T) {
		fixture := newPasswordChangePostgresFixture(t, 2)
		snapshot := takePasswordChangePostgresSnapshot(t, fixture)
		trace := &passwordChangePostgresTrace{}
		service := newPasswordChangePostgresFaultService(
			fixture,
			trace,
			passwordChangePostgresFailBeforeDelete,
		)

		result, err := service.ChangePassword(
			t.Context(),
			fixture.user.Principal(),
			passwordChangePostgresBrowserInput(fixture),
		)

		assertPasswordChangePostgresInternalFailure(t, result, err)
		if !errors.Is(err, errPasswordChangePostgresInjected) {
			t.Fatal("password-update rollback did not preserve its injected cause")
		}
		if !trace.updateObserved.Load() || trace.deleteAttempts.Load() != 1 ||
			trace.deletes.Load() != 0 || trace.createAttempts.Load() != 0 {
			t.Fatal("password-update rollback did not stop after the real update")
		}
		assertPasswordChangePostgresSnapshot(t, fixture, snapshot)
		assertPasswordChangePostgresPassword(t, fixture, fixture.oldPassword, true)
		assertPasswordChangePostgresPassword(t, fixture, fixture.newPassword, false)
	})

	t.Run("failure after all session deletion", func(t *testing.T) {
		fixture := newPasswordChangePostgresFixture(t, 2)
		snapshot := takePasswordChangePostgresSnapshot(t, fixture)
		trace := &passwordChangePostgresTrace{}
		service := newPasswordChangePostgresFaultService(
			fixture,
			trace,
			passwordChangePostgresFailAfterDelete,
		)

		result, err := service.ChangePassword(
			t.Context(),
			fixture.user.Principal(),
			passwordChangePostgresBrowserInput(fixture),
		)

		assertPasswordChangePostgresInternalFailure(t, result, err)
		if !errors.Is(err, errPasswordChangePostgresInjected) {
			t.Fatal("session-delete rollback did not preserve its injected cause")
		}
		if !trace.updateObserved.Load() || trace.deleteAttempts.Load() != 1 ||
			trace.deletes.Load() != 1 || trace.createAttempts.Load() != 0 {
			t.Fatal("session-delete rollback did not execute the real update and deletion")
		}
		assertPasswordChangePostgresSnapshot(t, fixture, snapshot)
		assertPasswordChangePostgresPassword(t, fixture, fixture.oldPassword, true)
		assertPasswordChangePostgresPassword(t, fixture, fixture.newPassword, false)
	})

	t.Run("replacement insert collision", func(t *testing.T) {
		fixture := newPasswordChangePostgresFixture(t, 2)
		other, err := fixture.service.CreateLocalUser(
			t.Context(),
			fixture.user.Principal(),
			application.CreateUserInput{
				Username: "password-change-collision-user",
				Password: "Collision-Owner-2026!",
			},
		)
		if err != nil {
			t.Fatal("could not prepare replacement-collision user")
		}
		collisionSession, err := fixture.service.Login(
			t.Context(),
			other.Username,
			"Collision-Owner-2026!",
		)
		if err != nil {
			t.Fatal("could not prepare replacement-collision session")
		}
		entropy, err := base64.RawURLEncoding.DecodeString(collisionSession.Secret)
		if err != nil || len(entropy) != 32 {
			t.Fatal("could not prepare replacement-collision entropy")
		}
		snapshot := takePasswordChangePostgresSnapshot(t, fixture)
		trace := &passwordChangePostgresTrace{}
		faultStore := &passwordChangePostgresFaultStore{
			Store:               fixture.store,
			trace:               trace,
			expectedNewPassword: fixture.newPassword,
			expectedUpdated:     fixture.clock.now,
		}
		service := application.NewService(
			faultStore,
			fixture.clock,
			application.WithPasswordChangeEntropy(bytes.NewReader(entropy)),
		)

		result, changeErr := service.ChangePassword(
			t.Context(),
			fixture.user.Principal(),
			passwordChangePostgresBrowserInput(fixture),
		)

		assertPasswordChangePostgresInternalFailure(t, result, changeErr)
		if !trace.updateObserved.Load() || trace.deletes.Load() != 1 ||
			trace.createAttempts.Load() != 1 || trace.creates.Load() != 0 {
			t.Fatal("replacement collision did not follow real update/delete/insert ordering")
		}
		assertPasswordChangePostgresSnapshot(t, fixture, snapshot)
		assertPasswordChangePostgresSession(t, fixture.service, collisionSession, true)
	})

	t.Run("browser deferred commit failure", func(t *testing.T) {
		fixture := newPasswordChangePostgresFixture(t, 2)
		snapshot := takePasswordChangePostgresSnapshot(t, fixture)
		installPasswordChangePostgresCommitFailure(t, fixture.db)
		trace := &passwordChangePostgresTrace{}
		service := newPasswordChangePostgresFaultService(
			fixture,
			trace,
			passwordChangePostgresNoFault,
		)

		result, err := service.ChangePassword(
			t.Context(),
			fixture.user.Principal(),
			passwordChangePostgresBrowserInput(fixture),
		)

		assertPasswordChangePostgresInternalFailure(t, result, err)
		if !trace.updateObserved.Load() || trace.deletes.Load() != 1 ||
			trace.creates.Load() != 1 {
			t.Fatal("browser COMMIT failure did not follow completed update/delete/create DML")
		}
		assertPasswordChangePostgresSnapshot(t, fixture, snapshot)
	})

	t.Run("token deferred commit failure", func(t *testing.T) {
		fixture := newPasswordChangePostgresFixture(t, 2)
		snapshot := takePasswordChangePostgresSnapshot(t, fixture)
		installPasswordChangePostgresCommitFailure(t, fixture.db)
		trace := &passwordChangePostgresTrace{}
		service := newPasswordChangePostgresFaultService(
			fixture,
			trace,
			passwordChangePostgresNoFault,
		)

		result, err := service.ChangePassword(
			t.Context(),
			fixture.user.Principal(),
			application.NewPasswordChangeInput(
				fixture.oldPassword,
				fixture.newPassword,
				application.APITokenPasswordChangeCredential(),
			),
		)

		assertPasswordChangePostgresInternalFailure(t, result, err)
		if !trace.updateObserved.Load() || trace.deletes.Load() != 1 ||
			trace.createAttempts.Load() != 0 {
			t.Fatal("Token COMMIT failure did not follow completed update/delete DML")
		}
		assertPasswordChangePostgresSnapshot(t, fixture, snapshot)
	})

	t.Run("successful browser rotation", func(t *testing.T) {
		fixture := newPasswordChangePostgresFixture(t, 2)
		tokenBefore := passwordChangePostgresTokenRow(t, fixture)
		trace := &passwordChangePostgresTrace{}
		service := newPasswordChangePostgresFaultService(
			fixture,
			trace,
			passwordChangePostgresNoFault,
		)

		result, err := service.ChangePassword(
			t.Context(),
			fixture.user.Principal(),
			passwordChangePostgresBrowserInput(fixture),
		)
		if err != nil || result == nil {
			t.Fatal("browser password change did not commit")
		}
		replacement, present := result.BrowserSession()
		if !present || passwordChangePostgresSessionHasNoMaterial(replacement) {
			t.Fatal("browser password change returned no replacement session")
		}
		if !trace.updateObserved.Load() || trace.deletes.Load() != 1 ||
			trace.creates.Load() != 1 {
			t.Fatal("browser password change did not execute update/delete/create once")
		}
		assertPasswordChangePostgresBrowserSuccess(t, fixture, replacement)
		assertPasswordChangePostgresTokenUnchanged(t, fixture, tokenBefore)
		assertPasswordChangePostgresTokenUsable(t, fixture)
	})

	t.Run("successful token invalidation", func(t *testing.T) {
		fixture := newPasswordChangePostgresFixture(t, 2)
		tokenBefore := passwordChangePostgresTokenRow(t, fixture)
		trace := &passwordChangePostgresTrace{}
		service := newPasswordChangePostgresFaultService(
			fixture,
			trace,
			passwordChangePostgresNoFault,
		)

		result, err := service.ChangePassword(
			t.Context(),
			fixture.user.Principal(),
			application.NewPasswordChangeInput(
				fixture.oldPassword,
				fixture.newPassword,
				application.APITokenPasswordChangeCredential(),
			),
		)
		if err != nil || result == nil {
			t.Fatal("Token-origin password change did not commit")
		}
		if session, present := result.BrowserSession(); present ||
			passwordChangePostgresSessionHasMaterial(session) {
			t.Fatal("Token-origin password change returned browser material")
		}
		if !trace.updateObserved.Load() || trace.deletes.Load() != 1 ||
			trace.createAttempts.Load() != 0 {
			t.Fatal("Token-origin password change did not execute update/delete only")
		}
		if sessionPostgresSessionCount(t, fixture.db) != 0 {
			t.Fatal("Token-origin password change left browser sessions")
		}
		assertPasswordChangePostgresPassword(t, fixture, fixture.oldPassword, false)
		assertPasswordChangePostgresPassword(t, fixture, fixture.newPassword, true)
		assertPasswordChangePostgresTokenUnchanged(t, fixture, tokenBefore)
		assertPasswordChangePostgresTokenUsable(t, fixture)
	})
}

func TestPostgresConcurrentPasswordChangesHaveOneWinner(t *testing.T) {
	t.Run("browser origin", func(t *testing.T) {
		fixture := newPasswordChangePostgresFixture(t, 2)
		firstStore, _ := sessionPostgresPinnedStore(t, fixture.db)
		secondStore, secondBackendPID := sessionPostgresPinnedStore(t, fixture.db)

		firstVerified := make(chan struct{}, 1)
		secondVerified := make(chan struct{}, 1)
		releaseFirstEntry := make(chan struct{})
		releaseSecondEntry := make(chan struct{})
		firstCommitReady := make(chan struct{}, 1)
		releaseFirstCommit := make(chan struct{})
		var releaseFirstEntryOnce, releaseSecondEntryOnce, releaseFirstCommitOnce sync.Once
		releaseFirst := func() { releaseFirstEntryOnce.Do(func() { close(releaseFirstEntry) }) }
		releaseSecond := func() { releaseSecondEntryOnce.Do(func() { close(releaseSecondEntry) }) }
		commitFirst := func() { releaseFirstCommitOnce.Do(func() { close(releaseFirstCommit) }) }
		defer releaseFirst()
		defer releaseSecond()
		defer commitFirst()

		firstHoldingStore := &sessionPostgresHoldCommitStore{
			Store:   firstStore,
			ready:   firstCommitReady,
			release: releaseFirstCommit,
		}
		firstBarrierStore := &sessionPostgresLoginTransactionBarrierStore{
			Store:    firstHoldingStore,
			verified: firstVerified,
			release:  releaseFirstEntry,
		}
		secondBarrierStore := &sessionPostgresLoginTransactionBarrierStore{
			Store:    secondStore,
			verified: secondVerified,
			release:  releaseSecondEntry,
		}
		firstService := application.NewService(firstBarrierStore, fixture.clock)
		secondService := application.NewService(secondBarrierStore, fixture.clock)
		firstNext := "Concurrent-Browser-First-2026!"
		secondNext := "Concurrent-Browser-Second-2026!"
		firstResult := make(chan passwordChangePostgresOutcome, 1)
		secondResult := make(chan passwordChangePostgresOutcome, 1)

		go func() {
			result, err := firstService.ChangePassword(
				t.Context(),
				fixture.user.Principal(),
				application.NewPasswordChangeInput(
					fixture.oldPassword,
					firstNext,
					application.BrowserSessionPasswordChangeCredential(
						fixture.sessions[0].Secret,
						fixture.sessions[0].CSRFToken,
					),
				),
			)
			firstResult <- passwordChangePostgresOutcome{result: result, err: err}
		}()
		go func() {
			result, err := secondService.ChangePassword(
				t.Context(),
				fixture.user.Principal(),
				application.NewPasswordChangeInput(
					fixture.oldPassword,
					secondNext,
					application.BrowserSessionPasswordChangeCredential(
						fixture.sessions[1].Secret,
						fixture.sessions[1].CSRFToken,
					),
				),
			)
			secondResult <- passwordChangePostgresOutcome{result: result, err: err}
		}()

		sessionPostgresReceive(t, firstVerified)
		sessionPostgresReceive(t, secondVerified)
		releaseFirst()
		sessionPostgresReceive(t, firstCommitReady)
		releaseSecond()
		sessionPostgresAwaitAdvisoryWait(t, fixture.db, secondBackendPID)
		commitFirst()

		winner := sessionPostgresReceive(t, firstResult)
		loser := sessionPostgresReceive(t, secondResult)
		if winner.err != nil || winner.result == nil {
			t.Fatal("selected first browser password change did not win")
		}
		if loser.result != nil || !passwordChangePostgresCurrentPasswordFailure(loser.err) {
			t.Fatal("serialized browser password-change loser had an unstable outcome")
		}
		replacement, present := winner.result.BrowserSession()
		if !present || passwordChangePostgresSessionHasNoMaterial(replacement) {
			t.Fatal("browser password-change winner returned no replacement")
		}
		if sessionPostgresSessionCount(t, fixture.db) != 1 {
			t.Fatal("concurrent browser password changes left an unexpected session set")
		}
		for _, oldSession := range fixture.sessions {
			assertPasswordChangePostgresSession(t, fixture.service, oldSession, false)
		}
		assertPasswordChangePostgresSession(t, fixture.service, replacement, true)
		assertPasswordChangePostgresPassword(t, fixture, fixture.oldPassword, false)
		assertPasswordChangePostgresPassword(t, fixture, firstNext, true)
		assertPasswordChangePostgresPassword(t, fixture, secondNext, false)
		assertPasswordChangePostgresTokenUsable(t, fixture)
	})

	t.Run("token origin", func(t *testing.T) {
		fixture := newPasswordChangePostgresFixture(t, 2)
		firstStore, _ := sessionPostgresPinnedStore(t, fixture.db)
		secondStore, secondBackendPID := sessionPostgresPinnedStore(t, fixture.db)

		firstVerified := make(chan struct{}, 1)
		secondVerified := make(chan struct{}, 1)
		releaseFirstEntry := make(chan struct{})
		releaseSecondEntry := make(chan struct{})
		firstCommitReady := make(chan struct{}, 1)
		releaseFirstCommit := make(chan struct{})
		var releaseFirstEntryOnce, releaseSecondEntryOnce, releaseFirstCommitOnce sync.Once
		releaseFirst := func() { releaseFirstEntryOnce.Do(func() { close(releaseFirstEntry) }) }
		releaseSecond := func() { releaseSecondEntryOnce.Do(func() { close(releaseSecondEntry) }) }
		commitFirst := func() { releaseFirstCommitOnce.Do(func() { close(releaseFirstCommit) }) }
		defer releaseFirst()
		defer releaseSecond()
		defer commitFirst()

		firstHoldingStore := &sessionPostgresHoldCommitStore{
			Store:   firstStore,
			ready:   firstCommitReady,
			release: releaseFirstCommit,
		}
		firstBarrierStore := &sessionPostgresLoginTransactionBarrierStore{
			Store:    firstHoldingStore,
			verified: firstVerified,
			release:  releaseFirstEntry,
		}
		secondBarrierStore := &sessionPostgresLoginTransactionBarrierStore{
			Store:    secondStore,
			verified: secondVerified,
			release:  releaseSecondEntry,
		}
		firstService := application.NewService(firstBarrierStore, fixture.clock)
		secondService := application.NewService(secondBarrierStore, fixture.clock)
		firstNext := "Concurrent-Token-First-2026!"
		secondNext := "Concurrent-Token-Second-2026!"
		firstResult := make(chan passwordChangePostgresOutcome, 1)
		secondResult := make(chan passwordChangePostgresOutcome, 1)

		go func() {
			result, err := firstService.ChangePassword(
				t.Context(),
				fixture.user.Principal(),
				application.NewPasswordChangeInput(
					fixture.oldPassword,
					firstNext,
					application.APITokenPasswordChangeCredential(),
				),
			)
			firstResult <- passwordChangePostgresOutcome{result: result, err: err}
		}()
		go func() {
			result, err := secondService.ChangePassword(
				t.Context(),
				fixture.user.Principal(),
				application.NewPasswordChangeInput(
					fixture.oldPassword,
					secondNext,
					application.APITokenPasswordChangeCredential(),
				),
			)
			secondResult <- passwordChangePostgresOutcome{result: result, err: err}
		}()

		sessionPostgresReceive(t, firstVerified)
		sessionPostgresReceive(t, secondVerified)
		releaseFirst()
		sessionPostgresReceive(t, firstCommitReady)
		releaseSecond()
		sessionPostgresAwaitAdvisoryWait(t, fixture.db, secondBackendPID)
		commitFirst()

		winner := sessionPostgresReceive(t, firstResult)
		loser := sessionPostgresReceive(t, secondResult)
		if winner.err != nil || winner.result == nil {
			t.Fatal("selected first Token password change did not win")
		}
		if session, present := winner.result.BrowserSession(); present ||
			passwordChangePostgresSessionHasMaterial(session) {
			t.Fatal("Token password-change winner returned browser material")
		}
		if loser.result != nil || !passwordChangePostgresCurrentPasswordFailure(loser.err) {
			t.Fatal("serialized Token password-change loser had an unstable outcome")
		}
		if sessionPostgresSessionCount(t, fixture.db) != 0 {
			t.Fatal("concurrent Token password changes left a browser session")
		}
		assertPasswordChangePostgresPassword(t, fixture, fixture.oldPassword, false)
		assertPasswordChangePostgresPassword(t, fixture, firstNext, true)
		assertPasswordChangePostgresPassword(t, fixture, secondNext, false)
		assertPasswordChangePostgresTokenUsable(t, fixture)
	})
}

func TestPostgresPasswordChangeSerializesWithLogin(t *testing.T) {
	t.Run("password change wins", func(t *testing.T) {
		fixture := newPasswordChangePostgresFixture(t, 1)
		passwordStore, _ := sessionPostgresPinnedStore(t, fixture.db)
		loginStore, loginBackendPID := sessionPostgresPinnedStore(t, fixture.db)

		passwordVerified := make(chan struct{}, 1)
		loginVerified := make(chan struct{}, 1)
		releasePasswordEntry := make(chan struct{})
		releaseLoginEntry := make(chan struct{})
		passwordCommitReady := make(chan struct{}, 1)
		releasePasswordCommit := make(chan struct{})
		var releasePasswordEntryOnce, releaseLoginEntryOnce, releasePasswordCommitOnce sync.Once
		enterPassword := func() { releasePasswordEntryOnce.Do(func() { close(releasePasswordEntry) }) }
		enterLogin := func() { releaseLoginEntryOnce.Do(func() { close(releaseLoginEntry) }) }
		commitPassword := func() { releasePasswordCommitOnce.Do(func() { close(releasePasswordCommit) }) }
		defer enterPassword()
		defer enterLogin()
		defer commitPassword()

		passwordHoldingStore := &sessionPostgresHoldCommitStore{
			Store:   passwordStore,
			ready:   passwordCommitReady,
			release: releasePasswordCommit,
		}
		passwordBarrierStore := &sessionPostgresLoginTransactionBarrierStore{
			Store:    passwordHoldingStore,
			verified: passwordVerified,
			release:  releasePasswordEntry,
		}
		loginBarrierStore := &sessionPostgresLoginTransactionBarrierStore{
			Store:    loginStore,
			verified: loginVerified,
			release:  releaseLoginEntry,
		}
		passwordService := application.NewService(passwordBarrierStore, fixture.clock)
		loginService := application.NewService(loginBarrierStore, fixture.clock)
		passwordResult := make(chan passwordChangePostgresOutcome, 1)
		loginResult := make(chan passwordChangePostgresLoginOutcome, 1)

		go func() {
			result, err := passwordService.ChangePassword(
				t.Context(),
				fixture.user.Principal(),
				passwordChangePostgresBrowserInput(fixture),
			)
			passwordResult <- passwordChangePostgresOutcome{result: result, err: err}
		}()
		go func() {
			session, err := loginService.Login(
				t.Context(),
				fixture.user.Username,
				fixture.oldPassword,
			)
			loginResult <- passwordChangePostgresLoginOutcome{session: session, err: err}
		}()

		sessionPostgresReceive(t, passwordVerified)
		sessionPostgresReceive(t, loginVerified)
		enterPassword()
		sessionPostgresReceive(t, passwordCommitReady)
		enterLogin()
		sessionPostgresAwaitAdvisoryWait(t, fixture.db, loginBackendPID)
		commitPassword()

		passwordOutcome := sessionPostgresReceive(t, passwordResult)
		loginOutcome := sessionPostgresReceive(t, loginResult)
		if passwordOutcome.err != nil || passwordOutcome.result == nil {
			t.Fatal("password-change-first serialization did not commit password change")
		}
		if loginOutcome.err == nil || shared.ReasonOf(loginOutcome.err) != shared.ErrorReasonUnauthenticated ||
			passwordChangePostgresSessionHasMaterial(loginOutcome.session) {
			t.Fatal("stale preverified login did not fail without session material")
		}
		replacement, present := passwordOutcome.result.BrowserSession()
		if !present || passwordChangePostgresSessionHasNoMaterial(replacement) {
			t.Fatal("password-change-first serialization returned no replacement")
		}
		if sessionPostgresSessionCount(t, fixture.db) != 1 {
			t.Fatal("password-change-first serialization left an unexpected session set")
		}
		assertPasswordChangePostgresSession(t, fixture.service, fixture.sessions[0], false)
		assertPasswordChangePostgresSession(t, fixture.service, replacement, true)
		assertPasswordChangePostgresPassword(t, fixture, fixture.oldPassword, false)
		assertPasswordChangePostgresPassword(t, fixture, fixture.newPassword, true)
		assertPasswordChangePostgresTokenUsable(t, fixture)
	})

	t.Run("login wins", func(t *testing.T) {
		fixture := newPasswordChangePostgresFixture(t, 1)
		loginStore, _ := sessionPostgresPinnedStore(t, fixture.db)
		passwordStore, passwordBackendPID := sessionPostgresPinnedStore(t, fixture.db)

		loginVerified := make(chan struct{}, 1)
		passwordVerified := make(chan struct{}, 1)
		releaseLoginEntry := make(chan struct{})
		releasePasswordEntry := make(chan struct{})
		loginCommitReady := make(chan struct{}, 1)
		releaseLoginCommit := make(chan struct{})
		var releaseLoginEntryOnce, releasePasswordEntryOnce, releaseLoginCommitOnce sync.Once
		enterLogin := func() { releaseLoginEntryOnce.Do(func() { close(releaseLoginEntry) }) }
		enterPassword := func() { releasePasswordEntryOnce.Do(func() { close(releasePasswordEntry) }) }
		commitLogin := func() { releaseLoginCommitOnce.Do(func() { close(releaseLoginCommit) }) }
		defer enterLogin()
		defer enterPassword()
		defer commitLogin()

		loginHoldingStore := &sessionPostgresHoldCommitStore{
			Store:   loginStore,
			ready:   loginCommitReady,
			release: releaseLoginCommit,
		}
		loginBarrierStore := &sessionPostgresLoginTransactionBarrierStore{
			Store:    loginHoldingStore,
			verified: loginVerified,
			release:  releaseLoginEntry,
		}
		passwordBarrierStore := &sessionPostgresLoginTransactionBarrierStore{
			Store:    passwordStore,
			verified: passwordVerified,
			release:  releasePasswordEntry,
		}
		loginService := application.NewService(loginBarrierStore, fixture.clock)
		passwordService := application.NewService(passwordBarrierStore, fixture.clock)
		loginResult := make(chan passwordChangePostgresLoginOutcome, 1)
		passwordResult := make(chan passwordChangePostgresOutcome, 1)

		go func() {
			session, err := loginService.Login(
				t.Context(),
				fixture.user.Username,
				fixture.oldPassword,
			)
			loginResult <- passwordChangePostgresLoginOutcome{session: session, err: err}
		}()
		go func() {
			result, err := passwordService.ChangePassword(
				t.Context(),
				fixture.user.Principal(),
				passwordChangePostgresBrowserInput(fixture),
			)
			passwordResult <- passwordChangePostgresOutcome{result: result, err: err}
		}()

		sessionPostgresReceive(t, loginVerified)
		sessionPostgresReceive(t, passwordVerified)
		enterLogin()
		sessionPostgresReceive(t, loginCommitReady)
		enterPassword()
		sessionPostgresAwaitAdvisoryWait(t, fixture.db, passwordBackendPID)
		commitLogin()

		loginOutcome := sessionPostgresReceive(t, loginResult)
		passwordOutcome := sessionPostgresReceive(t, passwordResult)
		if loginOutcome.err != nil || passwordChangePostgresSessionHasNoMaterial(loginOutcome.session) {
			t.Fatal("login-first serialization did not return its committed session")
		}
		if passwordOutcome.err != nil || passwordOutcome.result == nil {
			t.Fatal("login-first serialization did not commit password change")
		}
		replacement, present := passwordOutcome.result.BrowserSession()
		if !present || passwordChangePostgresSessionHasNoMaterial(replacement) {
			t.Fatal("login-first serialization returned no password-change replacement")
		}
		if sessionPostgresSessionCount(t, fixture.db) != 1 {
			t.Fatal("login-first serialization left an unexpected session set")
		}
		assertPasswordChangePostgresSession(t, fixture.service, fixture.sessions[0], false)
		assertPasswordChangePostgresSession(t, fixture.service, loginOutcome.session, false)
		assertPasswordChangePostgresSession(t, fixture.service, replacement, true)
		assertPasswordChangePostgresPassword(t, fixture, fixture.oldPassword, false)
		assertPasswordChangePostgresPassword(t, fixture, fixture.newPassword, true)
		assertPasswordChangePostgresTokenUsable(t, fixture)
	})
}

func newPasswordChangePostgresFixture(
	t *testing.T,
	sessionCount int,
) *passwordChangePostgresFixture {
	t.Helper()
	db := newCredentialPostgres(t)
	clock := &testClock{now: time.Date(2026, 8, 19, 8, 0, 0, 0, time.UTC)}
	store := identitypostgres.NewStore(db)
	service := application.NewService(store, clock)
	oldPassword := "Postgres-Current-2026!"
	newPassword := "Postgres-Replacement-2026!"
	user, err := service.BootstrapAdministrator(
		t.Context(),
		"password-change-postgres-admin",
		"",
		oldPassword,
	)
	if err != nil {
		t.Fatal("could not prepare PostgreSQL password-change user")
	}
	sessions := make([]domain.BrowserSession, 0, sessionCount)
	for range sessionCount {
		session, loginErr := service.Login(t.Context(), user.Username, oldPassword)
		if loginErr != nil {
			t.Fatal("could not prepare PostgreSQL password-change session")
		}
		sessions = append(sessions, session)
	}
	token, err := service.CreateToken(t.Context(), user.Principal(), application.CreateTokenInput{
		Description:  "password change PostgreSQL",
		WriteEnabled: true,
	})
	if err != nil || token.Secret == "" {
		t.Fatal("could not prepare PostgreSQL password-change API token")
	}
	clock.now = clock.now.Add(time.Hour)
	return &passwordChangePostgresFixture{
		db:          db,
		store:       store,
		service:     service,
		clock:       clock,
		user:        user,
		sessions:    sessions,
		token:       token,
		oldPassword: oldPassword,
		newPassword: newPassword,
	}
}

func newPasswordChangePostgresFaultService(
	fixture *passwordChangePostgresFixture,
	trace *passwordChangePostgresTrace,
	fault passwordChangePostgresFault,
) *application.Service {
	return application.NewService(&passwordChangePostgresFaultStore{
		Store:               fixture.store,
		fault:               fault,
		trace:               trace,
		expectedNewPassword: fixture.newPassword,
		expectedUpdated:     fixture.clock.now,
	}, fixture.clock)
}

func passwordChangePostgresBrowserInput(
	fixture *passwordChangePostgresFixture,
) application.ChangePasswordInput {
	return application.NewPasswordChangeInput(
		fixture.oldPassword,
		fixture.newPassword,
		application.BrowserSessionPasswordChangeCredential(
			fixture.sessions[0].Secret,
			fixture.sessions[0].CSRFToken,
		),
	)
}

func takePasswordChangePostgresSnapshot(
	t *testing.T,
	fixture *passwordChangePostgresFixture,
) passwordChangePostgresSnapshot {
	t.Helper()
	user, hash, err := fixture.store.UserByID(t.Context(), fixture.user.ID)
	if err != nil {
		t.Fatal("could not snapshot PostgreSQL password-change user")
	}
	var rows []identitypostgres.SessionRow
	if err := fixture.db.Order("secret_hash").Find(&rows).Error; err != nil {
		t.Fatal("could not snapshot PostgreSQL password-change sessions")
	}
	sessions := make([]application.SessionRecord, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, application.SessionRecord{
			SecretHash: append([]byte(nil), row.SecretHash...),
			CSRFHash:   append([]byte(nil), row.CSRFHash...),
			UserID:     row.UserID,
			Created:    row.Created,
			Expires:    row.Expires,
			LastSeen:   row.LastSeen,
		})
	}
	token, found := passwordChangePostgresFindTokenRow(t, fixture)
	return passwordChangePostgresSnapshot{
		user:       user,
		hash:       hash,
		sessions:   sessions,
		token:      token,
		tokenFound: found,
	}
}

func assertPasswordChangePostgresSnapshot(
	t *testing.T,
	fixture *passwordChangePostgresFixture,
	snapshot passwordChangePostgresSnapshot,
) {
	t.Helper()
	freshDB := passwordChangePostgresFreshDatabase(t, fixture.db)
	freshStore := identitypostgres.NewStore(freshDB)
	user, hash, err := freshStore.UserByID(t.Context(), fixture.user.ID)
	if err != nil || user.ID != snapshot.user.ID || user.IsActive != snapshot.user.IsActive ||
		!user.Updated.Equal(snapshot.user.Updated) ||
		!sessionPostgresEqualText(hash, snapshot.hash) {
		t.Fatal("failed password change did not restore the prior user state")
	}
	if passwordChangePostgresSessionCount(t, freshDB) != int64(len(snapshot.sessions)) {
		t.Fatal("failed password change did not restore the prior session count")
	}
	for _, expected := range snapshot.sessions {
		actual, loadErr := freshStore.SessionByHash(t.Context(), expected.SecretHash)
		if loadErr != nil || !passwordChangePostgresRecordEqual(actual, expected) {
			t.Fatal("failed password change did not restore the prior session set")
		}
	}
	token, found := passwordChangePostgresFindTokenRowAt(
		t,
		freshDB,
		fixture.token.Token.ID,
	)
	if found != snapshot.tokenFound ||
		(found && !passwordChangePostgresTokenRowEqual(token, snapshot.token)) {
		t.Fatal("failed password change mutated the independent API token")
	}
}

func assertPasswordChangePostgresInternalFailure(
	t *testing.T,
	result application.ChangePasswordResult,
	err error,
) {
	t.Helper()
	if result != nil {
		t.Fatal("failed password change returned result material")
	}
	if err == nil || shared.ReasonOf(err) != shared.ErrorReasonInternal {
		t.Fatal("PostgreSQL password-change failure was not classified Internal")
	}
}

func assertPasswordChangePostgresBrowserSuccess(
	t *testing.T,
	fixture *passwordChangePostgresFixture,
	replacement domain.BrowserSession,
) {
	t.Helper()
	if replacement.User.ID != fixture.user.ID ||
		!replacement.Expires.Equal(fixture.clock.now.Add(application.BrowserSessionLifetime)) {
		t.Fatal("browser password-change replacement has incorrect ownership or lifetime")
	}
	for _, oldSession := range fixture.sessions {
		if sessionPostgresEqualText(oldSession.Secret, replacement.Secret) ||
			sessionPostgresEqualText(oldSession.CSRFToken, replacement.CSRFToken) {
			t.Fatal("browser password change reused old credential material")
		}
		assertPasswordChangePostgresSession(t, fixture.service, oldSession, false)
	}
	if sessionPostgresSessionCount(t, fixture.db) != 1 {
		t.Fatal("browser password change did not leave exactly one replacement session")
	}
	assertPasswordChangePostgresSession(t, fixture.service, replacement, true)
	record, err := fixture.store.SessionByHash(
		t.Context(),
		sessionPostgresDigest(replacement.Secret),
	)
	if err != nil || record.UserID != fixture.user.ID ||
		!record.Created.Equal(fixture.clock.now) ||
		!record.LastSeen.Equal(fixture.clock.now) ||
		!record.Expires.Equal(fixture.clock.now.Add(application.BrowserSessionLifetime)) ||
		subtle.ConstantTimeCompare(record.CSRFHash, sessionPostgresDigest(replacement.CSRFToken)) != 1 {
		t.Fatal("browser password-change replacement row has incorrect durable state")
	}
	assertPasswordChangePostgresPassword(t, fixture, fixture.oldPassword, false)
	assertPasswordChangePostgresPassword(t, fixture, fixture.newPassword, true)
}

func assertPasswordChangePostgresSession(
	t *testing.T,
	service *application.Service,
	session domain.BrowserSession,
	wantAccepted bool,
) {
	t.Helper()
	_, authErr := service.AuthenticateSession(t.Context(), session.Secret)
	csrfErr := service.VerifyCSRF(t.Context(), session.Secret, session.CSRFToken)
	if (authErr == nil) != wantAccepted || (csrfErr == nil) != wantAccepted {
		t.Fatal("browser session has an unexpected authentication or CSRF outcome")
	}
}

func assertPasswordChangePostgresPassword(
	t *testing.T,
	fixture *passwordChangePostgresFixture,
	password string,
	wantAccepted bool,
) {
	t.Helper()
	_, err := fixture.service.AuthenticatePassword(
		t.Context(),
		fixture.user.Username,
		password,
	)
	if (err == nil) != wantAccepted {
		t.Fatal("PostgreSQL password state has an unexpected authentication outcome")
	}
}

func assertPasswordChangePostgresTokenUsable(
	t *testing.T,
	fixture *passwordChangePostgresFixture,
) {
	t.Helper()
	if _, err := fixture.service.AuthenticateToken(
		t.Context(),
		fixture.token.Secret,
		"127.0.0.1:443",
		true,
	); err != nil {
		t.Fatal("password change made the independent API token unusable")
	}
}

func assertPasswordChangePostgresTokenUnchanged(
	t *testing.T,
	fixture *passwordChangePostgresFixture,
	expected identitypostgres.TokenRow,
) {
	t.Helper()
	actual := passwordChangePostgresTokenRow(t, fixture)
	if !passwordChangePostgresTokenRowEqual(actual, expected) {
		t.Fatal("password change mutated the independent API-token row")
	}
}

func passwordChangePostgresTokenRow(
	t *testing.T,
	fixture *passwordChangePostgresFixture,
) identitypostgres.TokenRow {
	t.Helper()
	row, found := passwordChangePostgresFindTokenRow(t, fixture)
	if !found {
		t.Fatal("PostgreSQL password-change token row is missing")
	}
	return row
}

func passwordChangePostgresFindTokenRow(
	t *testing.T,
	fixture *passwordChangePostgresFixture,
) (identitypostgres.TokenRow, bool) {
	t.Helper()
	return passwordChangePostgresFindTokenRowAt(t, fixture.db, fixture.token.Token.ID)
}

func passwordChangePostgresFindTokenRowAt(
	t *testing.T,
	db *gorm.DB,
	tokenID int64,
) (identitypostgres.TokenRow, bool) {
	t.Helper()
	var row identitypostgres.TokenRow
	err := db.Where("id = ?", tokenID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return identitypostgres.TokenRow{}, false
	}
	if err != nil {
		t.Fatal("could not inspect PostgreSQL password-change token state")
	}
	return row, true
}

func passwordChangePostgresSessionCount(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var count int64
	if err := db.Model(&identitypostgres.SessionRow{}).Count(&count).Error; err != nil {
		t.Fatal("could not inspect PostgreSQL password-change session count")
	}
	return count
}

func passwordChangePostgresFreshDatabase(t *testing.T, db *gorm.DB) *gorm.DB {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal("could not prepare fresh PostgreSQL password-change observer")
	}
	connection, err := sqlDB.Conn(t.Context())
	if err != nil {
		t.Fatal("could not prepare fresh PostgreSQL password-change observer")
	}
	t.Cleanup(func() {
		if connection.Close() != nil {
			t.Error("could not close fresh PostgreSQL password-change observer")
		}
	})
	freshDB, err := gorm.Open(
		gormpostgres.New(gormpostgres.Config{Conn: connection}),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		t.Fatal("could not prepare fresh PostgreSQL password-change observer")
	}
	return freshDB
}

func passwordChangePostgresRecordEqual(
	left application.SessionRecord,
	right application.SessionRecord,
) bool {
	return passwordChangePostgresBytesEqual(left.SecretHash, right.SecretHash) &&
		passwordChangePostgresBytesEqual(left.CSRFHash, right.CSRFHash) &&
		left.UserID == right.UserID &&
		left.Created.Equal(right.Created) &&
		left.Expires.Equal(right.Expires) &&
		left.LastSeen.Equal(right.LastSeen)
}

func passwordChangePostgresTokenRowEqual(
	left identitypostgres.TokenRow,
	right identitypostgres.TokenRow,
) bool {
	return left.ID == right.ID &&
		left.UserID == right.UserID &&
		left.Display == right.Display &&
		passwordChangePostgresBytesEqual(left.SecretHash, right.SecretHash) &&
		left.Description == right.Description &&
		left.WriteEnabled == right.WriteEnabled &&
		bytes.Equal(left.AllowedIPs, right.AllowedIPs) &&
		left.Created.Equal(right.Created) &&
		passwordChangePostgresOptionalTimeEqual(left.Expires, right.Expires) &&
		passwordChangePostgresOptionalTimeEqual(left.LastUsed, right.LastUsed) &&
		passwordChangePostgresOptionalTimeEqual(left.RevokedAt, right.RevokedAt)
}

func passwordChangePostgresBytesEqual(left, right []byte) bool {
	return len(left) == len(right) && subtle.ConstantTimeCompare(left, right) == 1
}

func passwordChangePostgresOptionalTimeEqual(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
}

func passwordChangePostgresSessionHasMaterial(session domain.BrowserSession) bool {
	return session.Secret != "" || session.CSRFToken != "" ||
		session.User.ID != 0 || !session.Expires.IsZero()
}

func passwordChangePostgresSessionHasNoMaterial(session domain.BrowserSession) bool {
	return !passwordChangePostgresSessionHasMaterial(session)
}

func passwordChangePostgresCurrentPasswordFailure(err error) bool {
	if err == nil || shared.ReasonOf(err) != shared.ErrorReasonValidation {
		return false
	}
	violations := shared.ViolationsOf(err)
	return len(violations) == 1 &&
		violations[0].Field == "current_password" &&
		violations[0].Reason == "invalid" &&
		violations[0].Description == "Current password is incorrect."
}

func installPasswordChangePostgresCommitFailure(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec(`
CREATE FUNCTION i4_fail_password_change_commit()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  RAISE EXCEPTION 'injected deferred password-change commit failure';
END;
$$`).Error; err != nil {
		t.Fatal("could not install deferred password-change failure function")
	}
	if err := db.Exec(`
CREATE CONSTRAINT TRIGGER i4_password_change_commit_failure
AFTER UPDATE OF password_hash ON go_identity_users
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION i4_fail_password_change_commit()`).Error; err != nil {
		t.Fatal("could not install deferred password-change failure trigger")
	}
	t.Cleanup(func() {
		if err := db.Exec(`
DROP TRIGGER IF EXISTS i4_password_change_commit_failure ON go_identity_users`).Error; err != nil {
			t.Error("could not remove deferred password-change failure trigger")
		}
		if err := db.Exec(`DROP FUNCTION IF EXISTS i4_fail_password_change_commit()`).Error; err != nil {
			t.Error("could not remove deferred password-change failure function")
		}
	})
}
