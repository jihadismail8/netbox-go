package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGenerateAndValidateJWT(t *testing.T) {
	cfg := DefaultJWTConfig("test-secret-key")

	// Generate JWT
	user := &NetBoxClaims{
		UserID:      42,
		Username:    "testuser",
		IsSuperuser: true,
		IsStaff:     false,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user)
	tokenStr, err := token.SignedString([]byte(cfg.SecretKey))
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	// Validate
	uid, username, isSuper, write := validateJWT(tokenStr, cfg)
	assert.Equal(t, uint64(42), uid)
	assert.Equal(t, "testuser", username)
	assert.True(t, isSuper)
	assert.True(t, write)
}

func TestValidateJWT_BadSecret(t *testing.T) {
	cfg := DefaultJWTConfig("correct-secret")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, NetBoxClaims{
		UserID:   1,
		Username: "user",
	})
	tokenStr, _ := token.SignedString([]byte("wrong-secret"))

	uid, _, _, _ := validateJWT(tokenStr, cfg)
	assert.Equal(t, uint64(0), uid, "should fail with wrong secret")
}

func TestNetBoxAuthMiddleware_NoAuth(t *testing.T) {
	r := gin.New()
	r.Use(NetBoxAuthMiddleware(nil))
	handlerCalled := false
	r.GET("/test", func(c *gin.Context) {
		handlerCalled = true
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authentication credentials were not provided")
	assert.False(t, handlerCalled)
}

func TestNetBoxAuthMiddleware_MalformedTokenFailsClosed(t *testing.T) {
	r := gin.New()
	r.Use(NetBoxAuthMiddleware(nil))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Token too-short")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")
}

func TestNetBoxAuthMiddleware_BearerDisabled(t *testing.T) {
	r := gin.New()
	r.Use(NetBoxAuthMiddleware(nil))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer arbitrary")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "JWT authentication is not enabled")
}

func TestNetBoxAuthMiddleware_BadHeader(t *testing.T) {
	r := gin.New()
	r.Use(NetBoxAuthMiddleware(nil))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic abc123")
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid Authorization header")
}

func TestNetBoxAuthMiddleware_JWT(t *testing.T) {
	cfg := DefaultJWTConfig("test-jwt-secret")

	// Generate valid JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, NetBoxClaims{
		UserID:      7,
		Username:    "jwtuser",
		IsSuperuser: true,
	})
	tokenStr, _ := token.SignedString([]byte(cfg.SecretKey))

	r := gin.New()
	r.Use(NetBoxAuthMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		uid := CurrentUserID(c)
		username := CurrentUsername(c)
		c.JSON(200, gin.H{"uid": uid, "username": username})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"uid":7`)
	assert.Contains(t, w.Body.String(), `"username":"jwtuser"`)
}

func TestNetBoxAuthMiddleware_SkipPaths(t *testing.T) {
	r := gin.New()
	r.Use(NetBoxAuthMiddleware(nil, "/public"))
	r.GET("/public", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/public", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestRequireAuth(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		// Simulate no auth
		c.Set(string(CtxKeyUserID), uint64(0))
		c.Next()
	})
	r.Use(RequireAuth())
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestRequireSuperuser_NotSuperuser(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(CtxKeyUserID), uint64(1))
		c.Set(string(CtxKeyIsSuperuser), false)
		c.Next()
	})
	r.Use(RequireSuperuser())
	r.GET("/admin", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}
