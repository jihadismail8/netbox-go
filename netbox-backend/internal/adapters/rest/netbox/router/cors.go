package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	corsAllowMethods = "DELETE, GET, OPTIONS, PATCH, POST, PUT"
	corsAllowHeaders = "accept, authorization, content-type, user-agent, x-csrftoken, x-requested-with"
	corsMaxAge       = "86400"
)

const (
	headerAllowCredentials = "Access-Control-Allow-Credentials"
	headerAllowHeaders     = "Access-Control-Allow-Headers"
	headerAllowMethods     = "Access-Control-Allow-Methods"
	headerAllowOrigin      = "Access-Control-Allow-Origin"
	headerAllowPrivateNet  = "Access-Control-Allow-Private-Network"
	headerExposeHeaders    = "Access-Control-Expose-Headers"
	headerMaxAge           = "Access-Control-Max-Age"
	headerRequestMethod    = "Access-Control-Request-Method"
	headerVary             = "Vary"
)

// trustedOriginCORS applies the fixed credentialed CORS contract to the
// canonical HTTP runtime. allowedOrigins is copied into an exact-match set so
// callers cannot change the policy after construction.
func trustedOriginCORS(allowedOrigins []string) gin.HandlerFunc {
	trustedOrigins := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		trustedOrigins[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		origin, trusted := trustedRequestOrigin(c.Request.Header, trustedOrigins)
		writer := &corsResponseWriter{
			ResponseWriter: c.Writer,
			method:         c.Request.Method,
			origin:         origin,
			trusted:        trusted,
		}
		c.Writer = writer
		writer.apply()

		if c.Request.Method == http.MethodOptions && headerFieldPresent(c.Request.Header, headerRequestMethod) {
			c.Header("Content-Length", "0")
			c.AbortWithStatus(http.StatusOK)
			return
		}

		defer writer.apply()
		c.Next()
	}
}

// corsResponseWriter reapplies the policy immediately before a response is
// committed as well as after downstream returns. That keeps this middleware
// authoritative if a later handler attempts to replace its headers.
type corsResponseWriter struct {
	gin.ResponseWriter
	method  string
	origin  string
	trusted bool
}

func (w *corsResponseWriter) WriteHeader(code int) {
	w.apply()
	w.ResponseWriter.WriteHeader(code)
}

func (w *corsResponseWriter) WriteHeaderNow() {
	w.apply()
	w.ResponseWriter.WriteHeaderNow()
}

func (w *corsResponseWriter) Write(data []byte) (int, error) {
	w.apply()
	return w.ResponseWriter.Write(data)
}

func (w *corsResponseWriter) WriteString(data string) (int, error) {
	w.apply()
	return w.ResponseWriter.WriteString(data)
}

func (w *corsResponseWriter) Flush() {
	w.apply()
	w.ResponseWriter.Flush()
}

func (w *corsResponseWriter) apply() {
	applyCORSResponseHeaders(w.Header(), w.method, w.origin, w.trusted)
}

func applyCORSResponseHeaders(headers http.Header, method, origin string, trusted bool) {
	for _, name := range []string{
		headerAllowCredentials,
		headerAllowHeaders,
		headerAllowMethods,
		headerAllowOrigin,
		headerAllowPrivateNet,
		headerExposeHeaders,
		headerMaxAge,
	} {
		deleteHeaderFields(headers, name)
	}
	mergeVaryOrigin(headers)
	if !trusted {
		return
	}

	headers.Set(headerAllowOrigin, origin)
	headers.Set(headerAllowCredentials, "true")
	if method == http.MethodOptions {
		setCORSOptionsHeaders(headers)
	}
}

func trustedRequestOrigin(headers http.Header, trustedOrigins map[string]struct{}) (string, bool) {
	values := headerValues(headers, "Origin")
	if len(values) != 1 {
		return "", false
	}
	_, trusted := trustedOrigins[values[0]]
	return values[0], trusted
}

func headerValues(headers http.Header, name string) []string {
	var values []string
	for key, candidates := range headers {
		if strings.EqualFold(key, name) {
			values = append(values, candidates...)
		}
	}
	return values
}

func headerFieldPresent(headers http.Header, name string) bool {
	for key := range headers {
		if strings.EqualFold(key, name) {
			return true
		}
	}
	return false
}

func deleteHeaderFields(headers http.Header, name string) {
	for key := range headers {
		if strings.EqualFold(key, name) {
			delete(headers, key)
		}
	}
}

func setCORSOptionsHeaders(headers http.Header) {
	headers.Set(headerAllowMethods, corsAllowMethods)
	headers.Set(headerAllowHeaders, corsAllowHeaders)
	headers.Set(headerMaxAge, corsMaxAge)
}

func mergeVaryOrigin(headers http.Header) {
	values := headerValues(headers, headerVary)
	deleteHeaderFields(headers, headerVary)
	tokens := make([]string, 0, len(values)+1)
	seen := make(map[string]struct{}, len(values)+1)

	for _, value := range values {
		for _, rawToken := range strings.Split(value, ",") {
			token := strings.TrimSpace(rawToken)
			if token == "" {
				continue
			}
			if token == "*" {
				headers.Set(headerVary, "*")
				return
			}
			key := strings.ToLower(token)
			if _, duplicate := seen[key]; duplicate {
				continue
			}
			seen[key] = struct{}{}
			tokens = append(tokens, token)
		}
	}

	if _, exists := seen["origin"]; !exists {
		tokens = append(tokens, "Origin")
	}
	headers.Set(headerVary, strings.Join(tokens, ", "))
}
