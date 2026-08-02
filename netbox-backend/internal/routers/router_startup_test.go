// Package routers — startup regression tests.
//
// These guard against route-registration panics that only surface at server
// boot (e.g. duplicate path registration), which the per-handler integration
// tests don't exercise.
package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"netbox-go/internal/config"
)

// TestNewRouterDoesNotPanic ensures the full router builds without panicking.
//
// Regression test for a startup crash where /api/users/tokens was registered
// twice: once by the generic DRF autogen (usersToken model) and again by
// auth_routes.go's bespoke handlers. Gin panics on duplicate route registration,
// taking the whole server down. The fix (UnregisterModelEndpoint in
// routers.go) keeps the auth-aware handlers as the single source of truth.
//
// We don't issue any requests here — the goal is only to confirm the route
// tree assembles cleanly. DB access happens lazily inside handlers, so this
// works without a real database.
func TestNewRouterDoesNotPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// NewRouter() reads config (HTTP timeout, feature flags). Seed a minimal
	// zero-value config so it doesn't panic on config.Get().
	config.Set(&config.Config{})

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("NewRouter() panicked: %v", r)
		}
	}()

	r := NewRouter()
	if r == nil {
		t.Fatal("NewRouter() returned nil")
	}
}

func TestNewRouterContainsOnlySafePublicRuntimeSurface(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	cfg.App.EnableMetrics = true
	cfg.App.EnableHTTPProfile = true
	config.Set(cfg)
	r := NewRouter()

	registered := make(map[string]struct{}, len(r.Routes()))
	for _, route := range r.Routes() {
		registered[route.Method+" "+route.Path] = struct{}{}
	}

	for _, route := range []string{"GET /health", "GET /ping"} {
		if _, ok := registered[route]; !ok {
			t.Errorf("expected safe public route %q to be registered", route)
		}
	}

	for _, route := range []string{
		"GET /config",
		"GET /codes",
		"GET /metrics",
		"GET /api/status",
		"POST /api/auth/login",
		"POST /api/auth/provision",
		"GET /api/users/tokens",
		"POST /api/users/tokens",
		"GET /api/dcim/sites/:id",
		"GET /api/dcim/sites",
		"POST /api/dcim/sites",
	} {
		if _, ok := registered[route]; ok {
			t.Errorf("unsafe transitional route %q must not be registered", route)
		}
	}
}

func TestNewRouterDoesNotServeDiagnosticsOrBusinessAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config.Set(&config.Config{})
	r := NewRouter()

	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/config"},
		{method: http.MethodGet, path: "/codes"},
		{method: http.MethodGet, path: "/metrics"},
		{method: http.MethodGet, path: "/api/status/"},
		{method: http.MethodGet, path: "/api/dcim/sites/"},
		{method: http.MethodPost, path: "/api/dcim/sites/"},
		{method: http.MethodPost, path: "/api/auth/provision/"},
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("%s %s returned %d, want 404", tc.method, tc.path, w.Code)
		}
	}
}
