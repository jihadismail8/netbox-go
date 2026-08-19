package identity_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

const sessionMatrixCSRFDomainTag = "netbox-go/browser-csrf/v1"

type sessionMatrixUserResult struct {
	user domain.User
	hash string
	err  error
}

type sessionMatrixStore struct {
	application.Store

	sessions        map[string]application.SessionRecord
	usernameResults []sessionMatrixUserResult
	users           map[int64]sessionMatrixUserResult

	sessionLookupErr error
	createErr        error
	deleteErr        error
	updateErr        error
	commitErr        error

	events            []string
	transactionCalls  int
	usernameLookups   int
	sessionLookups    int
	userLookups       int
	createCalls       int
	deleteCalls       int
	updateCalls       int
	transactionDepth  int
	mutationOutsideTx bool
}

func newSessionMatrixStore() *sessionMatrixStore {
	return &sessionMatrixStore{
		sessions: make(map[string]application.SessionRecord),
		users:    make(map[int64]sessionMatrixUserResult),
	}
}

func (store *sessionMatrixStore) Transaction(ctx context.Context, apply func(application.Store) error) error {
	store.transactionCalls++
	store.events = append(store.events, "transaction.begin")
	snapshot := cloneSessionMatrixRows(store.sessions)
	store.transactionDepth++
	err := apply(store)
	store.transactionDepth--
	if err != nil {
		store.sessions = snapshot
		store.events = append(store.events, "transaction.rollback")
		return err
	}
	if store.commitErr != nil {
		store.sessions = snapshot
		store.events = append(store.events, "transaction.rollback")
		return store.commitErr
	}
	store.events = append(store.events, "transaction.commit")
	return nil
}

func (store *sessionMatrixStore) UserByUsername(context.Context, string) (domain.User, string, error) {
	store.usernameLookups++
	store.events = append(store.events, "user.by_username")
	if len(store.usernameResults) == 0 {
		return domain.User{}, "", application.ErrNotFound
	}
	index := store.usernameLookups - 1
	if index >= len(store.usernameResults) {
		index = len(store.usernameResults) - 1
	}
	result := store.usernameResults[index]
	return result.user, result.hash, result.err
}

func (store *sessionMatrixStore) UserByID(_ context.Context, id int64) (domain.User, string, error) {
	store.userLookups++
	store.events = append(store.events, "user.by_id")
	result, ok := store.users[id]
	if !ok {
		return domain.User{}, "", application.ErrNotFound
	}
	return result.user, result.hash, result.err
}

func (store *sessionMatrixStore) SessionByHash(_ context.Context, sessionHash []byte) (application.SessionRecord, error) {
	store.sessionLookups++
	store.events = append(store.events, "session.by_hash")
	if store.sessionLookupErr != nil {
		return application.SessionRecord{}, store.sessionLookupErr
	}
	record, ok := store.sessions[string(sessionHash)]
	if !ok {
		return application.SessionRecord{}, application.ErrNotFound
	}
	return cloneSessionMatrixRecord(record), nil
}

func (store *sessionMatrixStore) CreateSession(_ context.Context, record application.SessionRecord) error {
	store.createCalls++
	store.events = append(store.events, "session.create")
	if store.transactionDepth == 0 {
		store.mutationOutsideTx = true
	}
	if store.createErr != nil {
		return store.createErr
	}
	store.sessions[string(record.SecretHash)] = cloneSessionMatrixRecord(record)
	return nil
}

func (store *sessionMatrixStore) DeleteSession(_ context.Context, sessionHash []byte) error {
	store.deleteCalls++
	store.events = append(store.events, "session.delete")
	if store.transactionDepth == 0 {
		store.mutationOutsideTx = true
	}
	if store.deleteErr != nil {
		return store.deleteErr
	}
	key := string(sessionHash)
	if _, ok := store.sessions[key]; !ok {
		return application.ErrNotFound
	}
	delete(store.sessions, key)
	return nil
}

func (store *sessionMatrixStore) UpdateSessionCSRF(_ context.Context, sessionHash, csrfHash []byte) error {
	store.updateCalls++
	store.events = append(store.events, "session.update_csrf")
	if store.transactionDepth == 0 {
		store.mutationOutsideTx = true
	}
	if store.updateErr != nil {
		return store.updateErr
	}
	key := string(sessionHash)
	record, ok := store.sessions[key]
	if !ok {
		return application.ErrNotFound
	}
	record.CSRFHash = append([]byte(nil), csrfHash...)
	store.sessions[key] = record
	return nil
}

type sessionMatrixClock struct {
	now   time.Time
	calls int
}

func (clock *sessionMatrixClock) Now() time.Time {
	clock.calls++
	return clock.now
}

type sessionCSRFRecoveryAPI interface {
	CSRFForSession(context.Context, string) (string, error)
}

type transactionalSessionLogoutAPI interface {
	Logout(context.Context, string, string) error
}

func cloneSessionMatrixRows(source map[string]application.SessionRecord) map[string]application.SessionRecord {
	clone := make(map[string]application.SessionRecord, len(source))
	for key, record := range source {
		clone[key] = cloneSessionMatrixRecord(record)
	}
	return clone
}

func cloneSessionMatrixRecord(record application.SessionRecord) application.SessionRecord {
	record.SecretHash = append([]byte(nil), record.SecretHash...)
	record.CSRFHash = append([]byte(nil), record.CSRFHash...)
	return record
}

func sessionMatrixOpaqueValue(seed byte) string {
	material := make([]byte, 32)
	for index := range material {
		material[index] = seed + byte(index)
	}
	return base64.RawURLEncoding.EncodeToString(material)
}

func sessionMatrixDigest(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}

func sessionMatrixDerivedCSRF(sessionSecret string) string {
	mac := hmac.New(sha256.New, []byte(sessionSecret))
	_, _ = mac.Write([]byte(sessionMatrixCSRFDomainTag))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func sessionMatrixRecord(sessionSecret, csrf string, userID int64, expires time.Time) application.SessionRecord {
	created := expires.Add(-time.Hour)
	return application.SessionRecord{
		SecretHash: sessionMatrixDigest(sessionSecret),
		CSRFHash:   sessionMatrixDigest(csrf),
		UserID:     userID,
		Created:    created,
		LastSeen:   created,
		Expires:    expires,
	}
}

func sessionMatrixBytesEqual(left, right []byte) bool {
	return len(left) == len(right) && subtle.ConstantTimeCompare(left, right) == 1
}

func sessionMatrixStringsEqual(left, right string) bool {
	return len(left) == len(right) && subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

func sessionMatrixHasRow(store *sessionMatrixStore, sessionSecret string) bool {
	_, ok := store.sessions[string(sessionMatrixDigest(sessionSecret))]
	return ok
}

func sessionMatrixRow(store *sessionMatrixStore, sessionSecret string) (application.SessionRecord, bool) {
	record, ok := store.sessions[string(sessionMatrixDigest(sessionSecret))]
	return record, ok
}

func assertSessionMatrixFailure(t *testing.T, err error, kind application.SessionCredentialFailureKind) {
	t.Helper()
	if err == nil {
		t.Error("expected classified session rejection")
		return
	}
	if shared.ReasonOf(err) != shared.ErrorReasonUnauthenticated {
		t.Error("session rejection did not retain the unauthenticated reason")
	}
	if err.Error() != "Authentication credentials were not provided." {
		t.Error("session rejection did not retain the generic public message")
	}
	var failure *application.SessionCredentialFailure
	if !errors.As(err, &failure) {
		t.Error("session rejection did not expose a typed transport-neutral cause")
		return
	}
	if failure.Kind != kind {
		t.Error("session rejection used the wrong transport-neutral kind")
	}
	if !application.SessionCredentialAllowsTokenFallback(err) {
		t.Error("expected classified session rejection to permit credential fallback")
	}
}

func assertSessionMatrixReason(t *testing.T, err error, reason shared.ErrorReason, cause error) {
	t.Helper()
	if err == nil {
		t.Error("expected application error")
		return
	}
	if shared.ReasonOf(err) != reason {
		t.Error("application error used the wrong reason")
	}
	var failure *application.SessionCredentialFailure
	if errors.As(err, &failure) {
		t.Error("non-session outcome unexpectedly exposed a session credential cause")
	}
	if application.SessionCredentialAllowsTokenFallback(err) {
		t.Error("non-session outcome unexpectedly permitted credential fallback")
	}
	if cause != nil && !errors.Is(err, cause) {
		t.Error("application error did not preserve its infrastructure cause")
	}
}

func assertNoSessionMatrixMaterial(t *testing.T, session domain.BrowserSession) {
	t.Helper()
	if session.User.ID != 0 || session.Secret != "" || session.CSRFToken != "" || !session.Expires.IsZero() {
		t.Error("failed login returned browser-session material")
	}
}

func assertNoSessionMatrixString(t *testing.T, value string) {
	t.Helper()
	if value != "" {
		t.Error("failed operation returned credential material")
	}
}

func requireSessionMatrixNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal("expected operation to succeed")
	}
}

func assertSessionMatrixRowCount(t *testing.T, store *sessionMatrixStore, want int) {
	t.Helper()
	if len(store.sessions) != want {
		t.Error("session store contained an unexpected number of rows")
	}
}

func TestSessionCredentialOutcomeClassification(t *testing.T) {
	now := time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC)
	sessionSecret := sessionMatrixOpaqueValue(0x11)
	csrf := sessionMatrixOpaqueValue(0x31)
	activeUser := domain.User{ID: 41, Username: "active-user", IsActive: true}
	inactiveUser := activeUser
	inactiveUser.IsActive = false
	lookupFailure := errors.New("session lookup unavailable")
	ownerFailure := errors.New("session owner unavailable")

	t.Run("fallback helper is limited to the four classified outcomes", func(t *testing.T) {
		for _, kind := range []application.SessionCredentialFailureKind{
			application.SessionCredentialFailureMissing,
			application.SessionCredentialFailureUnknown,
			application.SessionCredentialFailureExpired,
			application.SessionCredentialFailureInactiveOwner,
		} {
			err := shared.WrapError(
				shared.ErrorReasonUnauthenticated,
				"Authentication credentials were not provided.",
				&application.SessionCredentialFailure{Kind: kind},
			)
			if !application.SessionCredentialAllowsTokenFallback(err) {
				t.Error("known session rejection did not permit credential fallback")
			}
		}
		unknownKind := shared.WrapError(
			shared.ErrorReasonUnauthenticated,
			"Authentication credentials were not provided.",
			&application.SessionCredentialFailure{Kind: 255},
		)
		if application.SessionCredentialAllowsTokenFallback(unknownKind) {
			t.Error("unknown session rejection kind permitted credential fallback")
		}
		if application.SessionCredentialAllowsTokenFallback(shared.Unauthenticated()) {
			t.Error("generic unauthenticated outcome permitted credential fallback")
		}
		internalTyped := shared.WrapError(
			shared.ErrorReasonInternal,
			"An internal error occurred.",
			&application.SessionCredentialFailure{Kind: application.SessionCredentialFailureUnknown},
		)
		if application.SessionCredentialAllowsTokenFallback(internalTyped) {
			t.Error("internal outcome permitted credential fallback")
		}
	})

	authenticationCases := []struct {
		name             string
		secret           string
		hasSession       bool
		expires          time.Time
		sessionLookupErr error
		owner            sessionMatrixUserResult
		wantKind         application.SessionCredentialFailureKind
		wantReason       shared.ErrorReason
		wantCause        error
		wantSessionCalls int
		wantUserCalls    int
		wantClockCalls   int
		wantEvents       []string
		wantUser         bool
	}{
		{
			name:       "missing secret",
			wantKind:   application.SessionCredentialFailureMissing,
			wantReason: shared.ErrorReasonUnauthenticated,
			wantEvents: nil,
		},
		{
			name:             "unknown digest",
			secret:           sessionSecret,
			wantKind:         application.SessionCredentialFailureUnknown,
			wantReason:       shared.ErrorReasonUnauthenticated,
			wantSessionCalls: 1,
			wantEvents:       []string{"session.by_hash"},
		},
		{
			name:             "expired just before boundary",
			secret:           sessionSecret,
			hasSession:       true,
			expires:          now.Add(-time.Nanosecond),
			owner:            sessionMatrixUserResult{user: activeUser},
			wantKind:         application.SessionCredentialFailureExpired,
			wantReason:       shared.ErrorReasonUnauthenticated,
			wantSessionCalls: 1,
			wantClockCalls:   1,
			wantEvents:       []string{"session.by_hash"},
		},
		{
			name:             "expired at boundary",
			secret:           sessionSecret,
			hasSession:       true,
			expires:          now,
			owner:            sessionMatrixUserResult{user: activeUser},
			wantKind:         application.SessionCredentialFailureExpired,
			wantReason:       shared.ErrorReasonUnauthenticated,
			wantSessionCalls: 1,
			wantClockCalls:   1,
			wantEvents:       []string{"session.by_hash"},
		},
		{
			name:             "inactive owner after valid session",
			secret:           sessionSecret,
			hasSession:       true,
			expires:          now.Add(time.Nanosecond),
			owner:            sessionMatrixUserResult{user: inactiveUser},
			wantKind:         application.SessionCredentialFailureInactiveOwner,
			wantReason:       shared.ErrorReasonUnauthenticated,
			wantSessionCalls: 1,
			wantUserCalls:    1,
			wantClockCalls:   1,
			wantEvents:       []string{"session.by_hash", "user.by_id"},
		},
		{
			name:             "session infrastructure failure",
			secret:           sessionSecret,
			sessionLookupErr: lookupFailure,
			wantReason:       shared.ErrorReasonInternal,
			wantCause:        lookupFailure,
			wantSessionCalls: 1,
			wantEvents:       []string{"session.by_hash"},
		},
		{
			name:             "orphan owner",
			secret:           sessionSecret,
			hasSession:       true,
			expires:          now.Add(time.Nanosecond),
			owner:            sessionMatrixUserResult{err: application.ErrNotFound},
			wantReason:       shared.ErrorReasonInternal,
			wantCause:        application.ErrNotFound,
			wantSessionCalls: 1,
			wantUserCalls:    1,
			wantClockCalls:   1,
			wantEvents:       []string{"session.by_hash", "user.by_id"},
		},
		{
			name:             "owner infrastructure failure",
			secret:           sessionSecret,
			hasSession:       true,
			expires:          now.Add(time.Nanosecond),
			owner:            sessionMatrixUserResult{err: ownerFailure},
			wantReason:       shared.ErrorReasonInternal,
			wantCause:        ownerFailure,
			wantSessionCalls: 1,
			wantUserCalls:    1,
			wantClockCalls:   1,
			wantEvents:       []string{"session.by_hash", "user.by_id"},
		},
		{
			name:             "valid just after boundary",
			secret:           sessionSecret,
			hasSession:       true,
			expires:          now.Add(time.Nanosecond),
			owner:            sessionMatrixUserResult{user: activeUser},
			wantSessionCalls: 1,
			wantUserCalls:    1,
			wantClockCalls:   1,
			wantEvents:       []string{"session.by_hash", "user.by_id"},
			wantUser:         true,
		},
	}

	for _, test := range authenticationCases {
		t.Run("authenticate session "+test.name, func(t *testing.T) {
			store := newSessionMatrixStore()
			store.sessionLookupErr = test.sessionLookupErr
			if test.hasSession {
				store.sessions[string(sessionMatrixDigest(sessionSecret))] = sessionMatrixRecord(
					sessionSecret,
					csrf,
					activeUser.ID,
					test.expires,
				)
				store.users[activeUser.ID] = test.owner
			}
			clock := &sessionMatrixClock{now: now}
			service := application.NewService(store, clock)

			user, err := service.AuthenticateSession(t.Context(), test.secret)

			if test.wantUser {
				requireSessionMatrixNoError(t, err)
				require.Equal(t, activeUser, user)
			} else {
				require.Equal(t, domain.User{}, user)
				if test.wantKind != 0 {
					assertSessionMatrixFailure(t, err, test.wantKind)
				} else {
					assertSessionMatrixReason(t, err, test.wantReason, test.wantCause)
				}
			}
			require.Equal(t, test.wantSessionCalls, store.sessionLookups)
			require.Equal(t, test.wantUserCalls, store.userLookups)
			require.Equal(t, test.wantClockCalls, clock.calls)
			require.Equal(t, test.wantEvents, store.events)
		})
	}

	currentUserCases := []struct {
		name            string
		principal       domain.Principal
		owner           sessionMatrixUserResult
		wantReason      shared.ErrorReason
		wantCause       error
		wantUser        bool
		wantUserLookups int
	}{
		{
			name:       "unauthenticated principal",
			wantReason: shared.ErrorReasonUnauthenticated,
		},
		{
			name:            "active row",
			principal:       activeUser.Principal(),
			owner:           sessionMatrixUserResult{user: activeUser},
			wantUser:        true,
			wantUserLookups: 1,
		},
		{
			name:            "inactive row",
			principal:       activeUser.Principal(),
			owner:           sessionMatrixUserResult{user: inactiveUser},
			wantReason:      shared.ErrorReasonUnauthenticated,
			wantUserLookups: 1,
		},
		{
			name:            "missing row",
			principal:       activeUser.Principal(),
			owner:           sessionMatrixUserResult{err: application.ErrNotFound},
			wantReason:      shared.ErrorReasonInternal,
			wantCause:       application.ErrNotFound,
			wantUserLookups: 1,
		},
		{
			name:            "lookup failure",
			principal:       activeUser.Principal(),
			owner:           sessionMatrixUserResult{err: ownerFailure},
			wantReason:      shared.ErrorReasonInternal,
			wantCause:       ownerFailure,
			wantUserLookups: 1,
		},
	}

	for _, test := range currentUserCases {
		t.Run("current user "+test.name, func(t *testing.T) {
			store := newSessionMatrixStore()
			store.users[activeUser.ID] = test.owner
			clock := &sessionMatrixClock{now: now}
			service := application.NewService(store, clock)

			user, err := service.CurrentUser(t.Context(), test.principal)

			if test.wantUser {
				requireSessionMatrixNoError(t, err)
				require.Equal(t, activeUser, user)
			} else {
				require.Equal(t, domain.User{}, user)
				assertSessionMatrixReason(t, err, test.wantReason, test.wantCause)
			}
			require.Equal(t, test.wantUserLookups, store.userLookups)
			require.Zero(t, store.sessionLookups)
			require.Zero(t, clock.calls)
		})
	}
}

func TestSessionCSRFOutcomeClassification(t *testing.T) {
	now := time.Date(2026, 8, 17, 11, 0, 0, 0, time.UTC)
	sessionSecret := sessionMatrixOpaqueValue(0x12)
	storedCSRF := sessionMatrixOpaqueValue(0x32)
	mismatchedCSRF := sessionMatrixOpaqueValue(0x52)
	activeUser := domain.User{ID: 42, Username: "csrf-user", IsActive: true}
	inactiveUser := activeUser
	inactiveUser.IsActive = false
	lookupFailure := errors.New("session lookup unavailable")
	ownerFailure := errors.New("session owner unavailable")

	tests := []struct {
		name             string
		sessionSecret    string
		csrf             string
		hasSession       bool
		expires          time.Time
		sessionLookupErr error
		owner            sessionMatrixUserResult
		wantKind         application.SessionCredentialFailureKind
		wantReason       shared.ErrorReason
		wantCause        error
		wantSessionCalls int
		wantUserCalls    int
		wantClockCalls   int
		wantEvents       []string
	}{
		{
			name:       "missing session precedes csrf",
			csrf:       storedCSRF,
			wantKind:   application.SessionCredentialFailureMissing,
			wantReason: shared.ErrorReasonUnauthenticated,
			wantEvents: nil,
		},
		{
			name:             "unknown session precedes csrf",
			sessionSecret:    sessionSecret,
			csrf:             storedCSRF,
			wantKind:         application.SessionCredentialFailureUnknown,
			wantReason:       shared.ErrorReasonUnauthenticated,
			wantSessionCalls: 1,
			wantEvents:       []string{"session.by_hash"},
		},
		{
			name:             "expired just before boundary",
			sessionSecret:    sessionSecret,
			csrf:             storedCSRF,
			hasSession:       true,
			expires:          now.Add(-time.Nanosecond),
			owner:            sessionMatrixUserResult{user: activeUser},
			wantKind:         application.SessionCredentialFailureExpired,
			wantReason:       shared.ErrorReasonUnauthenticated,
			wantSessionCalls: 1,
			wantClockCalls:   1,
			wantEvents:       []string{"session.by_hash"},
		},
		{
			name:             "expired at boundary",
			sessionSecret:    sessionSecret,
			csrf:             storedCSRF,
			hasSession:       true,
			expires:          now,
			owner:            sessionMatrixUserResult{user: activeUser},
			wantKind:         application.SessionCredentialFailureExpired,
			wantReason:       shared.ErrorReasonUnauthenticated,
			wantSessionCalls: 1,
			wantClockCalls:   1,
			wantEvents:       []string{"session.by_hash"},
		},
		{
			name:             "inactive owner precedes csrf",
			sessionSecret:    sessionSecret,
			csrf:             storedCSRF,
			hasSession:       true,
			expires:          now.Add(time.Nanosecond),
			owner:            sessionMatrixUserResult{user: inactiveUser},
			wantKind:         application.SessionCredentialFailureInactiveOwner,
			wantReason:       shared.ErrorReasonUnauthenticated,
			wantSessionCalls: 1,
			wantUserCalls:    1,
			wantClockCalls:   1,
			wantEvents:       []string{"session.by_hash", "user.by_id"},
		},
		{
			name:             "session infrastructure failure",
			sessionSecret:    sessionSecret,
			csrf:             storedCSRF,
			sessionLookupErr: lookupFailure,
			wantReason:       shared.ErrorReasonInternal,
			wantCause:        lookupFailure,
			wantSessionCalls: 1,
			wantEvents:       []string{"session.by_hash"},
		},
		{
			name:             "orphan owner",
			sessionSecret:    sessionSecret,
			csrf:             storedCSRF,
			hasSession:       true,
			expires:          now.Add(time.Nanosecond),
			owner:            sessionMatrixUserResult{err: application.ErrNotFound},
			wantReason:       shared.ErrorReasonInternal,
			wantCause:        application.ErrNotFound,
			wantSessionCalls: 1,
			wantUserCalls:    1,
			wantClockCalls:   1,
			wantEvents:       []string{"session.by_hash", "user.by_id"},
		},
		{
			name:             "owner infrastructure failure",
			sessionSecret:    sessionSecret,
			csrf:             storedCSRF,
			hasSession:       true,
			expires:          now.Add(time.Nanosecond),
			owner:            sessionMatrixUserResult{err: ownerFailure},
			wantReason:       shared.ErrorReasonInternal,
			wantCause:        ownerFailure,
			wantSessionCalls: 1,
			wantUserCalls:    1,
			wantClockCalls:   1,
			wantEvents:       []string{"session.by_hash", "user.by_id"},
		},
		{
			name:             "empty csrf after valid session",
			sessionSecret:    sessionSecret,
			hasSession:       true,
			expires:          now.Add(time.Nanosecond),
			owner:            sessionMatrixUserResult{user: activeUser},
			wantReason:       shared.ErrorReasonForbidden,
			wantSessionCalls: 1,
			wantUserCalls:    1,
			wantClockCalls:   1,
			wantEvents:       []string{"session.by_hash", "user.by_id"},
		},
		{
			name:             "mismatched csrf after valid session",
			sessionSecret:    sessionSecret,
			csrf:             mismatchedCSRF,
			hasSession:       true,
			expires:          now.Add(time.Nanosecond),
			owner:            sessionMatrixUserResult{user: activeUser},
			wantReason:       shared.ErrorReasonForbidden,
			wantSessionCalls: 1,
			wantUserCalls:    1,
			wantClockCalls:   1,
			wantEvents:       []string{"session.by_hash", "user.by_id"},
		},
		{
			name:             "matching csrf after valid session",
			sessionSecret:    sessionSecret,
			csrf:             storedCSRF,
			hasSession:       true,
			expires:          now.Add(time.Nanosecond),
			owner:            sessionMatrixUserResult{user: activeUser},
			wantSessionCalls: 1,
			wantUserCalls:    1,
			wantClockCalls:   1,
			wantEvents:       []string{"session.by_hash", "user.by_id"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newSessionMatrixStore()
			store.sessionLookupErr = test.sessionLookupErr
			if test.hasSession {
				record := sessionMatrixRecord(sessionSecret, storedCSRF, activeUser.ID, test.expires)
				store.sessions[string(record.SecretHash)] = record
				store.users[activeUser.ID] = test.owner
			}
			before, hadBefore := sessionMatrixRow(store, sessionSecret)
			clock := &sessionMatrixClock{now: now}
			service := application.NewService(store, clock)

			err := service.VerifyCSRF(t.Context(), test.sessionSecret, test.csrf)

			if test.wantReason == "" {
				requireSessionMatrixNoError(t, err)
			} else if test.wantKind != 0 {
				assertSessionMatrixFailure(t, err, test.wantKind)
			} else {
				assertSessionMatrixReason(t, err, test.wantReason, test.wantCause)
			}
			after, hasAfter := sessionMatrixRow(store, sessionSecret)
			if hadBefore != hasAfter ||
				(hadBefore && !sessionMatrixBytesEqual(before.CSRFHash, after.CSRFHash)) {
				t.Error("csrf verification mutated the stored session")
			}
			require.Equal(t, test.wantSessionCalls, store.sessionLookups)
			require.Equal(t, test.wantUserCalls, store.userLookups)
			require.Equal(t, test.wantClockCalls, clock.calls)
			require.Equal(t, test.wantEvents, store.events)
			require.Zero(t, store.transactionCalls)
			require.False(t, store.mutationOutsideTx)
		})
	}
}

func TestLoginUsesOneTransactionForPasswordRevalidationAndRotation(t *testing.T) {
	now := time.Date(
		2026, 8, 17, 15, 0, 0, 0,
		time.FixedZone("login-test-offset", 3*60*60),
	)
	expectedNow := now.UTC()
	password := "application-login-password-2026"
	incorrectPassword := "nonmatching-login-credential-2026"
	passwordHash, hashErr := application.HashPassword(password)
	requireSessionMatrixNoError(t, hashErr)
	activeUser := domain.User{
		ID:       51,
		Username: "login-user",
		Email:    "before-revalidation@example.test",
		IsActive: true,
	}
	revalidatedUser := activeUser
	revalidatedUser.Email = "after-revalidation@example.test"
	priorSecret := sessionMatrixOpaqueValue(0x13)
	priorCSRF := sessionMatrixOpaqueValue(0x33)
	lookupFailure := errors.New("password lookup unavailable")
	createFailure := errors.New("session insert unavailable")
	deleteFailure := errors.New("session delete unavailable")
	commitFailure := errors.New("session transaction commit unavailable")

	newLoginStore := func() *sessionMatrixStore {
		store := newSessionMatrixStore()
		store.usernameResults = []sessionMatrixUserResult{
			{user: activeUser, hash: passwordHash},
			{user: revalidatedUser, hash: passwordHash},
		}
		return store
	}

	assertSuccessfulLogin := func(
		t *testing.T,
		store *sessionMatrixStore,
		clock *sessionMatrixClock,
		session domain.BrowserSession,
	) {
		t.Helper()
		if session.Secret == "" || session.CSRFToken == "" {
			t.Error("successful login omitted browser-session material")
			return
		}
		decodedSecret, decodeErr := base64.RawURLEncoding.DecodeString(session.Secret)
		if decodeErr != nil || len(decodedSecret) != 32 {
			t.Error("session secret was not a raw-url encoded 256-bit value")
		}
		expectedCSRF := sessionMatrixDerivedCSRF(session.Secret)
		if !sessionMatrixStringsEqual(session.CSRFToken, expectedCSRF) {
			t.Error("csrf value did not use the exact session-bound hmac derivation")
		}
		record, ok := sessionMatrixRow(store, session.Secret)
		if !ok {
			t.Error("successful login did not persist the returned session")
			return
		}
		if !sessionMatrixBytesEqual(record.SecretHash, sessionMatrixDigest(session.Secret)) {
			t.Error("stored session digest did not bind the returned session")
		}
		if !sessionMatrixBytesEqual(record.CSRFHash, sessionMatrixDigest(expectedCSRF)) {
			t.Error("stored csrf digest did not bind the derived csrf value")
		}
		if record.Created != expectedNow || record.LastSeen != expectedNow {
			t.Error("login did not use one clock sample for created and last-seen")
		}
		if record.Expires != expectedNow.Add(12*time.Hour) || session.Expires != record.Expires {
			t.Error("login did not apply the fixed twelve-hour session lifetime")
		}
		for _, timestamp := range []time.Time{
			record.Created,
			record.LastSeen,
			record.Expires,
			session.Expires,
		} {
			_, offset := timestamp.Zone()
			if timestamp.Location() != time.UTC || offset != 0 {
				t.Error("login timestamp was not normalized to UTC")
				break
			}
		}
		if record.UserID != revalidatedUser.ID {
			t.Error("login did not return the transactionally revalidated user")
		}
		require.Equal(t, revalidatedUser, session.User)
		if clock.calls != 1 {
			t.Error("successful login did not take exactly one application-clock sample")
		}
	}

	t.Run("fresh login revalidates and creates inside one transaction", func(t *testing.T) {
		store := newLoginStore()
		clock := &sessionMatrixClock{now: now}
		service := application.NewService(store, clock)

		session, err := service.LoginReplacing(t.Context(), activeUser.Username, password, "")

		requireSessionMatrixNoError(t, err)
		assertSuccessfulLogin(t, store, clock, session)
		require.Equal(t, 1, store.transactionCalls)
		require.Equal(t, 2, store.usernameLookups)
		require.Equal(t, 1, store.createCalls)
		require.Zero(t, store.deleteCalls)
		require.False(t, store.mutationOutsideTx)
		require.Equal(t, []string{
			"user.by_username",
			"transaction.begin",
			"user.by_username",
			"session.create",
			"transaction.commit",
		}, store.events)
	})

	t.Run("replacing login revalidates before delete and rotates both values", func(t *testing.T) {
		store := newLoginStore()
		prior := sessionMatrixRecord(priorSecret, priorCSRF, activeUser.ID, now.Add(time.Hour))
		store.sessions[string(prior.SecretHash)] = prior
		clock := &sessionMatrixClock{now: now}
		service := application.NewService(store, clock)

		session, err := service.LoginReplacing(t.Context(), activeUser.Username, password, priorSecret)

		requireSessionMatrixNoError(t, err)
		assertSuccessfulLogin(t, store, clock, session)
		if sessionMatrixStringsEqual(session.Secret, priorSecret) ||
			sessionMatrixStringsEqual(session.CSRFToken, priorCSRF) {
			t.Error("replacing login did not rotate browser-session material")
		}
		if sessionMatrixHasRow(store, priorSecret) {
			t.Error("replacing login retained the prior session after commit")
		}
		assertSessionMatrixRowCount(t, store, 1)
		require.Equal(t, 1, store.transactionCalls)
		require.Equal(t, 2, store.usernameLookups)
		require.Equal(t, 1, store.deleteCalls)
		require.Equal(t, 1, store.createCalls)
		require.False(t, store.mutationOutsideTx)
		require.Equal(t, []string{
			"user.by_username",
			"transaction.begin",
			"user.by_username",
			"session.delete",
			"session.create",
			"transaction.commit",
		}, store.events)
	})

	t.Run("initial credential outcomes do not enter a transaction", func(t *testing.T) {
		inactiveUser := activeUser
		inactiveUser.IsActive = false
		corruptVerifier := "unusable-verifier-fixture"
		tests := []struct {
			name        string
			username    string
			password    string
			lookup      sessionMatrixUserResult
			wantReason  shared.ErrorReason
			wantCause   error
			wantLookups int
		}{
			{
				name:       "empty username",
				password:   password,
				wantReason: shared.ErrorReasonUnauthenticated,
			},
			{
				name:       "empty credential",
				username:   activeUser.Username,
				wantReason: shared.ErrorReasonUnauthenticated,
			},
			{
				name:        "missing username",
				username:    activeUser.Username,
				password:    password,
				lookup:      sessionMatrixUserResult{err: application.ErrNotFound},
				wantReason:  shared.ErrorReasonUnauthenticated,
				wantLookups: 1,
			},
			{
				name:        "nonmatching credential",
				username:    activeUser.Username,
				password:    incorrectPassword,
				lookup:      sessionMatrixUserResult{user: activeUser, hash: passwordHash},
				wantReason:  shared.ErrorReasonUnauthenticated,
				wantLookups: 1,
			},
			{
				name:        "inactive owner with matching credential",
				username:    activeUser.Username,
				password:    password,
				lookup:      sessionMatrixUserResult{user: inactiveUser, hash: passwordHash},
				wantReason:  shared.ErrorReasonUnauthenticated,
				wantLookups: 1,
			},
			{
				name:        "inactive owner with nonmatching credential",
				username:    activeUser.Username,
				password:    incorrectPassword,
				lookup:      sessionMatrixUserResult{user: inactiveUser, hash: passwordHash},
				wantReason:  shared.ErrorReasonUnauthenticated,
				wantLookups: 1,
			},
			{
				name:        "active owner with corrupt verifier",
				username:    activeUser.Username,
				password:    password,
				lookup:      sessionMatrixUserResult{user: activeUser, hash: corruptVerifier},
				wantReason:  shared.ErrorReasonInternal,
				wantCause:   bcrypt.ErrHashTooShort,
				wantLookups: 1,
			},
			{
				name:        "inactive owner with corrupt verifier",
				username:    activeUser.Username,
				password:    password,
				lookup:      sessionMatrixUserResult{user: inactiveUser, hash: corruptVerifier},
				wantReason:  shared.ErrorReasonInternal,
				wantCause:   bcrypt.ErrHashTooShort,
				wantLookups: 1,
			},
			{
				name:        "infrastructure failure",
				username:    activeUser.Username,
				password:    password,
				lookup:      sessionMatrixUserResult{err: lookupFailure},
				wantReason:  shared.ErrorReasonInternal,
				wantCause:   lookupFailure,
				wantLookups: 1,
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				store := newSessionMatrixStore()
				store.usernameResults = []sessionMatrixUserResult{test.lookup}
				clock := &sessionMatrixClock{now: now}
				service := application.NewService(store, clock)

				session, err := service.LoginReplacing(t.Context(), test.username, test.password, "")

				assertSessionMatrixReason(t, err, test.wantReason, test.wantCause)
				assertNoSessionMatrixMaterial(t, session)
				require.Equal(t, test.wantLookups, store.usernameLookups)
				require.Zero(t, store.transactionCalls)
				require.Zero(t, store.createCalls)
				require.Zero(t, clock.calls)
			})
		}
	})

	t.Run("transaction rejects stale password identity state before mutation", func(t *testing.T) {
		changedSuffix := "A"
		if passwordHash[len(passwordHash)-1:] == changedSuffix {
			changedSuffix = "B"
		}
		changedHash := passwordHash[:len(passwordHash)-1] + changedSuffix
		changedIdentity := revalidatedUser
		changedIdentity.ID++
		inactiveUser := revalidatedUser
		inactiveUser.IsActive = false
		tests := []struct {
			name        string
			revalidated sessionMatrixUserResult
			wantReason  shared.ErrorReason
			wantCause   error
		}{
			{
				name:        "password hash changed",
				revalidated: sessionMatrixUserResult{user: revalidatedUser, hash: changedHash},
				wantReason:  shared.ErrorReasonUnauthenticated,
			},
			{
				name:        "identity changed",
				revalidated: sessionMatrixUserResult{user: changedIdentity, hash: passwordHash},
				wantReason:  shared.ErrorReasonUnauthenticated,
			},
			{
				name:        "owner became inactive",
				revalidated: sessionMatrixUserResult{user: inactiveUser, hash: passwordHash},
				wantReason:  shared.ErrorReasonUnauthenticated,
			},
			{
				name:        "owner disappeared",
				revalidated: sessionMatrixUserResult{err: application.ErrNotFound},
				wantReason:  shared.ErrorReasonUnauthenticated,
			},
			{
				name:        "revalidation infrastructure failure",
				revalidated: sessionMatrixUserResult{err: lookupFailure},
				wantReason:  shared.ErrorReasonInternal,
				wantCause:   lookupFailure,
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				store := newSessionMatrixStore()
				store.usernameResults = []sessionMatrixUserResult{
					{user: activeUser, hash: passwordHash},
					test.revalidated,
				}
				clock := &sessionMatrixClock{now: now}
				service := application.NewService(store, clock)

				session, err := service.LoginReplacing(t.Context(), activeUser.Username, password, "")

				assertSessionMatrixReason(t, err, test.wantReason, test.wantCause)
				assertNoSessionMatrixMaterial(t, session)
				require.Equal(t, 1, store.transactionCalls)
				require.Equal(t, 2, store.usernameLookups)
				require.Zero(t, store.createCalls)
				require.Zero(t, store.deleteCalls)
				require.Zero(t, clock.calls)
				assertSessionMatrixRowCount(t, store, 0)
				require.False(t, store.mutationOutsideTx)
				require.Equal(t, []string{
					"user.by_username",
					"transaction.begin",
					"user.by_username",
					"transaction.rollback",
				}, store.events)
			})
		}
	})

	t.Run("insert failure rolls replacing deletion back", func(t *testing.T) {
		store := newLoginStore()
		prior := sessionMatrixRecord(priorSecret, priorCSRF, activeUser.ID, now.Add(time.Hour))
		store.sessions[string(prior.SecretHash)] = prior
		store.createErr = createFailure
		clock := &sessionMatrixClock{now: now}
		service := application.NewService(store, clock)

		session, err := service.LoginReplacing(t.Context(), activeUser.Username, password, priorSecret)

		assertSessionMatrixReason(t, err, shared.ErrorReasonInternal, createFailure)
		assertNoSessionMatrixMaterial(t, session)
		if !sessionMatrixHasRow(store, priorSecret) || len(store.sessions) != 1 {
			t.Error("failed replacing login did not restore the prior session")
		}
		require.Equal(t, 1, store.transactionCalls)
		require.Equal(t, 2, store.usernameLookups)
		require.Equal(t, 1, store.deleteCalls)
		require.Equal(t, 1, store.createCalls)
		require.Equal(t, 1, clock.calls)
		require.False(t, store.mutationOutsideTx)
		require.Equal(t, []string{
			"user.by_username",
			"transaction.begin",
			"user.by_username",
			"session.delete",
			"session.create",
			"transaction.rollback",
		}, store.events)
	})

	t.Run("delete failure prevents replacement and preserves prior session", func(t *testing.T) {
		store := newLoginStore()
		prior := sessionMatrixRecord(priorSecret, priorCSRF, activeUser.ID, now.Add(time.Hour))
		store.sessions[string(prior.SecretHash)] = prior
		store.deleteErr = deleteFailure
		clock := &sessionMatrixClock{now: now}
		service := application.NewService(store, clock)

		session, err := service.LoginReplacing(t.Context(), activeUser.Username, password, priorSecret)

		assertSessionMatrixReason(t, err, shared.ErrorReasonInternal, deleteFailure)
		assertNoSessionMatrixMaterial(t, session)
		if !sessionMatrixHasRow(store, priorSecret) || len(store.sessions) != 1 {
			t.Error("failed replacement deletion changed durable session state")
		}
		require.Equal(t, 1, store.transactionCalls)
		require.Equal(t, 2, store.usernameLookups)
		require.Equal(t, 1, store.deleteCalls)
		require.Zero(t, store.createCalls)
		require.Equal(t, 1, clock.calls)
		require.False(t, store.mutationOutsideTx)
	})

	t.Run("commit failure rolls fresh insertion back and returns no material", func(t *testing.T) {
		store := newLoginStore()
		store.commitErr = commitFailure
		clock := &sessionMatrixClock{now: now}
		service := application.NewService(store, clock)

		session, err := service.LoginReplacing(t.Context(), activeUser.Username, password, "")

		assertSessionMatrixReason(t, err, shared.ErrorReasonInternal, commitFailure)
		assertNoSessionMatrixMaterial(t, session)
		assertSessionMatrixRowCount(t, store, 0)
		require.Equal(t, 1, store.transactionCalls)
		require.Equal(t, 2, store.usernameLookups)
		require.Equal(t, 1, store.createCalls)
		require.Equal(t, 1, clock.calls)
		require.False(t, store.mutationOutsideTx)
		require.Equal(t, []string{
			"user.by_username",
			"transaction.begin",
			"user.by_username",
			"session.create",
			"transaction.rollback",
		}, store.events)
	})
}

func TestSessionCSRFDerivationAndRecovery(t *testing.T) {
	now := time.Date(2026, 8, 17, 13, 0, 0, 0, time.UTC)
	sessionSecret := sessionMatrixOpaqueValue(0x14)
	legacyCSRF := sessionMatrixOpaqueValue(0x34)
	derivedCSRF := sessionMatrixDerivedCSRF(sessionSecret)
	activeUser := domain.User{ID: 61, Username: "recovery-user", IsActive: true}
	inactiveUser := activeUser
	inactiveUser.IsActive = false
	lookupFailure := errors.New("session lookup unavailable")
	updateFailure := errors.New("csrf update unavailable")
	commitFailure := errors.New("csrf transaction commit unavailable")

	probe := application.NewService(newSessionMatrixStore(), &sessionMatrixClock{now: now})
	if _, ok := any(probe).(sessionCSRFRecoveryAPI); !ok {
		t.Error("identity service does not implement CSRFForSession(context.Context, string)")
		return
	}

	newRecoveryStore := func(csrf string, owner domain.User) *sessionMatrixStore {
		store := newSessionMatrixStore()
		record := sessionMatrixRecord(sessionSecret, csrf, activeUser.ID, now.Add(time.Hour))
		store.sessions[string(record.SecretHash)] = record
		store.users[activeUser.ID] = sessionMatrixUserResult{user: owner}
		return store
	}

	t.Run("legacy digest heals atomically to exact derivation", func(t *testing.T) {
		store := newRecoveryStore(legacyCSRF, activeUser)
		clock := &sessionMatrixClock{now: now}
		api := any(application.NewService(store, clock)).(sessionCSRFRecoveryAPI)

		csrf, err := api.CSRFForSession(t.Context(), sessionSecret)

		requireSessionMatrixNoError(t, err)
		if !sessionMatrixStringsEqual(csrf, derivedCSRF) {
			t.Error("csrf recovery did not return the exact session-bound derivation")
		}
		record, ok := sessionMatrixRow(store, sessionSecret)
		if !ok || !sessionMatrixBytesEqual(record.CSRFHash, sessionMatrixDigest(derivedCSRF)) {
			t.Error("csrf recovery did not durably bind the derived digest")
		}
		require.Equal(t, 1, store.transactionCalls)
		require.Equal(t, 1, store.sessionLookups)
		require.Equal(t, 1, store.userLookups)
		require.Equal(t, 1, store.updateCalls)
		require.Equal(t, 1, clock.calls)
		require.False(t, store.mutationOutsideTx)
		require.Equal(t, []string{
			"transaction.begin",
			"session.by_hash",
			"user.by_id",
			"session.update_csrf",
			"transaction.commit",
		}, store.events)
	})

	t.Run("already derived digest is stable and does not write", func(t *testing.T) {
		store := newRecoveryStore(derivedCSRF, activeUser)
		clock := &sessionMatrixClock{now: now}
		api := any(application.NewService(store, clock)).(sessionCSRFRecoveryAPI)

		csrf, err := api.CSRFForSession(t.Context(), sessionSecret)

		requireSessionMatrixNoError(t, err)
		if !sessionMatrixStringsEqual(csrf, derivedCSRF) {
			t.Error("stable csrf recovery changed the session-bound value")
		}
		require.Zero(t, store.updateCalls)
		require.Equal(t, 1, store.transactionCalls)
		require.Equal(t, 1, store.sessionLookups)
		require.Equal(t, 1, store.userLookups)
		require.Equal(t, 1, clock.calls)
		require.False(t, store.mutationOutsideTx)
		require.Equal(t, []string{
			"transaction.begin",
			"session.by_hash",
			"user.by_id",
			"transaction.commit",
		}, store.events)
	})

	t.Run("validation outcomes return no derived material", func(t *testing.T) {
		tests := []struct {
			name             string
			hasSession       bool
			expires          time.Time
			owner            sessionMatrixUserResult
			sessionLookupErr error
			wantKind         application.SessionCredentialFailureKind
			wantReason       shared.ErrorReason
			wantCause        error
			wantSessionCalls int
			wantUserCalls    int
			wantClockCalls   int
			wantEvents       []string
		}{
			{
				name:             "unknown session",
				wantKind:         application.SessionCredentialFailureUnknown,
				wantReason:       shared.ErrorReasonUnauthenticated,
				wantSessionCalls: 1,
				wantEvents:       []string{"transaction.begin", "session.by_hash", "transaction.rollback"},
			},
			{
				name:             "expired session",
				hasSession:       true,
				expires:          now,
				owner:            sessionMatrixUserResult{user: activeUser},
				wantKind:         application.SessionCredentialFailureExpired,
				wantReason:       shared.ErrorReasonUnauthenticated,
				wantSessionCalls: 1,
				wantClockCalls:   1,
				wantEvents:       []string{"transaction.begin", "session.by_hash", "transaction.rollback"},
			},
			{
				name:             "inactive owner",
				hasSession:       true,
				expires:          now.Add(time.Hour),
				owner:            sessionMatrixUserResult{user: inactiveUser},
				wantKind:         application.SessionCredentialFailureInactiveOwner,
				wantReason:       shared.ErrorReasonUnauthenticated,
				wantSessionCalls: 1,
				wantUserCalls:    1,
				wantClockCalls:   1,
				wantEvents: []string{
					"transaction.begin",
					"session.by_hash",
					"user.by_id",
					"transaction.rollback",
				},
			},
			{
				name:             "session infrastructure failure",
				sessionLookupErr: lookupFailure,
				wantReason:       shared.ErrorReasonInternal,
				wantCause:        lookupFailure,
				wantSessionCalls: 1,
				wantEvents:       []string{"transaction.begin", "session.by_hash", "transaction.rollback"},
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				store := newSessionMatrixStore()
				store.sessionLookupErr = test.sessionLookupErr
				if test.hasSession {
					record := sessionMatrixRecord(sessionSecret, legacyCSRF, activeUser.ID, test.expires)
					store.sessions[string(record.SecretHash)] = record
					store.users[activeUser.ID] = test.owner
				}
				clock := &sessionMatrixClock{now: now}
				api := any(application.NewService(store, clock)).(sessionCSRFRecoveryAPI)

				csrf, err := api.CSRFForSession(t.Context(), sessionSecret)

				assertNoSessionMatrixString(t, csrf)
				if test.wantKind != 0 {
					assertSessionMatrixFailure(t, err, test.wantKind)
				} else {
					assertSessionMatrixReason(t, err, test.wantReason, test.wantCause)
				}
				require.Equal(t, 1, store.transactionCalls)
				require.Equal(t, test.wantSessionCalls, store.sessionLookups)
				require.Equal(t, test.wantUserCalls, store.userLookups)
				require.Equal(t, test.wantClockCalls, clock.calls)
				require.Zero(t, store.updateCalls)
				require.False(t, store.mutationOutsideTx)
				require.Equal(t, test.wantEvents, store.events)
			})
		}
	})

	t.Run("update disappearance maps to unknown and returns no value", func(t *testing.T) {
		store := newRecoveryStore(legacyCSRF, activeUser)
		store.updateErr = application.ErrNotFound
		before, _ := sessionMatrixRow(store, sessionSecret)
		clock := &sessionMatrixClock{now: now}
		api := any(application.NewService(store, clock)).(sessionCSRFRecoveryAPI)

		csrf, err := api.CSRFForSession(t.Context(), sessionSecret)

		assertNoSessionMatrixString(t, csrf)
		assertSessionMatrixFailure(t, err, application.SessionCredentialFailureUnknown)
		after, ok := sessionMatrixRow(store, sessionSecret)
		if !ok || !sessionMatrixBytesEqual(before.CSRFHash, after.CSRFHash) {
			t.Error("failed csrf recovery changed the stored digest")
		}
		require.Equal(t, 1, store.transactionCalls)
		require.Equal(t, 1, store.updateCalls)
		require.Equal(t, 1, clock.calls)
		require.False(t, store.mutationOutsideTx)
		require.Equal(t, []string{
			"transaction.begin",
			"session.by_hash",
			"user.by_id",
			"session.update_csrf",
			"transaction.rollback",
		}, store.events)
	})

	t.Run("update infrastructure failure rolls back and returns no value", func(t *testing.T) {
		store := newRecoveryStore(legacyCSRF, activeUser)
		store.updateErr = updateFailure
		before, _ := sessionMatrixRow(store, sessionSecret)
		clock := &sessionMatrixClock{now: now}
		api := any(application.NewService(store, clock)).(sessionCSRFRecoveryAPI)

		csrf, err := api.CSRFForSession(t.Context(), sessionSecret)

		assertNoSessionMatrixString(t, csrf)
		assertSessionMatrixReason(t, err, shared.ErrorReasonInternal, updateFailure)
		after, ok := sessionMatrixRow(store, sessionSecret)
		if !ok || !sessionMatrixBytesEqual(before.CSRFHash, after.CSRFHash) {
			t.Error("failed csrf update changed the stored digest")
		}
		require.Equal(t, 1, store.transactionCalls)
		require.Equal(t, 1, store.updateCalls)
		require.Equal(t, 1, clock.calls)
		require.False(t, store.mutationOutsideTx)
	})

	t.Run("commit failure restores legacy digest and returns no value", func(t *testing.T) {
		store := newRecoveryStore(legacyCSRF, activeUser)
		store.commitErr = commitFailure
		before, _ := sessionMatrixRow(store, sessionSecret)
		clock := &sessionMatrixClock{now: now}
		api := any(application.NewService(store, clock)).(sessionCSRFRecoveryAPI)

		csrf, err := api.CSRFForSession(t.Context(), sessionSecret)

		assertNoSessionMatrixString(t, csrf)
		assertSessionMatrixReason(t, err, shared.ErrorReasonInternal, commitFailure)
		after, ok := sessionMatrixRow(store, sessionSecret)
		if !ok || !sessionMatrixBytesEqual(before.CSRFHash, after.CSRFHash) {
			t.Error("failed csrf transaction did not restore the legacy digest")
		}
		require.Equal(t, 1, store.transactionCalls)
		require.Equal(t, 1, store.updateCalls)
		require.Equal(t, 1, clock.calls)
		require.False(t, store.mutationOutsideTx)
		require.Equal(t, []string{
			"transaction.begin",
			"session.by_hash",
			"user.by_id",
			"session.update_csrf",
			"transaction.rollback",
		}, store.events)
	})
}

func TestLogoutRevalidatesAndRevokesInOneTransaction(t *testing.T) {
	now := time.Date(2026, 8, 17, 14, 0, 0, 0, time.UTC)
	sessionSecret := sessionMatrixOpaqueValue(0x15)
	csrf := sessionMatrixOpaqueValue(0x35)
	mismatchedCSRF := sessionMatrixOpaqueValue(0x55)
	activeUser := domain.User{ID: 71, Username: "logout-user", IsActive: true}
	inactiveUser := activeUser
	inactiveUser.IsActive = false
	lookupFailure := errors.New("session lookup unavailable")
	ownerFailure := errors.New("session owner unavailable")
	deleteFailure := errors.New("session delete unavailable")
	commitFailure := errors.New("logout transaction commit unavailable")

	probe := application.NewService(newSessionMatrixStore(), &sessionMatrixClock{now: now})
	if _, ok := any(probe).(transactionalSessionLogoutAPI); !ok {
		t.Error("identity service does not implement Logout(context.Context, string, string)")
		return
	}

	newLogoutStore := func(expires time.Time, owner sessionMatrixUserResult) *sessionMatrixStore {
		store := newSessionMatrixStore()
		record := sessionMatrixRecord(sessionSecret, csrf, activeUser.ID, expires)
		store.sessions[string(record.SecretHash)] = record
		store.users[activeUser.ID] = owner
		return store
	}

	t.Run("matching csrf revokes only after revalidation and commit", func(t *testing.T) {
		store := newLogoutStore(now.Add(time.Hour), sessionMatrixUserResult{user: activeUser})
		clock := &sessionMatrixClock{now: now}
		api := any(application.NewService(store, clock)).(transactionalSessionLogoutAPI)

		err := api.Logout(t.Context(), sessionSecret, csrf)

		requireSessionMatrixNoError(t, err)
		if sessionMatrixHasRow(store, sessionSecret) {
			t.Error("successful logout retained the revoked session")
		}
		require.Equal(t, 1, store.transactionCalls)
		require.Equal(t, 1, store.sessionLookups)
		require.Equal(t, 1, store.userLookups)
		require.Equal(t, 1, store.deleteCalls)
		require.Equal(t, 1, clock.calls)
		require.False(t, store.mutationOutsideTx)
		require.Equal(t, []string{
			"transaction.begin",
			"session.by_hash",
			"user.by_id",
			"session.delete",
			"transaction.commit",
		}, store.events)
	})

	t.Run("credential and csrf failures leave the row unchanged", func(t *testing.T) {
		tests := []struct {
			name             string
			sessionSecret    string
			csrf             string
			hasSession       bool
			expires          time.Time
			owner            sessionMatrixUserResult
			sessionLookupErr error
			wantKind         application.SessionCredentialFailureKind
			wantReason       shared.ErrorReason
			wantCause        error
			wantSessionCalls int
			wantUserCalls    int
			wantClockCalls   int
			wantEvents       []string
		}{
			{
				name:       "missing session",
				csrf:       csrf,
				wantKind:   application.SessionCredentialFailureMissing,
				wantReason: shared.ErrorReasonUnauthenticated,
				wantEvents: []string{"transaction.begin", "transaction.rollback"},
			},
			{
				name:             "unknown session",
				sessionSecret:    sessionSecret,
				csrf:             csrf,
				wantKind:         application.SessionCredentialFailureUnknown,
				wantReason:       shared.ErrorReasonUnauthenticated,
				wantSessionCalls: 1,
				wantEvents:       []string{"transaction.begin", "session.by_hash", "transaction.rollback"},
			},
			{
				name:             "expired session",
				sessionSecret:    sessionSecret,
				csrf:             csrf,
				hasSession:       true,
				expires:          now,
				owner:            sessionMatrixUserResult{user: activeUser},
				wantKind:         application.SessionCredentialFailureExpired,
				wantReason:       shared.ErrorReasonUnauthenticated,
				wantSessionCalls: 1,
				wantClockCalls:   1,
				wantEvents:       []string{"transaction.begin", "session.by_hash", "transaction.rollback"},
			},
			{
				name:             "inactive owner",
				sessionSecret:    sessionSecret,
				csrf:             csrf,
				hasSession:       true,
				expires:          now.Add(time.Hour),
				owner:            sessionMatrixUserResult{user: inactiveUser},
				wantKind:         application.SessionCredentialFailureInactiveOwner,
				wantReason:       shared.ErrorReasonUnauthenticated,
				wantSessionCalls: 1,
				wantUserCalls:    1,
				wantClockCalls:   1,
				wantEvents: []string{
					"transaction.begin",
					"session.by_hash",
					"user.by_id",
					"transaction.rollback",
				},
			},
			{
				name:             "empty csrf",
				sessionSecret:    sessionSecret,
				hasSession:       true,
				expires:          now.Add(time.Hour),
				owner:            sessionMatrixUserResult{user: activeUser},
				wantReason:       shared.ErrorReasonForbidden,
				wantSessionCalls: 1,
				wantUserCalls:    1,
				wantClockCalls:   1,
				wantEvents: []string{
					"transaction.begin",
					"session.by_hash",
					"user.by_id",
					"transaction.rollback",
				},
			},
			{
				name:             "mismatched csrf",
				sessionSecret:    sessionSecret,
				csrf:             mismatchedCSRF,
				hasSession:       true,
				expires:          now.Add(time.Hour),
				owner:            sessionMatrixUserResult{user: activeUser},
				wantReason:       shared.ErrorReasonForbidden,
				wantSessionCalls: 1,
				wantUserCalls:    1,
				wantClockCalls:   1,
				wantEvents: []string{
					"transaction.begin",
					"session.by_hash",
					"user.by_id",
					"transaction.rollback",
				},
			},
			{
				name:             "session infrastructure failure",
				sessionSecret:    sessionSecret,
				csrf:             csrf,
				sessionLookupErr: lookupFailure,
				wantReason:       shared.ErrorReasonInternal,
				wantCause:        lookupFailure,
				wantSessionCalls: 1,
				wantEvents:       []string{"transaction.begin", "session.by_hash", "transaction.rollback"},
			},
			{
				name:             "owner infrastructure failure",
				sessionSecret:    sessionSecret,
				csrf:             csrf,
				hasSession:       true,
				expires:          now.Add(time.Hour),
				owner:            sessionMatrixUserResult{err: ownerFailure},
				wantReason:       shared.ErrorReasonInternal,
				wantCause:        ownerFailure,
				wantSessionCalls: 1,
				wantUserCalls:    1,
				wantClockCalls:   1,
				wantEvents: []string{
					"transaction.begin",
					"session.by_hash",
					"user.by_id",
					"transaction.rollback",
				},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				store := newSessionMatrixStore()
				store.sessionLookupErr = test.sessionLookupErr
				if test.hasSession {
					record := sessionMatrixRecord(sessionSecret, csrf, activeUser.ID, test.expires)
					store.sessions[string(record.SecretHash)] = record
					store.users[activeUser.ID] = test.owner
				}
				clock := &sessionMatrixClock{now: now}
				api := any(application.NewService(store, clock)).(transactionalSessionLogoutAPI)

				err := api.Logout(t.Context(), test.sessionSecret, test.csrf)

				if test.wantKind != 0 {
					assertSessionMatrixFailure(t, err, test.wantKind)
				} else {
					assertSessionMatrixReason(t, err, test.wantReason, test.wantCause)
				}
				if test.hasSession && !sessionMatrixHasRow(store, sessionSecret) {
					t.Error("failed logout removed the stored session")
				}
				require.Equal(t, 1, store.transactionCalls)
				require.Equal(t, test.wantSessionCalls, store.sessionLookups)
				require.Equal(t, test.wantUserCalls, store.userLookups)
				require.Equal(t, test.wantClockCalls, clock.calls)
				require.Zero(t, store.deleteCalls)
				require.False(t, store.mutationOutsideTx)
				require.Equal(t, test.wantEvents, store.events)
			})
		}
	})

	t.Run("delete disappearance becomes unknown and keeps rollback boundary", func(t *testing.T) {
		store := newLogoutStore(now.Add(time.Hour), sessionMatrixUserResult{user: activeUser})
		store.deleteErr = application.ErrNotFound
		clock := &sessionMatrixClock{now: now}
		api := any(application.NewService(store, clock)).(transactionalSessionLogoutAPI)

		err := api.Logout(t.Context(), sessionSecret, csrf)

		assertSessionMatrixFailure(t, err, application.SessionCredentialFailureUnknown)
		if !sessionMatrixHasRow(store, sessionSecret) {
			t.Error("delete disappearance changed the stored session")
		}
		require.Equal(t, 1, store.transactionCalls)
		require.Equal(t, 1, store.deleteCalls)
		require.False(t, store.mutationOutsideTx)
		require.Equal(t, []string{
			"transaction.begin",
			"session.by_hash",
			"user.by_id",
			"session.delete",
			"transaction.rollback",
		}, store.events)
	})

	t.Run("delete infrastructure failure preserves the session", func(t *testing.T) {
		store := newLogoutStore(now.Add(time.Hour), sessionMatrixUserResult{user: activeUser})
		store.deleteErr = deleteFailure
		clock := &sessionMatrixClock{now: now}
		api := any(application.NewService(store, clock)).(transactionalSessionLogoutAPI)

		err := api.Logout(t.Context(), sessionSecret, csrf)

		assertSessionMatrixReason(t, err, shared.ErrorReasonInternal, deleteFailure)
		if !sessionMatrixHasRow(store, sessionSecret) {
			t.Error("failed delete changed the stored session")
		}
		require.Equal(t, 1, store.transactionCalls)
		require.Equal(t, 1, store.deleteCalls)
		require.False(t, store.mutationOutsideTx)
	})

	t.Run("commit failure restores the revoked session", func(t *testing.T) {
		store := newLogoutStore(now.Add(time.Hour), sessionMatrixUserResult{user: activeUser})
		store.commitErr = commitFailure
		clock := &sessionMatrixClock{now: now}
		api := any(application.NewService(store, clock)).(transactionalSessionLogoutAPI)

		err := api.Logout(t.Context(), sessionSecret, csrf)

		assertSessionMatrixReason(t, err, shared.ErrorReasonInternal, commitFailure)
		if !sessionMatrixHasRow(store, sessionSecret) {
			t.Error("failed logout commit did not restore the stored session")
		}
		require.Equal(t, 1, store.transactionCalls)
		require.Equal(t, 1, store.deleteCalls)
		require.False(t, store.mutationOutsideTx)
		require.Equal(t, []string{
			"transaction.begin",
			"session.by_hash",
			"user.by_id",
			"session.delete",
			"transaction.rollback",
		}, store.events)
	})

	t.Run("second revocation observes unknown session", func(t *testing.T) {
		store := newLogoutStore(now.Add(time.Hour), sessionMatrixUserResult{user: activeUser})
		clock := &sessionMatrixClock{now: now}
		api := any(application.NewService(store, clock)).(transactionalSessionLogoutAPI)

		firstErr := api.Logout(t.Context(), sessionSecret, csrf)
		secondErr := api.Logout(t.Context(), sessionSecret, csrf)

		requireSessionMatrixNoError(t, firstErr)
		assertSessionMatrixFailure(t, secondErr, application.SessionCredentialFailureUnknown)
		if sessionMatrixHasRow(store, sessionSecret) {
			t.Error("completed revocations retained the stored session")
		}
		require.Equal(t, 2, store.transactionCalls)
		require.Equal(t, 2, store.sessionLookups)
		require.Equal(t, 1, store.userLookups)
		require.Equal(t, 1, store.deleteCalls)
		require.Equal(t, 1, clock.calls)
		require.False(t, store.mutationOutsideTx)
		require.Equal(t, []string{
			"transaction.begin",
			"session.by_hash",
			"user.by_id",
			"session.delete",
			"transaction.commit",
			"transaction.begin",
			"session.by_hash",
			"transaction.rollback",
		}, store.events)
	})
}
