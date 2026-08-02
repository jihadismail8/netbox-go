package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	identitypostgres "netbox-go/internal/adapters/postgres/identity"
	workflowhttp "netbox-go/internal/adapters/rest/netbox/workflow"
	"netbox-go/internal/config"
	"netbox-go/internal/platform/composition"
)

func TestBaselineAndIdentityAuthenticationStatusesRemainDistinct(t *testing.T) {
	config.Set(&config.Config{})
	db, err := gorm.Open(sqlite.Open("file:runtime_auth_status?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	core := composition.NewCore(db)
	router := New(core.Identity, core.Sites, false, runtimeWorkflowOptions(core)...)

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

func TestRuntimePublicProbeBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("NETBOX_WEB_ASSETS_PATH", filepath.Join(t.TempDir(), "missing"))
	config.Set(&config.Config{})
	db, err := gorm.Open(sqlite.Open("file:runtime_public_probes?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	core := composition.NewCore(db)
	runtime := New(core.Identity, core.Sites, false, runtimeWorkflowOptions(core)...)

	routes := make(map[string]struct{}, len(runtime.Routes()))
	for _, route := range runtime.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}
	require.Contains(t, routes, http.MethodGet+" /health")
	require.Contains(t, routes, http.MethodGet+" /ready")
	require.NotContains(t, routes, http.MethodGet+" /ping")

	for _, test := range []struct {
		path string
		want int
	}{
		{path: "/health", want: http.StatusOK},
		{path: "/ready", want: http.StatusOK},
		{path: "/ping", want: http.StatusNotFound},
	} {
		response := runtimeRequest(runtime, http.MethodGet, test.path, nil, nil)
		require.Equal(t, test.want, response.Code, "GET %s: %s", test.path, response.Body.String())
	}
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
	New(core.Identity, core.Sites, false, runtimeWorkflowOptions(core)...).ServeHTTP(
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
	require.Equal(t, csrfBody.Token, initialCSRF.Value)

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
	session := responseCookie(loginResponse.Result(), "netbox_session")
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
