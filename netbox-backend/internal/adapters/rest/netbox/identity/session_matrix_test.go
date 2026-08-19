package identity

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
)

const (
	i3SessionCookieName = "sessionid"
	i3CSRFCookieName    = "csrftoken"
	i3SessionA          = "opaque-session-fixture-a"
	i3SessionUnknown    = "opaque-session-fixture-unknown"
	i3CSRFA             = "opaque-csrf-fixture-a"
	i3CSRFOther         = "opaque-csrf-fixture-other"
	i3TokenB            = "opaque-token-fixture-b"
	i3CSRFDomainTag     = "netbox-go/browser-csrf/v1"
	i3LoginUsername     = "browser-login-fixture"
	i3LoginPassword     = "browser-password-fixture"
)

type i3RESTClock struct{ now time.Time }

func (clock i3RESTClock) Now() time.Time { return clock.now }

type i3DigestKey [sha256.Size]byte

type i3RESTStore struct {
	application.Store

	sessions map[i3DigestKey]application.SessionRecord
	users    map[int64]domain.User

	loginUser         domain.User
	loginPasswordHash string

	sessionLookupErr error
	userLookupErr    error
	tokenLookupErr   error
	updateErr        error
	deleteErr        error
	transactionErr   error
	commitErr        error

	tokenRecord application.TokenRecord
	tokenUser   domain.User

	beforeTransaction func(*i3RESTStore)
	beforeSessionLoad func(*i3RESTStore, int)

	transactions    int
	sessionLookups  int
	userLookups     int
	usernameLookups int
	tokenLookups    int
	tokenTouches    int
	csrfUpdates     int
	deletes         int
	sessionCreates  int
	events          []string
}

func (store *i3RESTStore) Transaction(_ context.Context, fn func(application.Store) error) error {
	store.transactions++
	store.events = append(store.events, "transaction")
	if store.transactionErr != nil {
		return store.transactionErr
	}
	if store.beforeTransaction != nil {
		before := store.beforeTransaction
		store.beforeTransaction = nil
		before(store)
	}
	snapshot := i3CloneSessions(store.sessions)
	if err := fn(store); err != nil {
		store.sessions = snapshot
		return err
	}
	if store.commitErr != nil {
		store.sessions = snapshot
		return store.commitErr
	}
	return nil
}

func (store *i3RESTStore) SessionByHash(_ context.Context, hash []byte) (application.SessionRecord, error) {
	store.sessionLookups++
	store.events = append(store.events, "session")
	if store.beforeSessionLoad != nil {
		store.beforeSessionLoad(store, store.sessionLookups)
	}
	if store.sessionLookupErr != nil {
		return application.SessionRecord{}, store.sessionLookupErr
	}
	record, ok := store.sessions[i3Key(hash)]
	if !ok {
		return application.SessionRecord{}, application.ErrNotFound
	}
	return i3CloneSession(record), nil
}

func (store *i3RESTStore) UserByID(_ context.Context, id int64) (domain.User, string, error) {
	store.userLookups++
	store.events = append(store.events, "user")
	if store.userLookupErr != nil {
		return domain.User{}, "", store.userLookupErr
	}
	user, ok := store.users[id]
	if !ok {
		return domain.User{}, "", application.ErrNotFound
	}
	return user, "", nil
}

func (store *i3RESTStore) UserByUsername(_ context.Context, username string) (domain.User, string, error) {
	store.usernameLookups++
	store.events = append(store.events, "username")
	if username != i3LoginUsername || store.loginUser.ID == 0 {
		return domain.User{}, "", application.ErrNotFound
	}
	return store.loginUser, store.loginPasswordHash, nil
}

func (store *i3RESTStore) CreateSession(_ context.Context, record application.SessionRecord) error {
	store.sessionCreates++
	store.events = append(store.events, "create-session")
	key := i3Key(record.SecretHash)
	if _, exists := store.sessions[key]; exists {
		return errors.New("session already exists")
	}
	store.sessions[key] = i3CloneSession(record)
	return nil
}

func (store *i3RESTStore) TokenByHash(_ context.Context, hash []byte) (application.TokenRecord, domain.User, error) {
	store.tokenLookups++
	store.events = append(store.events, "token")
	if store.tokenLookupErr != nil {
		return application.TokenRecord{}, domain.User{}, store.tokenLookupErr
	}
	if subtle.ConstantTimeCompare(hash, i3Digest(i3TokenB)) != 1 {
		return application.TokenRecord{}, domain.User{}, application.ErrNotFound
	}
	return store.tokenRecord, store.tokenUser, nil
}

func (store *i3RESTStore) TouchToken(context.Context, int64, time.Time) error {
	store.tokenTouches++
	store.events = append(store.events, "token-touch")
	return nil
}

func (store *i3RESTStore) UpdateSessionCSRF(_ context.Context, sessionHash, csrfHash []byte) error {
	store.csrfUpdates++
	store.events = append(store.events, "csrf-update")
	if store.updateErr != nil {
		return store.updateErr
	}
	key := i3Key(sessionHash)
	record, ok := store.sessions[key]
	if !ok {
		return application.ErrNotFound
	}
	record.CSRFHash = append([]byte(nil), csrfHash...)
	store.sessions[key] = record
	return nil
}

func (store *i3RESTStore) DeleteSession(_ context.Context, hash []byte) error {
	store.deletes++
	store.events = append(store.events, "delete-session")
	if store.deleteErr != nil {
		return store.deleteErr
	}
	key := i3Key(hash)
	if _, ok := store.sessions[key]; !ok {
		return application.ErrNotFound
	}
	delete(store.sessions, key)
	return nil
}

type i3RESTBoundary struct {
	name         string
	wantRejected int
	middleware   func(*Handler) gin.HandlerFunc
}

type i3ProtectedResult struct {
	response  *httptest.ResponseRecorder
	called    bool
	principal domain.Principal
}

func TestRESTSessionFirstCredentialArbitration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 8, 17, 18, 0, 0, 0, time.UTC)

	for _, boundary := range i3RESTBoundaries() {
		t.Run(boundary.name, func(t *testing.T) {
			t.Run("valid session outranks a different valid token", func(t *testing.T) {
				store := i3ValidRESTStore(now, true)
				result := i3ServeProtected(store, now, boundary, http.MethodGet,
					[]*http.Cookie{i3Cookie(i3SessionCookieName, i3SessionA)},
					map[string][]string{"Authorization": {"Token " + i3TokenB}},
				)

				require.Equal(t, http.StatusNoContent, result.response.Code)
				require.True(t, result.called)
				require.Equal(t, int64(101), result.principal.ID)
				require.Equal(t, 1, store.sessionLookups)
				require.Equal(t, 1, store.userLookups)
				require.Zero(t, store.tokenLookups)
				require.Zero(t, store.tokenTouches)
				require.Equal(t, []string{"session", "user"}, store.events)
			})

			for _, authorization := range []struct {
				name   string
				values []string
			}{
				{name: "malformed authorization", values: []string{"Bearer rejected"}},
				{name: "duplicate authorization", values: []string{"Token " + i3TokenB, "Token " + i3TokenB}},
			} {
				t.Run("valid session ignores "+authorization.name, func(t *testing.T) {
					store := i3ValidRESTStore(now, true)
					result := i3ServeProtected(store, now, boundary, http.MethodGet,
						[]*http.Cookie{i3Cookie(i3SessionCookieName, i3SessionA)},
						map[string][]string{"Authorization": authorization.values},
					)

					require.Equal(t, http.StatusNoContent, result.response.Code)
					require.True(t, result.called)
					require.Equal(t, int64(101), result.principal.ID)
					require.Zero(t, store.tokenLookups)
					require.Equal(t, []string{"session", "user"}, store.events)
				})
			}

			t.Run("session csrf failure never falls through", func(t *testing.T) {
				store := i3ValidRESTStore(now, true)
				result := i3ServeProtected(store, now, boundary, http.MethodPost,
					[]*http.Cookie{
						i3Cookie(i3SessionCookieName, i3SessionA),
						i3Cookie(i3CSRFCookieName, i3CSRFA),
					},
					map[string][]string{
						"Authorization": {"Token " + i3TokenB},
						"X-CSRFToken":   {i3CSRFOther},
					},
				)

				i3RequireDetail(t, result.response, http.StatusForbidden, "You do not have permission to perform this action.")
				require.False(t, result.called)
				require.Zero(t, store.tokenLookups)
				require.Zero(t, store.tokenTouches)
			})

			for _, state := range []struct {
				name      string
				configure func(*i3RESTStore)
				wantUsers int
			}{
				{name: "unknown", configure: func(store *i3RESTStore) {
					delete(store.sessions, i3Key(i3Digest(i3SessionA)))
				}},
				{name: "expired", configure: func(store *i3RESTStore) {
					record := store.sessions[i3Key(i3Digest(i3SessionA))]
					record.Expires = now
					store.sessions[i3Key(i3Digest(i3SessionA))] = record
				}},
				{name: "inactive owner", configure: func(store *i3RESTStore) {
					user := store.users[101]
					user.IsActive = false
					store.users[101] = user
				}, wantUsers: 1},
			} {
				t.Run(state.name+" session falls through to token", func(t *testing.T) {
					store := i3ValidRESTStore(now, true)
					state.configure(store)
					result := i3ServeProtected(store, now, boundary, http.MethodGet,
						[]*http.Cookie{i3Cookie(i3SessionCookieName, i3SessionA)},
						map[string][]string{"Authorization": {"Token " + i3TokenB}},
					)

					require.Equal(t, http.StatusNoContent, result.response.Code)
					require.True(t, result.called)
					require.Equal(t, int64(202), result.principal.ID)
					require.Equal(t, 1, store.sessionLookups)
					require.Equal(t, state.wantUsers, store.userLookups)
					require.Equal(t, 1, store.tokenLookups)
					wantEvents := []string{"session"}
					if state.wantUsers == 1 {
						wantEvents = append(wantEvents, "user")
					}
					wantEvents = append(wantEvents, "token")
					require.Equal(t, wantEvents, store.events)
				})

				t.Run(state.name+" session preserves malformed token rejection", func(t *testing.T) {
					store := i3ValidRESTStore(now, true)
					state.configure(store)
					result := i3ServeProtected(store, now, boundary, http.MethodGet,
						[]*http.Cookie{i3Cookie(i3SessionCookieName, i3SessionA)},
						map[string][]string{"Authorization": {"Bearer rejected"}},
					)

					i3RequireDetail(t, result.response, boundary.wantRejected, "Authentication credentials were not provided.")
					require.False(t, result.called)
					require.Equal(t, 1, store.sessionLookups)
					require.Equal(t, state.wantUsers, store.userLookups)
					require.Zero(t, store.tokenLookups)
					require.Zero(t, store.tokenTouches)
					wantEvents := []string{"session"}
					if state.wantUsers == 1 {
						wantEvents = append(wantEvents, "user")
					}
					require.Equal(t, wantEvents, store.events)
				})
			}

			for _, failure := range []struct {
				name       string
				configure  func(*i3RESTStore)
				wantEvents []string
			}{
				{name: "session lookup", configure: func(store *i3RESTStore) {
					store.sessionLookupErr = errors.New("session storage unavailable")
				}, wantEvents: []string{"session"}},
				{name: "owner lookup", configure: func(store *i3RESTStore) {
					store.userLookupErr = errors.New("owner storage unavailable")
				}, wantEvents: []string{"session", "user"}},
			} {
				t.Run(failure.name+" failure stops token fallback", func(t *testing.T) {
					store := i3ValidRESTStore(now, true)
					failure.configure(store)
					result := i3ServeProtected(store, now, boundary, http.MethodGet,
						[]*http.Cookie{i3Cookie(i3SessionCookieName, i3SessionA)},
						map[string][]string{"Authorization": {"Token " + i3TokenB}},
					)

					i3RequireDetail(t, result.response, http.StatusInternalServerError, "An internal error occurred.")
					require.False(t, result.called)
					require.Zero(t, store.tokenLookups)
					require.Zero(t, store.tokenTouches)
					require.Equal(t, failure.wantEvents, store.events)
				})
			}

			t.Run("duplicate session cookies fail before either credential lookup", func(t *testing.T) {
				store := i3ValidRESTStore(now, true)
				result := i3ServeProtected(store, now, boundary, http.MethodGet,
					[]*http.Cookie{
						i3Cookie(i3SessionCookieName, i3SessionA),
						i3Cookie(i3SessionCookieName, i3SessionUnknown),
					},
					map[string][]string{"Authorization": {"Token " + i3TokenB}},
				)

				i3RequireDetail(t, result.response, boundary.wantRejected, "Authentication credentials were not provided.")
				require.False(t, result.called)
				require.Zero(t, store.sessionLookups)
				require.Zero(t, store.tokenLookups)
				require.Empty(t, store.events)
			})

			t.Run("token remains usable when no session is present", func(t *testing.T) {
				store := i3ValidRESTStore(now, true)
				result := i3ServeProtected(store, now, boundary, http.MethodGet, nil,
					map[string][]string{"Authorization": {"Token " + i3TokenB}},
				)

				require.Equal(t, http.StatusNoContent, result.response.Code)
				require.True(t, result.called)
				require.Equal(t, int64(202), result.principal.ID)
				require.Zero(t, store.sessionLookups)
				require.Equal(t, 1, store.tokenLookups)
				require.Equal(t, []string{"token"}, store.events)
			})

			t.Run("empty session cookie is treated as absent and token remains usable", func(t *testing.T) {
				store := i3ValidRESTStore(now, true)
				result := i3ServeProtected(store, now, boundary, http.MethodGet,
					[]*http.Cookie{i3Cookie(i3SessionCookieName, "")},
					map[string][]string{"Authorization": {"Token " + i3TokenB}},
				)

				require.Equal(t, http.StatusNoContent, result.response.Code)
				require.True(t, result.called)
				require.Equal(t, int64(202), result.principal.ID)
				require.Zero(t, store.sessionLookups)
				require.Equal(t, 1, store.tokenLookups)
				require.Equal(t, []string{"token"}, store.events)
			})
		})
	}

	t.Run("extension preserves duplicate authorization handling without a session", func(t *testing.T) {
		store := i3ValidRESTStore(now, true)
		result := i3ServeProtected(store, now, i3RESTBoundaries()[1], http.MethodGet, nil,
			map[string][]string{"Authorization": {"Token " + i3TokenB, "Token ignored-by-existing-extension-parser"}},
		)

		require.Equal(t, http.StatusNoContent, result.response.Code)
		require.True(t, result.called)
		require.Equal(t, int64(202), result.principal.ID)
		require.Zero(t, store.sessionLookups)
		require.Equal(t, 1, store.tokenLookups)
		require.Equal(t, []string{"token"}, store.events)
	})

	t.Run("login rejects non-exact csrf before binding or application work", func(t *testing.T) {
		for _, pair := range []struct {
			name    string
			cookies []*http.Cookie
			headers []string
		}{
			{name: "missing pair"},
			{name: "cookie only", cookies: []*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA)}},
			{name: "header only", headers: []string{i3CSRFA}},
			{name: "empty cookie", cookies: []*http.Cookie{i3Cookie(i3CSRFCookieName, "")}, headers: []string{i3CSRFA}},
			{name: "empty header", cookies: []*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA)}, headers: []string{""}},
			{name: "literal mismatch", cookies: []*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA)}, headers: []string{i3CSRFOther}},
			{name: "duplicate cookies", cookies: []*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA), i3Cookie(i3CSRFCookieName, i3CSRFA)}, headers: []string{i3CSRFA}},
			{name: "duplicate headers", cookies: []*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA)}, headers: []string{i3CSRFA, i3CSRFA}},
		} {
			t.Run(pair.name, func(t *testing.T) {
				store := i3LoginRESTStore(t, now)
				headers := map[string][]string{"Content-Type": {"application/json"}}
				if pair.headers != nil {
					headers["X-CSRFToken"] = pair.headers
				}
				response := i3ServeRegisteredBody(store, now, false, http.MethodPost, "/api/auth/login/", pair.cookies, headers, []byte("{"))

				i3RequireDetail(t, response, http.StatusForbidden, "You do not have permission to perform this action.")
				require.Zero(t, store.usernameLookups)
				require.Zero(t, store.transactions)
				require.Zero(t, store.sessionCreates)
				require.Zero(t, store.deletes)
				require.Empty(t, store.events)
				i3RequireSetCookieCount(t, response, 0)
			})
		}
	})

	t.Run("login accepts the exact pair before binding", func(t *testing.T) {
		store := i3LoginRESTStore(t, now)
		response := i3ServeRegisteredBody(store, now, false, http.MethodPost, "/api/auth/login/",
			[]*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA)},
			map[string][]string{
				"Content-Type": {"application/json"},
				"X-CSRFToken":  {i3CSRFA},
			},
			[]byte("{"),
		)

		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Zero(t, store.usernameLookups)
		require.Zero(t, store.transactions)
		require.Empty(t, store.events)
		i3RequireSetCookieCount(t, response, 0)
	})

	t.Run("login application failures emit no replacement cookies", func(t *testing.T) {
		for _, failure := range []struct {
			name             string
			body             string
			configure        func(*i3RESTStore)
			wantStatus       int
			wantDetail       string
			wantUsernames    int
			wantTransactions int
			wantCreates      int
			wantEvents       []string
		}{
			{
				name:          "credential rejection",
				body:          `{"username":"browser-login-fixture","password":"rejected-browser-password"}`,
				wantStatus:    http.StatusUnauthorized,
				wantDetail:    "Authentication credentials were not provided.",
				wantUsernames: 1,
				wantEvents:    []string{"username"},
			},
			{
				name: "transaction start failure",
				body: i3LoginJSON(),
				configure: func(store *i3RESTStore) {
					store.transactionErr = errors.New("transaction unavailable")
				},
				wantStatus:       http.StatusInternalServerError,
				wantDetail:       "An internal error occurred.",
				wantUsernames:    1,
				wantTransactions: 1,
				wantEvents:       []string{"username", "transaction"},
			},
			{
				name: "commit failure",
				body: i3LoginJSON(),
				configure: func(store *i3RESTStore) {
					store.commitErr = errors.New("commit unavailable")
				},
				wantStatus:       http.StatusInternalServerError,
				wantDetail:       "An internal error occurred.",
				wantUsernames:    2,
				wantTransactions: 1,
				wantCreates:      1,
				wantEvents:       []string{"username", "transaction", "username", "create-session"},
			},
		} {
			t.Run(failure.name, func(t *testing.T) {
				store := i3LoginRESTStore(t, now)
				if failure.configure != nil {
					failure.configure(store)
				}
				response := i3ServeRegisteredBody(store, now, false, http.MethodPost, "/api/auth/login/",
					[]*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA)},
					map[string][]string{
						"Content-Type": {"application/json"},
						"X-CSRFToken":  {i3CSRFA},
					},
					[]byte(failure.body),
				)

				i3RequireDetail(t, response, failure.wantStatus, failure.wantDetail)
				require.Equal(t, failure.wantUsernames, store.usernameLookups)
				require.Equal(t, failure.wantTransactions, store.transactions)
				require.Equal(t, failure.wantCreates, store.sessionCreates)
				require.Zero(t, store.deletes)
				require.Equal(t, failure.wantEvents, store.events)
				if got := len(store.sessions); got != 1 {
					t.Fatalf("session row count = %d, want 1", got)
				}
				i3RequireSetCookieCount(t, response, 0)
			})
		}
	})

	t.Run("rejected login csrf does not consume throttle budget", func(t *testing.T) {
		store := i3LoginRESTStore(t, now)
		handler := NewHandler(application.NewService(store, i3RESTClock{now: now}), false)
		router := gin.New()
		handler.Register(router)
		for attempt := 0; attempt < 6; attempt++ {
			response := httptest.NewRecorder()
			request := i3RequestBody(http.MethodPost, "/api/auth/login/",
				[]*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA)},
				map[string][]string{
					"Content-Type": {"application/json"},
					"X-CSRFToken":  {i3CSRFOther},
				},
				[]byte(i3LoginJSON()),
			)
			router.ServeHTTP(response, request)
			i3RequireDetail(t, response, http.StatusForbidden, "You do not have permission to perform this action.")
			i3RequireSetCookieCount(t, response, 0)
		}
		require.Zero(t, store.usernameLookups)
		require.Zero(t, store.transactions)
		require.Empty(t, store.events)

		response := httptest.NewRecorder()
		request := i3RequestBody(http.MethodPost, "/api/auth/login/",
			[]*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA)},
			map[string][]string{
				"Content-Type": {"application/json"},
				"X-CSRFToken":  {i3CSRFA},
			},
			[]byte(i3LoginJSON()),
		)
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code)
		require.Equal(t, 2, store.usernameLookups)
		require.Equal(t, 1, store.transactions)
		i3RequireSetCookieCount(t, response, 2)
	})

	t.Run("login duplicate sessions fail before binding or application work", func(t *testing.T) {
		store := i3LoginRESTStore(t, now)
		response := i3ServeRegisteredBody(store, now, false, http.MethodPost, "/api/auth/login/",
			[]*http.Cookie{
				i3Cookie(i3SessionCookieName, i3SessionA),
				i3Cookie(i3SessionCookieName, i3SessionUnknown),
				i3Cookie(i3CSRFCookieName, i3CSRFA),
			},
			map[string][]string{
				"Content-Type": {"application/json"},
				"X-CSRFToken":  {i3CSRFA},
			},
			[]byte(i3LoginJSON()),
		)

		i3RequireDetail(t, response, http.StatusForbidden, "You do not have permission to perform this action.")
		require.Zero(t, store.usernameLookups)
		require.Zero(t, store.transactions)
		require.Zero(t, store.sessionCreates)
		require.Zero(t, store.deletes)
		require.Empty(t, store.events)
		i3RequireSetCookieCount(t, response, 0)
	})

	for _, candidate := range []struct {
		name        string
		cookie      *http.Cookie
		wantDeletes int
		wantEvents  []string
	}{
		{
			name:       "one empty session is not a rotation candidate",
			cookie:     i3Cookie(i3SessionCookieName, ""),
			wantEvents: []string{"username", "transaction", "username", "create-session"},
		},
		{
			name:        "one nonempty session is the rotation candidate",
			cookie:      i3Cookie(i3SessionCookieName, i3SessionA),
			wantDeletes: 1,
			wantEvents:  []string{"username", "transaction", "username", "delete-session", "create-session"},
		},
	} {
		t.Run(candidate.name, func(t *testing.T) {
			store := i3LoginRESTStore(t, now)
			response := i3ServeRegisteredBody(store, now, false, http.MethodPost, "/api/auth/login/",
				[]*http.Cookie{candidate.cookie, i3Cookie(i3CSRFCookieName, i3CSRFA)},
				map[string][]string{
					"Content-Type": {"application/json"},
					"X-CSRFToken":  {i3CSRFA},
				},
				[]byte(i3LoginJSON()),
			)

			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, 2, store.usernameLookups)
			require.Equal(t, 1, store.transactions)
			require.Equal(t, candidate.wantDeletes, store.deletes)
			require.Equal(t, 1, store.sessionCreates)
			require.Equal(t, candidate.wantEvents, store.events)
			i3RequireSetCookieCount(t, response, 2)
			session := i3RequireCookie(t, response, i3SessionCookieName)
			csrf := i3RequireCookie(t, response, i3CSRFCookieName)
			i3RequireSameOpaque(t, csrf.Value, i3DerivedCSRF(session.Value))
			require.True(t, i3SessionCSRFMatches(store, session.Value, csrf.Value), "issued cookies did not match the persisted session binding")
			i3RequireBodyOmitsOpaque(t, response.Body.Bytes(), session.Value, csrf.Value)
			if candidate.wantDeletes == 1 {
				require.False(t, i3HasSession(store, i3SessionA))
			}
		})
	}
}

func TestRESTSessionCSRFPairsAndMethodSafety(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 8, 17, 18, 15, 0, 0, time.UTC)

	for _, boundary := range i3RESTBoundaries() {
		t.Run(boundary.name, func(t *testing.T) {
			for _, pair := range []struct {
				name       string
				cookies    []*http.Cookie
				headers    []string
				storedCSRF string
				wantOK     bool
			}{
				{name: "missing pair", storedCSRF: i3CSRFA},
				{name: "cookie only", cookies: []*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA)}, storedCSRF: i3CSRFA},
				{name: "header only", headers: []string{i3CSRFA}, storedCSRF: i3CSRFA},
				{name: "empty cookie", cookies: []*http.Cookie{i3Cookie(i3CSRFCookieName, "")}, headers: []string{i3CSRFA}, storedCSRF: i3CSRFA},
				{name: "empty header", cookies: []*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA)}, headers: []string{""}, storedCSRF: i3CSRFA},
				{name: "literal mismatch", cookies: []*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA)}, headers: []string{i3CSRFOther}, storedCSRF: i3CSRFA},
				{name: "duplicate cookies", cookies: []*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA), i3Cookie(i3CSRFCookieName, i3CSRFA)}, headers: []string{i3CSRFA}, storedCSRF: i3CSRFA},
				{name: "duplicate headers", cookies: []*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA)}, headers: []string{i3CSRFA, i3CSRFA}, storedCSRF: i3CSRFA},
				{name: "stored digest mismatch", cookies: []*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFOther)}, headers: []string{i3CSRFOther}, storedCSRF: i3CSRFA},
				{name: "exact pair", cookies: []*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA)}, headers: []string{i3CSRFA}, storedCSRF: i3CSRFA, wantOK: true},
			} {
				t.Run(pair.name, func(t *testing.T) {
					store := i3ValidRESTStore(now, true)
					record := store.sessions[i3Key(i3Digest(i3SessionA))]
					record.CSRFHash = i3Digest(pair.storedCSRF)
					store.sessions[i3Key(i3Digest(i3SessionA))] = record
					cookies := append([]*http.Cookie{i3Cookie(i3SessionCookieName, i3SessionA)}, pair.cookies...)
					headers := map[string][]string{"Authorization": {"Token " + i3TokenB}}
					if pair.headers != nil {
						headers["X-CSRFToken"] = pair.headers
					}
					result := i3ServeProtected(store, now, boundary, http.MethodPost, cookies, headers)

					require.Zero(t, store.tokenLookups)
					require.Zero(t, store.tokenTouches)
					if pair.wantOK {
						require.Equal(t, http.StatusNoContent, result.response.Code)
						require.True(t, result.called)
						require.Equal(t, int64(101), result.principal.ID)
						return
					}
					i3RequireDetail(t, result.response, http.StatusForbidden, "You do not have permission to perform this action.")
					require.False(t, result.called)
				})
			}

			for _, race := range []struct {
				name            string
				configureSecond func(*i3RESTStore)
				wantStatus      int
				wantDetail      string
				wantUsers       int
				wantEvents      []string
			}{
				{
					name: "session disappears before stored csrf verification",
					configureSecond: func(store *i3RESTStore) {
						delete(store.sessions, i3Key(i3Digest(i3SessionA)))
					},
					wantStatus: http.StatusForbidden,
					wantDetail: "Authentication credentials were not provided.",
					wantUsers:  1,
					wantEvents: []string{"session", "user", "session"},
				},
				{
					name: "session expires before stored csrf verification",
					configureSecond: func(store *i3RESTStore) {
						record := store.sessions[i3Key(i3Digest(i3SessionA))]
						record.Expires = now
						store.sessions[i3Key(i3Digest(i3SessionA))] = record
					},
					wantStatus: http.StatusForbidden,
					wantDetail: "Authentication credentials were not provided.",
					wantUsers:  1,
					wantEvents: []string{"session", "user", "session"},
				},
				{
					name: "owner becomes inactive before stored csrf verification",
					configureSecond: func(store *i3RESTStore) {
						user := store.users[101]
						user.IsActive = false
						store.users[101] = user
					},
					wantStatus: http.StatusForbidden,
					wantDetail: "Authentication credentials were not provided.",
					wantUsers:  2,
					wantEvents: []string{"session", "user", "session", "user"},
				},
				{
					name: "session lookup fails during stored csrf verification",
					configureSecond: func(store *i3RESTStore) {
						store.sessionLookupErr = errors.New("session storage unavailable")
					},
					wantStatus: http.StatusInternalServerError,
					wantDetail: "An internal error occurred.",
					wantUsers:  1,
					wantEvents: []string{"session", "user", "session"},
				},
				{
					name: "owner lookup fails during stored csrf verification",
					configureSecond: func(store *i3RESTStore) {
						store.userLookupErr = errors.New("owner storage unavailable")
					},
					wantStatus: http.StatusInternalServerError,
					wantDetail: "An internal error occurred.",
					wantUsers:  2,
					wantEvents: []string{"session", "user", "session", "user"},
				},
			} {
				t.Run(race.name, func(t *testing.T) {
					store := i3ValidRESTStore(now, true)
					store.beforeSessionLoad = func(store *i3RESTStore, call int) {
						if call == 2 {
							race.configureSecond(store)
						}
					}
					result := i3ServeProtected(store, now, boundary, http.MethodPost,
						[]*http.Cookie{
							i3Cookie(i3SessionCookieName, i3SessionA),
							i3Cookie(i3CSRFCookieName, i3CSRFA),
						},
						map[string][]string{
							"Authorization": {"Token " + i3TokenB},
							"X-CSRFToken":   {i3CSRFA},
						},
					)

					i3RequireDetail(t, result.response, race.wantStatus, race.wantDetail)
					require.False(t, result.called)
					require.Equal(t, 2, store.sessionLookups)
					require.Equal(t, race.wantUsers, store.userLookups)
					require.Zero(t, store.tokenLookups)
					require.Zero(t, store.tokenTouches)
					require.Equal(t, race.wantEvents, store.events)
				})
			}

			t.Run("token authentication ignores csrf inputs", func(t *testing.T) {
				store := i3ValidRESTStore(now, true)
				result := i3ServeProtected(store, now, boundary, http.MethodPost,
					[]*http.Cookie{
						i3Cookie(i3CSRFCookieName, i3CSRFA),
						i3Cookie(i3CSRFCookieName, i3CSRFOther),
					},
					map[string][]string{
						"Authorization": {"Token " + i3TokenB},
						"X-CSRFToken":   {i3CSRFA, i3CSRFOther},
					},
				)

				require.Equal(t, http.StatusNoContent, result.response.Code)
				require.True(t, result.called)
				require.Equal(t, int64(202), result.principal.ID)
				require.Equal(t, 1, store.tokenLookups)
			})

			for _, methodCase := range []struct {
				name           string
				method         string
				session        bool
				csrf           bool
				wantStatus     int
				wantHandler    bool
				wantPrincipal  int64
				wantTokenCalls int
			}{
				{name: "session TRACE is csrf safe", method: http.MethodTrace, session: true, wantStatus: http.StatusNoContent, wantHandler: true, wantPrincipal: 101},
				{name: "session PROPFIND requires csrf", method: "PROPFIND", session: true, wantStatus: http.StatusForbidden},
				{name: "session PROPFIND accepts exact csrf", method: "PROPFIND", session: true, csrf: true, wantStatus: http.StatusNoContent, wantHandler: true, wantPrincipal: 101},
				{name: "read only token OPTIONS is safe", method: http.MethodOptions, wantStatus: http.StatusNoContent, wantHandler: true, wantPrincipal: 202, wantTokenCalls: 1},
				{name: "read only token TRACE is a write", method: http.MethodTrace, wantStatus: http.StatusForbidden, wantTokenCalls: 1},
				{name: "read only token unknown method is a write", method: "PROPFIND", wantStatus: http.StatusForbidden, wantTokenCalls: 1},
			} {
				t.Run(methodCase.name, func(t *testing.T) {
					store := i3ValidRESTStore(now, false)
					var cookies []*http.Cookie
					headers := map[string][]string{"Authorization": {"Token " + i3TokenB}}
					if methodCase.session {
						cookies = append(cookies, i3Cookie(i3SessionCookieName, i3SessionA))
					}
					if methodCase.csrf {
						cookies = append(cookies, i3Cookie(i3CSRFCookieName, i3CSRFA))
						headers["X-CSRFToken"] = []string{i3CSRFA}
					}
					result := i3ServeProtected(store, now, boundary, methodCase.method, cookies, headers)

					require.Equal(t, methodCase.wantStatus, result.response.Code)
					require.Equal(t, methodCase.wantHandler, result.called)
					require.Equal(t, methodCase.wantTokenCalls, store.tokenLookups)
					if methodCase.wantHandler {
						require.Equal(t, methodCase.wantPrincipal, result.principal.ID)
					}
				})
			}
		})
	}
}

func TestRESTCSRFBootstrapRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 8, 17, 18, 30, 0, 0, time.UTC)

	t.Run("anonymous bootstrap emits one fresh matching pair without storage", func(t *testing.T) {
		store := i3ValidRESTStore(now, true)
		response := i3ServeRegistered(store, now, false, http.MethodGet, "/api/auth/csrf/", nil, nil)

		require.Equal(t, http.StatusOK, response.Code)
		bodyToken := i3BootstrapBodyToken(t, response)
		cookie := i3RequireCookie(t, response, i3CSRFCookieName)
		i3RequireSameOpaque(t, bodyToken, cookie.Value)
		decoded, err := base64.RawURLEncoding.DecodeString(bodyToken)
		require.NoError(t, err)
		if len(decoded) != 32 {
			t.Fatalf("decoded CSRF length = %d, want 32", len(decoded))
		}
		i3RequireSetCookieCount(t, response, 1)

		second := i3ServeRegistered(store, now, false, http.MethodGet, "/api/auth/csrf/", nil, nil)
		require.Equal(t, http.StatusOK, second.Code)
		secondBodyToken := i3BootstrapBodyToken(t, second)
		i3RequireSameOpaque(t, secondBodyToken, i3RequireCookie(t, second, i3CSRFCookieName).Value)
		require.False(t, i3OpaqueEqual(bodyToken, secondBodyToken), "anonymous bootstrap values were not fresh")
		i3RequireSetCookieCount(t, second, 1)
		require.Zero(t, store.transactions)
		require.Zero(t, store.sessionLookups)
	})

	t.Run("active bootstrap heals once and then reuses the derived value", func(t *testing.T) {
		store := i3ValidRESTStore(now, true)
		record := store.sessions[i3Key(i3Digest(i3SessionA))]
		record.CSRFHash = i3Digest(i3CSRFOther)
		store.sessions[i3Key(i3Digest(i3SessionA))] = record
		expected := i3DerivedCSRF(i3SessionA)

		first := i3ServeRegistered(store, now, false, http.MethodGet, "/api/auth/csrf/",
			[]*http.Cookie{i3Cookie(i3SessionCookieName, i3SessionA)}, nil)
		require.Equal(t, http.StatusOK, first.Code)
		i3RequireSameOpaque(t, i3BootstrapBodyToken(t, first), expected)
		i3RequireSameOpaque(t, i3RequireCookie(t, first, i3CSRFCookieName).Value, expected)
		require.Equal(t, 1, store.transactions)
		require.Equal(t, 1, store.csrfUpdates)
		require.True(t, i3SessionCSRFMatches(store, i3SessionA, expected), "stored CSRF digest did not converge")

		second := i3ServeRegistered(store, now, false, http.MethodGet, "/api/auth/csrf/",
			[]*http.Cookie{i3Cookie(i3SessionCookieName, i3SessionA)}, nil)
		require.Equal(t, http.StatusOK, second.Code)
		i3RequireSameOpaque(t, i3BootstrapBodyToken(t, second), expected)
		i3RequireSameOpaque(t, i3RequireCookie(t, second, i3CSRFCookieName).Value, expected)
		require.Equal(t, 2, store.transactions)
		require.Equal(t, 1, store.csrfUpdates)
	})

	for _, state := range []struct {
		name      string
		configure func(*i3RESTStore)
		wantUsers int
	}{
		{name: "unknown", configure: func(store *i3RESTStore) {
			delete(store.sessions, i3Key(i3Digest(i3SessionA)))
		}},
		{name: "expired", configure: func(store *i3RESTStore) {
			record := store.sessions[i3Key(i3Digest(i3SessionA))]
			record.Expires = now
			store.sessions[i3Key(i3Digest(i3SessionA))] = record
		}},
		{name: "inactive owner", configure: func(store *i3RESTStore) {
			user := store.users[101]
			user.IsActive = false
			store.users[101] = user
		}, wantUsers: 1},
	} {
		t.Run(state.name+" session receives anonymous bootstrap", func(t *testing.T) {
			store := i3ValidRESTStore(now, true)
			state.configure(store)
			response := i3ServeRegistered(store, now, false, http.MethodGet, "/api/auth/csrf/",
				[]*http.Cookie{i3Cookie(i3SessionCookieName, i3SessionA)}, nil)

			require.Equal(t, http.StatusOK, response.Code)
			bodyToken := i3BootstrapBodyToken(t, response)
			i3RequireSameOpaque(t, bodyToken, i3RequireCookie(t, response, i3CSRFCookieName).Value)
			require.False(t, i3OpaqueEqual(bodyToken, i3DerivedCSRF(i3SessionA)), "invalid session received a session-derived CSRF value")
			require.Equal(t, 1, store.transactions)
			require.Equal(t, 1, store.sessionLookups)
			require.Equal(t, state.wantUsers, store.userLookups)
			require.Zero(t, store.csrfUpdates)
		})
	}

	for _, failure := range []struct {
		name      string
		configure func(*i3RESTStore)
	}{
		{name: "transaction start", configure: func(store *i3RESTStore) {
			store.transactionErr = errors.New("session transaction unavailable")
		}},
		{name: "session lookup", configure: func(store *i3RESTStore) {
			store.sessionLookupErr = errors.New("session storage unavailable")
		}},
		{name: "owner lookup", configure: func(store *i3RESTStore) {
			store.userLookupErr = errors.New("owner storage unavailable")
		}},
		{name: "commit", configure: func(store *i3RESTStore) {
			store.commitErr = errors.New("session transaction unavailable")
		}},
	} {
		t.Run(failure.name+" failure emits no replacement", func(t *testing.T) {
			store := i3ValidRESTStore(now, true)
			record := store.sessions[i3Key(i3Digest(i3SessionA))]
			record.CSRFHash = i3Digest(i3CSRFOther)
			store.sessions[i3Key(i3Digest(i3SessionA))] = record
			failure.configure(store)
			response := i3ServeRegistered(store, now, false, http.MethodGet, "/api/auth/csrf/",
				[]*http.Cookie{i3Cookie(i3SessionCookieName, i3SessionA)}, nil)

			i3RequireDetail(t, response, http.StatusInternalServerError, "An internal error occurred.")
			i3RequireSetCookieCount(t, response, 0)
		})
	}

	t.Run("disappearing row takes the anonymous path", func(t *testing.T) {
		store := i3ValidRESTStore(now, true)
		record := store.sessions[i3Key(i3Digest(i3SessionA))]
		record.CSRFHash = i3Digest(i3CSRFOther)
		store.sessions[i3Key(i3Digest(i3SessionA))] = record
		store.updateErr = application.ErrNotFound
		response := i3ServeRegistered(store, now, false, http.MethodGet, "/api/auth/csrf/",
			[]*http.Cookie{i3Cookie(i3SessionCookieName, i3SessionA)}, nil)

		require.Equal(t, http.StatusOK, response.Code)
		bodyToken := i3BootstrapBodyToken(t, response)
		i3RequireSameOpaque(t, bodyToken, i3RequireCookie(t, response, i3CSRFCookieName).Value)
		require.False(t, i3OpaqueEqual(bodyToken, i3DerivedCSRF(i3SessionA)), "failed update returned session-derived material")
		require.Equal(t, 1, store.csrfUpdates)
	})

	t.Run("duplicate session cookies fail closed without replacement", func(t *testing.T) {
		store := i3ValidRESTStore(now, true)
		response := i3ServeRegistered(store, now, false, http.MethodGet, "/api/auth/csrf/",
			[]*http.Cookie{
				i3Cookie(i3SessionCookieName, i3SessionA),
				i3Cookie(i3SessionCookieName, i3SessionUnknown),
			}, nil)

		i3RequireDetail(t, response, http.StatusForbidden, "You do not have permission to perform this action.")
		i3RequireSetCookieCount(t, response, 0)
		require.Zero(t, store.transactions)
		require.Zero(t, store.sessionLookups)
	})
}

func TestRESTSessionCookieLifecycleContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 8, 17, 18, 45, 0, 0, time.UTC)
	expires := time.Date(2035, 1, 2, 3, 4, 5, 0, time.UTC)

	for _, secure := range []bool{false, true} {
		name := "development"
		if secure {
			name = "production"
		}
		t.Run(name+" issue cookies", func(t *testing.T) {
			handler := NewHandler(application.NewService(i3ValidRESTStore(now, true), i3RESTClock{now: now}), secure)
			context, response := i3ResponseContext()
			handler.setSessionCookie(context, i3SessionA, expires)
			handler.setCSRFCookie(context, i3CSRFA)

			i3RequireSetCookieCount(t, response, 2)
			session := i3RequireCookie(t, response, i3SessionCookieName)
			csrf := i3RequireCookie(t, response, i3CSRFCookieName)
			i3RequireSameOpaque(t, session.Value, i3SessionA)
			i3RequireSameOpaque(t, csrf.Value, i3CSRFA)
			require.Equal(t, "/", session.Path)
			require.Equal(t, "/", csrf.Path)
			require.Empty(t, session.Domain)
			require.Empty(t, csrf.Domain)
			require.True(t, session.HttpOnly)
			require.False(t, csrf.HttpOnly)
			require.Equal(t, secure, session.Secure)
			require.Equal(t, secure, csrf.Secure)
			require.Equal(t, http.SameSiteLaxMode, session.SameSite)
			require.Equal(t, http.SameSiteLaxMode, csrf.SameSite)
			require.Equal(t, 43_200, session.MaxAge)
			require.Equal(t, 43_200, csrf.MaxAge)
			require.True(t, session.Expires.Equal(expires), "session expiry did not preserve the application value")
		})
	}

	for _, secure := range []bool{false, true} {
		name := "development"
		if secure {
			name = "production"
		}
		t.Run(name+" tombstones preserve scope and delete immediately", func(t *testing.T) {
			handler := NewHandler(application.NewService(i3ValidRESTStore(now, true), i3RESTClock{now: now}), secure)
			context, response := i3ResponseContext()
			handler.clearCookies(context)

			i3RequireSetCookieCount(t, response, 2)
			session := i3RequireCookie(t, response, i3SessionCookieName)
			csrf := i3RequireCookie(t, response, i3CSRFCookieName)
			for _, cookie := range []*http.Cookie{session, csrf} {
				require.Zero(t, len(cookie.Value))
				require.Equal(t, "/", cookie.Path)
				require.Empty(t, cookie.Domain)
				require.Equal(t, secure, cookie.Secure)
				require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
				require.Equal(t, -1, cookie.MaxAge)
				require.True(t, cookie.Expires.Before(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)), "cookie tombstone expiry was not in the past")
			}
			require.True(t, session.HttpOnly)
			require.False(t, csrf.HttpOnly)
			for _, header := range response.Header().Values("Set-Cookie") {
				require.True(t, strings.Contains(header, "Max-Age=0"), "cookie tombstone did not emit Max-Age=0")
			}
		})
	}
}

func TestRESTLogoutIsSessionOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 8, 17, 19, 0, 0, 0, time.UTC)

	t.Run("token alone cannot log out or clear cookies", func(t *testing.T) {
		store := i3ValidRESTStore(now, true)
		response := i3ServeRegistered(store, now, true, http.MethodPost, "/api/auth/logout/", nil,
			map[string][]string{"Authorization": {"Token " + i3TokenB}})

		i3RequireDetail(t, response, http.StatusUnauthorized, "Authentication credentials were not provided.")
		require.Zero(t, store.tokenLookups)
		require.Zero(t, store.tokenTouches)
		require.Zero(t, store.deletes)
		i3RequireSetCookieCount(t, response, 0)
	})

	t.Run("invalid session cannot fall through to token", func(t *testing.T) {
		store := i3ValidRESTStore(now, true)
		response := i3ServeRegistered(store, now, true, http.MethodPost, "/api/auth/logout/",
			[]*http.Cookie{i3Cookie(i3SessionCookieName, i3SessionUnknown)},
			map[string][]string{"Authorization": {"Token " + i3TokenB}})

		i3RequireDetail(t, response, http.StatusUnauthorized, "Authentication credentials were not provided.")
		require.Zero(t, store.tokenLookups)
		require.Zero(t, store.tokenTouches)
		require.Zero(t, store.deletes)
		i3RequireSetCookieCount(t, response, 0)
	})

	t.Run("duplicate sessions stop before storage", func(t *testing.T) {
		store := i3ValidRESTStore(now, true)
		response := i3ServeRegistered(store, now, true, http.MethodPost, "/api/auth/logout/",
			[]*http.Cookie{
				i3Cookie(i3SessionCookieName, i3SessionA),
				i3Cookie(i3SessionCookieName, i3SessionUnknown),
			}, map[string][]string{"Authorization": {"Token " + i3TokenB}})

		i3RequireDetail(t, response, http.StatusUnauthorized, "Authentication credentials were not provided.")
		require.Zero(t, store.sessionLookups)
		require.Zero(t, store.tokenLookups)
		require.Zero(t, store.deletes)
		i3RequireSetCookieCount(t, response, 0)
	})

	t.Run("csrf failure never revokes or clears", func(t *testing.T) {
		store := i3ValidRESTStore(now, true)
		response := i3ServeRegistered(store, now, true, http.MethodPost, "/api/auth/logout/",
			[]*http.Cookie{
				i3Cookie(i3SessionCookieName, i3SessionA),
				i3Cookie(i3CSRFCookieName, i3CSRFA),
			}, map[string][]string{
				"Authorization": {"Token " + i3TokenB},
				"X-CSRFToken":   {i3CSRFOther},
			})

		i3RequireDetail(t, response, http.StatusForbidden, "You do not have permission to perform this action.")
		require.Zero(t, store.tokenLookups)
		require.Zero(t, store.transactions)
		require.Zero(t, store.deletes)
		i3RequireSetCookieCount(t, response, 0)
	})

	for _, duplicate := range []struct {
		name    string
		cookies []*http.Cookie
		headers []string
	}{
		{
			name: "duplicate csrf cookies",
			cookies: []*http.Cookie{
				i3Cookie(i3CSRFCookieName, i3CSRFA),
				i3Cookie(i3CSRFCookieName, i3CSRFA),
			},
			headers: []string{i3CSRFA},
		},
		{
			name:    "duplicate csrf headers",
			cookies: []*http.Cookie{i3Cookie(i3CSRFCookieName, i3CSRFA)},
			headers: []string{i3CSRFA, i3CSRFA},
		},
	} {
		t.Run(duplicate.name+" never revoke or clear", func(t *testing.T) {
			store := i3ValidRESTStore(now, true)
			cookies := append([]*http.Cookie{i3Cookie(i3SessionCookieName, i3SessionA)}, duplicate.cookies...)
			response := i3ServeRegistered(store, now, true, http.MethodPost, "/api/auth/logout/", cookies,
				map[string][]string{"X-CSRFToken": duplicate.headers})

			i3RequireDetail(t, response, http.StatusForbidden, "You do not have permission to perform this action.")
			require.Zero(t, store.transactions)
			require.Zero(t, store.deletes)
			i3RequireSetCookieCount(t, response, 0)
		})
	}

	t.Run("valid session ignores authorization and commits one revocation", func(t *testing.T) {
		store := i3ValidRESTStore(now, true)
		response := i3ServeRegistered(store, now, true, http.MethodPost, "/api/auth/logout/",
			[]*http.Cookie{
				i3Cookie(i3SessionCookieName, i3SessionA),
				i3Cookie(i3CSRFCookieName, i3CSRFA),
			}, map[string][]string{
				"Authorization": {"Bearer rejected"},
				"X-CSRFToken":   {i3CSRFA},
			})

		require.Equal(t, http.StatusNoContent, response.Code)
		require.Zero(t, response.Body.Len())
		require.Zero(t, store.tokenLookups)
		require.Equal(t, 1, store.transactions)
		require.Equal(t, 1, store.deletes)
		require.False(t, i3HasSession(store, i3SessionA))
		i3RequireSetCookieCount(t, response, 2)
		for _, name := range []string{i3SessionCookieName, i3CSRFCookieName} {
			cookie := i3RequireCookie(t, response, name)
			require.Zero(t, len(cookie.Value))
			require.Equal(t, -1, cookie.MaxAge)
		}
	})

	t.Run("csrf is revalidated inside the revocation transaction", func(t *testing.T) {
		store := i3ValidRESTStore(now, true)
		store.beforeTransaction = func(store *i3RESTStore) {
			record := store.sessions[i3Key(i3Digest(i3SessionA))]
			record.CSRFHash = i3Digest(i3CSRFOther)
			store.sessions[i3Key(i3Digest(i3SessionA))] = record
		}
		response := i3ServeRegistered(store, now, true, http.MethodPost, "/api/auth/logout/",
			[]*http.Cookie{
				i3Cookie(i3SessionCookieName, i3SessionA),
				i3Cookie(i3CSRFCookieName, i3CSRFA),
			}, map[string][]string{"X-CSRFToken": {i3CSRFA}})

		i3RequireDetail(t, response, http.StatusForbidden, "You do not have permission to perform this action.")
		require.Equal(t, 1, store.transactions)
		require.Zero(t, store.deletes)
		require.True(t, i3HasSession(store, i3SessionA))
		i3RequireSetCookieCount(t, response, 0)
	})

	for _, race := range []struct {
		name       string
		configure  func(*i3RESTStore)
		wantStatus int
		wantDetail string
	}{
		{
			name: "session disappears after middleware authentication",
			configure: func(store *i3RESTStore) {
				delete(store.sessions, i3Key(i3Digest(i3SessionA)))
			},
			wantStatus: http.StatusUnauthorized,
			wantDetail: "Authentication credentials were not provided.",
		},
		{
			name: "session lookup fails after middleware authentication",
			configure: func(store *i3RESTStore) {
				store.sessionLookupErr = errors.New("session storage unavailable")
			},
			wantStatus: http.StatusInternalServerError,
			wantDetail: "An internal error occurred.",
		},
	} {
		t.Run(race.name+" emits no tombstone", func(t *testing.T) {
			store := i3ValidRESTStore(now, true)
			store.beforeTransaction = race.configure
			response := i3ServeRegistered(store, now, true, http.MethodPost, "/api/auth/logout/",
				[]*http.Cookie{
					i3Cookie(i3SessionCookieName, i3SessionA),
					i3Cookie(i3CSRFCookieName, i3CSRFA),
				}, map[string][]string{
					"Authorization": {"Token " + i3TokenB},
					"X-CSRFToken":   {i3CSRFA},
				})

			i3RequireDetail(t, response, race.wantStatus, race.wantDetail)
			require.Equal(t, 1, store.transactions)
			require.Zero(t, store.tokenLookups)
			require.Zero(t, store.tokenTouches)
			require.Zero(t, store.deletes)
			i3RequireSetCookieCount(t, response, 0)
		})
	}

	t.Run("failed commit emits no false tombstone", func(t *testing.T) {
		store := i3ValidRESTStore(now, true)
		store.commitErr = errors.New("session commit unavailable")
		response := i3ServeRegistered(store, now, true, http.MethodPost, "/api/auth/logout/",
			[]*http.Cookie{
				i3Cookie(i3SessionCookieName, i3SessionA),
				i3Cookie(i3CSRFCookieName, i3CSRFA),
			}, map[string][]string{"X-CSRFToken": {i3CSRFA}})

		i3RequireDetail(t, response, http.StatusInternalServerError, "An internal error occurred.")
		require.Equal(t, 1, store.transactions)
		require.Equal(t, 1, store.deletes)
		require.True(t, i3HasSession(store, i3SessionA))
		i3RequireSetCookieCount(t, response, 0)
	})
}

func i3RESTBoundaries() []i3RESTBoundary {
	return []i3RESTBoundary{
		{
			name:         "baseline",
			wantRejected: http.StatusForbidden,
			middleware:   func(handler *Handler) gin.HandlerFunc { return handler.BaselineMiddleware() },
		},
		{
			name:         "extension",
			wantRejected: http.StatusUnauthorized,
			middleware:   func(handler *Handler) gin.HandlerFunc { return handler.Middleware() },
		},
	}
}

func i3ValidRESTStore(now time.Time, tokenWriteEnabled bool) *i3RESTStore {
	lastUsed := now
	store := &i3RESTStore{
		sessions: make(map[i3DigestKey]application.SessionRecord),
		users: map[int64]domain.User{
			101: {ID: 101, Username: "session-user", IsActive: true},
			202: {ID: 202, Username: "token-user", IsActive: true},
		},
		tokenRecord: application.TokenRecord{Token: domain.APIToken{
			ID: 301, UserID: 202, WriteEnabled: tokenWriteEnabled, LastUsed: &lastUsed,
		}},
		tokenUser: domain.User{ID: 202, Username: "token-user", IsActive: true},
	}
	store.sessions[i3Key(i3Digest(i3SessionA))] = application.SessionRecord{
		SecretHash: i3Digest(i3SessionA),
		CSRFHash:   i3Digest(i3CSRFA),
		UserID:     101,
		Created:    now.Add(-time.Hour),
		LastSeen:   now.Add(-time.Hour),
		Expires:    now.Add(time.Hour),
	}
	return store
}

func i3LoginRESTStore(t *testing.T, now time.Time) *i3RESTStore {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(i3LoginPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatal("could not prepare login credential fixture")
	}
	store := i3ValidRESTStore(now, true)
	store.loginUser = domain.User{ID: 303, Username: i3LoginUsername, IsActive: true}
	store.loginPasswordHash = string(hash)
	store.users[store.loginUser.ID] = store.loginUser
	return store
}

func i3LoginJSON() string {
	return `{"username":"browser-login-fixture","password":"browser-password-fixture"}`
}

func i3ServeProtected(
	store *i3RESTStore,
	now time.Time,
	boundary i3RESTBoundary,
	method string,
	cookies []*http.Cookie,
	headers map[string][]string,
) i3ProtectedResult {
	handler := NewHandler(application.NewService(store, i3RESTClock{now: now}), false)
	result := i3ProtectedResult{response: httptest.NewRecorder()}
	router := gin.New()
	router.Handle(method, "/protected", boundary.middleware(handler), func(c *gin.Context) {
		result.called = true
		result.principal, _ = domain.PrincipalFromContext(c.Request.Context())
		c.Status(http.StatusNoContent)
	})
	request := i3Request(method, "/protected", cookies, headers)
	router.ServeHTTP(result.response, request)
	return result
}

func i3ServeRegistered(
	store *i3RESTStore,
	now time.Time,
	secure bool,
	method, path string,
	cookies []*http.Cookie,
	headers map[string][]string,
) *httptest.ResponseRecorder {
	return i3ServeRegisteredBody(store, now, secure, method, path, cookies, headers, nil)
}

func i3ServeRegisteredBody(
	store *i3RESTStore,
	now time.Time,
	secure bool,
	method, path string,
	cookies []*http.Cookie,
	headers map[string][]string,
	body []byte,
) *httptest.ResponseRecorder {
	handler := NewHandler(application.NewService(store, i3RESTClock{now: now}), secure)
	router := gin.New()
	handler.Register(router)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, i3RequestBody(method, path, cookies, headers, body))
	return response
}

func i3Request(method, path string, cookies []*http.Cookie, headers map[string][]string) *http.Request {
	return i3RequestBody(method, path, cookies, headers, nil)
}

func i3RequestBody(method, path string, cookies []*http.Cookie, headers map[string][]string, body []byte) *http.Request {
	var request *http.Request
	if body == nil {
		request = httptest.NewRequest(method, path, nil)
	} else {
		request = httptest.NewRequest(method, path, bytes.NewReader(body))
	}
	request.RemoteAddr = "192.0.2.10:443"
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	for name, values := range headers {
		for _, value := range values {
			request.Header.Add(name, value)
		}
	}
	return request
}

func i3Cookie(name, value string) *http.Cookie {
	return &http.Cookie{Name: name, Value: value, Path: "/"}
}

func i3ResponseContext() (*gin.Context, *httptest.ResponseRecorder) {
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return context, response
}

func i3RequireDetail(t *testing.T, response *httptest.ResponseRecorder, status int, detail string) {
	t.Helper()
	require.Equal(t, status, response.Code)
	require.Empty(t, response.Header().Get("WWW-Authenticate"))
	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, map[string]string{"detail": detail}, body)
}

func i3BootstrapBodyToken(t *testing.T, response *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Token string `json:"csrf_token"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	if body.Token == "" {
		t.Fatal("CSRF bootstrap returned no body value")
	}
	return body.Token
}

func i3RequireCookie(t *testing.T, response *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("response did not contain the %s cookie", name)
	return nil
}

func i3RequireSetCookieCount(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if got := len(response.Header().Values("Set-Cookie")); got != want {
		t.Fatalf("Set-Cookie count = %d, want %d", got, want)
	}
}

func i3RequireSameOpaque(t *testing.T, left, right string) {
	t.Helper()
	require.True(t, i3OpaqueEqual(left, right), "opaque values did not match")
}

func i3RequireBodyOmitsOpaque(t *testing.T, body []byte, values ...string) {
	t.Helper()
	for _, value := range values {
		if value != "" && bytes.Contains(body, []byte(value)) {
			t.Fatal("response body exposed opaque credential material")
		}
	}
}

func i3OpaqueEqual(left, right string) bool {
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

func i3DerivedCSRF(sessionSecret string) string {
	mac := hmac.New(sha256.New, []byte(sessionSecret))
	_, _ = mac.Write([]byte(i3CSRFDomainTag))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func i3Digest(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return append([]byte(nil), sum[:]...)
}

func i3Key(hash []byte) i3DigestKey {
	var key i3DigestKey
	copy(key[:], hash)
	return key
}

func i3CloneSession(record application.SessionRecord) application.SessionRecord {
	record.SecretHash = append([]byte(nil), record.SecretHash...)
	record.CSRFHash = append([]byte(nil), record.CSRFHash...)
	return record
}

func i3CloneSessions(source map[i3DigestKey]application.SessionRecord) map[i3DigestKey]application.SessionRecord {
	clone := make(map[i3DigestKey]application.SessionRecord, len(source))
	for key, record := range source {
		clone[key] = i3CloneSession(record)
	}
	return clone
}

func i3HasSession(store *i3RESTStore, secret string) bool {
	_, ok := store.sessions[i3Key(i3Digest(secret))]
	return ok
}

func i3SessionCSRFMatches(store *i3RESTStore, sessionSecret, csrf string) bool {
	record, ok := store.sessions[i3Key(i3Digest(sessionSecret))]
	return ok && subtle.ConstantTimeCompare(record.CSRFHash, i3Digest(csrf)) == 1
}
