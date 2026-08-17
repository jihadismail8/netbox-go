package identity

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
)

type baselineTokenStore struct {
	application.Store
	record         application.TokenRecord
	user           domain.User
	lookupErr      error
	touchErr       error
	lookupCalls    int
	touchCalls     int
	expectedDigest []byte
	digestMatched  bool
}

func (store *baselineTokenStore) TokenByHash(_ context.Context, digest []byte) (application.TokenRecord, domain.User, error) {
	store.lookupCalls++
	if store.expectedDigest != nil {
		store.digestMatched = subtle.ConstantTimeCompare(store.expectedDigest, digest) == 1
	}
	return store.record, store.user, store.lookupErr
}

func (store *baselineTokenStore) TouchToken(context.Context, int64, time.Time) error {
	store.touchCalls++
	return store.touchErr
}

func TestBaselineTokenAuthorizationGrammar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 8, 17, 7, 30, 0, 0, time.UTC)
	invalidCredential := string([]byte{'T', 'o', 'k', 'e', 'n', ' ', 0xff})
	invalidCredentialWithExtraField := string([]byte{'T', 'o', 'k', 'e', 'n', ' ', 0xff, ' ', 'x'})

	const (
		missingDetail          = "Authentication credentials were not provided."
		missingTokenDetail     = "Invalid token header. No credentials provided."
		spacedTokenDetail      = "Invalid token header. Token string should not contain spaces."
		invalidCharacterDetail = "Invalid token header. Token string should not contain invalid characters."
		fixtureCredential      = "opaque-fixture"
	)

	tests := []struct {
		name           string
		headerValues   []string
		wantDetail     string
		wantCredential string
		wantAccepted   bool
	}{
		{name: "absent", wantDetail: missingDetail},
		{name: "empty", headerValues: []string{""}, wantDetail: missingDetail},
		{name: "ASCII whitespace only", headerValues: []string{" \t\v\f"}, wantDetail: missingDetail},
		{name: "unsupported bearer scheme", headerValues: []string{"Bearer " + fixtureCredential}, wantDetail: missingDetail},
		{name: "value without scheme", headerValues: []string{fixtureCredential}, wantDetail: missingDetail},
		{name: "duplicate field values", headerValues: []string{"Token " + fixtureCredential, "Token " + fixtureCredential}, wantDetail: missingDetail},
		{name: "scheme without credential", headerValues: []string{"Token"}, wantDetail: missingTokenDetail},
		{name: "scheme followed only by separators", headerValues: []string{"Token \t\v\f"}, wantDetail: missingTokenDetail},
		{name: "more than two fields", headerValues: []string{"Token " + fixtureCredential + " extra"}, wantDetail: spacedTokenDetail},
		{name: "field count precedes UTF-8 validation", headerValues: []string{invalidCredentialWithExtraField}, wantDetail: spacedTokenDetail},
		{name: "invalid UTF-8 credential", headerValues: []string{invalidCredential}, wantDetail: invalidCharacterDetail},
		{name: "non-ASCII delimiter is not a separator", headerValues: []string{"Token\u00a0" + fixtureCredential}, wantDetail: missingDetail},
		{name: "ASCII case-insensitive scheme", headerValues: []string{"tOkEn " + fixtureCredential}, wantCredential: fixtureCredential, wantAccepted: true},
		{name: "leading trailing and repeated separators", headerValues: []string{"\t  ToKeN  " + fixtureCredential + " \t"}, wantCredential: fixtureCredential, wantAccepted: true},
		{name: "vertical tab and form feed separators", headerValues: []string{"Token\v\f" + fixtureCredential}, wantCredential: fixtureCredential, wantAccepted: true},
		{name: "valid UTF-8 remains opaque", headerValues: []string{"Token " + fixtureCredential + "\u00a0"}, wantCredential: fixtureCredential + "\u00a0", wantAccepted: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &baselineTokenStore{
				record: application.TokenRecord{Token: baselineToken(now)},
				user:   baselineTokenUser(),
			}
			if test.wantAccepted {
				wantDigest := sha256.Sum256([]byte(test.wantCredential))
				store.expectedDigest = wantDigest[:]
			}

			response, handlerCalled := performBaselineTokenRequest(
				t,
				store,
				now,
				http.MethodGet,
				"192.0.2.1:443",
				test.headerValues,
			)

			require.Empty(t, response.Header().Get("WWW-Authenticate"))
			if test.wantAccepted {
				require.Equal(t, http.StatusNoContent, response.Code)
				require.True(t, handlerCalled)
				require.Equal(t, 1, store.lookupCalls)
				require.Zero(t, store.touchCalls)
				require.True(t, store.digestMatched, "application received altered opaque credential")
				return
			}

			require.Equal(t, http.StatusForbidden, response.Code)
			require.False(t, handlerCalled)
			require.Zero(t, store.lookupCalls)
			require.Zero(t, store.touchCalls)
			requireTokenDetail(t, response, test.wantDetail)
		})
	}

	t.Run("baseline grammar remains isolated from identity extension", func(t *testing.T) {
		headerValue := "tOkEn   " + fixtureCredential
		wantDigest := sha256.Sum256([]byte(fixtureCredential))
		baselineStore := &baselineTokenStore{
			record:         application.TokenRecord{Token: baselineToken(now)},
			user:           baselineTokenUser(),
			expectedDigest: wantDigest[:],
		}

		baselineResponse, baselineHandlerCalled := performBaselineTokenRequest(
			t,
			baselineStore,
			now,
			http.MethodGet,
			"192.0.2.1:443",
			[]string{headerValue},
		)

		require.Equal(t, http.StatusNoContent, baselineResponse.Code)
		require.True(t, baselineHandlerCalled)
		require.Equal(t, 1, baselineStore.lookupCalls)
		require.Zero(t, baselineStore.touchCalls)
		require.True(t, baselineStore.digestMatched, "baseline middleware altered opaque credential")

		extensionStore := &baselineTokenStore{
			record: application.TokenRecord{Token: baselineToken(now)},
			user:   baselineTokenUser(),
		}
		service := application.NewService(extensionStore, restCredentialClock{now: now})
		handler := NewHandler(service, false)
		extensionHandlerCalled := false
		router := gin.New()
		router.GET("/protected", handler.Middleware(), func(c *gin.Context) {
			extensionHandlerCalled = true
			c.Status(http.StatusNoContent)
		})
		request := httptest.NewRequest(http.MethodGet, "/protected", nil)
		request.RemoteAddr = "192.0.2.1:443"
		request.Header.Add("Authorization", headerValue)
		extensionResponse := httptest.NewRecorder()

		router.ServeHTTP(extensionResponse, request)

		require.Equal(t, http.StatusUnauthorized, extensionResponse.Code)
		require.False(t, extensionHandlerCalled)
		require.Zero(t, extensionStore.lookupCalls)
		require.Zero(t, extensionStore.touchCalls)
		require.Empty(t, extensionResponse.Header().Get("WWW-Authenticate"))
		requireTokenDetail(t, extensionResponse, missingDetail)
	})
}

func TestBaselineTokenOutcomeRendering(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 8, 17, 7, 45, 0, 0, time.UTC)
	stale := now.Add(-2 * time.Minute)
	expired := now
	revoked := now.Add(-time.Hour)
	lookupFailure := errors.New("credential lookup unavailable")
	touchFailure := errors.New("credential touch unavailable")
	activeUser := baselineTokenUser()

	const (
		invalidTokenDetail      = "Invalid token"
		expiredTokenDetail      = "Token expired"
		inactiveUserDetail      = "User inactive"
		sourceUnavailableDetail = "Client IP address could not be determined for validation. Check that the HTTP server is correctly configured to pass the required header(s)."
		sourceDeniedDetail      = "Source IP 198.51.100.1 is not permitted to authenticate using this token."
		forbiddenDetail         = "You do not have permission to perform this action."
		internalDetail          = "An internal error occurred."
	)

	tests := []struct {
		name        string
		method      string
		remote      string
		record      application.TokenRecord
		user        domain.User
		lookupErr   error
		touchErr    error
		wantStatus  int
		wantDetail  string
		wantLookups int
		wantTouches int
		wantHandler bool
	}{
		{
			name: "unknown", method: http.MethodGet, remote: "192.0.2.1:443",
			lookupErr: application.ErrNotFound, wantStatus: http.StatusForbidden,
			wantDetail: invalidTokenDetail, wantLookups: 1,
		},
		{
			name: "revoked", method: http.MethodGet, remote: "192.0.2.1:443",
			record: application.TokenRecord{Token: baselineToken(stale), RevokedAt: &revoked}, user: activeUser,
			wantStatus: http.StatusForbidden, wantDetail: invalidTokenDetail, wantLookups: 1,
		},
		{
			name: "expired", method: http.MethodGet, remote: "192.0.2.1:443",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := baselineToken(stale)
				token.Expires = &expired
				return token
			}()},
			user: activeUser, wantStatus: http.StatusForbidden, wantDetail: expiredTokenDetail,
			wantLookups: 1, wantTouches: 1,
		},
		{
			name: "inactive owner", method: http.MethodGet, remote: "192.0.2.1:443",
			record:     application.TokenRecord{Token: baselineToken(stale)},
			user:       domain.User{ID: activeUser.ID, Username: "inactive"},
			wantStatus: http.StatusForbidden, wantDetail: inactiveUserDetail, wantLookups: 1, wantTouches: 1,
		},
		{
			name: "restricted source unavailable when absent", method: http.MethodGet, remote: "",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := baselineToken(stale)
				token.AllowedIPs = []string{"192.0.2.0/24"}
				return token
			}()},
			user: activeUser, wantStatus: http.StatusForbidden, wantDetail: sourceUnavailableDetail,
			wantLookups: 1, wantTouches: 1,
		},
		{
			name: "restricted source unavailable when malformed", method: http.MethodGet, remote: "not-an-address",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := baselineToken(stale)
				token.AllowedIPs = []string{"192.0.2.0/24"}
				return token
			}()},
			user: activeUser, wantStatus: http.StatusForbidden, wantDetail: sourceUnavailableDetail,
			wantLookups: 1, wantTouches: 1,
		},
		{
			name: "restricted source denied", method: http.MethodGet, remote: "198.51.100.1:443",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := baselineToken(stale)
				token.AllowedIPs = []string{"192.0.2.0/24"}
				return token
			}()},
			user: activeUser, wantStatus: http.StatusForbidden, wantDetail: sourceDeniedDetail,
			wantLookups: 1, wantTouches: 1,
		},
		{
			name: "write disabled", method: http.MethodPost, remote: "192.0.2.1:443",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := baselineToken(stale)
				token.WriteEnabled = false
				return token
			}()},
			user: activeUser, wantStatus: http.StatusForbidden, wantDetail: forbiddenDetail,
			wantLookups: 1, wantTouches: 1,
		},
		{
			name: "lookup infrastructure failure", method: http.MethodGet, remote: "192.0.2.1:443",
			lookupErr: lookupFailure, wantStatus: http.StatusInternalServerError,
			wantDetail: internalDetail, wantLookups: 1,
		},
		{
			name: "touch infrastructure failure", method: http.MethodGet, remote: "192.0.2.1:443",
			record: baselineTokenRecord(stale), user: activeUser, touchErr: touchFailure,
			wantStatus: http.StatusInternalServerError, wantDetail: internalDetail,
			wantLookups: 1, wantTouches: 1,
		},
		{
			name: "valid", method: http.MethodGet, remote: "192.0.2.1:443",
			record: baselineTokenRecord(stale), user: activeUser,
			wantStatus: http.StatusNoContent, wantLookups: 1, wantTouches: 1, wantHandler: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &baselineTokenStore{
				record:    test.record,
				user:      test.user,
				lookupErr: test.lookupErr,
				touchErr:  test.touchErr,
			}

			response, handlerCalled := performBaselineTokenRequest(
				t,
				store,
				now,
				test.method,
				test.remote,
				[]string{"Token opaque-fixture"},
			)

			require.Equal(t, test.wantStatus, response.Code)
			require.Equal(t, test.wantHandler, handlerCalled)
			require.Equal(t, test.wantLookups, store.lookupCalls)
			require.Equal(t, test.wantTouches, store.touchCalls)
			require.Empty(t, response.Header().Get("WWW-Authenticate"))
			if test.wantHandler {
				require.Empty(t, response.Body.String())
				return
			}
			requireTokenDetail(t, response, test.wantDetail)
		})
	}
}

func performBaselineTokenRequest(
	t *testing.T,
	store *baselineTokenStore,
	now time.Time,
	method string,
	remote string,
	authorizationValues []string,
) (*httptest.ResponseRecorder, bool) {
	t.Helper()
	service := application.NewService(store, restCredentialClock{now: now})
	handler := NewHandler(service, false)
	handlerCalled := false
	router := gin.New()
	router.Handle(method, "/protected", handler.BaselineMiddleware(), func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(method, "/protected", nil)
	request.RemoteAddr = remote
	for _, value := range authorizationValues {
		request.Header.Add("Authorization", value)
	}
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	return response, handlerCalled
}

func requireTokenDetail(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, map[string]string{"detail": want}, body)
}

func baselineTokenRecord(lastUsed time.Time) application.TokenRecord {
	return application.TokenRecord{Token: baselineToken(lastUsed)}
}

func baselineToken(lastUsed time.Time) domain.APIToken {
	return domain.APIToken{
		ID: 17, UserID: 41, WriteEnabled: true, LastUsed: &lastUsed,
	}
}

func baselineTokenUser() domain.User {
	return domain.User{ID: 41, Username: "rest-user", IsActive: true}
}
