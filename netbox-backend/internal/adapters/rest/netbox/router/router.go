// Package router composes the supported HTTP runtime without importing the
// frozen generated router/handler/model tree.
package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/errcode"
	"github.com/go-dev-frame/sponge/pkg/gin/handlerfunc"
	"github.com/go-dev-frame/sponge/pkg/gin/middleware"
	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/utils"
	"go.uber.org/zap"

	"netbox-go/api/openapi"
	identityhttp "netbox-go/internal/adapters/rest/netbox/identity"
	workflowhttp "netbox-go/internal/adapters/rest/netbox/workflow"
	dcimapp "netbox-go/internal/application/dcim"
	identityapp "netbox-go/internal/application/identity"
	"netbox-go/internal/config"
	"netbox-go/internal/platform/readiness"
)

const defaultWebAssetsPath = "/app/web/dist"

// New constructs the entire public HTTP surface for the current Capability
// Profile. Generated legacy routes cannot self-register into this package.
func New(
	identityService *identityapp.Service,
	siteService *dcimapp.SiteService,
	secureCookies bool,
	corsAllowedOrigins []string,
	readinessChecker readiness.Checker,
	workflowOptions ...workflowhttp.HandlerOption,
) *gin.Engine {
	return newWithLogger(
		identityService,
		siteService,
		secureCookies,
		corsAllowedOrigins,
		readinessChecker,
		logger.Get(),
		workflowOptions...,
	)
}

func newWithLogger(
	identityService *identityapp.Service,
	siteService *dcimapp.SiteService,
	secureCookies bool,
	corsAllowedOrigins []string,
	readinessChecker readiness.Checker,
	log *zap.Logger,
	workflowOptions ...workflowhttp.HandlerOption,
) *gin.Engine {
	if identityService == nil {
		panic("runtime router requires an identity service")
	}
	if readinessChecker == nil {
		panic("runtime router requires a readiness checker")
	}

	r := gin.New()
	r.RedirectTrailingSlash = false
	r.Use(gin.Recovery())
	r.Use(trustedOriginCORS(corsAllowedOrigins))

	if config.Get().HTTP.Timeout > 0 {
		r.Use(middleware.Timeout(time.Second * time.Duration(config.Get().HTTP.Timeout)))
	}
	r.Use(middleware.RequestID())
	r.Use(pathOnlyAccessLog(log, "/metrics"))
	if config.Get().App.EnableLimit {
		r.Use(middleware.RateLimit())
	}
	if config.Get().App.EnableCircuitBreaker {
		r.Use(middleware.CircuitBreaker(
			middleware.WithValidCode(errcode.InternalServerError.Code(), errcode.ServiceUnavailable.Code()),
		))
	}
	if config.Get().App.EnableTrace {
		r.Use(middleware.Tracing(config.Get().App.Name))
	}

	r.GET("/health", handlerfunc.CheckHealth)
	r.GET("/ready", readinessHandler(readinessChecker))

	identityHandler := identityhttp.NewHandler(identityService, secureCookies)
	identityHandler.Register(r)
	authenticate := identityHandler.BaselineMiddleware()
	workflowhttp.NewHandler(siteService, workflowOptions...).Register(r, authenticate)
	r.GET("/api/schema/", authenticate, func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", openapi.Schema)
	})

	registerSPA(r, webAssetsPath())
	return r
}

func readinessHandler(checker readiness.Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := checker.Check(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, handlerfunc.CheckHealthReply{
				Status:   "DOWN",
				Hostname: utils.GetHostname(),
			})
			return
		}
		handlerfunc.CheckHealth(c)
	}
}

// pathOnlyAccessLog intentionally records only request/response metadata. In
// particular, it never reads bodies or headers and never logs the raw URL,
// whose query string may contain credentials.
func pathOnlyAccessLog(log *zap.Logger, ignoredPaths ...string) gin.HandlerFunc {
	if log == nil {
		log = zap.NewNop()
	}
	ignored := make(map[string]struct{}, len(ignoredPaths))
	for _, path := range ignoredPaths {
		ignored[path] = struct{}{}
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if _, ok := ignored[path]; ok {
			c.Next()
			return
		}
		method := c.Request.Method
		requestID := middleware.GCtxRequestID(c)
		started := time.Now()

		c.Next()

		fields := []zap.Field{
			zap.Int("code", c.Writer.Status()),
			zap.String("method", method),
			zap.String("url", path),
			zap.Int64("time_us", time.Since(started).Microseconds()),
			zap.Int("size", c.Writer.Size()),
		}
		if requestID != "" {
			fields = append(fields, zap.String(middleware.ContextRequestIDKey, requestID))
		}
		if c.Writer.Status() >= http.StatusInternalServerError {
			log.Error("Gin response", fields...)
			return
		}
		log.Info("Gin response", fields...)
	}
}

func webAssetsPath() string {
	if value := strings.TrimSpace(os.Getenv("NETBOX_WEB_ASSETS_PATH")); value != "" {
		return value
	}
	return defaultWebAssetsPath
}

func registerSPA(r *gin.Engine, root string) {
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return
	}
	r.Use(serveSPA(root))
}

func serveSPA(root string) gin.HandlerFunc {
	filesystem := http.Dir(root)
	server := http.FileServer(filesystem)
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") ||
			strings.HasPrefix(path, "/apis/") ||
			path == "/health" ||
			path == "/ready" ||
			path == "/codes" ||
			path == "/metrics" {
			c.Next()
			return
		}

		fullPath := filepath.Join(root, filepath.Clean(path))
		if _, err := os.Stat(fullPath); err == nil {
			server.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}
		c.Request.URL.Path = "/"
		server.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
