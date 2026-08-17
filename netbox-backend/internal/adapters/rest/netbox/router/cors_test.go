package router

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const corsTestTrustedOrigin = "https://console.example.test"

func TestCORSActualRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	allowedOrigins := []string{corsTestTrustedOrigin}
	runtime := gin.New()
	runtime.Use(gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, _ any) {
		c.String(http.StatusInternalServerError, "recovered body")
	}))
	runtime.Use(trustedOriginCORS(allowedOrigins))
	runtime.GET("/actual", func(c *gin.Context) {
		if c.Query("panic") == "true" {
			panic("focused CORS recovery")
		}
		if c.Query("overwrite") == "true" {
			setOverwritingCORSHeaders(c.Writer.Header())
			c.String(http.StatusAccepted, "overwriting downstream body")
			return
		}
		status, err := strconv.Atoi(c.Query("status"))
		require.NoError(t, err)
		c.Header("X-Downstream", "preserved")
		c.String(status, "downstream body")
	})

	t.Run("trusted contract is reasserted after downstream", func(t *testing.T) {
		response := performCORSRequest(runtime, http.MethodGet, "/actual?overwrite=true", []string{corsTestTrustedOrigin}, nil, nil)
		require.Equal(t, http.StatusAccepted, response.Code)
		require.Equal(t, "overwriting downstream body", response.Body.String())
		wireHeaders := response.Result().Header
		requireTrustedCORS(t, wireHeaders, corsTestTrustedOrigin)
		require.Empty(t, wireHeaders.Get(headerAllowMethods))
		require.Empty(t, wireHeaders.Get(headerAllowHeaders))
		require.Empty(t, wireHeaders.Get(headerMaxAge))
		requireNoForbiddenCORSHeaders(t, wireHeaders)
	})

	t.Run("untrusted downstream cannot create a grant", func(t *testing.T) {
		response := performCORSRequest(runtime, http.MethodGet, "/actual?overwrite=true", []string{"https://untrusted.example.test"}, nil, nil)
		require.Equal(t, http.StatusAccepted, response.Code)
		require.Equal(t, "overwriting downstream body", response.Body.String())
		wireHeaders := response.Result().Header
		requireNoCORSGrant(t, wireHeaders)
		requireNoForbiddenCORSHeaders(t, wireHeaders)
	})

	// Construction takes ownership of the values rather than retaining the
	// caller's mutable slice.
	allowedOrigins[0] = "https://mutated.example.test"

	for _, status := range []int{
		http.StatusOK,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
	} {
		t.Run(fmt.Sprintf("trusted status %d", status), func(t *testing.T) {
			response := performCORSRequest(runtime, http.MethodGet, "/actual?status="+strconv.Itoa(status), []string{corsTestTrustedOrigin}, nil, nil)
			require.Equal(t, status, response.Code)
			require.Equal(t, "downstream body", response.Body.String())
			require.Equal(t, "preserved", response.Header().Get("X-Downstream"))
			requireTrustedCORS(t, response.Header(), corsTestTrustedOrigin)
			require.Empty(t, response.Header().Get(headerAllowMethods))
			require.Empty(t, response.Header().Get(headerAllowHeaders))
			require.Empty(t, response.Header().Get(headerMaxAge))
			requireNoForbiddenCORSHeaders(t, response.Header())
		})
	}

	t.Run("trusted recovered status 500", func(t *testing.T) {
		response := performCORSRequest(runtime, http.MethodGet, "/actual?panic=true", []string{corsTestTrustedOrigin}, nil, nil)
		require.Equal(t, http.StatusInternalServerError, response.Code)
		require.Equal(t, "recovered body", response.Body.String())
		requireTrustedCORS(t, response.Header(), corsTestTrustedOrigin)
		requireNoForbiddenCORSHeaders(t, response.Header())
	})

	untrustedCases := []struct {
		name    string
		origins []string
	}{
		{name: "missing origin"},
		{name: "untrusted", origins: []string{"https://untrusted.example.test"}},
		{name: "scheme mismatch", origins: []string{"http://console.example.test"}},
		{name: "subdomain", origins: []string{"https://sub.console.example.test"}},
		{name: "suffix confusable", origins: []string{"https://console.example.test.attacker.invalid"}},
		{name: "port mismatch", origins: []string{"https://console.example.test:8443"}},
		{name: "malformed", origins: []string{"://console.example.test"}},
		{name: "multiple fields", origins: []string{corsTestTrustedOrigin, corsTestTrustedOrigin}},
		{name: "comma combined", origins: []string{corsTestTrustedOrigin + ", https://untrusted.example.test"}},
		{name: "leading whitespace", origins: []string{" " + corsTestTrustedOrigin}},
		{name: "trailing whitespace", origins: []string{corsTestTrustedOrigin + "\t"}},
		{name: "null", origins: []string{"null"}},
		{name: "wildcard", origins: []string{"*"}},
		{name: "mutated constructor value", origins: []string{"https://mutated.example.test"}},
	}
	for _, test := range untrustedCases {
		t.Run(test.name, func(t *testing.T) {
			response := performCORSRequest(runtime, http.MethodGet, "/actual?status=418", test.origins, nil, nil)
			require.Equal(t, http.StatusTeapot, response.Code)
			require.Equal(t, "downstream body", response.Body.String())
			require.Equal(t, "preserved", response.Header().Get("X-Downstream"))
			requireNoCORSGrant(t, response.Header())
			requireVaryTokenCount(t, response.Header(), "Origin", 1)
			requireNoForbiddenCORSHeaders(t, response.Header())
		})
	}
}

func TestCORSPreflightRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	downstreamCalls := 0
	runtime := gin.New()
	runtime.Use(trustedOriginCORS([]string{corsTestTrustedOrigin}))
	runtime.Use(func(c *gin.Context) {
		downstreamCalls++
		if c.GetHeader("X-Test-Overwrite-CORS") == "true" {
			setOverwritingCORSHeaders(c.Writer.Header())
		}
		c.Next()
	})
	runtime.GET("/sentinel", func(c *gin.Context) {
		c.String(http.StatusOK, "downstream sentinel")
	})

	trustedCases := []struct {
		name          string
		requestMethod string
		requestHeader string
	}{
		{name: "ordinary", requestMethod: http.MethodPost, requestHeader: "Content-Type, X-CSRFToken"},
		{name: "empty request method field"},
		{name: "unsupported values are not reflected", requestMethod: http.MethodTrace, requestHeader: "X-Unknown"},
		{name: "request header casing does not alter response", requestMethod: http.MethodPatch, requestHeader: "CONTENT-TYPE, X-CsRfToKeN"},
	}
	for _, test := range trustedCases {
		t.Run(test.name, func(t *testing.T) {
			requestMethod := test.requestMethod
			extra := http.Header{}
			extra.Set("Access-Control-Request-Headers", test.requestHeader)
			extra.Set("Access-Control-Request-Private-Network", "true")
			response := performCORSRequest(runtime, http.MethodOptions, "/sentinel", []string{corsTestTrustedOrigin}, &requestMethod, extra)
			require.Equal(t, http.StatusOK, response.Code)
			require.Empty(t, response.Body.String())
			require.Equal(t, "0", response.Header().Get("Content-Length"))
			requireTrustedCORS(t, response.Header(), corsTestTrustedOrigin)
			requireFixedCORSOptions(t, response.Header())
			require.NotContains(t, response.Header().Get(headerAllowMethods), http.MethodTrace)
			require.NotContains(t, response.Header().Get(headerAllowHeaders), "X-Unknown")
			requireNoForbiddenCORSHeaders(t, response.Header())
			require.Zero(t, downstreamCalls, "preflight reached downstream middleware")
		})
	}

	untrustedCases := []struct {
		name    string
		origins []string
	}{
		{name: "missing origin"},
		{name: "malformed", origins: []string{"https://console.example.test/path"}},
		{name: "multiple", origins: []string{corsTestTrustedOrigin, corsTestTrustedOrigin}},
		{name: "null", origins: []string{"null"}},
		{name: "untrusted", origins: []string{"https://untrusted.example.test"}},
	}
	for _, test := range untrustedCases {
		t.Run(test.name, func(t *testing.T) {
			requestMethod := http.MethodPost
			response := performCORSRequest(runtime, http.MethodOptions, "/sentinel", test.origins, &requestMethod, nil)
			require.Equal(t, http.StatusOK, response.Code)
			require.Empty(t, response.Body.String())
			require.Equal(t, "0", response.Header().Get("Content-Length"))
			requireNoCORSGrant(t, response.Header())
			requireVaryTokenCount(t, response.Header(), "Origin", 1)
			requireNoForbiddenCORSHeaders(t, response.Header())
			require.Zero(t, downstreamCalls, "untrusted preflight reached downstream middleware")
		})
	}

	t.Run("trusted OPTIONS without request method is not preflight", func(t *testing.T) {
		response := performCORSRequest(runtime, http.MethodOptions, "/sentinel", []string{corsTestTrustedOrigin}, nil, http.Header{"X-Test-Overwrite-CORS": []string{"true"}})
		require.Equal(t, http.StatusNotFound, response.Code)
		require.Equal(t, "404 page not found", response.Body.String())
		wireHeaders := response.Result().Header
		requireTrustedCORS(t, wireHeaders, corsTestTrustedOrigin)
		requireFixedCORSOptions(t, wireHeaders)
		requireNoForbiddenCORSHeaders(t, wireHeaders)
		require.Equal(t, 1, downstreamCalls)
	})

	t.Run("untrusted OPTIONS without request method is not preflight", func(t *testing.T) {
		response := performCORSRequest(runtime, http.MethodOptions, "/sentinel", []string{"https://untrusted.example.test"}, nil, http.Header{"X-Test-Overwrite-CORS": []string{"true"}})
		require.Equal(t, http.StatusNotFound, response.Code)
		require.Equal(t, "404 page not found", response.Body.String())
		wireHeaders := response.Result().Header
		requireNoCORSGrant(t, wireHeaders)
		requireVaryTokenCount(t, wireHeaders, "Origin", 1)
		requireNoForbiddenCORSHeaders(t, wireHeaders)
		require.Equal(t, 2, downstreamCalls)
	})
}

func TestCORSVaryHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	runtime := gin.New()
	runtime.Use(func(c *gin.Context) {
		switch c.GetHeader("X-Test-Vary") {
		case "existing":
			c.Writer.Header()["vary"] = []string{"Accept-Encoding", "origin, Accept-Language"}
		case "wildcard":
			c.Writer.Header()["vArY"] = []string{"Accept-Encoding, *", "Origin"}
		}
		c.Next()
	})
	runtime.Use(trustedOriginCORS([]string{corsTestTrustedOrigin}))
	runtime.GET("/vary", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	t.Run("merges without duplicate origin", func(t *testing.T) {
		response := performCORSRequest(runtime, http.MethodGet, "/vary", nil, nil, http.Header{"X-Test-Vary": []string{"existing"}})
		require.Equal(t, http.StatusNoContent, response.Code)
		requireVaryTokenCount(t, response.Header(), "Accept-Encoding", 1)
		requireVaryTokenCount(t, response.Header(), "Accept-Language", 1)
		requireVaryTokenCount(t, response.Header(), "Origin", 1)
		requireNoCORSGrant(t, response.Header())
	})

	t.Run("wildcard remains the sole semantic", func(t *testing.T) {
		response := performCORSRequest(runtime, http.MethodGet, "/vary", []string{corsTestTrustedOrigin}, nil, http.Header{"X-Test-Vary": []string{"wildcard"}})
		require.Equal(t, http.StatusNoContent, response.Code)
		require.Equal(t, []string{"*"}, headerValues(response.Header(), headerVary))
		require.Equal(t, []string{corsTestTrustedOrigin}, headerValues(response.Header(), headerAllowOrigin))
		require.Equal(t, []string{"true"}, headerValues(response.Header(), headerAllowCredentials))
	})
}

func performCORSRequest(
	handler http.Handler,
	method string,
	path string,
	origins []string,
	requestMethod *string,
	extraHeaders http.Header,
) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	for _, origin := range origins {
		request.Header.Add("Origin", origin)
	}
	if requestMethod != nil {
		request.Header[headerRequestMethod] = []string{*requestMethod}
	}
	for key, values := range extraHeaders {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func requireTrustedCORS(t *testing.T, headers http.Header, origin string) {
	t.Helper()
	require.Equal(t, []string{origin}, headerValues(headers, headerAllowOrigin))
	require.Equal(t, []string{"true"}, headerValues(headers, headerAllowCredentials))
	requireVaryTokenCount(t, headers, "Origin", 1)
}

func requireFixedCORSOptions(t *testing.T, headers http.Header) {
	t.Helper()
	require.Equal(t, []string{corsAllowMethods}, headerValues(headers, headerAllowMethods))
	require.Equal(t, []string{corsAllowHeaders}, headerValues(headers, headerAllowHeaders))
	require.Equal(t, []string{corsMaxAge}, headerValues(headers, headerMaxAge))
}

func requireNoCORSGrant(t *testing.T, headers http.Header) {
	t.Helper()
	for _, name := range []string{
		headerAllowOrigin,
		headerAllowCredentials,
		headerAllowMethods,
		headerAllowHeaders,
		headerMaxAge,
	} {
		require.Empty(t, headerValues(headers, name), name)
	}
}

func requireNoForbiddenCORSHeaders(t *testing.T, headers http.Header) {
	t.Helper()
	require.NotContains(t, headerValues(headers, headerAllowOrigin), "*")
	require.NotContains(t, headerValues(headers, headerAllowHeaders), "*")
	require.Empty(t, headerValues(headers, headerExposeHeaders))
	require.Empty(t, headerValues(headers, headerAllowPrivateNet))
}

func setOverwritingCORSHeaders(headers http.Header) {
	headers.Set(headerAllowOrigin, "*")
	headers.Set(headerAllowCredentials, "false")
	headers.Set(headerAllowMethods, http.MethodTrace)
	headers.Set(headerAllowHeaders, "*")
	headers.Set(headerMaxAge, "1")
	headers.Set(headerExposeHeaders, "x-secret")
	headers.Set(headerAllowPrivateNet, "true")
	headers["access-control-allow-origin"] = []string{"https://raw-map-overwrite.example.test"}
	headers["access-control-expose-headers"] = []string{"x-raw-secret"}
}

func requireVaryTokenCount(t *testing.T, headers http.Header, wanted string, count int) {
	t.Helper()
	actual := 0
	values := headerValues(headers, headerVary)
	for _, value := range values {
		for _, token := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(token), wanted) {
				actual++
			}
		}
	}
	require.Equal(t, count, actual, "Vary values: %v", values)
}
