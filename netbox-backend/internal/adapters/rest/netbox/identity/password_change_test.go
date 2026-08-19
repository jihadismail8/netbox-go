package identity

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
)

const (
	i4RESTCurrentPassword = "transport-current-password"
	i4RESTNextPassword    = "transport-replacement-password"
	i4RESTSession         = "transport-browser-session"
	i4RESTSiblingSession  = "transport-browser-sibling"
	i4RESTTokenSession    = "transport-token-user-session"
	i4RESTToken           = "transport-api-token"

	i4RESTPasswordChangeContextFixtureKey = "netbox-go.identity.password-change-credential"
)

type i4RESTClock struct{ now time.Time }

func (clock i4RESTClock) Now() time.Time { return clock.now }

type i4RESTDigest [sha256.Size]byte

type i4RESTStore struct {
	application.Store

	users    map[int64]domain.User
	hashes   map[int64]string
	sessions map[i4RESTDigest]application.SessionRecord

	tokenRecord application.TokenRecord
	tokenUser   domain.User

	sessionLookupErr error
	userLookupErr    error
	tokenLookupErr   error
	transactionErr   error
	updateErr        error
	deleteErr        error
	createErr        error
	commitErr        error

	beforeTransaction func(*i4RESTStore)

	transactions   int
	userLookups    int
	sessionLookups int
	tokenLookups   int
	tokenTouches   int
	updates        int
	deletes        int
	creates        int
	updatedUsers   []int64
	deletedUsers   []int64
}

func (store *i4RESTStore) Transaction(_ context.Context, fn func(application.Store) error) error {
	store.transactions++
	if store.transactionErr != nil {
		return store.transactionErr
	}
	if store.beforeTransaction != nil {
		hook := store.beforeTransaction
		store.beforeTransaction = nil
		hook(store)
	}
	users := i4RESTCloneUsers(store.users)
	hashes := i4RESTCloneHashes(store.hashes)
	sessions := i4RESTCloneSessions(store.sessions)
	if err := fn(store); err != nil {
		store.users = users
		store.hashes = hashes
		store.sessions = sessions
		return err
	}
	if store.commitErr != nil {
		store.users = users
		store.hashes = hashes
		store.sessions = sessions
		return store.commitErr
	}
	return nil
}

func (store *i4RESTStore) UserByID(_ context.Context, id int64) (domain.User, string, error) {
	store.userLookups++
	if store.userLookupErr != nil {
		return domain.User{}, "", store.userLookupErr
	}
	user, ok := store.users[id]
	if !ok {
		return domain.User{}, "", application.ErrNotFound
	}
	return user, store.hashes[id], nil
}

func (store *i4RESTStore) SessionByHash(_ context.Context, hash []byte) (application.SessionRecord, error) {
	store.sessionLookups++
	if store.sessionLookupErr != nil {
		return application.SessionRecord{}, store.sessionLookupErr
	}
	record, ok := store.sessions[i4RESTKey(hash)]
	if !ok {
		return application.SessionRecord{}, application.ErrNotFound
	}
	return i4RESTCloneSession(record), nil
}

func (store *i4RESTStore) TokenByHash(_ context.Context, hash []byte) (application.TokenRecord, domain.User, error) {
	store.tokenLookups++
	if store.tokenLookupErr != nil {
		return application.TokenRecord{}, domain.User{}, store.tokenLookupErr
	}
	if subtle.ConstantTimeCompare(hash, i4RESTDigestValue(i4RESTToken)) != 1 {
		return application.TokenRecord{}, domain.User{}, application.ErrNotFound
	}
	return store.tokenRecord, store.tokenUser, nil
}

func (store *i4RESTStore) TouchToken(context.Context, int64, time.Time) error {
	store.tokenTouches++
	return nil
}

func (store *i4RESTStore) UpdatePassword(_ context.Context, id int64, hash string, changedAt time.Time) error {
	store.updates++
	store.updatedUsers = append(store.updatedUsers, id)
	if store.updateErr != nil {
		return store.updateErr
	}
	if _, ok := store.users[id]; !ok {
		return application.ErrNotFound
	}
	store.hashes[id] = hash
	user := store.users[id]
	user.Updated = changedAt
	store.users[id] = user
	return nil
}

func (store *i4RESTStore) DeleteSessionsForUser(_ context.Context, userID int64) error {
	store.deletes++
	store.deletedUsers = append(store.deletedUsers, userID)
	if store.deleteErr != nil {
		return store.deleteErr
	}
	for key, record := range store.sessions {
		if record.UserID == userID {
			delete(store.sessions, key)
		}
	}
	return nil
}

func (store *i4RESTStore) CreateSession(_ context.Context, record application.SessionRecord) error {
	store.creates++
	if store.createErr != nil {
		return store.createErr
	}
	key := i4RESTKey(record.SecretHash)
	if _, exists := store.sessions[key]; exists {
		return errors.New("duplicate session fixture")
	}
	store.sessions[key] = i4RESTCloneSession(record)
	return nil
}

func TestRESTPasswordChangeCredentialProvenance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 8, 19, 8, 0, 0, 0, time.UTC)

	t.Run("provenance guard precedes application call", func(t *testing.T) {
		i4RESTRequireProvenanceGuardBeforeService(t)
	})

	for _, authorization := range []struct {
		name  string
		value string
	}{
		{name: "absent authorization"},
		{name: "valid authorization ignored", value: "Token " + i4RESTToken},
		{name: "invalid authorization ignored", value: "Token unknown"},
		{name: "malformed authorization ignored", value: "Bearer malformed"},
	} {
		t.Run("valid session wins/"+authorization.name, func(t *testing.T) {
			store := i4RESTFixture(t, now)
			headers := map[string][]string{
				"X-CSRFToken": {i3DerivedCSRF(i4RESTSession)},
			}
			if authorization.value != "" {
				headers["Authorization"] = []string{authorization.value}
			}
			response := i4RESTPasswordRequest(store, now, false,
				[]*http.Cookie{
					i3Cookie(sessionCookie, i4RESTSession),
					i3Cookie(csrfCookie, i3DerivedCSRF(i4RESTSession)),
				},
				headers,
				i4RESTValidBody(),
			)

			require.Equal(t, http.StatusNoContent, response.Code)
			require.Zero(t, response.Body.Len())
			require.Equal(t, []int64{101}, store.updatedUsers)
			require.Equal(t, []int64{101}, store.deletedUsers)
			require.Zero(t, store.tokenLookups)
			require.Equal(t, 1, store.creates)
			i4RESTRequirePassword(t, store.hashes[101], i4RESTNextPassword, true)
			i4RESTRequirePassword(t, store.hashes[202], i4RESTCurrentPassword, true)
		})
	}

	t.Run("session csrf refusal never falls through", func(t *testing.T) {
		store := i4RESTFixture(t, now)
		response := i4RESTPasswordRequest(store, now, false,
			[]*http.Cookie{
				i3Cookie(sessionCookie, i4RESTSession),
				i3Cookie(csrfCookie, i3DerivedCSRF(i4RESTSession)),
			},
			map[string][]string{
				"Authorization": {"Token " + i4RESTToken},
				"X-CSRFToken":   {"different-csrf-candidate"},
			},
			i4RESTValidBody(),
		)

		i3RequireDetail(t, response, http.StatusForbidden, "You do not have permission to perform this action.")
		require.Zero(t, store.tokenLookups)
		require.Zero(t, store.transactions)
		require.Zero(t, store.updates)
		i3RequireSetCookieCount(t, response, 0)
	})

	for _, state := range []struct {
		name      string
		cookies   []*http.Cookie
		configure func(*i4RESTStore)
	}{
		{name: "no session"},
		{name: "one empty session", cookies: []*http.Cookie{i3Cookie(sessionCookie, "")}},
		{name: "unknown session", cookies: []*http.Cookie{i3Cookie(sessionCookie, i4RESTSession)}, configure: func(store *i4RESTStore) {
			delete(store.sessions, i4RESTKey(i4RESTDigestValue(i4RESTSession)))
		}},
		{name: "expired session", cookies: []*http.Cookie{i3Cookie(sessionCookie, i4RESTSession)}, configure: func(store *i4RESTStore) {
			record := store.sessions[i4RESTKey(i4RESTDigestValue(i4RESTSession))]
			record.Expires = now
			store.sessions[i4RESTKey(i4RESTDigestValue(i4RESTSession))] = record
		}},
		{name: "inactive session owner", cookies: []*http.Cookie{i3Cookie(sessionCookie, i4RESTSession)}, configure: func(store *i4RESTStore) {
			user := store.users[101]
			user.IsActive = false
			store.users[101] = user
		}},
	} {
		t.Run("token wins after permitted fallthrough/"+state.name, func(t *testing.T) {
			store := i4RESTFixture(t, now)
			if state.configure != nil {
				state.configure(store)
			}
			response := i4RESTPasswordRequest(store, now, false, state.cookies,
				map[string][]string{"Authorization": {"Token " + i4RESTToken}}, i4RESTValidBody())

			require.Equal(t, http.StatusNoContent, response.Code)
			require.Equal(t, []int64{202}, store.updatedUsers)
			require.Equal(t, []int64{202}, store.deletedUsers)
			require.Equal(t, 1, store.tokenLookups)
			require.Zero(t, store.creates)
			require.Zero(t, i4RESTSessionCount(store, 202))
			i4RESTRequirePassword(t, store.hashes[202], i4RESTNextPassword, true)
			i3RequireSetCookieCount(t, response, 0)
		})
	}

	t.Run("duplicate session cookies stop before token or application", func(t *testing.T) {
		store := i4RESTFixture(t, now)
		response := i4RESTPasswordRequest(store, now, false,
			[]*http.Cookie{
				i3Cookie(sessionCookie, i4RESTSession),
				i3Cookie(sessionCookie, i4RESTSiblingSession),
			},
			map[string][]string{"Authorization": {"Token " + i4RESTToken}}, i4RESTValidBody())

		i3RequireDetail(t, response, http.StatusUnauthorized, "Authentication credentials were not provided.")
		require.Zero(t, store.sessionLookups)
		require.Zero(t, store.tokenLookups)
		require.Zero(t, store.transactions)
		i3RequireSetCookieCount(t, response, 0)
	})

	t.Run("session infrastructure failure stops token fallback", func(t *testing.T) {
		store := i4RESTFixture(t, now)
		store.sessionLookupErr = errors.New("session backend unavailable")
		response := i4RESTPasswordRequest(store, now, false,
			[]*http.Cookie{i3Cookie(sessionCookie, i4RESTSession)},
			map[string][]string{"Authorization": {"Token " + i4RESTToken}}, i4RESTValidBody())

		i3RequireDetail(t, response, http.StatusInternalServerError, "An internal error occurred.")
		require.Zero(t, store.tokenLookups)
		require.Zero(t, store.transactions)
		i3RequireSetCookieCount(t, response, 0)
	})

	for _, contextCase := range []struct {
		name      string
		configure func(*gin.Context)
	}{
		{name: "missing provenance context"},
		{name: "invalid provenance context", configure: func(c *gin.Context) {
			c.Set(i4RESTPasswordChangeContextFixtureKey, struct{}{})
		}},
	} {
		t.Run(contextCase.name+" is an adapter failure", func(t *testing.T) {
			store := i4RESTFixture(t, now)
			handler := NewHandler(i4RESTService(store, now), false)
			router := gin.New()
			router.POST("/direct", func(c *gin.Context) {
				setPrincipal(c, store.users[101])
				if contextCase.configure != nil {
					contextCase.configure(c)
				}
				c.Next()
			}, handler.changePassword)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, i3RequestBody(http.MethodPost, "/direct", nil,
				map[string][]string{"Content-Type": {"application/json"}}, i4RESTValidBody()))

			i3RequireDetail(t, response, http.StatusInternalServerError, "An internal error occurred.")
			require.Zero(t, store.userLookups)
			require.Zero(t, store.transactions)
			require.Zero(t, store.updates)
			i3RequireSetCookieCount(t, response, 0)
		})
	}
}

func TestRESTPasswordChangeSessionCookieRotation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 8, 19, 9, 0, 0, 0, time.UTC)

	for _, secure := range []bool{false, true} {
		name := "development"
		if secure {
			name = "production"
		}
		t.Run(name+" browser rotation", func(t *testing.T) {
			store := i4RESTFixture(t, now)
			handler := NewHandler(i4RESTService(store, now), secure)
			router := gin.New()
			handler.Register(router)
			router.GET("/i4/session", handler.Middleware(), func(c *gin.Context) { c.Status(http.StatusNoContent) })
			router.POST("/i4/csrf", handler.Middleware(), func(c *gin.Context) { c.Status(http.StatusNoContent) })

			response := httptest.NewRecorder()
			router.ServeHTTP(response, i3RequestBody(http.MethodPost, "/api/auth/password/change/",
				[]*http.Cookie{
					i3Cookie(sessionCookie, i4RESTSession),
					i3Cookie(csrfCookie, i3DerivedCSRF(i4RESTSession)),
				},
				map[string][]string{
					"Content-Type": {"application/json"},
					"X-CSRFToken":  {i3DerivedCSRF(i4RESTSession)},
				}, i4RESTValidBody()))

			require.Equal(t, http.StatusNoContent, response.Code)
			require.Zero(t, response.Body.Len())
			i3RequireSetCookieCount(t, response, 2)
			session := i3RequireCookie(t, response, sessionCookie)
			csrf := i3RequireCookie(t, response, csrfCookie)
			require.NotEmpty(t, session.Value)
			require.NotEmpty(t, csrf.Value)
			require.False(t, i3OpaqueEqual(session.Value, i4RESTSession), "replacement session reused prior material")
			require.False(t, i3OpaqueEqual(csrf.Value, i3DerivedCSRF(i4RESTSession)), "replacement csrf reused prior material")
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
			require.True(t, session.Expires.Equal(now.Add(application.BrowserSessionLifetime)), "session expiry did not use the application result")
			require.True(t, csrf.Expires.IsZero(), "csrf cookie unexpectedly carried an expiry")
			require.Equal(t, 1, i4RESTSessionCount(store, 101))
			require.False(t, i4RESTHasSession(store, i4RESTSession))
			require.False(t, i4RESTHasSession(store, i4RESTSiblingSession))
			require.True(t, i4RESTHasSession(store, session.Value), "replacement session was not persisted")
			require.True(t, i4RESTSessionCSRFMatches(store, session.Value, csrf.Value), "replacement csrf did not match persisted state")

			usable := httptest.NewRecorder()
			router.ServeHTTP(usable, i3Request(http.MethodPost, "/i4/csrf",
				[]*http.Cookie{i3Cookie(sessionCookie, session.Value), i3Cookie(csrfCookie, csrf.Value)},
				map[string][]string{"X-CSRFToken": {csrf.Value}}))
			require.Equal(t, http.StatusNoContent, usable.Code)

			for _, old := range []string{i4RESTSession, i4RESTSiblingSession} {
				rejected := httptest.NewRecorder()
				router.ServeHTTP(rejected, i3Request(http.MethodGet, "/i4/session",
					[]*http.Cookie{i3Cookie(sessionCookie, old)}, nil))
				require.Equal(t, http.StatusUnauthorized, rejected.Code)
			}
		})
	}

	for _, ambient := range []struct {
		name      string
		cookies   []*http.Cookie
		configure func(*i4RESTStore)
	}{
		{name: "without ambient session"},
		{name: "with unknown ambient session", cookies: []*http.Cookie{i3Cookie(sessionCookie, "unknown-ambient-session")}},
		{name: "with expired ambient session", cookies: []*http.Cookie{i3Cookie(sessionCookie, i4RESTSession)}, configure: func(store *i4RESTStore) {
			record := store.sessions[i4RESTKey(i4RESTDigestValue(i4RESTSession))]
			record.Expires = now
			store.sessions[i4RESTKey(i4RESTDigestValue(i4RESTSession))] = record
		}},
		{name: "with inactive-owner ambient session", cookies: []*http.Cookie{i3Cookie(sessionCookie, i4RESTSession)}, configure: func(store *i4RESTStore) {
			user := store.users[101]
			user.IsActive = false
			store.users[101] = user
		}},
	} {
		t.Run("token success emits no cookies/"+ambient.name, func(t *testing.T) {
			store := i4RESTFixture(t, now)
			if ambient.configure != nil {
				ambient.configure(store)
			}
			response := i4RESTPasswordRequest(store, now, false, ambient.cookies,
				map[string][]string{"Authorization": {"Token " + i4RESTToken}}, i4RESTValidBody())

			require.Equal(t, http.StatusNoContent, response.Code)
			require.Zero(t, response.Body.Len())
			i3RequireSetCookieCount(t, response, 0)
			require.Zero(t, i4RESTSessionCount(store, 202))
			require.Zero(t, store.creates)
			require.Equal(t, 1, store.tokenLookups)
			require.Equal(t, int64(303), store.tokenRecord.Token.ID, "api token state changed")
		})
	}
}

func TestRESTPasswordChangeFailureAndNoCookieMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	backendFailure := errors.New("password-change backend unavailable")

	tests := []struct {
		name       string
		cookies    func() []*http.Cookie
		headers    func() map[string][]string
		body       []byte
		configure  func(*i4RESTStore)
		entropy    *bytes.Reader
		wantStatus int
		wantDetail string
		wantField  string
		wantValue  string
	}{
		{
			name: "malformed json", headers: i4RESTTokenHeaders, body: []byte("{"),
			wantStatus: http.StatusBadRequest, wantField: "non_field_errors", wantValue: "Expected current_password and new_password.",
		},
		{
			name: "current password validation", headers: i4RESTTokenHeaders,
			body:       []byte(`{"current_password":"incorrect-current-password","new_password":"transport-replacement-password"}`),
			wantStatus: http.StatusBadRequest, wantField: "current_password", wantValue: "Current password is incorrect.",
		},
		{
			name: "new password validation", headers: i4RESTTokenHeaders,
			body:       []byte(`{"current_password":"transport-current-password","new_password":"short"}`),
			wantStatus: http.StatusBadRequest, wantField: "new_password", wantValue: "Password must contain at least 12 characters.",
		},
		{
			name: "missing credential", body: i4RESTValidBody(),
			wantStatus: http.StatusUnauthorized, wantDetail: "Authentication credentials were not provided.",
		},
		{
			name: "middleware csrf refusal", cookies: i4RESTBrowserCookies, headers: func() map[string][]string {
				return map[string][]string{"Authorization": {"Token " + i4RESTToken}, "X-CSRFToken": {"different-csrf-candidate"}}
			}, body: i4RESTValidBody(), wantStatus: http.StatusForbidden, wantDetail: "You do not have permission to perform this action.",
		},
		{
			name: "session disappears in transaction", cookies: i4RESTBrowserCookies, headers: i4RESTBrowserHeaders,
			body: i4RESTValidBody(), configure: func(store *i4RESTStore) {
				store.beforeTransaction = func(store *i4RESTStore) {
					delete(store.sessions, i4RESTKey(i4RESTDigestValue(i4RESTSession)))
				}
			}, wantStatus: http.StatusUnauthorized, wantDetail: "Authentication credentials were not provided.",
		},
		{
			name: "session expires in transaction", cookies: i4RESTBrowserCookies, headers: i4RESTBrowserHeaders,
			body: i4RESTValidBody(), configure: func(store *i4RESTStore) {
				store.beforeTransaction = func(store *i4RESTStore) {
					record := store.sessions[i4RESTKey(i4RESTDigestValue(i4RESTSession))]
					record.Expires = now
					store.sessions[i4RESTKey(i4RESTDigestValue(i4RESTSession))] = record
				}
			}, wantStatus: http.StatusUnauthorized, wantDetail: "Authentication credentials were not provided.",
		},
		{
			name: "session owner changes in transaction", cookies: i4RESTBrowserCookies, headers: i4RESTBrowserHeaders,
			body: i4RESTValidBody(), configure: func(store *i4RESTStore) {
				store.beforeTransaction = func(store *i4RESTStore) {
					record := store.sessions[i4RESTKey(i4RESTDigestValue(i4RESTSession))]
					record.UserID = 202
					store.sessions[i4RESTKey(i4RESTDigestValue(i4RESTSession))] = record
				}
			}, wantStatus: http.StatusUnauthorized, wantDetail: "Authentication credentials were not provided.",
		},
		{
			name: "session owner becomes inactive in transaction", cookies: i4RESTBrowserCookies, headers: i4RESTBrowserHeaders,
			body: i4RESTValidBody(), configure: func(store *i4RESTStore) {
				store.beforeTransaction = func(store *i4RESTStore) {
					user := store.users[101]
					user.IsActive = false
					store.users[101] = user
				}
			}, wantStatus: http.StatusUnauthorized, wantDetail: "Authentication credentials were not provided.",
		},
		{
			name: "stored csrf changes in transaction", cookies: i4RESTBrowserCookies, headers: i4RESTBrowserHeaders,
			body: i4RESTValidBody(), configure: func(store *i4RESTStore) {
				store.beforeTransaction = func(store *i4RESTStore) {
					record := store.sessions[i4RESTKey(i4RESTDigestValue(i4RESTSession))]
					record.CSRFHash = i4RESTDigestValue("different-csrf-candidate")
					store.sessions[i4RESTKey(i4RESTDigestValue(i4RESTSession))] = record
				}
			}, wantStatus: http.StatusForbidden, wantDetail: "You do not have permission to perform this action.",
		},
		{
			name: "session lookup fails in transaction", cookies: i4RESTBrowserCookies, headers: i4RESTBrowserHeaders,
			body: i4RESTValidBody(), configure: func(store *i4RESTStore) {
				store.beforeTransaction = func(store *i4RESTStore) {
					store.sessionLookupErr = backendFailure
				}
			}, wantStatus: http.StatusInternalServerError, wantDetail: "An internal error occurred.",
		},
		{
			name: "transaction start failure", headers: i4RESTTokenHeaders, body: i4RESTValidBody(),
			configure:  func(store *i4RESTStore) { store.transactionErr = backendFailure },
			wantStatus: http.StatusInternalServerError, wantDetail: "An internal error occurred.",
		},
		{
			name: "replacement entropy failure", cookies: i4RESTBrowserCookies, headers: i4RESTBrowserHeaders,
			body: i4RESTValidBody(), entropy: bytes.NewReader([]byte("short")),
			wantStatus: http.StatusInternalServerError, wantDetail: "An internal error occurred.",
		},
		{
			name: "password update failure", headers: i4RESTTokenHeaders, body: i4RESTValidBody(),
			configure:  func(store *i4RESTStore) { store.updateErr = backendFailure },
			wantStatus: http.StatusInternalServerError, wantDetail: "An internal error occurred.",
		},
		{
			name: "session delete failure", headers: i4RESTTokenHeaders, body: i4RESTValidBody(),
			configure:  func(store *i4RESTStore) { store.deleteErr = backendFailure },
			wantStatus: http.StatusInternalServerError, wantDetail: "An internal error occurred.",
		},
		{
			name: "replacement insert failure", cookies: i4RESTBrowserCookies, headers: i4RESTBrowserHeaders,
			body: i4RESTValidBody(), configure: func(store *i4RESTStore) { store.createErr = backendFailure },
			wantStatus: http.StatusInternalServerError, wantDetail: "An internal error occurred.",
		},
		{
			name: "commit failure after dml", cookies: i4RESTBrowserCookies, headers: i4RESTBrowserHeaders,
			body: i4RESTValidBody(), configure: func(store *i4RESTStore) { store.commitErr = backendFailure },
			wantStatus: http.StatusInternalServerError, wantDetail: "An internal error occurred.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := i4RESTFixture(t, now)
			if test.configure != nil {
				test.configure(store)
			}
			var cookies []*http.Cookie
			if test.cookies != nil {
				cookies = test.cookies()
			}
			var headers map[string][]string
			if test.headers != nil {
				headers = test.headers()
			}
			response := i4RESTPasswordRequestWithEntropy(store, now, false, cookies, headers, test.body, test.entropy)

			if test.wantField != "" {
				i4RESTRequireFieldError(t, response, test.wantStatus, test.wantField, test.wantValue)
			} else {
				i3RequireDetail(t, response, test.wantStatus, test.wantDetail)
			}
			i3RequireSetCookieCount(t, response, 0)
			i4RESTRequireBodyConcealsCredentials(t, response.Body.Bytes())
		})
	}
}

func i4RESTFixture(t *testing.T, now time.Time) *i4RESTStore {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(i4RESTCurrentPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatal("could not prepare password-change fixture")
	}
	lastUsed := now
	store := &i4RESTStore{
		users: map[int64]domain.User{
			101: {ID: 101, Username: "browser-user", IsActive: true},
			202: {ID: 202, Username: "token-user", IsActive: true},
		},
		hashes:   map[int64]string{101: string(hash), 202: string(hash)},
		sessions: make(map[i4RESTDigest]application.SessionRecord),
		tokenRecord: application.TokenRecord{Token: domain.APIToken{
			ID: 303, UserID: 202, WriteEnabled: true, LastUsed: &lastUsed,
		}},
		tokenUser: domain.User{ID: 202, Username: "token-user", IsActive: true},
	}
	for _, session := range []struct {
		secret string
		userID int64
	}{
		{secret: i4RESTSession, userID: 101},
		{secret: i4RESTSiblingSession, userID: 101},
		{secret: i4RESTTokenSession, userID: 202},
	} {
		store.sessions[i4RESTKey(i4RESTDigestValue(session.secret))] = application.SessionRecord{
			SecretHash: i4RESTDigestValue(session.secret),
			CSRFHash:   i4RESTDigestValue(i3DerivedCSRF(session.secret)),
			UserID:     session.userID,
			Created:    now.Add(-time.Hour),
			LastSeen:   now.Add(-time.Hour),
			Expires:    now.Add(time.Hour),
		}
	}
	return store
}

func i4RESTService(store *i4RESTStore, now time.Time) *application.Service {
	return application.NewService(
		store,
		i4RESTClock{now: now},
		application.WithPasswordChangeEntropy(bytes.NewReader(bytes.Repeat([]byte{0x4a}, 32))),
	)
}

func i4RESTPasswordRequest(
	store *i4RESTStore,
	now time.Time,
	secure bool,
	cookies []*http.Cookie,
	headers map[string][]string,
	body []byte,
) *httptest.ResponseRecorder {
	return i4RESTPasswordRequestWithEntropy(store, now, secure, cookies, headers, body, nil)
}

func i4RESTPasswordRequestWithEntropy(
	store *i4RESTStore,
	now time.Time,
	secure bool,
	cookies []*http.Cookie,
	headers map[string][]string,
	body []byte,
	entropy *bytes.Reader,
) *httptest.ResponseRecorder {
	if headers == nil {
		headers = make(map[string][]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = []string{"application/json"}
	}
	if entropy == nil {
		entropy = bytes.NewReader(bytes.Repeat([]byte{0x4a}, 32))
	}
	service := application.NewService(store, i4RESTClock{now: now}, application.WithPasswordChangeEntropy(entropy))
	handler := NewHandler(service, secure)
	router := gin.New()
	handler.Register(router)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, i3RequestBody(http.MethodPost, "/api/auth/password/change/", cookies, headers, body))
	return response
}

func i4RESTValidBody() []byte {
	return []byte(`{"current_password":"transport-current-password","new_password":"transport-replacement-password"}`)
}

func i4RESTTokenHeaders() map[string][]string {
	return map[string][]string{"Authorization": {"Token " + i4RESTToken}}
}

func i4RESTBrowserCookies() []*http.Cookie {
	return []*http.Cookie{
		i3Cookie(sessionCookie, i4RESTSession),
		i3Cookie(csrfCookie, i3DerivedCSRF(i4RESTSession)),
	}
}

func i4RESTBrowserHeaders() map[string][]string {
	return map[string][]string{"X-CSRFToken": {i3DerivedCSRF(i4RESTSession)}}
}

func i4RESTDigestValue(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return append([]byte(nil), sum[:]...)
}

func i4RESTKey(hash []byte) i4RESTDigest {
	var key i4RESTDigest
	copy(key[:], hash)
	return key
}

func i4RESTCloneSession(record application.SessionRecord) application.SessionRecord {
	record.SecretHash = append([]byte(nil), record.SecretHash...)
	record.CSRFHash = append([]byte(nil), record.CSRFHash...)
	return record
}

func i4RESTCloneSessions(source map[i4RESTDigest]application.SessionRecord) map[i4RESTDigest]application.SessionRecord {
	result := make(map[i4RESTDigest]application.SessionRecord, len(source))
	for key, record := range source {
		result[key] = i4RESTCloneSession(record)
	}
	return result
}

func i4RESTCloneUsers(source map[int64]domain.User) map[int64]domain.User {
	result := make(map[int64]domain.User, len(source))
	for id, user := range source {
		result[id] = user
	}
	return result
}

func i4RESTCloneHashes(source map[int64]string) map[int64]string {
	result := make(map[int64]string, len(source))
	for id, hash := range source {
		result[id] = hash
	}
	return result
}

func i4RESTSessionCount(store *i4RESTStore, userID int64) int {
	count := 0
	for _, record := range store.sessions {
		if record.UserID == userID {
			count++
		}
	}
	return count
}

func i4RESTHasSession(store *i4RESTStore, secret string) bool {
	_, ok := store.sessions[i4RESTKey(i4RESTDigestValue(secret))]
	return ok
}

func i4RESTSessionCSRFMatches(store *i4RESTStore, secret, csrf string) bool {
	record, ok := store.sessions[i4RESTKey(i4RESTDigestValue(secret))]
	return ok && subtle.ConstantTimeCompare(record.CSRFHash, i4RESTDigestValue(csrf)) == 1
}

func i4RESTRequirePassword(t *testing.T, hash, password string, want bool) {
	t.Helper()
	got := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
	if got != want {
		t.Error("stored password state did not match the expected outcome")
	}
}

func i4RESTRequireFieldError(t *testing.T, response *httptest.ResponseRecorder, status int, field, description string) {
	t.Helper()
	require.Equal(t, status, response.Code)
	var body map[string][]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, map[string][]string{field: {description}}, body)
}

func i4RESTRequireBodyConcealsCredentials(t *testing.T, body []byte) {
	t.Helper()
	for _, value := range []string{
		i4RESTCurrentPassword,
		i4RESTNextPassword,
		i4RESTSession,
		i4RESTSiblingSession,
		i4RESTTokenSession,
		i4RESTToken,
	} {
		if bytes.Contains(body, []byte(value)) {
			t.Fatal("password-change response exposed reusable credential material")
		}
	}
}

func i4RESTRequireProvenanceGuardBeforeService(t *testing.T) {
	t.Helper()
	files := token.NewFileSet()
	parsed, err := parser.ParseFile(files, "http.go", nil, 0)
	if err != nil {
		t.Fatal("could not parse the password-change REST adapter")
	}

	var handler *ast.FuncDecl
	for _, declaration := range parsed.Decls {
		candidate, ok := declaration.(*ast.FuncDecl)
		if ok && candidate.Name.Name == "changePassword" && candidate.Recv != nil {
			handler = candidate
			break
		}
	}
	if handler == nil || handler.Body == nil {
		t.Fatal("password-change REST handler was not found")
	}

	var serviceCall token.Pos
	ast.Inspect(handler.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		method, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || method.Sel.Name != "ChangePassword" {
			return true
		}
		receiver, ok := method.X.(*ast.SelectorExpr)
		if ok && receiver.Sel.Name == "service" {
			serviceCall = call.Pos()
		}
		return true
	})
	if serviceCall == token.NoPos {
		t.Fatal("password-change application call was not found")
	}

	constants := i4RESTStringConstants(parsed)
	var contextGet token.Pos
	var contextValue, existenceFlag string
	var typeAssertion token.Pos
	var acceptedFlag string
	var contextKey string

	ast.Inspect(handler.Body, func(node ast.Node) bool {
		assignment, ok := node.(*ast.AssignStmt)
		if !ok || assignment.Pos() >= serviceCall || len(assignment.Rhs) != 1 {
			return true
		}

		if call, ok := assignment.Rhs[0].(*ast.CallExpr); ok {
			method, selectorOK := call.Fun.(*ast.SelectorExpr)
			context, contextOK := method.X.(*ast.Ident)
			if selectorOK && contextOK && context.Name == "c" && method.Sel.Name == "Get" && len(call.Args) == 1 && len(assignment.Lhs) == 2 {
				contextGet = assignment.Pos()
				contextValue = i4RESTIdentifierName(assignment.Lhs[0])
				existenceFlag = i4RESTIdentifierName(assignment.Lhs[1])
				contextKey = i4RESTStringExpression(call.Args[0], constants)
			}
		}

		assertion, ok := assignment.Rhs[0].(*ast.TypeAssertExpr)
		if !ok || len(assignment.Lhs) != 2 {
			return true
		}
		value, ok := assertion.X.(*ast.Ident)
		if ok && contextValue != "" && value.Name == contextValue {
			typeAssertion = assignment.Pos()
			acceptedFlag = i4RESTIdentifierName(assignment.Lhs[1])
		}
		return true
	})

	if contextGet == token.NoPos || contextValue == "" || existenceFlag == "" {
		t.Fatal("password-change handler does not extract private provenance before calling the application")
	}
	if contextKey != i4RESTPasswordChangeContextFixtureKey {
		t.Fatal("password-change handler and invalid-context fixture do not share the private provenance key")
	}
	if typeAssertion == token.NoPos || acceptedFlag == "" || typeAssertion <= contextGet {
		t.Fatal("password-change handler does not validate the private provenance type before calling the application")
	}

	guarded := map[string]bool{existenceFlag: false, acceptedFlag: false}
	ast.Inspect(handler.Body, func(node ast.Node) bool {
		guard, ok := node.(*ast.IfStmt)
		if !ok || guard.Pos() >= serviceCall || !i4RESTContainsReturn(guard.Body) {
			return true
		}
		for flag := range guarded {
			if i4RESTExpressionRejectsFalseFlag(guard.Cond, flag) {
				guarded[flag] = true
			}
		}
		return true
	})
	if !guarded[existenceFlag] || !guarded[acceptedFlag] {
		t.Fatal("password-change handler can reach the application without guarding missing and invalid provenance")
	}
}

func i4RESTStringConstants(file *ast.File) map[string]string {
	constants := make(map[string]string)
	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok || group.Tok != token.CONST {
			continue
		}
		for _, specification := range group.Specs {
			values, ok := specification.(*ast.ValueSpec)
			if !ok || len(values.Names) != len(values.Values) {
				continue
			}
			for index, name := range values.Names {
				if value := i4RESTStringExpression(values.Values[index], constants); value != "" {
					constants[name.Name] = value
				}
			}
		}
	}
	return constants
}

func i4RESTStringExpression(expression ast.Expr, constants map[string]string) string {
	switch value := expression.(type) {
	case *ast.BasicLit:
		if value.Kind != token.STRING {
			return ""
		}
		decoded, err := strconv.Unquote(value.Value)
		if err != nil {
			return ""
		}
		return decoded
	case *ast.Ident:
		return constants[value.Name]
	default:
		return ""
	}
}

func i4RESTIdentifierName(expression ast.Expr) string {
	identifier, ok := expression.(*ast.Ident)
	if !ok {
		return ""
	}
	return identifier.Name
}

func i4RESTContainsReturn(node ast.Node) bool {
	found := false
	ast.Inspect(node, func(candidate ast.Node) bool {
		if _, ok := candidate.(*ast.ReturnStmt); ok {
			found = true
			return false
		}
		return !found
	})
	return found
}

func i4RESTExpressionRejectsFalseFlag(expression ast.Expr, name string) bool {
	found := false
	ast.Inspect(expression, func(node ast.Node) bool {
		negation, ok := node.(*ast.UnaryExpr)
		if !ok || negation.Op != token.NOT {
			return !found
		}
		identifier, ok := negation.X.(*ast.Ident)
		if ok && identifier.Name == name {
			found = true
			return false
		}
		return !found
	})
	return found
}
