package router

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	identitypostgres "netbox-go/internal/adapters/postgres/identity"
	identityhttp "netbox-go/internal/adapters/rest/netbox/identity"
	workflowhttp "netbox-go/internal/adapters/rest/netbox/workflow"
	identityapp "netbox-go/internal/application/identity"
	"netbox-go/internal/config"
	identitydomain "netbox-go/internal/domain/identity"
	"netbox-go/internal/platform/composition"
	"netbox-go/internal/platform/readiness"
)

func TestBaselineAndIdentityAuthenticationStatusesRemainDistinct(t *testing.T) {
	config.Set(&config.Config{})
	db, err := gorm.Open(sqlite.Open("file:runtime_auth_status?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	core := composition.NewCore(db)
	router := New(core.Identity, core.Sites, false, nil, alwaysReadyChecker(), runtimeWorkflowOptions(core)...)

	for _, test := range []struct {
		path string
		want int
	}{
		{path: "/api/dcim/sites/?limit=1", want: http.StatusForbidden},
		{path: "/api/auth/session/", want: http.StatusUnauthorized},
	} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
		if recorder.Code != test.want {
			t.Fatalf("GET %s = %d, want %d: %s", test.path, recorder.Code, test.want, recorder.Body.String())
		}
	}
}

func TestBaselineTokenHTTPMethodSafety(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(
		sqlite.Open("file:baseline_token_method_safety?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(identitypostgres.Models()...))

	core := composition.NewCore(db)
	administrator, err := core.Identity.BootstrapAdministrator(
		t.Context(),
		"method-safety-admin",
		"",
		"Method-Safety-Password-2026!",
	)
	require.NoError(t, err)

	token, err := core.Identity.CreateToken(
		t.Context(),
		administrator.Principal(),
		identityapp.CreateTokenInput{
			Description:  "HTTP method safety",
			WriteEnabled: false,
		},
	)
	require.NoError(t, err)

	methods := []struct {
		name   string
		method string
		safe   bool
	}{
		{name: "get", method: http.MethodGet, safe: true},
		{name: "head", method: http.MethodHead, safe: true},
		{name: "options", method: http.MethodOptions, safe: true},
		{name: "post", method: http.MethodPost},
		{name: "put", method: http.MethodPut},
		{name: "patch", method: http.MethodPatch},
		{name: "delete", method: http.MethodDelete},
		{name: "extension method", method: "PROPFIND"},
	}

	handler := identityhttp.NewHandler(core.Identity, false)
	engine := gin.New()
	handlerCalls := make(map[string]int, len(methods))
	principals := make(map[string]identitydomain.Principal, len(methods))
	recordHandler := func(method string) gin.HandlerFunc {
		return func(c *gin.Context) {
			handlerCalls[method]++
			principals[method], _ = identitydomain.PrincipalFromContext(c.Request.Context())
			c.Status(http.StatusNoContent)
		}
	}
	for _, methodCase := range methods {
		engine.Handle(
			methodCase.method,
			"/protected",
			handler.BaselineMiddleware(),
			recordHandler(methodCase.method),
		)
	}

	headerForms := []struct {
		name   string
		format func(string) string
	}{
		{
			name: "canonical",
			format: func(credential string) string {
				return "Token " + credential
			},
		},
		{
			name: "baseline casefold and repeated separator",
			format: func(credential string) string {
				return "tOkEn   " + credential
			},
		},
	}

	for _, headerForm := range headerForms {
		for _, methodCase := range methods {
			t.Run(headerForm.name+"/"+methodCase.name, func(t *testing.T) {
				before := handlerCalls[methodCase.method]
				request := httptest.NewRequest(methodCase.method, "/protected", nil)
				request.RemoteAddr = "192.0.2.10:443"
				request.Header.Set("Authorization", headerForm.format(token.Secret))
				response := httptest.NewRecorder()

				engine.ServeHTTP(response, request)

				require.Empty(t, response.Header().Get("WWW-Authenticate"))
				if methodCase.safe {
					require.Equal(t, http.StatusNoContent, response.Code)
					require.Equal(t, before+1, handlerCalls[methodCase.method])
					require.Equal(t, administrator.ID, principals[methodCase.method].ID)
					return
				}

				require.Equal(t, http.StatusForbidden, response.Code)
				require.JSONEq(
					t,
					`{"detail":"You do not have permission to perform this action."}`,
					response.Body.String(),
				)
				require.Equal(t, before, handlerCalls[methodCase.method])
			})
		}
	}
}

func TestRuntimePublicProbesTrackPostgreSQLReadiness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("NETBOX_WEB_ASSETS_PATH", filepath.Join(t.TempDir(), "missing"))
	config.Set(&config.Config{})
	db, err := gorm.Open(sqlite.Open("file:runtime_public_probes?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	core := composition.NewCore(db)
	dependencyCause := errors.New("database endpoint secret must not escape")
	checker := &scriptedReadinessChecker{results: []error{nil, dependencyCause, nil}}
	runtime := New(core.Identity, core.Sites, false, nil, checker, runtimeWorkflowOptions(core)...)

	routes := make(map[string]struct{}, len(runtime.Routes()))
	for _, route := range runtime.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}
	require.Contains(t, routes, http.MethodGet+" /health")
	require.Contains(t, routes, http.MethodGet+" /ready")
	require.NotContains(t, routes, http.MethodGet+" /ping")

	for _, test := range []struct {
		name       string
		readyCode  int
		readyState string
	}{
		{name: "available", readyCode: http.StatusOK, readyState: "UP"},
		{name: "lost", readyCode: http.StatusServiceUnavailable, readyState: "DOWN"},
		{name: "recovered", readyCode: http.StatusOK, readyState: "UP"},
	} {
		t.Run(test.name, func(t *testing.T) {
			callsBeforeHealth := checker.Calls()
			health := runtimeRequest(runtime, http.MethodGet, "/health", nil, nil)
			require.Equal(t, http.StatusOK, health.Code, health.Body.String())
			require.Equal(t, callsBeforeHealth, checker.Calls(), "liveness must not probe PostgreSQL")
			requireProbeResponse(t, health, "UP")

			ready := runtimeRequest(runtime, http.MethodGet, "/ready", nil, nil)
			require.Equal(t, test.readyCode, ready.Code, ready.Body.String())
			requireProbeResponse(t, ready, test.readyState)
			require.NotContains(t, ready.Body.String(), dependencyCause.Error())
		})
	}
	require.Equal(t, 3, checker.Calls())

	ping := runtimeRequest(runtime, http.MethodGet, "/ping", nil, nil)
	require.Equal(t, http.StatusNotFound, ping.Code, ping.Body.String())
}

func TestRuntimeRouterServesSPAHistoryFallback(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("profile application"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NETBOX_WEB_ASSETS_PATH", root)
	config.Set(&config.Config{})
	db, err := gorm.Open(sqlite.Open("file:runtime_spa?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	core := composition.NewCore(db)
	recorder := httptest.NewRecorder()
	New(core.Identity, core.Sites, false, nil, alwaysReadyChecker(), runtimeWorkflowOptions(core)...).ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/dcim/sites/", nil),
	)
	if recorder.Code != http.StatusOK || recorder.Body.String() != "profile application" {
		t.Fatalf("SPA fallback = %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestRuntimeAccessLogDoesNotRecordIdentitySecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config.Set(&config.Config{})
	db, err := gorm.Open(sqlite.Open("file:runtime_safe_access_log?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(identitypostgres.Models()...))
	core := composition.NewCore(db)

	const password = "Router-Log-Password-2026!"
	_, err = core.Identity.BootstrapAdministrator(t.Context(), "audit-admin", "", password)
	require.NoError(t, err)

	logCore, observed := observer.New(zap.InfoLevel)
	runtime := newWithLogger(
		core.Identity,
		core.Sites,
		false,
		nil,
		alwaysReadyChecker(),
		zap.New(logCore),
		runtimeWorkflowOptions(core)...,
	)

	const querySecret = "future-query-secret-2026"
	requestIDs := []string{"csrf-request-id", "login-request-id", "token-request-id"}
	csrfResponse := runtimeRequest(
		runtime,
		http.MethodGet,
		"/api/auth/csrf/?access_key="+querySecret,
		nil,
		map[string]string{"X-Request-Id": requestIDs[0]},
	)
	require.Equal(t, http.StatusOK, csrfResponse.Code)
	var csrfBody struct {
		Token string `json:"csrf_token"`
	}
	require.NoError(t, json.Unmarshal(csrfResponse.Body.Bytes(), &csrfBody))
	require.NotEmpty(t, csrfBody.Token)
	initialCSRF := responseCookie(csrfResponse.Result(), "csrftoken")
	require.NotNil(t, initialCSRF)
	if subtle.ConstantTimeCompare([]byte(csrfBody.Token), []byte(initialCSRF.Value)) != 1 {
		t.Fatal("CSRF response body and cookie did not match")
	}

	loginBody, err := json.Marshal(map[string]string{"username": "audit-admin", "password": password})
	require.NoError(t, err)
	loginResponse := runtimeRequest(
		runtime,
		http.MethodPost,
		"/api/auth/login/",
		loginBody,
		map[string]string{
			"Content-Type": "application/json",
			"Cookie":       initialCSRF.String(),
			"X-CSRFToken":  initialCSRF.Value,
			"X-Request-Id": requestIDs[1],
		},
	)
	require.Equal(t, http.StatusOK, loginResponse.Code)
	session := responseCookie(loginResponse.Result(), "sessionid")
	rotatedCSRF := responseCookie(loginResponse.Result(), "csrftoken")
	require.NotNil(t, session)
	require.NotNil(t, rotatedCSRF)

	tokenResponse := runtimeRequest(
		runtime,
		http.MethodPost,
		"/api/auth/tokens/",
		[]byte(`{"description":"log regression","write_enabled":true}`),
		map[string]string{
			"Content-Type": "application/json",
			"Cookie":       session.String() + "; " + rotatedCSRF.String(),
			"X-CSRFToken":  rotatedCSRF.Value,
			"X-Request-Id": requestIDs[2],
		},
	)
	require.Equal(t, http.StatusCreated, tokenResponse.Code)
	var tokenBody map[string]any
	require.NoError(t, json.Unmarshal(tokenResponse.Body.Bytes(), &tokenBody))
	tokenSecret, ok := tokenBody["secret"].(string)
	require.True(t, ok)
	require.NotEmpty(t, tokenSecret)

	entries := observed.All()
	require.Len(t, entries, 3)
	expected := []struct {
		method string
		path   string
		status int
	}{
		{method: http.MethodGet, path: "/api/auth/csrf/", status: http.StatusOK},
		{method: http.MethodPost, path: "/api/auth/login/", status: http.StatusOK},
		{method: http.MethodPost, path: "/api/auth/tokens/", status: http.StatusCreated},
	}
	serializable := make([]map[string]any, 0, len(entries))
	for index, entry := range entries {
		fields := entry.ContextMap()
		require.Equal(t, "Gin response", entry.Message)
		require.Equal(t, expected[index].method, fields["method"])
		require.Equal(t, expected[index].path, fields["url"])
		require.EqualValues(t, expected[index].status, fields["code"])
		require.Equal(t, requestIDs[index], fields["request_id"])
		require.Contains(t, fields, "time_us")
		require.Contains(t, fields, "size")
		require.NotContains(t, fields, "body")
		require.NotContains(t, fields, "headers")
		require.False(t, strings.Contains(fields["url"].(string), "?"))
		serializable = append(serializable, map[string]any{"message": entry.Message, "fields": fields})
	}

	captured, err := json.Marshal(serializable)
	require.NoError(t, err)
	for label, secret := range map[string]string{
		"login password":        password,
		"initial CSRF value":    initialCSRF.Value,
		"rotated CSRF value":    rotatedCSRF.Value,
		"browser session value": session.Value,
		"one-time token secret": tokenSecret,
		"query-string secret":   querySecret,
	} {
		if secret == "" {
			t.Fatalf("test did not obtain %s", label)
		}
		if bytes.Contains(captured, []byte(secret)) {
			t.Fatalf("access log exposed %s", label)
		}
	}
}

func TestSPAServesAssetsAndFallsBackWithoutInterceptingAPI(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("profile application"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "app.js"), []byte("application asset"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		path, body string
	}{
		{path: "/app.js", body: "application asset"},
		{path: "/dcim/sites/1/", body: "profile application"},
		{path: "/ping", body: "profile application"},
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, test.path, nil)
		serveSPA(root)((testContext{request: request, recorder: recorder}).ginContext())
		if recorder.Code != http.StatusOK || recorder.Body.String() != test.body {
			t.Fatalf("GET %s = %d %q, want 200 %q", test.path, recorder.Code, recorder.Body.String(), test.body)
		}
	}
}

func TestRuntimeCORSCSRFAndToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("NETBOX_WEB_ASSETS_PATH", filepath.Join(t.TempDir(), "missing"))
	config.Set(&config.Config{})
	db, err := gorm.Open(sqlite.Open("file:runtime_cors_identity?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(identitypostgres.Models()...))
	core := composition.NewCore(db)
	administrator, err := core.Identity.BootstrapAdministrator(t.Context(), "cors-admin", "", "CORS-Admin-Password-2026!")
	require.NoError(t, err)
	readOnlyToken, err := core.Identity.CreateToken(t.Context(), administrator.Principal(), identityapp.CreateTokenInput{Description: "CORS read only"})
	require.NoError(t, err)
	writeToken, err := core.Identity.CreateToken(t.Context(), administrator.Principal(), identityapp.CreateTokenInput{Description: "CORS write", WriteEnabled: true})
	require.NoError(t, err)

	runtime := New(core.Identity, core.Sites, false, []string{corsTestTrustedOrigin}, alwaysReadyChecker(), runtimeWorkflowOptions(core)...)

	csrfResponse := runtimeRequest(
		runtime,
		http.MethodGet,
		"/api/auth/csrf/",
		nil,
		map[string]string{"Origin": corsTestTrustedOrigin},
	)
	require.Equal(t, http.StatusOK, csrfResponse.Code)
	requireTrustedCORS(t, csrfResponse.Header(), corsTestTrustedOrigin)
	initialCSRF := responseCookie(csrfResponse.Result(), "csrftoken")
	require.NotNil(t, initialCSRF)

	loginBody, err := json.Marshal(map[string]string{
		"username": "cors-admin",
		"password": "CORS-Admin-Password-2026!",
	})
	require.NoError(t, err)
	loginResponse := runtimeRequest(
		runtime,
		http.MethodPost,
		"/api/auth/login/",
		loginBody,
		map[string]string{
			"Content-Type": "application/json",
			"Cookie":       initialCSRF.String(),
			"Origin":       corsTestTrustedOrigin,
			"X-CSRFToken":  initialCSRF.Value,
		},
	)
	require.Equal(t, http.StatusOK, loginResponse.Code)
	requireTrustedCORS(t, loginResponse.Header(), corsTestTrustedOrigin)
	session := responseCookie(loginResponse.Result(), "sessionid")
	rotatedCSRF := responseCookie(loginResponse.Result(), "csrftoken")
	require.NotNil(t, session)
	require.NotNil(t, rotatedCSRF)
	authCookies := session.String() + "; " + rotatedCSRF.String()

	csrfMutationCases := []struct {
		name      string
		csrfToken string
		want      int
	}{
		{name: "missing CSRF", want: http.StatusForbidden},
		{name: "wrong CSRF", csrfToken: "wrong-csrf-value", want: http.StatusForbidden},
		{name: "matching CSRF", csrfToken: rotatedCSRF.Value, want: http.StatusCreated},
	}
	for _, test := range csrfMutationCases {
		t.Run(test.name, func(t *testing.T) {
			headers := map[string]string{
				"Content-Type": "application/json",
				"Cookie":       authCookies,
				"Origin":       corsTestTrustedOrigin,
			}
			if test.csrfToken != "" {
				headers["X-CSRFToken"] = test.csrfToken
			}
			response := runtimeRequest(
				runtime,
				http.MethodPost,
				"/api/auth/tokens/",
				[]byte(`{"description":"CORS CSRF regression","write_enabled":true}`),
				headers,
			)
			require.Equal(t, test.want, response.Code, test.name)
			requireTrustedCORS(t, response.Header(), corsTestTrustedOrigin)
		})
	}

	tokenCases := []struct {
		name   string
		method string
		path   string
		body   []byte
		token  string
		want   int
	}{
		{name: "read-only token safe request", method: http.MethodGet, path: "/api/auth/session/", token: readOnlyToken.Secret, want: http.StatusOK},
		{name: "read-only token unsafe request", method: http.MethodPost, path: "/api/auth/tokens/", body: []byte(`{"description":"read-only denial"}`), token: readOnlyToken.Secret, want: http.StatusForbidden},
		{name: "write token needs no CSRF", method: http.MethodPost, path: "/api/auth/tokens/", body: []byte(`{"description":"write-token success"}`), token: writeToken.Secret, want: http.StatusCreated},
	}
	for _, test := range tokenCases {
		t.Run(test.name, func(t *testing.T) {
			response := runtimeRequest(
				runtime,
				test.method,
				test.path,
				test.body,
				map[string]string{
					"Authorization": "Token " + test.token,
					"Content-Type":  "application/json",
					"Origin":        corsTestTrustedOrigin,
				},
			)
			require.Equal(t, test.want, response.Code, test.name)
			requireTrustedCORS(t, response.Header(), corsTestTrustedOrigin)
		})
	}

	noOriginResponse := runtimeRequest(runtime, http.MethodGet, "/api/auth/session/", nil, nil)
	untrustedResponse := runtimeRequest(
		runtime,
		http.MethodGet,
		"/api/auth/session/",
		nil,
		map[string]string{"Origin": "https://untrusted.example.test"},
	)
	require.Equal(t, noOriginResponse.Code, untrustedResponse.Code)
	require.Equal(t, noOriginResponse.Body.String(), untrustedResponse.Body.String())
	requireNoCORSGrant(t, noOriginResponse.Header())
	requireNoCORSGrant(t, untrustedResponse.Header())
	requireVaryTokenCount(t, noOriginResponse.Header(), "Origin", 1)
	requireVaryTokenCount(t, untrustedResponse.Header(), "Origin", 1)

	emptyRuntime := New(core.Identity, core.Sites, false, nil, alwaysReadyChecker(), runtimeWorkflowOptions(core)...)
	emptyPolicyResponse := runtimeRequest(
		emptyRuntime,
		http.MethodGet,
		"/api/auth/session/",
		nil,
		map[string]string{"X-Request-Id": "empty-cors-policy"},
	)
	trustedPolicySameOriginResponse := runtimeRequest(
		runtime,
		http.MethodGet,
		"/api/auth/session/",
		nil,
		map[string]string{"X-Request-Id": "empty-cors-policy"},
	)
	require.Equal(t, trustedPolicySameOriginResponse.Code, emptyPolicyResponse.Code)
	require.Equal(t, trustedPolicySameOriginResponse.Body.String(), emptyPolicyResponse.Body.String())
	require.Equal(t, trustedPolicySameOriginResponse.Header(), emptyPolicyResponse.Header())
	requireNoCORSGrant(t, emptyPolicyResponse.Header())
}

func TestRuntimeCORSRouteInventory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("NETBOX_WEB_ASSETS_PATH", filepath.Join(t.TempDir(), "missing"))
	config.Set(&config.Config{})
	db, err := gorm.Open(sqlite.Open("file:runtime_cors_inventory?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	core := composition.NewCore(db)

	emptyRuntime := New(core.Identity, core.Sites, false, nil, alwaysReadyChecker(), runtimeWorkflowOptions(core)...)
	allowedOrigins := []string{corsTestTrustedOrigin}
	corsRuntime := New(core.Identity, core.Sites, false, allowedOrigins, alwaysReadyChecker(), runtimeWorkflowOptions(core)...)
	allowedOrigins[0] = "https://mutated.example.test"

	require.Equal(t, semanticRouteInventory(emptyRuntime), semanticRouteInventory(corsRuntime))
	for _, route := range corsRuntime.Routes() {
		require.NotEqual(t, http.MethodOptions, route.Method, "%s unexpectedly registered OPTIONS", route.Path)
		require.NotEqual(t, "/ping", route.Path)
	}

	requestMethod := http.MethodPost
	preflight := performCORSRequest(
		corsRuntime,
		http.MethodOptions,
		"/api/auth/login/",
		[]string{corsTestTrustedOrigin},
		&requestMethod,
		nil,
	)
	require.Equal(t, http.StatusOK, preflight.Code)
	requireTrustedCORS(t, preflight.Header(), corsTestTrustedOrigin)
	requireFixedCORSOptions(t, preflight.Header())

	trusted := runtimeRequest(corsRuntime, http.MethodGet, "/health", nil, map[string]string{"Origin": corsTestTrustedOrigin})
	require.Equal(t, http.StatusOK, trusted.Code)
	requireTrustedCORS(t, trusted.Header(), corsTestTrustedOrigin)

	mutated := runtimeRequest(corsRuntime, http.MethodGet, "/health", nil, map[string]string{"Origin": "https://mutated.example.test"})
	require.Equal(t, http.StatusOK, mutated.Code)
	requireNoCORSGrant(t, mutated.Header())
}

func semanticRouteInventory(runtime *gin.Engine) []string {
	routes := runtime.Routes()
	inventory := make([]string, 0, len(routes))
	for _, route := range routes {
		inventory = append(inventory, route.Method+" "+route.Path+" "+route.Handler)
	}
	return inventory
}

// Keep the package test focused on the middleware by building the smallest
// Gin context required by serveSPA.
type testContext struct {
	request  *http.Request
	recorder *httptest.ResponseRecorder
}

func (value testContext) ginContext() *gin.Context {
	context, _ := gin.CreateTestContext(value.recorder)
	context.Request = value.request
	return context
}

func runtimeRequest(
	handler http.Handler,
	method string,
	path string,
	body []byte,
	headers map[string]string,
) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func responseCookie(response *http.Response, name string) *http.Cookie {
	for _, candidate := range response.Cookies() {
		if candidate.Name == name {
			return candidate
		}
	}
	return nil
}

func runtimeWorkflowOptions(core composition.Core) []workflowhttp.HandlerOption {
	return []workflowhttp.HandlerOption{
		workflowhttp.WithOrganizationServices(core.Manufacturers, core.RackRoles),
		workflowhttp.WithRackTypeService(core.RackTypes),
		workflowhttp.WithRackService(core.Racks),
		workflowhttp.WithDeviceRoleService(core.DeviceRoles),
		workflowhttp.WithDeviceTypeService(core.DeviceTypes),
		workflowhttp.WithInterfaceTemplateService(core.InterfaceTemplates),
		workflowhttp.WithDeviceService(core.Devices),
		workflowhttp.WithInterfaceService(core.Interfaces),
		workflowhttp.WithVRFService(core.VRFs),
		workflowhttp.WithPrefixService(core.Prefixes),
		workflowhttp.WithIPAddressService(core.IPAddresses),
	}
}

type scriptedReadinessChecker struct {
	mu      sync.Mutex
	results []error
	calls   int
}

func (checker *scriptedReadinessChecker) Check(context.Context) error {
	checker.mu.Lock()
	defer checker.mu.Unlock()
	if checker.calls >= len(checker.results) {
		return errors.New("unexpected readiness check")
	}
	result := checker.results[checker.calls]
	checker.calls++
	return result
}

func (checker *scriptedReadinessChecker) Calls() int {
	checker.mu.Lock()
	defer checker.mu.Unlock()
	return checker.calls
}

func alwaysReadyChecker() readiness.Checker {
	return readinessCheckerFunc(func(context.Context) error { return nil })
}

type readinessCheckerFunc func(context.Context) error

func (check readinessCheckerFunc) Check(ctx context.Context) error { return check(ctx) }

func requireProbeResponse(t *testing.T, response *httptest.ResponseRecorder, status string) {
	t.Helper()
	var body struct {
		Status   string `json:"status"`
		Hostname string `json:"hostname"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, status, body.Status)
	require.NotEmpty(t, body.Hostname)
	require.Equal(t, "application/json; charset=utf-8", response.Header().Get("Content-Type"))
}
