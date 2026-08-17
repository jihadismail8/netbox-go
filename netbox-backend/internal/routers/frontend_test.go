package routers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestServeSPASeparatesOperationalAPIAndFrontendPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	writeFrontendTestFile(t, filepath.Join(root, "index.html"), "profile application")
	writeFrontendTestFile(t, filepath.Join(root, "app.js"), "console.log('netbox')")

	router := gin.New()
	router.Use(serveSPA(root))
	for _, path := range []string{
		"/api/widgets",
		"/apis/schema",
		"/health",
		"/ping",
		"/codes",
		"/metrics",
	} {
		router.GET(path, func(c *gin.Context) {
			c.String(http.StatusAccepted, "downstream")
		})
	}

	for _, tt := range []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{name: "API route continues", path: "/api/widgets", wantStatus: http.StatusAccepted, wantBody: "downstream"},
		{name: "API schema continues", path: "/apis/schema", wantStatus: http.StatusAccepted, wantBody: "downstream"},
		{name: "health continues", path: "/health", wantStatus: http.StatusAccepted, wantBody: "downstream"},
		{name: "ping continues", path: "/ping", wantStatus: http.StatusAccepted, wantBody: "downstream"},
		{name: "codes continues", path: "/codes", wantStatus: http.StatusAccepted, wantBody: "downstream"},
		{name: "metrics continues", path: "/metrics", wantStatus: http.StatusAccepted, wantBody: "downstream"},
		{name: "existing asset is served", path: "/app.js", wantStatus: http.StatusOK, wantBody: "console.log('netbox')"},
		{name: "history route falls back to index", path: "/dcim/devices/42", wantStatus: http.StatusOK, wantBody: "profile application"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			router.ServeHTTP(response, request)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if got := response.Body.String(); got != tt.wantBody {
				t.Fatalf("body = %q, want %q", got, tt.wantBody)
			}
		})
	}
}

func TestRegisterFrontendStaticRequiresDirectory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	originalPath := webAssetsPath
	t.Cleanup(func() { webAssetsPath = originalPath })

	root := t.TempDir()
	indexPath := filepath.Join(root, "index.html")
	writeFrontendTestFile(t, indexPath, "registered frontend")

	for _, tt := range []struct {
		name       string
		assetsPath string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "missing path leaves API only router",
			assetsPath: filepath.Join(root, "missing"),
			wantStatus: http.StatusNotFound,
			wantBody:   "404 page not found",
		},
		{
			name:       "regular file is not an asset root",
			assetsPath: indexPath,
			wantStatus: http.StatusNotFound,
			wantBody:   "404 page not found",
		},
		{
			name:       "directory installs SPA middleware",
			assetsPath: root,
			wantStatus: http.StatusOK,
			wantBody:   "registered frontend",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			webAssetsPath = tt.assetsPath
			router := gin.New()
			registerFrontendStatic(router)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/profile/route", nil))
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if got := response.Body.String(); got != tt.wantBody {
				t.Fatalf("body = %q, want %q", got, tt.wantBody)
			}
		})
	}
}

func writeFrontendTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write frontend fixture %s: %v", path, err)
	}
}
