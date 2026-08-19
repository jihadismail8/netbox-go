package identity

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	domain "netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

const (
	passwordChangeMatrixCurrentPassword = "matrix-current-password-2026"
	passwordChangeMatrixNextPassword    = "matrix-next-password-2026"
	passwordChangeMatrixOtherPassword   = "matrix-other-password-2026"
)

type passwordChangeMatrixUserResult struct {
	user domain.User
	hash string
	err  error
}

type passwordChangeMatrixState struct {
	user         domain.User
	passwordHash string
	sessions     map[string]SessionRecord
	tokenMarker  int
}

type passwordChangeMatrixStore struct {
	Store

	user         domain.User
	passwordHash string
	sessions     map[string]SessionRecord
	tokenMarker  int

	userResults       []passwordChangeMatrixUserResult
	userResultCursor  int
	sessionLookupErr  error
	sessionLookupErrs map[int]error
	updateErr         error
	deleteAllErr      error
	createErr         error
	finalizeErr       error
	commitErr         error

	events            []string
	transactionCalls  int
	userLookups       int
	sessionLookups    int
	updateCalls       int
	deleteAllCalls    int
	createCalls       int
	tokenLookups      int
	tokenTouches      int
	transactionDepth  int
	mutationOutsideTx bool
	updatedAt         time.Time
}

func newPasswordChangeMatrixStore(user domain.User, hash string) *passwordChangeMatrixStore {
	return &passwordChangeMatrixStore{
		user:         user,
		passwordHash: hash,
		sessions:     make(map[string]SessionRecord),
		tokenMarker:  1,
	}
}

func (store *passwordChangeMatrixStore) Transaction(ctx context.Context, apply func(Store) error) error {
	store.transactionCalls++
	store.events = append(store.events, "transaction.begin")
	before := store.snapshot()
	store.transactionDepth++
	err := apply(store)
	store.transactionDepth--
	if err != nil {
		store.restore(before)
		store.events = append(store.events, "transaction.rollback")
		return err
	}
	if store.finalizeErr != nil {
		store.restore(before)
		store.events = append(store.events, "transaction.finalize", "transaction.rollback")
		return store.finalizeErr
	}
	if store.commitErr != nil {
		store.restore(before)
		store.events = append(store.events, "transaction.rollback")
		return store.commitErr
	}
	store.events = append(store.events, "transaction.commit")
	return nil
}

func (store *passwordChangeMatrixStore) UserByID(context.Context, int64) (domain.User, string, error) {
	store.userLookups++
	store.events = append(store.events, "user.by_id")
	if store.userResultCursor < len(store.userResults) {
		result := store.userResults[store.userResultCursor]
		store.userResultCursor++
		return result.user, result.hash, result.err
	}
	return clonePasswordChangeMatrixUser(store.user), store.passwordHash, nil
}

func (store *passwordChangeMatrixStore) SessionByHash(_ context.Context, sessionHash []byte) (SessionRecord, error) {
	store.sessionLookups++
	store.events = append(store.events, "session.by_hash")
	if lookupErr, ok := store.sessionLookupErrs[store.sessionLookups]; ok {
		return SessionRecord{}, lookupErr
	}
	if store.sessionLookupErr != nil {
		return SessionRecord{}, store.sessionLookupErr
	}
	record, ok := store.sessions[string(sessionHash)]
	if !ok {
		return SessionRecord{}, ErrNotFound
	}
	return clonePasswordChangeMatrixSession(record), nil
}

func (store *passwordChangeMatrixStore) UpdatePassword(_ context.Context, id int64, hash string, changedAt time.Time) error {
	store.updateCalls++
	store.events = append(store.events, "password.update")
	store.updatedAt = changedAt
	if store.transactionDepth == 0 {
		store.mutationOutsideTx = true
	}
	if id == store.user.ID {
		store.passwordHash = hash
		store.user.Updated = changedAt
	}
	if store.updateErr != nil {
		return store.updateErr
	}
	return nil
}

func (store *passwordChangeMatrixStore) DeleteSessionsForUser(_ context.Context, userID int64) error {
	store.deleteAllCalls++
	store.events = append(store.events, "sessions.delete_all")
	if store.transactionDepth == 0 {
		store.mutationOutsideTx = true
	}
	for key, record := range store.sessions {
		if record.UserID == userID {
			delete(store.sessions, key)
		}
	}
	if store.deleteAllErr != nil {
		return store.deleteAllErr
	}
	return nil
}

func (store *passwordChangeMatrixStore) CreateSession(_ context.Context, record SessionRecord) error {
	store.createCalls++
	store.events = append(store.events, "session.create")
	if store.transactionDepth == 0 {
		store.mutationOutsideTx = true
	}
	store.sessions[string(record.SecretHash)] = clonePasswordChangeMatrixSession(record)
	if store.createErr != nil {
		return store.createErr
	}
	return nil
}

func (store *passwordChangeMatrixStore) TokenByHash(context.Context, []byte) (TokenRecord, domain.User, error) {
	store.tokenLookups++
	return TokenRecord{}, domain.User{}, ErrNotFound
}

func (store *passwordChangeMatrixStore) TouchToken(context.Context, int64, time.Time) error {
	store.tokenTouches++
	return nil
}

func (store *passwordChangeMatrixStore) snapshot() passwordChangeMatrixState {
	return passwordChangeMatrixState{
		user:         clonePasswordChangeMatrixUser(store.user),
		passwordHash: store.passwordHash,
		sessions:     clonePasswordChangeMatrixSessions(store.sessions),
		tokenMarker:  store.tokenMarker,
	}
}

func (store *passwordChangeMatrixStore) restore(state passwordChangeMatrixState) {
	store.user = clonePasswordChangeMatrixUser(state.user)
	store.passwordHash = state.passwordHash
	store.sessions = clonePasswordChangeMatrixSessions(state.sessions)
	store.tokenMarker = state.tokenMarker
}

type passwordChangeMatrixClock struct {
	now    time.Time
	calls  int
	events *[]string
}

func (clock *passwordChangeMatrixClock) Now() time.Time {
	clock.calls++
	if clock.events != nil {
		*clock.events = append(*clock.events, "clock.now")
	}
	return clock.now
}

type passwordChangeMatrixEntropy struct {
	material []byte
	err      error
	short    int
	maxChunk int
	offset   int
	calls    int
	events   *[]string
}

func (reader *passwordChangeMatrixEntropy) Read(target []byte) (int, error) {
	reader.calls++
	if reader.events != nil {
		*reader.events = append(*reader.events, "entropy.read")
	}
	if reader.err != nil {
		return 0, reader.err
	}
	if reader.short > 0 {
		count := reader.short
		if count > len(target) {
			count = len(target)
		}
		for index := 0; index < count; index++ {
			target[index] = byte(0x80 + index)
		}
		return count, io.ErrUnexpectedEOF
	}
	material := reader.material[reader.offset:]
	if len(material) == 0 {
		return 0, io.EOF
	}
	if reader.maxChunk > 0 && len(material) > reader.maxChunk {
		material = material[:reader.maxChunk]
	}
	count := copy(target, material)
	reader.offset += count
	return count, nil
}

type passwordChangeMatrixUnknownCredential struct{}

func (*passwordChangeMatrixUnknownCredential) passwordChangeCredential() {}
func (*passwordChangeMatrixUnknownCredential) String() string {
	return "<unknown-password-change-credential:redacted>"
}
func (credential *passwordChangeMatrixUnknownCredential) GoString() string {
	return credential.String()
}
func (credential *passwordChangeMatrixUnknownCredential) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, credential.String())
}
func (*passwordChangeMatrixUnknownCredential) MarshalJSON() ([]byte, error) {
	return nil, errors.New("unknown password-change credential serialization is disabled")
}

type passwordChangeMatrixUnknownInput struct{}

func (*passwordChangeMatrixUnknownInput) passwordChangeInput() {}
func (*passwordChangeMatrixUnknownInput) String() string {
	return "<unknown-password-change-input:redacted>"
}
func (input *passwordChangeMatrixUnknownInput) GoString() string { return input.String() }
func (input *passwordChangeMatrixUnknownInput) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, input.String())
}
func (*passwordChangeMatrixUnknownInput) MarshalJSON() ([]byte, error) {
	return nil, errors.New("unknown password-change input serialization is disabled")
}

type passwordChangeMatrixContextKey struct{}

func passwordChangeMatrixUser() domain.User {
	return domain.User{
		ID:        901,
		Username:  "password-matrix-user",
		Email:     "password-matrix@example.test",
		IsActive:  true,
		Created:   time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC),
		Updated:   time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC),
		IsStaff:   true,
		FirstName: "Matrix",
	}
}

func passwordChangeMatrixHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal("password fixture hashing failed")
	}
	return hash
}

func passwordChangeMatrixOpaque(seed byte) string {
	material := make([]byte, 32)
	for index := range material {
		material[index] = seed + byte(index)
	}
	return base64.RawURLEncoding.EncodeToString(material)
}

func passwordChangeMatrixDigest(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}

func passwordChangeMatrixCSRF(secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte("netbox-go/browser-csrf/v1"))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func passwordChangeMatrixSession(secret, csrf string, userID int64, expires time.Time) SessionRecord {
	created := expires.Add(-time.Hour)
	return SessionRecord{
		SecretHash: passwordChangeMatrixDigest(secret),
		CSRFHash:   passwordChangeMatrixDigest(csrf),
		UserID:     userID,
		Created:    created,
		LastSeen:   created,
		Expires:    expires,
	}
}

func passwordChangeMatrixAddSession(store *passwordChangeMatrixStore, secret, csrf string, userID int64, expires time.Time) {
	record := passwordChangeMatrixSession(secret, csrf, userID, expires)
	store.sessions[string(record.SecretHash)] = record
}

func passwordChangeMatrixHasSession(store *passwordChangeMatrixStore, secret string) bool {
	_, ok := store.sessions[string(passwordChangeMatrixDigest(secret))]
	return ok
}

func passwordChangeMatrixSessionForSecret(store *passwordChangeMatrixStore, secret string) (SessionRecord, bool) {
	record, ok := store.sessions[string(passwordChangeMatrixDigest(secret))]
	return clonePasswordChangeMatrixSession(record), ok
}

func clonePasswordChangeMatrixUser(user domain.User) domain.User {
	user.Permissions = append([]string(nil), user.Permissions...)
	if user.ObjectVisibility != nil {
		visibility := make(map[string]map[int64]struct{}, len(user.ObjectVisibility))
		for permission, ids := range user.ObjectVisibility {
			copied := make(map[int64]struct{}, len(ids))
			for id := range ids {
				copied[id] = struct{}{}
			}
			visibility[permission] = copied
		}
		user.ObjectVisibility = visibility
	}
	return user
}

func clonePasswordChangeMatrixSession(record SessionRecord) SessionRecord {
	record.SecretHash = append([]byte(nil), record.SecretHash...)
	record.CSRFHash = append([]byte(nil), record.CSRFHash...)
	return record
}

func clonePasswordChangeMatrixSessions(source map[string]SessionRecord) map[string]SessionRecord {
	cloned := make(map[string]SessionRecord, len(source))
	for key, record := range source {
		cloned[key] = clonePasswordChangeMatrixSession(record)
	}
	return cloned
}

func passwordChangeMatrixTextEqual(left, right string) bool {
	return len(left) == len(right) && subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

func passwordChangeMatrixBytesEqual(left, right []byte) bool {
	return len(left) == len(right) && subtle.ConstantTimeCompare(left, right) == 1
}

func passwordChangeMatrixCall(
	ctx context.Context,
	service *Service,
	principal domain.Principal,
	input ChangePasswordInput,
) (result ChangePasswordResult, err error, panicked bool) {
	defer func() {
		if recover() != nil {
			result = nil
			err = nil
			panicked = true
		}
	}()
	result, err = service.ChangePassword(ctx, principal, input)
	return result, err, false
}

func assertPasswordChangeMatrixReason(t *testing.T, err error, reason shared.ErrorReason, cause error) {
	t.Helper()
	if err == nil {
		t.Error("expected a classified application error")
		return
	}
	if shared.ReasonOf(err) != reason {
		t.Error("application error used the wrong stable reason")
	}
	if cause != nil && !errors.Is(err, cause) {
		t.Error("application error did not retain its internal cause")
	}
	if reason == shared.ErrorReasonInternal && err.Error() != "An internal error occurred." {
		t.Error("internal application error exposed a non-generic message")
	}
}

func assertPasswordChangeMatrixValidation(t *testing.T, err error, field, description string) {
	t.Helper()
	assertPasswordChangeMatrixReason(t, err, shared.ErrorReasonValidation, nil)
	violations := shared.ViolationsOf(err)
	if len(violations) != 1 {
		t.Error("validation outcome did not contain exactly one field violation")
		return
	}
	if violations[0].Field != field || violations[0].Reason != "invalid" || violations[0].Description != description {
		t.Error("validation outcome did not retain the exact field contract")
	}
}

func assertPasswordChangeMatrixNilResult(t *testing.T, result ChangePasswordResult) {
	t.Helper()
	if result != nil {
		t.Error("failed password change returned a result")
	}
}

func assertPasswordChangeMatrixNoBrowserResult(t *testing.T, result ChangePasswordResult) {
	t.Helper()
	if result == nil {
		t.Error("successful password change returned a nil result")
		return
	}
	session, ok := result.BrowserSession()
	if ok || session.User.ID != 0 || session.Secret != "" || session.CSRFToken != "" || !session.Expires.IsZero() {
		t.Error("token-origin password change returned browser-session material")
	}
}

func assertPasswordChangeMatrixStateEqual(t *testing.T, want passwordChangeMatrixState, store *passwordChangeMatrixStore) {
	t.Helper()
	if !passwordChangeMatrixTextEqual(want.passwordHash, store.passwordHash) {
		t.Error("failed password change did not restore the prior password hash")
	}
	if want.user.Updated != store.user.Updated {
		t.Error("failed password change did not restore the prior user timestamp")
	}
	if want.tokenMarker != store.tokenMarker {
		t.Error("password change modified independent API-token state")
	}
	if len(want.sessions) != len(store.sessions) {
		t.Error("failed password change did not restore the complete session set")
		return
	}
	for key, expected := range want.sessions {
		actual, ok := store.sessions[key]
		if !ok || actual.UserID != expected.UserID || actual.Created != expected.Created ||
			actual.LastSeen != expected.LastSeen || actual.Expires != expected.Expires ||
			!passwordChangeMatrixBytesEqual(actual.SecretHash, expected.SecretHash) ||
			!passwordChangeMatrixBytesEqual(actual.CSRFHash, expected.CSRFHash) {
			t.Error("failed password change altered prior session state")
			return
		}
	}
}

func assertPasswordChangeMatrixNoMutation(t *testing.T, store *passwordChangeMatrixStore) {
	t.Helper()
	if store.updateCalls != 0 || store.deleteAllCalls != 0 || store.createCalls != 0 || store.mutationOutsideTx {
		t.Error("rejected password change attempted a mutation")
	}
}

func passwordChangeMatrixEntropyMaterial(seed byte) []byte {
	material := make([]byte, 32)
	for index := range material {
		material[index] = seed + byte(index)
	}
	return material
}

func passwordChangeMatrixFunctionSource(source, signature string) (string, bool) {
	start := strings.Index(source, signature)
	if start < 0 {
		return "", false
	}
	remainder := source[start+len(signature):]
	next := strings.Index(remainder, "\nfunc ")
	if next < 0 {
		return source[start:], true
	}
	return source[start : start+len(signature)+next], true
}

func TestPasswordChangeCredentialProvenanceMatrix(t *testing.T) {
	now := time.Date(2026, 8, 19, 10, 0, 0, 0, time.FixedZone("matrix-offset", 3*60*60))
	user := passwordChangeMatrixUser()
	hash := passwordChangeMatrixHash(t, passwordChangeMatrixCurrentPassword)
	originSecret := passwordChangeMatrixOpaque(0x11)
	originCSRF := passwordChangeMatrixCSRF(originSecret)

	t.Run("invalid principals and provenance fail before lookup", func(t *testing.T) {
		var typedNilCredential *passwordChangeCredential
		cases := []struct {
			name       string
			principal  domain.Principal
			credential PasswordChangeCredential
		}{
			{
				name:       "unauthenticated principal",
				credential: APITokenPasswordChangeCredential(),
			},
			{
				name:      "nil provenance",
				principal: user.Principal(),
			},
			{
				name:       "typed nil provenance",
				principal:  user.Principal(),
				credential: typedNilCredential,
			},
			{
				name:       "unknown sealed provenance",
				principal:  user.Principal(),
				credential: &passwordChangeMatrixUnknownCredential{},
			},
			{
				name:       "invalid private kind",
				principal:  user.Principal(),
				credential: &passwordChangeCredential{kind: passwordChangeCredentialInvalid},
			},
			{
				name:      "inconsistent token payload",
				principal: user.Principal(),
				credential: &passwordChangeCredential{
					kind:          passwordChangeCredentialAPIToken,
					sessionSecret: originSecret,
					csrf:          originCSRF,
				},
			},
		}

		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				store := newPasswordChangeMatrixStore(user, hash)
				clock := &passwordChangeMatrixClock{now: now}
				service := NewService(store, clock)
				input := NewPasswordChangeInput(
					passwordChangeMatrixCurrentPassword,
					passwordChangeMatrixNextPassword,
					test.credential,
				)

				result, err, panicked := passwordChangeMatrixCall(t.Context(), service, test.principal, input)

				if panicked {
					t.Error("invalid password-change provenance caused a panic")
				}
				assertPasswordChangeMatrixNilResult(t, result)
				assertPasswordChangeMatrixReason(t, err, shared.ErrorReasonUnauthenticated, nil)
				if store.userLookups != 0 || store.transactionCalls != 0 || clock.calls != 0 {
					t.Error("invalid password-change provenance crossed the no-I/O boundary")
				}
				assertPasswordChangeMatrixNoMutation(t, store)
			})
		}
	})

	t.Run("browser session provenance is revalidated inside the transaction", func(t *testing.T) {
		mismatchedCSRF := passwordChangeMatrixOpaque(0x41)
		ownerMismatch := user.ID + 1
		lookupFailure := errors.New("session lookup unavailable")
		cases := []struct {
			name             string
			sessionSecret    string
			csrf             string
			hasSession       bool
			sessionUserID    int64
			expires          time.Time
			sessionLookupErr error
			wantReason       shared.ErrorReason
			wantKind         SessionCredentialFailureKind
			wantCause        error
			wantSession      bool
		}{
			{
				name:          "empty session secret",
				csrf:          originCSRF,
				wantReason:    shared.ErrorReasonUnauthenticated,
				wantKind:      SessionCredentialFailureMissing,
				expires:       now.Add(time.Hour),
				sessionUserID: user.ID,
			},
			{
				name:          "unknown session",
				sessionSecret: originSecret,
				csrf:          originCSRF,
				wantReason:    shared.ErrorReasonUnauthenticated,
				wantKind:      SessionCredentialFailureUnknown,
			},
			{
				name:          "expired session",
				sessionSecret: originSecret,
				csrf:          originCSRF,
				hasSession:    true,
				sessionUserID: user.ID,
				expires:       now.UTC(),
				wantReason:    shared.ErrorReasonUnauthenticated,
				wantKind:      SessionCredentialFailureExpired,
			},
			{
				name:          "owner mismatch",
				sessionSecret: originSecret,
				csrf:          originCSRF,
				hasSession:    true,
				sessionUserID: ownerMismatch,
				expires:       now.Add(time.Hour),
				wantReason:    shared.ErrorReasonUnauthenticated,
			},
			{
				name:          "empty csrf",
				sessionSecret: originSecret,
				hasSession:    true,
				sessionUserID: user.ID,
				expires:       now.Add(time.Hour),
				wantReason:    shared.ErrorReasonForbidden,
			},
			{
				name:          "mismatched csrf",
				sessionSecret: originSecret,
				csrf:          mismatchedCSRF,
				hasSession:    true,
				sessionUserID: user.ID,
				expires:       now.Add(time.Hour),
				wantReason:    shared.ErrorReasonForbidden,
			},
			{
				name:             "session infrastructure",
				sessionSecret:    originSecret,
				csrf:             originCSRF,
				sessionLookupErr: lookupFailure,
				wantReason:       shared.ErrorReasonInternal,
				wantCause:        lookupFailure,
			},
			{
				name:          "valid session and csrf",
				sessionSecret: originSecret,
				csrf:          originCSRF,
				hasSession:    true,
				sessionUserID: user.ID,
				expires:       now.Add(time.Hour),
				wantSession:   true,
			},
		}

		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				store := newPasswordChangeMatrixStore(user, hash)
				store.sessionLookupErr = test.sessionLookupErr
				if test.hasSession {
					passwordChangeMatrixAddSession(
						store,
						test.sessionSecret,
						originCSRF,
						test.sessionUserID,
						test.expires,
					)
				}
				clock := &passwordChangeMatrixClock{now: now, events: &store.events}
				entropy := &passwordChangeMatrixEntropy{
					material: passwordChangeMatrixEntropyMaterial(0x71),
					events:   &store.events,
				}
				service := NewService(store, clock, WithPasswordChangeEntropy(entropy))
				input := NewPasswordChangeInput(
					passwordChangeMatrixCurrentPassword,
					passwordChangeMatrixNextPassword,
					BrowserSessionPasswordChangeCredential(test.sessionSecret, test.csrf),
				)

				result, err, panicked := passwordChangeMatrixCall(t.Context(), service, user.Principal(), input)

				if panicked {
					t.Error("browser password-change provenance caused a panic")
				}
				if test.wantSession {
					if err != nil {
						t.Error("valid browser provenance was rejected")
					}
					if result == nil {
						t.Error("valid browser provenance returned no result")
					} else if _, ok := result.BrowserSession(); !ok {
						t.Error("valid browser provenance omitted its replacement result")
					}
					return
				}

				assertPasswordChangeMatrixNilResult(t, result)
				if test.wantKind != 0 {
					assertPasswordChangeMatrixReason(t, err, test.wantReason, nil)
					var failure *SessionCredentialFailure
					if !errors.As(err, &failure) || failure.Kind != test.wantKind {
						t.Error("browser provenance used the wrong typed session cause")
					}
				} else {
					assertPasswordChangeMatrixReason(t, err, test.wantReason, test.wantCause)
				}
				if store.transactionCalls != 1 {
					t.Error("preverified browser provenance did not use exactly one transaction")
				}
				assertPasswordChangeMatrixNoMutation(t, store)
			})
		}
	})

	t.Run("token provenance has no browser result and ignores ambient transport-like state", func(t *testing.T) {
		store := newPasswordChangeMatrixStore(user, hash)
		passwordChangeMatrixAddSession(store, originSecret, originCSRF, user.ID, now.Add(time.Hour))
		clock := &passwordChangeMatrixClock{now: now, events: &store.events}
		service := NewService(store, clock)
		ctx := context.WithValue(t.Context(), passwordChangeMatrixContextKey{}, struct {
			cookie, authorization string
		}{cookie: originSecret, authorization: passwordChangeMatrixOpaque(0x66)})
		input := NewPasswordChangeInput(
			passwordChangeMatrixCurrentPassword,
			passwordChangeMatrixNextPassword,
			APITokenPasswordChangeCredential(),
		)

		result, err, panicked := passwordChangeMatrixCall(ctx, service, user.Principal(), input)

		if panicked {
			t.Error("token-origin password change caused a panic")
		}
		if err != nil {
			t.Error("valid token provenance was rejected")
		}
		assertPasswordChangeMatrixNoBrowserResult(t, result)
		if len(store.sessions) != 0 {
			t.Error("token-origin password change retained browser sessions")
		}
		if store.tokenMarker != 1 || store.tokenLookups != 0 || store.tokenTouches != 0 {
			t.Error("password change modified or re-authenticated the independent API token")
		}
	})

	t.Run("formatting and serialization are fixed redactions", func(t *testing.T) {
		browserCredential := BrowserSessionPasswordChangeCredential(originSecret, originCSRF)
		tokenCredential := APITokenPasswordChangeCredential()
		input := NewPasswordChangeInput(
			passwordChangeMatrixCurrentPassword,
			passwordChangeMatrixNextPassword,
			browserCredential,
		)
		result := ChangePasswordResult(&changePasswordResult{browserSession: &domain.BrowserSession{
			User:      user,
			Secret:    passwordChangeMatrixOpaque(0x77),
			CSRFToken: passwordChangeMatrixOpaque(0x97),
			Expires:   now.Add(BrowserSessionLifetime),
		}})

		values := []struct {
			value interface {
				String() string
				GoString() string
				Format(fmt.State, rune)
				MarshalJSON() ([]byte, error)
			}
			want string
		}{
			{value: browserCredential, want: "<password-change-credential:redacted>"},
			{value: tokenCredential, want: "<password-change-credential:redacted>"},
			{value: input, want: "<change-password-input:redacted>"},
			{value: result, want: "<change-password-result:redacted>"},
		}
		formats := []string{
			"%v", "%+v", "%#v", "%s", "%q", "%x", "%X", "%d", "%o", "%b",
			"%c", "%U", "%e", "%f", "%g", "%t", "%+100.50v", "%#-100.50x",
		}
		contained := []string{
			originSecret,
			originCSRF,
			passwordChangeMatrixCurrentPassword,
			passwordChangeMatrixNextPassword,
		}

		for _, test := range values {
			if test.value.String() != test.want || test.value.GoString() != test.want {
				t.Error("password-change value did not use its fixed redacted string")
			}
			for _, format := range formats {
				formatted := fmt.Sprintf(format, test.value)
				if formatted != test.want {
					t.Error("password-change formatter honored a revealing verb or flag")
				}
				for _, candidate := range contained {
					if strings.Contains(formatted, candidate) {
						t.Error("password-change formatter exposed contained material")
					}
				}
			}

			payload, marshalErr := test.value.MarshalJSON()
			if marshalErr == nil || payload != nil {
				t.Error("password-change value allowed direct JSON serialization")
			}
			payload, marshalErr = json.Marshal(test.value)
			if marshalErr == nil || payload != nil {
				t.Error("password-change value allowed encoding/json serialization")
			}
		}

		if fmt.Sprintf("%T", browserCredential) != fmt.Sprintf("%T", tokenCredential) {
			t.Error("credential concrete type disclosed the winning provenance")
		}
		for _, value := range []any{browserCredential, tokenCredential, input, result} {
			pointerText := fmt.Sprintf("%p", value)
			typeText := fmt.Sprintf("%T", value)
			if !strings.HasPrefix(pointerText, "0x") || typeText == "" {
				t.Error("password-change value was not represented by a private pointer")
			}
			for _, candidate := range contained {
				if strings.Contains(pointerText, candidate) || strings.Contains(typeText, candidate) {
					t.Error("pointer or type formatting exposed contained material")
				}
			}
			concreteType := reflect.TypeOf(value)
			if concreteType.Kind() != reflect.Pointer || concreteType.Elem().Kind() != reflect.Struct {
				t.Error("password-change interface did not contain a private struct pointer")
				continue
			}
			for index := 0; index < concreteType.Elem().NumField(); index++ {
				field := concreteType.Elem().Field(index)
				if field.PkgPath == "" || field.Tag != "" {
					t.Error("password-change private representation exported or tagged a contained field")
				}
			}
		}
	})

	t.Run("scoped source audit keeps contained values out of diagnostic call sites", func(t *testing.T) {
		_, currentFile, _, ok := runtime.Caller(0)
		if !ok {
			t.Fatal("could not locate the application test source")
		}
		serviceSource, err := os.ReadFile(filepath.Join(filepath.Dir(currentFile), "service.go"))
		if err != nil {
			t.Fatal("could not read the password-change application source")
		}
		source := string(serviceSource)
		contractStart := strings.Index(source, "type passwordChangeCredentialKind")
		contractEnd := strings.Index(source, "type CreateTokenInput")
		useCaseStart := strings.Index(source, "func (s *Service) ChangePassword")
		useCaseEnd := strings.Index(source, "func (s *Service) BootstrapAdministrator")
		if contractStart < 0 || contractEnd <= contractStart || useCaseStart < 0 || useCaseEnd <= useCaseStart {
			t.Fatal("could not isolate the password-change application source")
		}
		scopedSource := source[contractStart:contractEnd] + source[useCaseStart:useCaseEnd]
		for _, forbidden := range []string{
			"return passwordChangeCredential{",
			"return passwordChangeInput{",
			"return changePasswordResult{",
			"fmt.Printf(",
			"fmt.Sprintf(",
			"fmt.Errorf(",
			"log.Printf(",
			"logger.",
			"slog.",
			"MarshalText(",
			"MarshalBinary(",
			"LogValue(",
			"json:\"",
			"protobuf:\"",
			"gorm:\"",
			"%p",
		} {
			if strings.Contains(scopedSource, forbidden) {
				t.Error("password-change application source contains an unsafe diagnostic or value representation")
			}
		}
		for _, declaration := range []string{
			"type passwordChangeCredential struct",
			"type passwordChangeInput struct",
			"type changePasswordResult struct",
		} {
			if !strings.Contains(scopedSource, declaration) {
				t.Error("password-change application source omitted a private concrete contract type")
			}
		}

		internalRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
		adapterFunctions := []struct {
			path       string
			signatures []string
		}{
			{
				path: filepath.Join(internalRoot, "adapters", "rest", "netbox", "identity", "http.go"),
				signatures: []string{
					"func (h *Handler) middleware(",
					"func (h *Handler) changePassword(",
				},
			},
			{
				path:       filepath.Join(internalRoot, "adapters", "grpc", "identity", "server.go"),
				signatures: []string{"func (s *Server) ChangePassword("},
			},
			{
				path:       filepath.Join(internalRoot, "adapters", "postgres", "identity", "store.go"),
				signatures: []string{"func (s *Store) UpdatePassword("},
			},
		}
		for _, target := range adapterFunctions {
			adapterSource, readErr := os.ReadFile(target.path)
			if readErr != nil {
				t.Fatal("could not read a password-change adapter source")
			}
			for _, signature := range target.signatures {
				functionSource, found := passwordChangeMatrixFunctionSource(string(adapterSource), signature)
				if !found {
					t.Error("password-change adapter source omitted an expected call site")
					continue
				}
				functionSource = strings.ReplaceAll(
					functionSource,
					`logger.Info("identity password changed", logger.Int64("userID", principal.ID))`,
					"",
				)
				for _, diagnostic := range []string{
					"logger.",
					"slog.",
					"log.Printf(",
					"fmt.Printf(",
					"fmt.Sprintf(",
					"fmt.Errorf(",
					"%p",
				} {
					if strings.Contains(functionSource, diagnostic) {
						t.Error("password-change adapter call site contains a credential-risk diagnostic")
					}
				}
			}
		}

		testSource, readErr := os.ReadFile(currentFile)
		if readErr != nil {
			t.Fatal("could not read the password-change matrix source")
		}
		for _, unsafeAssertion := range []string{
			"\tt.Error" + "f(",
			"\tt.Fatal" + "f(",
			"\tt.Lo" + "g(",
			"\tt.Log" + "f(",
			"requ" + "ire.",
			"ass" + "ert.",
		} {
			if strings.Contains(string(testSource), unsafeAssertion) {
				t.Error("password-change matrix contains a value-capable assertion diagnostic")
			}
		}
	})
}

func TestPasswordChangeVerificationClassification(t *testing.T) {
	now := time.Date(2026, 8, 19, 11, 0, 0, 0, time.FixedZone("verification-offset", -4*60*60))
	user := passwordChangeMatrixUser()
	activeHash := passwordChangeMatrixHash(t, passwordChangeMatrixCurrentPassword)
	changedHash := passwordChangeMatrixHash(t, passwordChangeMatrixOtherPassword)
	lookupFailure := errors.New("password lookup unavailable")
	transactionLookupFailure := errors.New("password revalidation unavailable")

	t.Run("nil typed nil and unknown input fail closed without panic", func(t *testing.T) {
		var typedNilInput *passwordChangeInput
		cases := []struct {
			name  string
			input ChangePasswordInput
		}{
			{name: "nil input"},
			{name: "typed nil private input", input: typedNilInput},
			{name: "unknown sealed input", input: &passwordChangeMatrixUnknownInput{}},
		}
		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				store := newPasswordChangeMatrixStore(user, activeHash)
				clock := &passwordChangeMatrixClock{now: now}
				service := NewService(store, clock)

				result, err, panicked := passwordChangeMatrixCall(t.Context(), service, user.Principal(), test.input)

				if panicked {
					t.Error("invalid password-change input caused a panic")
				}
				assertPasswordChangeMatrixNilResult(t, result)
				assertPasswordChangeMatrixReason(t, err, shared.ErrorReasonInternal, nil)
				if store.userLookups != 0 || store.transactionCalls != 0 || clock.calls != 0 {
					t.Error("invalid password-change input crossed the zero-I/O boundary")
				}
				assertPasswordChangeMatrixNoMutation(t, store)
			})
		}
	})

	t.Run("initial verification classifies identity and verifier state before transaction", func(t *testing.T) {
		inactive := user
		inactive.IsActive = false
		differentID := user
		differentID.ID++
		malformedHash := "malformed-password-verifier"
		cases := []struct {
			name            string
			lookup          passwordChangeMatrixUserResult
			currentPassword string
			newPassword     string
			wantReason      shared.ErrorReason
			wantField       string
			wantDescription string
			wantCause       error
			wantSuccess     bool
		}{
			{
				name:            "authenticated user disappeared",
				lookup:          passwordChangeMatrixUserResult{err: ErrNotFound},
				currentPassword: passwordChangeMatrixCurrentPassword,
				newPassword:     passwordChangeMatrixNextPassword,
				wantReason:      shared.ErrorReasonInternal,
				wantCause:       ErrNotFound,
			},
			{
				name:            "lookup infrastructure failure",
				lookup:          passwordChangeMatrixUserResult{err: lookupFailure},
				currentPassword: passwordChangeMatrixCurrentPassword,
				newPassword:     passwordChangeMatrixNextPassword,
				wantReason:      shared.ErrorReasonInternal,
				wantCause:       lookupFailure,
			},
			{
				name:            "returned identity differs after verifier work",
				lookup:          passwordChangeMatrixUserResult{user: differentID, hash: activeHash},
				currentPassword: passwordChangeMatrixCurrentPassword,
				newPassword:     passwordChangeMatrixNextPassword,
				wantReason:      shared.ErrorReasonInternal,
			},
			{
				name:            "malformed verifier precedes returned identity check",
				lookup:          passwordChangeMatrixUserResult{user: differentID, hash: malformedHash},
				currentPassword: passwordChangeMatrixCurrentPassword,
				newPassword:     passwordChangeMatrixNextPassword,
				wantReason:      shared.ErrorReasonInternal,
				wantCause:       bcrypt.ErrHashTooShort,
			},
			{
				name:            "active current password mismatch",
				lookup:          passwordChangeMatrixUserResult{user: user, hash: activeHash},
				currentPassword: passwordChangeMatrixOtherPassword,
				newPassword:     passwordChangeMatrixNextPassword,
				wantReason:      shared.ErrorReasonValidation,
				wantField:       "current_password",
				wantDescription: "Current password is incorrect.",
			},
			{
				name:            "inactive owner with matching password",
				lookup:          passwordChangeMatrixUserResult{user: inactive, hash: activeHash},
				currentPassword: passwordChangeMatrixCurrentPassword,
				newPassword:     passwordChangeMatrixNextPassword,
				wantReason:      shared.ErrorReasonUnauthenticated,
			},
			{
				name:            "inactive owner with mismatched password",
				lookup:          passwordChangeMatrixUserResult{user: inactive, hash: activeHash},
				currentPassword: passwordChangeMatrixOtherPassword,
				newPassword:     passwordChangeMatrixNextPassword,
				wantReason:      shared.ErrorReasonUnauthenticated,
			},
			{
				name:            "active malformed verifier",
				lookup:          passwordChangeMatrixUserResult{user: user, hash: malformedHash},
				currentPassword: passwordChangeMatrixCurrentPassword,
				newPassword:     passwordChangeMatrixNextPassword,
				wantReason:      shared.ErrorReasonInternal,
				wantCause:       bcrypt.ErrHashTooShort,
			},
			{
				name:            "inactive malformed verifier still internal",
				lookup:          passwordChangeMatrixUserResult{user: inactive, hash: malformedHash},
				currentPassword: passwordChangeMatrixCurrentPassword,
				newPassword:     passwordChangeMatrixNextPassword,
				wantReason:      shared.ErrorReasonInternal,
				wantCause:       bcrypt.ErrHashTooShort,
			},
			{
				name:            "unchanged short new-password validation",
				lookup:          passwordChangeMatrixUserResult{user: user, hash: activeHash},
				currentPassword: passwordChangeMatrixCurrentPassword,
				newPassword:     "short",
				wantReason:      shared.ErrorReasonValidation,
				wantField:       "new_password",
				wantDescription: "Password must contain at least 12 characters.",
			},
			{
				name:            "unchanged maximum new-password validation",
				lookup:          passwordChangeMatrixUserResult{user: user, hash: activeHash},
				currentPassword: passwordChangeMatrixCurrentPassword,
				newPassword:     strings.Repeat("x", 1025),
				wantReason:      shared.ErrorReasonValidation,
				wantField:       "new_password",
				wantDescription: "Password is too long.",
			},
			{
				name:            "successful verifier enters one transaction",
				lookup:          passwordChangeMatrixUserResult{user: user, hash: activeHash},
				currentPassword: passwordChangeMatrixCurrentPassword,
				newPassword:     passwordChangeMatrixNextPassword,
				wantSuccess:     true,
			},
		}

		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				store := newPasswordChangeMatrixStore(user, activeHash)
				store.userResults = []passwordChangeMatrixUserResult{test.lookup}
				clock := &passwordChangeMatrixClock{now: now, events: &store.events}
				service := NewService(store, clock)
				input := NewPasswordChangeInput(
					test.currentPassword,
					test.newPassword,
					APITokenPasswordChangeCredential(),
				)

				result, err, panicked := passwordChangeMatrixCall(t.Context(), service, user.Principal(), input)

				if panicked {
					t.Error("initial password verification caused a panic")
				}
				if test.wantSuccess {
					if err != nil {
						t.Error("valid initial verifier was rejected")
					}
					assertPasswordChangeMatrixNoBrowserResult(t, result)
					if store.transactionCalls != 1 || store.userLookups != 2 {
						t.Error("successful verification did not enter exactly one revalidating transaction")
					}
					return
				}

				assertPasswordChangeMatrixNilResult(t, result)
				if test.wantField != "" {
					assertPasswordChangeMatrixValidation(t, err, test.wantField, test.wantDescription)
				} else {
					assertPasswordChangeMatrixReason(t, err, test.wantReason, test.wantCause)
				}
				if store.userLookups != 1 || store.transactionCalls != 0 || clock.calls != 0 {
					t.Error("initial password-verification failure crossed the transaction boundary")
				}
				assertPasswordChangeMatrixNoMutation(t, store)
			})
		}
	})

	t.Run("transaction revalidation classifies stale identity before session or mutation", func(t *testing.T) {
		inactive := user
		inactive.IsActive = false
		differentID := user
		differentID.ID++
		cases := []struct {
			name            string
			revalidated     passwordChangeMatrixUserResult
			wantReason      shared.ErrorReason
			wantField       string
			wantDescription string
			wantCause       error
		}{
			{
				name:        "transactional user disappeared",
				revalidated: passwordChangeMatrixUserResult{err: ErrNotFound},
				wantReason:  shared.ErrorReasonInternal,
				wantCause:   ErrNotFound,
			},
			{
				name:        "transactional lookup infrastructure",
				revalidated: passwordChangeMatrixUserResult{err: transactionLookupFailure},
				wantReason:  shared.ErrorReasonInternal,
				wantCause:   transactionLookupFailure,
			},
			{
				name:        "transactional identity differs",
				revalidated: passwordChangeMatrixUserResult{user: differentID, hash: activeHash},
				wantReason:  shared.ErrorReasonInternal,
			},
			{
				name:        "transactional inactive owner with same hash",
				revalidated: passwordChangeMatrixUserResult{user: inactive, hash: activeHash},
				wantReason:  shared.ErrorReasonUnauthenticated,
			},
			{
				name:        "transactional inactive owner with changed hash",
				revalidated: passwordChangeMatrixUserResult{user: inactive, hash: changedHash},
				wantReason:  shared.ErrorReasonUnauthenticated,
			},
			{
				name:            "transactional active owner with changed hash",
				revalidated:     passwordChangeMatrixUserResult{user: user, hash: changedHash},
				wantReason:      shared.ErrorReasonValidation,
				wantField:       "current_password",
				wantDescription: "Current password is incorrect.",
			},
		}

		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				store := newPasswordChangeMatrixStore(user, activeHash)
				store.userResults = []passwordChangeMatrixUserResult{
					{user: user, hash: activeHash},
					test.revalidated,
				}
				clock := &passwordChangeMatrixClock{now: now, events: &store.events}
				service := NewService(store, clock)
				input := NewPasswordChangeInput(
					passwordChangeMatrixCurrentPassword,
					passwordChangeMatrixNextPassword,
					BrowserSessionPasswordChangeCredential(passwordChangeMatrixOpaque(0x22), passwordChangeMatrixOpaque(0x42)),
				)

				result, err, panicked := passwordChangeMatrixCall(t.Context(), service, user.Principal(), input)

				if panicked {
					t.Error("transactional password revalidation caused a panic")
				}
				assertPasswordChangeMatrixNilResult(t, result)
				if test.wantField != "" {
					assertPasswordChangeMatrixValidation(t, err, test.wantField, test.wantDescription)
				} else {
					assertPasswordChangeMatrixReason(t, err, test.wantReason, test.wantCause)
				}
				if store.userLookups != 2 || store.transactionCalls != 1 || clock.calls != 1 {
					t.Error("stale password state did not use the exact revalidation boundary")
				}
				if store.sessionLookups != 0 {
					t.Error("session state was inspected before final password-state acceptance")
				}
				assertPasswordChangeMatrixNoMutation(t, store)
				wantEvents := []string{
					"user.by_id",
					"transaction.begin",
					"clock.now",
					"user.by_id",
					"transaction.rollback",
				}
				if !reflect.DeepEqual(store.events, wantEvents) {
					t.Error("transactional password revalidation used the wrong event order")
				}
			})
		}
	})
}

func TestPasswordChangeSessionRotationAndInvalidation(t *testing.T) {
	now := time.Date(2026, 8, 19, 12, 0, 0, 0, time.FixedZone("rotation-offset", 5*60*60+30*60))
	expectedNow := now.UTC()
	user := passwordChangeMatrixUser()
	hash := passwordChangeMatrixHash(t, passwordChangeMatrixCurrentPassword)
	originSecret := passwordChangeMatrixOpaque(0x13)
	originCSRF := passwordChangeMatrixCSRF(originSecret)
	siblingOneSecret := passwordChangeMatrixOpaque(0x33)
	siblingTwoSecret := passwordChangeMatrixOpaque(0x53)
	otherUserSecret := passwordChangeMatrixOpaque(0x73)

	t.Run("browser origin atomically revokes all old sessions and returns one replacement", func(t *testing.T) {
		store := newPasswordChangeMatrixStore(user, hash)
		passwordChangeMatrixAddSession(store, originSecret, originCSRF, user.ID, now.Add(time.Hour))
		passwordChangeMatrixAddSession(
			store,
			siblingOneSecret,
			passwordChangeMatrixCSRF(siblingOneSecret),
			user.ID,
			now.Add(2*time.Hour),
		)
		passwordChangeMatrixAddSession(
			store,
			siblingTwoSecret,
			passwordChangeMatrixCSRF(siblingTwoSecret),
			user.ID,
			now.Add(3*time.Hour),
		)
		passwordChangeMatrixAddSession(
			store,
			otherUserSecret,
			passwordChangeMatrixCSRF(otherUserSecret),
			user.ID+1,
			now.Add(4*time.Hour),
		)
		clock := &passwordChangeMatrixClock{now: now, events: &store.events}
		entropyMaterial := passwordChangeMatrixEntropyMaterial(0xa1)
		entropy := &passwordChangeMatrixEntropy{material: entropyMaterial, events: &store.events}
		service := NewService(store, clock, WithPasswordChangeEntropy(entropy))
		input := NewPasswordChangeInput(
			passwordChangeMatrixCurrentPassword,
			passwordChangeMatrixNextPassword,
			BrowserSessionPasswordChangeCredential(originSecret, originCSRF),
		)

		result, err, panicked := passwordChangeMatrixCall(t.Context(), service, user.Principal(), input)

		if panicked {
			t.Error("browser-origin password change caused a panic")
		}
		if err != nil {
			t.Error("valid browser-origin password change was rejected")
		}
		if result == nil {
			t.Fatal("browser-origin password change returned no typed result")
		}
		replacement, ok := result.BrowserSession()
		if !ok {
			t.Fatal("browser-origin password change omitted replacement material")
		}
		if passwordChangeMatrixTextEqual(replacement.Secret, originSecret) ||
			passwordChangeMatrixTextEqual(replacement.CSRFToken, originCSRF) {
			t.Error("browser-origin password change reused prior credential material")
		}
		decoded, decodeErr := base64.RawURLEncoding.DecodeString(replacement.Secret)
		if decodeErr != nil || len(decoded) != 32 || !passwordChangeMatrixBytesEqual(decoded, entropyMaterial) {
			t.Error("replacement session was not generated from exactly 256 bits of injected entropy")
		}
		expectedCSRF := passwordChangeMatrixCSRF(replacement.Secret)
		if !passwordChangeMatrixTextEqual(replacement.CSRFToken, expectedCSRF) {
			t.Error("replacement csrf was not derived from the replacement session")
		}
		if replacement.User.ID != user.ID || replacement.User.Username != user.Username {
			t.Error("replacement session did not retain the transactionally revalidated user")
		}

		for _, oldSecret := range []string{originSecret, siblingOneSecret, siblingTwoSecret} {
			if passwordChangeMatrixHasSession(store, oldSecret) {
				t.Error("browser-origin password change retained an old user session")
			}
		}
		if !passwordChangeMatrixHasSession(store, otherUserSecret) {
			t.Error("browser-origin password change deleted another user's session")
		}
		record, exists := passwordChangeMatrixSessionForSecret(store, replacement.Secret)
		if !exists || len(store.sessions) != 2 {
			t.Fatal("browser-origin password change did not leave exactly one user replacement")
		}
		if record.UserID != user.ID || record.Created != expectedNow || record.LastSeen != expectedNow ||
			record.Expires != expectedNow.Add(BrowserSessionLifetime) || replacement.Expires != record.Expires {
			t.Error("replacement session did not use the fixed one-clock lifetime")
		}
		if record.Created.Location() != time.UTC || record.LastSeen.Location() != time.UTC || record.Expires.Location() != time.UTC {
			t.Error("replacement session timestamps were not normalized to UTC")
		}
		if !passwordChangeMatrixBytesEqual(record.SecretHash, passwordChangeMatrixDigest(replacement.Secret)) ||
			!passwordChangeMatrixBytesEqual(record.CSRFHash, passwordChangeMatrixDigest(expectedCSRF)) {
			t.Error("replacement session persisted the wrong session or csrf digest")
		}
		if clock.calls != 1 || entropy.calls != 1 {
			t.Error("browser-origin password change did not sample clock and entropy exactly once")
		}
		if store.transactionCalls != 1 || store.updateCalls != 1 || store.deleteAllCalls != 1 || store.createCalls != 1 {
			t.Error("browser-origin password change did not use one complete mutation transaction")
		}
		if store.mutationOutsideTx {
			t.Error("browser-origin password change mutated state outside its transaction")
		}
		if store.tokenMarker != 1 || store.tokenLookups != 0 || store.tokenTouches != 0 {
			t.Error("browser-origin password change modified independent API-token state")
		}
		if bcrypt.CompareHashAndPassword([]byte(store.passwordHash), []byte(passwordChangeMatrixCurrentPassword)) == nil ||
			bcrypt.CompareHashAndPassword([]byte(store.passwordHash), []byte(passwordChangeMatrixNextPassword)) != nil {
			t.Error("browser-origin password change committed the wrong password state")
		}

		beforeFollowupClockCalls := clock.calls
		if _, authenticateErr := service.AuthenticateSession(t.Context(), replacement.Secret); authenticateErr != nil {
			t.Error("replacement session did not authenticate after commit")
		}
		if csrfErr := service.VerifyCSRF(t.Context(), replacement.Secret, replacement.CSRFToken); csrfErr != nil {
			t.Error("replacement session rejected its bound csrf after commit")
		}
		if _, authenticateErr := service.AuthenticateSession(t.Context(), originSecret); authenticateErr == nil {
			t.Error("originating session remained usable after rotation")
		}
		if clock.calls != beforeFollowupClockCalls+2 {
			t.Error("follow-up replacement checks used an unexpected clock boundary")
		}
	})

	t.Run("token origin revokes browser sessions without replacement or token mutation", func(t *testing.T) {
		store := newPasswordChangeMatrixStore(user, hash)
		passwordChangeMatrixAddSession(store, originSecret, originCSRF, user.ID, now.Add(time.Hour))
		passwordChangeMatrixAddSession(
			store,
			siblingOneSecret,
			passwordChangeMatrixCSRF(siblingOneSecret),
			user.ID,
			now.Add(2*time.Hour),
		)
		clock := &passwordChangeMatrixClock{now: now, events: &store.events}
		entropyFailure := errors.New("token origin must not request session entropy")
		entropy := &passwordChangeMatrixEntropy{err: entropyFailure, events: &store.events}
		service := NewService(store, clock, WithPasswordChangeEntropy(entropy))
		input := NewPasswordChangeInput(
			passwordChangeMatrixCurrentPassword,
			passwordChangeMatrixNextPassword,
			APITokenPasswordChangeCredential(),
		)

		result, err, panicked := passwordChangeMatrixCall(t.Context(), service, user.Principal(), input)

		if panicked {
			t.Error("token-origin password change caused a panic")
		}
		if err != nil {
			t.Error("valid token-origin password change was rejected")
		}
		assertPasswordChangeMatrixNoBrowserResult(t, result)
		if len(store.sessions) != 0 {
			t.Error("token-origin password change retained browser sessions")
		}
		if entropy.calls != 0 {
			t.Error("token-origin password change requested browser-session entropy")
		}
		if clock.calls != 1 || store.transactionCalls != 1 || store.updateCalls != 1 ||
			store.deleteAllCalls != 1 || store.createCalls != 0 {
			t.Error("token-origin password change used the wrong transaction effects")
		}
		if store.tokenMarker != 1 || store.tokenLookups != 0 || store.tokenTouches != 0 {
			t.Error("token-origin password change modified or re-authenticated API-token state")
		}
		if bcrypt.CompareHashAndPassword([]byte(store.passwordHash), []byte(passwordChangeMatrixCurrentPassword)) == nil ||
			bcrypt.CompareHashAndPassword([]byte(store.passwordHash), []byte(passwordChangeMatrixNextPassword)) != nil {
			t.Error("token-origin password change committed the wrong password state")
		}
	})
}

func TestPasswordChangeTransactionRollbackAndStateRevalidation(t *testing.T) {
	now := time.Date(2026, 8, 19, 13, 0, 0, 0, time.FixedZone("rollback-offset", -7*60*60))
	user := passwordChangeMatrixUser()
	hash := passwordChangeMatrixHash(t, passwordChangeMatrixCurrentPassword)
	changedHash := passwordChangeMatrixHash(t, passwordChangeMatrixOtherPassword)
	originSecret := passwordChangeMatrixOpaque(0x15)
	originCSRF := passwordChangeMatrixCSRF(originSecret)
	siblingSecret := passwordChangeMatrixOpaque(0x35)

	newBrowserStore := func() *passwordChangeMatrixStore {
		store := newPasswordChangeMatrixStore(user, hash)
		passwordChangeMatrixAddSession(store, originSecret, originCSRF, user.ID, now.Add(time.Hour))
		passwordChangeMatrixAddSession(
			store,
			siblingSecret,
			passwordChangeMatrixCSRF(siblingSecret),
			user.ID,
			now.Add(2*time.Hour),
		)
		return store
	}

	newBrowserInput := func() ChangePasswordInput {
		return NewPasswordChangeInput(
			passwordChangeMatrixCurrentPassword,
			passwordChangeMatrixNextPassword,
			BrowserSessionPasswordChangeCredential(originSecret, originCSRF),
		)
	}

	t.Run("password update port requires an application-owned timestamp", func(t *testing.T) {
		method, ok := reflect.TypeOf((*Store)(nil)).Elem().MethodByName("UpdatePassword")
		if !ok {
			t.Fatal("identity store omitted the password-update port")
		}
		wantTimestamp := reflect.TypeOf(time.Time{})
		if method.Type.NumIn() != 4 || method.Type.In(3) != wantTimestamp {
			t.Error("password-update port does not require the application timestamp")
		}
	})

	t.Run("exact success order contains all authorization and mutation in one transaction", func(t *testing.T) {
		store := newBrowserStore()
		clock := &passwordChangeMatrixClock{now: now, events: &store.events}
		entropy := &passwordChangeMatrixEntropy{
			material: passwordChangeMatrixEntropyMaterial(0xb1),
			events:   &store.events,
		}
		service := NewService(store, clock, WithPasswordChangeEntropy(entropy))

		result, err, panicked := passwordChangeMatrixCall(t.Context(), service, user.Principal(), newBrowserInput())

		if panicked {
			t.Error("ordered browser password change caused a panic")
		}
		if err != nil || result == nil {
			t.Fatal("ordered browser password change did not succeed")
		}
		if _, ok := result.BrowserSession(); !ok {
			t.Error("ordered browser password change omitted replacement material")
		}
		wantEvents := []string{
			"user.by_id",
			"transaction.begin",
			"clock.now",
			"user.by_id",
			"session.by_hash",
			"entropy.read",
			"session.by_hash",
			"password.update",
			"sessions.delete_all",
			"session.create",
			"transaction.commit",
		}
		if !reflect.DeepEqual(store.events, wantEvents) {
			t.Error("password change used the wrong revalidation or mutation event order")
		}
		if clock.calls != 1 || entropy.calls != 1 || store.transactionCalls != 1 {
			t.Error("password change did not own exactly one clock, entropy, and transaction boundary")
		}
		if store.updatedAt != now.UTC() {
			t.Error("password update did not receive the single UTC application timestamp")
		}
	})

	t.Run("default entropy produces one raw-url 256-bit replacement", func(t *testing.T) {
		store := newBrowserStore()
		clock := &passwordChangeMatrixClock{now: now, events: &store.events}
		service := NewService(store, clock)

		result, err, panicked := passwordChangeMatrixCall(t.Context(), service, user.Principal(), newBrowserInput())

		if panicked {
			t.Error("default password-change entropy caused a panic")
		}
		if err != nil || result == nil {
			t.Fatal("default password-change entropy did not produce a result")
		}
		session, ok := result.BrowserSession()
		if !ok {
			t.Fatal("default password-change entropy omitted browser material")
		}
		decoded, decodeErr := base64.RawURLEncoding.DecodeString(session.Secret)
		if decodeErr != nil || len(decoded) != 32 {
			t.Error("default password-change entropy did not produce 256 raw-url bits")
		}
	})

	t.Run("chunked entropy is filled to exactly 32 bytes", func(t *testing.T) {
		store := newBrowserStore()
		clock := &passwordChangeMatrixClock{now: now, events: &store.events}
		material := passwordChangeMatrixEntropyMaterial(0xd1)
		entropy := &passwordChangeMatrixEntropy{
			material: material,
			maxChunk: 5,
			events:   &store.events,
		}
		service := NewService(store, clock, WithPasswordChangeEntropy(entropy))

		result, err, panicked := passwordChangeMatrixCall(t.Context(), service, user.Principal(), newBrowserInput())

		if panicked {
			t.Error("chunked password-change entropy caused a panic")
		}
		if err != nil || result == nil {
			t.Fatal("chunked password-change entropy did not produce a result")
		}
		session, ok := result.BrowserSession()
		if !ok {
			t.Fatal("chunked password-change entropy omitted browser material")
		}
		decoded, decodeErr := base64.RawURLEncoding.DecodeString(session.Secret)
		if decodeErr != nil || len(decoded) != 32 || !passwordChangeMatrixBytesEqual(decoded, material) {
			t.Error("password change did not fill the complete 256-bit entropy buffer")
		}
		if entropy.calls <= 1 {
			t.Error("chunked entropy did not exercise repeated io.Reader reads")
		}
	})

	t.Run("entropy option rejects nil without mutating global state", func(t *testing.T) {
		panicked := false
		func() {
			defer func() {
				panicked = recover() != nil
			}()
			_ = WithPasswordChangeEntropy(nil)
		}()
		if !panicked {
			t.Error("nil password-change entropy was accepted")
		}

		failingStore := newBrowserStore()
		failingClock := &passwordChangeMatrixClock{now: now, events: &failingStore.events}
		failure := errors.New("injected entropy unavailable")
		failingService := NewService(
			failingStore,
			failingClock,
			WithPasswordChangeEntropy(&passwordChangeMatrixEntropy{err: failure, events: &failingStore.events}),
		)
		failedResult, failedErr, failedPanic := passwordChangeMatrixCall(
			t.Context(),
			failingService,
			user.Principal(),
			newBrowserInput(),
		)
		if failedPanic {
			t.Error("injected entropy failure caused a panic")
		}
		assertPasswordChangeMatrixNilResult(t, failedResult)
		assertPasswordChangeMatrixReason(t, failedErr, shared.ErrorReasonInternal, failure)

		defaultStore := newBrowserStore()
		defaultClock := &passwordChangeMatrixClock{now: now, events: &defaultStore.events}
		defaultService := NewService(defaultStore, defaultClock)
		defaultResult, defaultErr, defaultPanic := passwordChangeMatrixCall(
			t.Context(),
			defaultService,
			user.Principal(),
			newBrowserInput(),
		)
		if defaultPanic || defaultErr != nil || defaultResult == nil {
			t.Error("per-service entropy injection mutated the production default")
		}
	})

	t.Run("short and failed entropy happen after final credential revalidation and before writes", func(t *testing.T) {
		entropyFailure := errors.New("injected entropy read unavailable")
		cases := []struct {
			name      string
			reader    *passwordChangeMatrixEntropy
			wantCause error
		}{
			{
				name:      "short read",
				reader:    &passwordChangeMatrixEntropy{short: 7},
				wantCause: io.ErrUnexpectedEOF,
			},
			{
				name:      "reader error",
				reader:    &passwordChangeMatrixEntropy{err: entropyFailure},
				wantCause: entropyFailure,
			},
		}
		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				store := newBrowserStore()
				before := store.snapshot()
				clock := &passwordChangeMatrixClock{now: now, events: &store.events}
				test.reader.events = &store.events
				service := NewService(store, clock, WithPasswordChangeEntropy(test.reader))

				result, err, panicked := passwordChangeMatrixCall(t.Context(), service, user.Principal(), newBrowserInput())

				if panicked {
					t.Error("password-change entropy failure caused a panic")
				}
				assertPasswordChangeMatrixNilResult(t, result)
				assertPasswordChangeMatrixReason(t, err, shared.ErrorReasonInternal, test.wantCause)
				assertPasswordChangeMatrixStateEqual(t, before, store)
				if store.userLookups != 2 || store.sessionLookups != 1 || store.transactionCalls != 1 ||
					clock.calls != 1 || test.reader.calls != 1 {
					t.Error("entropy failure did not occur after complete final revalidation")
				}
				assertPasswordChangeMatrixNoMutation(t, store)
				wantEvents := []string{
					"user.by_id",
					"transaction.begin",
					"clock.now",
					"user.by_id",
					"session.by_hash",
					"entropy.read",
					"transaction.rollback",
				}
				if !reflect.DeepEqual(store.events, wantEvents) {
					t.Error("entropy failure crossed the wrong transaction event boundary")
				}
			})
		}
	})

	t.Run("same-user replacement collisions and probe failures precede every write", func(t *testing.T) {
		probeFailure := errors.New("replacement collision probe unavailable")
		cases := []struct {
			name       string
			material   []byte
			probeError error
		}{
			{
				name:     "originating session collision",
				material: passwordChangeMatrixEntropyMaterial(0x15),
			},
			{
				name:     "sibling session collision",
				material: passwordChangeMatrixEntropyMaterial(0x35),
			},
			{
				name:       "collision probe infrastructure",
				material:   passwordChangeMatrixEntropyMaterial(0xe1),
				probeError: probeFailure,
			},
		}

		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				store := newBrowserStore()
				if test.probeError != nil {
					store.sessionLookupErrs = map[int]error{2: test.probeError}
				}
				before := store.snapshot()
				clock := &passwordChangeMatrixClock{now: now, events: &store.events}
				entropy := &passwordChangeMatrixEntropy{
					material: test.material,
					events:   &store.events,
				}
				service := NewService(store, clock, WithPasswordChangeEntropy(entropy))

				result, err, panicked := passwordChangeMatrixCall(t.Context(), service, user.Principal(), newBrowserInput())

				if panicked {
					t.Error("replacement collision handling caused a panic")
				}
				assertPasswordChangeMatrixNilResult(t, result)
				assertPasswordChangeMatrixReason(t, err, shared.ErrorReasonInternal, test.probeError)
				assertPasswordChangeMatrixStateEqual(t, before, store)
				assertPasswordChangeMatrixNoMutation(t, store)
				if store.userLookups != 2 || store.sessionLookups != 2 || store.transactionCalls != 1 ||
					clock.calls != 1 || entropy.calls != 1 {
					t.Error("replacement collision was not checked at the final pre-write boundary")
				}
				wantEvents := []string{
					"user.by_id",
					"transaction.begin",
					"clock.now",
					"user.by_id",
					"session.by_hash",
					"entropy.read",
					"session.by_hash",
					"transaction.rollback",
				}
				if !reflect.DeepEqual(store.events, wantEvents) {
					t.Error("replacement collision crossed the wrong transaction event boundary")
				}
			})
		}
	})

	t.Run("dml and transaction finalization failures roll back password sessions and result", func(t *testing.T) {
		updateFailure := errors.New("password update unavailable")
		deleteFailure := errors.New("session revocation unavailable")
		insertFailure := errors.New("replacement insert unavailable")
		finalizeFailure := errors.New("transaction finalization unavailable")
		commitFailure := errors.New("transaction commit unavailable")
		cases := []struct {
			name             string
			configure        func(*passwordChangeMatrixStore)
			wantCause        error
			wantUpdateCalls  int
			wantDeleteCalls  int
			wantCreateCalls  int
			wantFinalizeStep bool
		}{
			{
				name: "update failure after write",
				configure: func(store *passwordChangeMatrixStore) {
					store.updateErr = updateFailure
				},
				wantCause:       updateFailure,
				wantUpdateCalls: 1,
			},
			{
				name: "delete failure after password and deletion",
				configure: func(store *passwordChangeMatrixStore) {
					store.deleteAllErr = deleteFailure
				},
				wantCause:       deleteFailure,
				wantUpdateCalls: 1,
				wantDeleteCalls: 1,
			},
			{
				name: "insert failure after password and deletion",
				configure: func(store *passwordChangeMatrixStore) {
					store.createErr = insertFailure
				},
				wantCause:       insertFailure,
				wantUpdateCalls: 1,
				wantDeleteCalls: 1,
				wantCreateCalls: 1,
			},
			{
				name: "post dml finalization failure",
				configure: func(store *passwordChangeMatrixStore) {
					store.finalizeErr = finalizeFailure
				},
				wantCause:        finalizeFailure,
				wantUpdateCalls:  1,
				wantDeleteCalls:  1,
				wantCreateCalls:  1,
				wantFinalizeStep: true,
			},
			{
				name: "commit failure after complete callback",
				configure: func(store *passwordChangeMatrixStore) {
					store.commitErr = commitFailure
				},
				wantCause:       commitFailure,
				wantUpdateCalls: 1,
				wantDeleteCalls: 1,
				wantCreateCalls: 1,
			},
		}

		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				store := newBrowserStore()
				before := store.snapshot()
				test.configure(store)
				clock := &passwordChangeMatrixClock{now: now, events: &store.events}
				entropy := &passwordChangeMatrixEntropy{
					material: passwordChangeMatrixEntropyMaterial(0xc1),
					events:   &store.events,
				}
				service := NewService(store, clock, WithPasswordChangeEntropy(entropy))

				result, err, panicked := passwordChangeMatrixCall(t.Context(), service, user.Principal(), newBrowserInput())

				if panicked {
					t.Error("password-change rollback path caused a panic")
				}
				assertPasswordChangeMatrixNilResult(t, result)
				assertPasswordChangeMatrixReason(t, err, shared.ErrorReasonInternal, test.wantCause)
				assertPasswordChangeMatrixStateEqual(t, before, store)
				if store.updateCalls != test.wantUpdateCalls || store.deleteAllCalls != test.wantDeleteCalls ||
					store.createCalls != test.wantCreateCalls || store.transactionCalls != 1 || clock.calls != 1 {
					t.Error("password-change failure did not traverse the intended real mutation boundary")
				}
				if store.mutationOutsideTx {
					t.Error("password-change failure mutated state outside its transaction")
				}
				if test.wantFinalizeStep && !strings.Contains(strings.Join(store.events, ","), "transaction.finalize") {
					t.Error("post-dml finalization failure did not occur after the callback")
				}
			})
		}
	})

	t.Run("token commit failure rolls back without replacement entropy", func(t *testing.T) {
		store := newBrowserStore()
		before := store.snapshot()
		commitFailure := errors.New("token-origin transaction commit unavailable")
		store.commitErr = commitFailure
		clock := &passwordChangeMatrixClock{now: now, events: &store.events}
		entropy := &passwordChangeMatrixEntropy{
			err:    errors.New("token origin must not request entropy"),
			events: &store.events,
		}
		service := NewService(store, clock, WithPasswordChangeEntropy(entropy))
		input := NewPasswordChangeInput(
			passwordChangeMatrixCurrentPassword,
			passwordChangeMatrixNextPassword,
			APITokenPasswordChangeCredential(),
		)

		result, err, panicked := passwordChangeMatrixCall(t.Context(), service, user.Principal(), input)

		if panicked {
			t.Error("token-origin commit failure caused a panic")
		}
		assertPasswordChangeMatrixNilResult(t, result)
		assertPasswordChangeMatrixReason(t, err, shared.ErrorReasonInternal, commitFailure)
		assertPasswordChangeMatrixStateEqual(t, before, store)
		if entropy.calls != 0 || store.createCalls != 0 || store.updateCalls != 1 || store.deleteAllCalls != 1 {
			t.Error("token-origin rollback used browser entropy or the wrong mutation path")
		}
	})

	t.Run("changed hash wins over missing session and blocks every write", func(t *testing.T) {
		store := newPasswordChangeMatrixStore(user, hash)
		store.userResults = []passwordChangeMatrixUserResult{
			{user: user, hash: hash},
			{user: user, hash: changedHash},
		}
		before := store.snapshot()
		clock := &passwordChangeMatrixClock{now: now, events: &store.events}
		service := NewService(store, clock)

		result, err, panicked := passwordChangeMatrixCall(t.Context(), service, user.Principal(), newBrowserInput())

		if panicked {
			t.Error("stale password revalidation caused a panic")
		}
		assertPasswordChangeMatrixNilResult(t, result)
		assertPasswordChangeMatrixValidation(
			t,
			err,
			"current_password",
			"Current password is incorrect.",
		)
		assertPasswordChangeMatrixStateEqual(t, before, store)
		if store.sessionLookups != 0 || store.updateCalls != 0 || store.deleteAllCalls != 0 || store.createCalls != 0 {
			t.Error("stale password state reached session validation or mutation")
		}
		wantEvents := []string{
			"user.by_id",
			"transaction.begin",
			"clock.now",
			"user.by_id",
			"transaction.rollback",
		}
		if !reflect.DeepEqual(store.events, wantEvents) {
			t.Error("stale password state was not rejected before session validation")
		}
	})
}
