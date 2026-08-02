package identity

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	postgres "netbox-go/internal/adapters/postgres/identity"
	application "netbox-go/internal/application/identity"
)

func TestBrowserSessionCSRFAndTokenSecretLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:identity_http?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(postgres.Models()...))
	service := application.NewService(postgres.NewStore(db), application.RealClock{})
	_, err = service.BootstrapAdministrator(t.Context(), "admin", "", "Correct-Horse-2026!")
	require.NoError(t, err)
	router := gin.New()
	NewHandler(service, false).Register(router)

	unauthorized := perform(router, http.MethodGet, "/api/auth/session/", nil, nil)
	require.Equal(t, http.StatusUnauthorized, unauthorized.Code)
	csrf := perform(router, http.MethodGet, "/api/auth/csrf/", nil, nil)
	require.Equal(t, http.StatusOK, csrf.Code)
	csrfValue := cookie(csrf.Result(), csrfCookie)
	require.NotNil(t, csrfValue)
	loginHeaders := map[string]string{"Content-Type": "application/json", "X-CSRFToken": csrfValue.Value, "Cookie": csrfValue.String()}
	login := perform(router, http.MethodPost, "/api/auth/login/", strings.NewReader(`{"username":"admin","password":"Correct-Horse-2026!"}`), loginHeaders)
	require.Equal(t, http.StatusOK, login.Code)
	session := cookie(login.Result(), sessionCookie)
	csrfValue = cookie(login.Result(), csrfCookie)
	require.NotNil(t, session)
	require.True(t, session.HttpOnly)
	require.NotNil(t, csrfValue)
	authCookies := session.String() + "; " + csrfValue.String()

	badCSRF := perform(router, http.MethodPost, "/api/auth/tokens/", strings.NewReader(`{"description":"bad"}`), map[string]string{"Content-Type": "application/json", "Cookie": authCookies})
	require.Equal(t, http.StatusForbidden, badCSRF.Code)
	created := perform(router, http.MethodPost, "/api/auth/tokens/", strings.NewReader(`{"description":"automation","write_enabled":true}`), map[string]string{"Content-Type": "application/json", "Cookie": authCookies, "X-CSRFToken": csrfValue.Value})
	require.Equal(t, http.StatusCreated, created.Code)
	var createdBody map[string]any
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), &createdBody))
	require.NotEmpty(t, createdBody["secret"])
	listed := perform(router, http.MethodGet, "/api/auth/tokens/", nil, map[string]string{"Cookie": authCookies})
	require.Equal(t, http.StatusOK, listed.Code)
	require.NotContains(t, listed.Body.String(), createdBody["secret"].(string))
}

func TestLoginThrottleAndProductionCookieFlags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:identity_http_security?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(postgres.Models()...))
	service := application.NewService(postgres.NewStore(db), application.RealClock{})
	_, err = service.BootstrapAdministrator(t.Context(), "admin", "", "Correct-Horse-2026!")
	require.NoError(t, err)
	router := gin.New()
	NewHandler(service, true).Register(router)

	csrf := perform(router, http.MethodGet, "/api/auth/csrf/", nil, nil)
	csrfValue := cookie(csrf.Result(), csrfCookie)
	require.NotNil(t, csrfValue)
	require.True(t, csrfValue.Secure)
	require.Equal(t, http.SameSiteLaxMode, csrfValue.SameSite)
	headers := map[string]string{
		"Content-Type": "application/json", "X-CSRFToken": csrfValue.Value,
		"Cookie": csrfValue.String(), "Origin": "https://untrusted.invalid",
	}
	successful := perform(router, http.MethodPost, "/api/auth/login/", strings.NewReader(`{"username":"admin","password":"Correct-Horse-2026!"}`), headers)
	require.Equal(t, http.StatusOK, successful.Code)
	sessionValue := cookie(successful.Result(), sessionCookie)
	require.NotNil(t, sessionValue)
	require.True(t, sessionValue.Secure)
	require.True(t, sessionValue.HttpOnly)
	require.Equal(t, http.SameSiteLaxMode, sessionValue.SameSite)
	require.Empty(t, successful.Header().Get("Access-Control-Allow-Origin"))
	for attempt := 0; attempt < 5; attempt++ {
		response := perform(router, http.MethodPost, "/api/auth/login/", strings.NewReader(`{"username":"admin","password":"wrong-password"}`), headers)
		require.Equal(t, http.StatusUnauthorized, response.Code)
		require.Empty(t, response.Header().Get("Access-Control-Allow-Origin"))
	}
	limited := perform(router, http.MethodPost, "/api/auth/login/", strings.NewReader(`{"username":"admin","password":"Correct-Horse-2026!"}`), headers)
	require.Equal(t, http.StatusTooManyRequests, limited.Code)
}

func TestLoginRotatesPresentedSessionAndNeverAcceptsFixedMaterial(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:identity_http_rotation?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(postgres.Models()...))
	service := application.NewService(postgres.NewStore(db), application.RealClock{})
	_, err = service.BootstrapAdministrator(t.Context(), "admin", "", "Correct-Horse-2026!")
	require.NoError(t, err)
	router := gin.New()
	NewHandler(service, false).Register(router)

	csrf := perform(router, http.MethodGet, "/api/auth/csrf/", nil, nil)
	csrfValue := cookie(csrf.Result(), csrfCookie)
	firstLogin := perform(router, http.MethodPost, "/api/auth/login/", strings.NewReader(`{"username":"admin","password":"Correct-Horse-2026!"}`), map[string]string{
		"Content-Type": "application/json", "X-CSRFToken": csrfValue.Value, "Cookie": csrfValue.String(),
	})
	require.Equal(t, http.StatusOK, firstLogin.Code)
	firstSession := cookie(firstLogin.Result(), sessionCookie)
	firstCSRF := cookie(firstLogin.Result(), csrfCookie)
	require.NotNil(t, firstSession)
	require.NotNil(t, firstCSRF)

	secondLogin := perform(router, http.MethodPost, "/api/auth/login/", strings.NewReader(`{"username":"admin","password":"Correct-Horse-2026!"}`), map[string]string{
		"Content-Type": "application/json", "X-CSRFToken": firstCSRF.Value,
		"Cookie": firstSession.String() + "; " + firstCSRF.String(),
	})
	require.Equal(t, http.StatusOK, secondLogin.Code)
	secondSession := cookie(secondLogin.Result(), sessionCookie)
	require.NotNil(t, secondSession)
	require.NotEqual(t, firstSession.Value, secondSession.Value)
	require.Equal(t, http.StatusUnauthorized, perform(router, http.MethodGet, "/api/auth/session/", nil, map[string]string{"Cookie": firstSession.String()}).Code)
	require.Equal(t, http.StatusOK, perform(router, http.MethodGet, "/api/auth/session/", nil, map[string]string{"Cookie": secondSession.String()}).Code)

	// An attacker-selected cookie is deleted if it happens to identify a row,
	// and is never reused as the newly authenticated session key.
	freshCSRFResponse := perform(router, http.MethodGet, "/api/auth/csrf/", nil, nil)
	freshCSRF := cookie(freshCSRFResponse.Result(), csrfCookie)
	fixed := &http.Cookie{Name: sessionCookie, Value: "attacker-selected-session", Path: "/"}
	fixedLogin := perform(router, http.MethodPost, "/api/auth/login/", strings.NewReader(`{"username":"admin","password":"Correct-Horse-2026!"}`), map[string]string{
		"Content-Type": "application/json", "X-CSRFToken": freshCSRF.Value,
		"Cookie": fixed.String() + "; " + freshCSRF.String(),
	})
	require.Equal(t, http.StatusOK, fixedLogin.Code)
	require.NotEqual(t, fixed.Value, cookie(fixedLogin.Result(), sessionCookie).Value)
}

func TestAuthenticationAndCSRFSecurityMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:identity_http_matrix?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(postgres.Models()...))
	service := application.NewService(postgres.NewStore(db), application.RealClock{})
	admin, err := service.BootstrapAdministrator(t.Context(), "admin", "", "Correct-Horse-2026!")
	require.NoError(t, err)
	readOnly, err := service.CreateToken(t.Context(), admin.Principal(), application.CreateTokenInput{WriteEnabled: false})
	require.NoError(t, err)
	writeToken, err := service.CreateToken(t.Context(), admin.Principal(), application.CreateTokenInput{WriteEnabled: true})
	require.NoError(t, err)
	handler := NewHandler(service, false)
	router := gin.New()
	handler.Register(router)
	router.GET("/protected", handler.Middleware(), func(c *gin.Context) { c.Status(http.StatusOK) })
	router.POST("/protected", handler.Middleware(), func(c *gin.Context) { c.Status(http.StatusNoContent) })

	csrfResponse := perform(router, http.MethodGet, "/api/auth/csrf/", nil, nil)
	csrfValue := cookie(csrfResponse.Result(), csrfCookie)
	login := perform(router, http.MethodPost, "/api/auth/login/", strings.NewReader(`{"username":"admin","password":"Correct-Horse-2026!"}`), map[string]string{
		"Content-Type": "application/json", "X-CSRFToken": csrfValue.Value, "Cookie": csrfValue.String(),
	})
	session := cookie(login.Result(), sessionCookie)
	csrfValue = cookie(login.Result(), csrfCookie)
	cookies := session.String() + "; " + csrfValue.String()

	tests := []struct {
		name, method string
		headers      map[string]string
		want         int
	}{
		{name: "missing credential", method: http.MethodGet, want: http.StatusUnauthorized},
		{name: "unsupported authorization scheme", method: http.MethodGet, headers: map[string]string{"Authorization": "Bearer material"}, want: http.StatusUnauthorized},
		{name: "cookie safe request needs no csrf header", method: http.MethodGet, headers: map[string]string{"Cookie": cookies}, want: http.StatusOK},
		{name: "cookie unsafe request missing csrf", method: http.MethodPost, headers: map[string]string{"Cookie": cookies}, want: http.StatusForbidden},
		{name: "cookie unsafe request wrong csrf", method: http.MethodPost, headers: map[string]string{"Cookie": cookies, "X-CSRFToken": "wrong"}, want: http.StatusForbidden},
		{name: "cookie unsafe request matching csrf", method: http.MethodPost, headers: map[string]string{"Cookie": cookies, "X-CSRFToken": csrfValue.Value}, want: http.StatusNoContent},
		{name: "read only token safe request", method: http.MethodGet, headers: map[string]string{"Authorization": "Token " + readOnly.Secret}, want: http.StatusOK},
		{name: "read only token unsafe request", method: http.MethodPost, headers: map[string]string{"Authorization": "Token " + readOnly.Secret}, want: http.StatusForbidden},
		{name: "write token does not require csrf", method: http.MethodPost, headers: map[string]string{"Authorization": "Token " + writeToken.Secret}, want: http.StatusNoContent},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.headers == nil {
				test.headers = map[string]string{}
			}
			test.headers["Origin"] = "https://untrusted.invalid"
			response := perform(router, test.method, "/protected", nil, test.headers)
			require.Equal(t, test.want, response.Code)
			require.Empty(t, response.Header().Get("Access-Control-Allow-Origin"), "no origin is implicitly trusted")
			require.Empty(t, response.Header().Get("Access-Control-Allow-Credentials"))
		})
	}
}

func perform(router http.Handler, method, path string, body *strings.Reader, headers map[string]string) *httptest.ResponseRecorder {
	var request *http.Request
	if body == nil {
		request = httptest.NewRequest(method, path, nil)
	} else {
		request = httptest.NewRequest(method, path, body)
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
func cookie(response *http.Response, name string) *http.Cookie {
	for _, candidate := range response.Cookies() {
		if candidate.Name == name {
			return candidate
		}
	}
	return nil
}
