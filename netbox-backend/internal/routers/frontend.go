package routers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// webAssetsPath is the directory where the compiled Vue.js frontend lives.
// This is set at build time or via Docker COPY. If it doesn't exist, the
// backend serves API-only (no frontend), which is useful for development.
var webAssetsPath = "/app/web/dist"

// registerFrontendStatic serves the Vue.js SPA from the dist directory.
// It handles SPA routing by falling back to index.html for non-API routes.
func registerFrontendStatic(r *gin.Engine) {
	// Check if the frontend dist directory exists
	info, err := os.Stat(webAssetsPath)
	if err != nil || !info.IsDir() {
		// Frontend not built yet — API-only mode
		return
	}

	// Serve static assets (JS, CSS, images, etc.)
	r.Use(serveSPA(webAssetsPath))
}

// serveSPA returns middleware that serves the Vue.js SPA.
// Non-API, non-file routes fall back to index.html for client-side routing.
func serveSPA(root string) gin.HandlerFunc {
	fs := http.Dir(root)
	fileServer := http.FileServer(fs)

	return func(c *gin.Context) {
		// Skip API routes — those are handled by registered handlers
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") ||
			strings.HasPrefix(path, "/apis/") ||
			path == "/health" ||
			path == "/ping" ||
			path == "/codes" ||
			path == "/metrics" {
			c.Next()
			return
		}

		// Check if the requested file exists
		fullPath := filepath.Join(root, filepath.Clean(path))
		if _, err := os.Stat(fullPath); err == nil {
			// File exists, serve it
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		// File doesn't exist — serve index.html for SPA routing
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
