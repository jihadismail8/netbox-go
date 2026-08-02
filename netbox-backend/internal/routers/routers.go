package routers

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-dev-frame/sponge/pkg/errcode"
	"github.com/go-dev-frame/sponge/pkg/gin/handlerfunc"
	"github.com/go-dev-frame/sponge/pkg/gin/middleware"
	"github.com/go-dev-frame/sponge/pkg/logger"

	"netbox-go/internal/config"
)

type routeFns = []func(r *gin.Engine, groupPathMiddlewares map[string][]gin.HandlerFunc, singlePathMiddlewares map[string][]gin.HandlerFunc)

var (
	// all route functions
	allRouteFns = make(routeFns, 0)
	// all middleware functions
	allMiddlewareFns = []func(c *middlewareConfig){}
)

// NewRouter create a new router
func NewRouter() *gin.Engine { //nolint
	return newRouter()
}

func newRouter() *gin.Engine { //nolint
	r := gin.New()

	// NetBox API uses trailing slashes on all endpoints (both list and detail).
	// Disable Gin's automatic 301/307 redirects — they lose request bodies on
	// PATCH/PUT. Instead, we strip trailing slashes in middleware below so that
	// /api/dcim/sites/3/ matches the registered /api/dcim/sites/:id route.
	r.RedirectTrailingSlash = false

	// Strip trailing slash from all requests except the root "/".
	// This ensures /api/dcim/sites/3/ is handled by /api/dcim/sites/:id
	// without a redirect, preserving PATCH/PUT request bodies.
	r.Use(func(c *gin.Context) {
		p := c.Request.URL.Path
		if len(p) > 1 && strings.HasSuffix(p, "/") {
			c.Request.URL.Path = strings.TrimRight(p, "/")
		}
		c.Next()
	})

	r.Use(gin.Recovery())
	// The Vue application is served from this listener and uses same-origin
	// cookies. Cross-origin credentialed requests are disabled by default.

	if config.Get().HTTP.Timeout > 0 {
		// if you need more fine-grained control over your routes, set the timeout in your routes, unsetting the timeout globally here.
		r.Use(middleware.Timeout(time.Second * time.Duration(config.Get().HTTP.Timeout)))
	}

	// request id middleware
	r.Use(middleware.RequestID())

	// logger middleware, to print simple messages, replace middleware.Logging with middleware.SimpleLog
	r.Use(middleware.Logging(
		middleware.WithLog(logger.Get()),
		middleware.WithRequestIDFromContext(),
		middleware.WithIgnoreRoutes("/metrics"), // ignore path
	))

	// limit middleware
	if config.Get().App.EnableLimit {
		r.Use(middleware.RateLimit(
		//middleware.WithWindow(time.Second*5), // default 10s
		//middleware.WithBucket(200), // default 100
		//middleware.WithCPUThreshold(900), // default 800
		))
	}

	// circuit breaker middleware
	if config.Get().App.EnableCircuitBreaker {
		r.Use(middleware.CircuitBreaker(
			//middleware.WithBreakerOption(
			//circuitbreaker.WithSuccess(75),           // default 60
			//circuitbreaker.WithRequest(100),          // default 100
			//circuitbreaker.WithBucket(20),            // default 10
			//circuitbreaker.WithWindow(time.Second*3), // default 3s
			//),
			//middleware.WithDegradeHandler(handler),              // Add degradation processing
			middleware.WithValidCode( // Add error codes to trigger circuit breaking
				errcode.InternalServerError.Code(),
				errcode.ServiceUnavailable.Code(),
			),
		))
	}

	// trace middleware
	if config.Get().App.EnableTrace {
		r.Use(middleware.Tracing(config.Get().App.Name))
	}

	// profile performance analysis
	if config.Get().App.EnableHTTPProfile {
		// implemented on port 8283
	}

	r.GET("/health", handlerfunc.CheckHealth)
	r.GET("/ping", handlerfunc.Ping)
	r.GET("/ready", handlerfunc.CheckHealth)

	c := newMiddlewareConfig()

	// set up all middlewares
	for _, fn := range allMiddlewareFns {
		fn(c)
	}

	// serve embedded Vue.js frontend (SPA) if the dist directory exists
	registerFrontendStatic(r)

	return r
}

type middlewareConfig struct {
	groupPathMiddlewares  map[string][]gin.HandlerFunc // middleware functions corresponding to route group
	singlePathMiddlewares map[string][]gin.HandlerFunc // middleware functions corresponding to a single route
}

func newMiddlewareConfig() *middlewareConfig {
	return &middlewareConfig{
		groupPathMiddlewares:  make(map[string][]gin.HandlerFunc),
		singlePathMiddlewares: make(map[string][]gin.HandlerFunc),
	}
}

func (c *middlewareConfig) setGroupPath(groupPath string, handlers ...gin.HandlerFunc) { //nolint
	if groupPath == "" {
		return
	}
	if groupPath[0] != '/' {
		groupPath = "/" + groupPath
	}

	handlerFns, ok := c.groupPathMiddlewares[groupPath]
	if !ok {
		c.groupPathMiddlewares[groupPath] = handlers
		return
	}

	c.groupPathMiddlewares[groupPath] = append(handlerFns, handlers...)
}

func (c *middlewareConfig) setSinglePath(method string, singlePath string, handlers ...gin.HandlerFunc) { //nolint
	if method == "" || singlePath == "" {
		return
	}

	key := getSinglePathKey(method, singlePath)
	handlerFns, ok := c.singlePathMiddlewares[key]
	if !ok {
		c.singlePathMiddlewares[key] = handlers
		return
	}

	c.singlePathMiddlewares[key] = append(handlerFns, handlers...)
}

func getSinglePathKey(method string, singlePath string) string { //nolint
	return strings.ToUpper(method) + "->" + singlePath
}
