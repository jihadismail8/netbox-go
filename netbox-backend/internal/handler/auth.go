// Package handler provides authentication middleware for NetBox-compatible API.
package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"netbox-go/internal/database"
	"netbox-go/internal/model"
)

// contextKey is a private type for context keys in this package.
type contextKey string

const (
	// CtxKeyUserID is the context key for the authenticated user ID.
	CtxKeyUserID = contextKey("userID")
	// CtxKeyUsername is the context key for the authenticated username.
	CtxKeyUsername = contextKey("username")
	// CtxKeyIsSuperuser is the context key for the superuser flag.
	CtxKeyIsSuperuser = contextKey("isSuperuser")
	// CtxKeyTokenWrite is whether the token allows write operations.
	CtxKeyTokenWrite = contextKey("tokenWriteEnabled")

	// Header constants
	authHeader = "Authorization"
)

// JWTConfig holds JWT signing configuration.
type JWTConfig struct {
	SecretKey     string
	SigningMethod jwt.SigningMethod
}

// DefaultJWTConfig returns a default JWT config using HS256.
func DefaultJWTConfig(secret string) *JWTConfig {
	return &JWTConfig{
		SecretKey:     secret,
		SigningMethod: jwt.SigningMethodHS256,
	}
}

// NetBoxClaims defines the JWT claims structure matching NetBox's Token auth.
type NetBoxClaims struct {
	UserID      uint64 `json:"user_id"`
	Username    string `json:"username"`
	IsSuperuser bool   `json:"is_superuser"`
	IsStaff     bool   `json:"is_staff"`
	jwt.RegisteredClaims
}

// ---- Auth Modes ----

// AuthMode represents the authentication method that was used.
type AuthMode string

const (
	AuthModeNone  AuthMode = "none"
	AuthModeToken AuthMode = "token"
	AuthModeJWT   AuthMode = "jwt"
)

// ---- Middleware ----

// NetBoxAuthMiddleware supports both NetBox Token auth and JWT Bearer auth.
// If jwtConfig is nil, JWT auth is disabled.
// Public paths in skipPaths bypass authentication entirely.
func NetBoxAuthMiddleware(jwtConfig *JWTConfig, skipPaths ...string) gin.HandlerFunc {
	skipSet := make(map[string]bool, len(skipPaths))
	for _, p := range skipPaths {
		skipSet[p] = true
	}

	return func(c *gin.Context) {
		// Skip auth for public paths
		if skipSet[c.Request.URL.Path] {
			c.Next()
			return
		}

		authHeaderVal := c.GetHeader(authHeader)
		if authHeaderVal == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"detail": "Authentication credentials were not provided.",
			})
			return
		}

		var mode AuthMode
		var userID uint64
		var username string
		var isSuperuser bool
		var writeEnabled bool

		switch {
		case strings.HasPrefix(authHeaderVal, "Token "):
			mode = AuthModeToken
			tokenKey := strings.TrimPrefix(authHeaderVal, "Token ")
			userID, username, isSuperuser, writeEnabled = validateToken(c.Request.Context(), tokenKey)
		case strings.HasPrefix(authHeaderVal, "Bearer "):
			mode = AuthModeJWT
			if jwtConfig == nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "JWT authentication is not enabled."})
				return
			}
			tokenStr := strings.TrimPrefix(authHeaderVal, "Bearer ")
			userID, username, isSuperuser, writeEnabled = validateJWT(tokenStr, jwtConfig)
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Invalid Authorization header. Expected 'Token <key>' or 'Bearer <jwt>'."})
			return
		}

		_ = mode // mode is available for logging/auditing

		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Invalid token."})
			return
		}

		// Enforce write-enabled for non-GET requests when using Token auth
		if mode == AuthModeToken && !writeEnabled && c.Request.Method != http.MethodGet {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"detail": "This token does not allow write operations."})
			return
		}

		c.Set(string(CtxKeyUserID), userID)
		c.Set(string(CtxKeyUsername), username)
		c.Set(string(CtxKeyIsSuperuser), isSuperuser)
		c.Set(string(CtxKeyTokenWrite), writeEnabled)
		c.Next()
	}
}

// validateToken looks up a NetBox API token in the database.
func validateToken(ctx context.Context, key string) (userID uint64, username string, isSuperuser bool, writeEnabled bool) {
	if len(key) != 40 {
		return 0, "", false, false
	}

	db := database.GetDB()
	if db == nil {
		return 0, "", false, false
	}

	var token model.UsersToken
	if err := db.Where("key = ?", key).First(&token).Error; err != nil {
		return 0, "", false, false
	}

	// Check expiration
	if token.Expires != nil && token.Expires.Before(time.Now()) {
		return 0, "", false, false
	}

	// Fetch user
	var user model.UsersUser
	if err := db.Where("id = ?", token.UserID).First(&user).Error; err != nil {
		return 0, "", false, false
	}

	// Check user is active
	if user.IsActive == nil || !bool(*user.IsActive) {
		return 0, "", false, false
	}

	// Update last_used (fire-and-forget)
	go func() {
		now := time.Now()
		db.Model(&model.UsersToken{}).Where("id = ?", token.ID).Update("last_used", now)
	}()

	writeEnabled = true
	if token.WriteEnabled != nil {
		writeEnabled = bool(*token.WriteEnabled)
	}
	isSuperuser = user.IsSuperuser != nil && bool(*user.IsSuperuser)

	return user.ID, user.Username, isSuperuser, writeEnabled
}

// validateJWT parses and validates a JWT token.
func validateJWT(tokenStr string, cfg *JWTConfig) (userID uint64, username string, isSuperuser bool, writeEnabled bool) {
	claims := &NetBoxClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != cfg.SigningMethod.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(cfg.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return 0, "", false, false
	}

	// JWT tokens allow write by default
	return claims.UserID, claims.Username, claims.IsSuperuser, true
}

// GenerateJWT creates a signed JWT for a given user.
func GenerateJWT(user *model.UsersUser, cfg *JWTConfig) (string, error) {
	claims := NetBoxClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}
	if user.IsSuperuser != nil {
		claims.IsSuperuser = bool(*user.IsSuperuser)
	}
	if user.IsStaff != nil {
		claims.IsStaff = bool(*user.IsStaff)
	}

	token := jwt.NewWithClaims(cfg.SigningMethod, claims)
	return token.SignedString([]byte(cfg.SecretKey))
}

// CurrentUserID extracts the authenticated user ID from the gin context.
func CurrentUserID(c *gin.Context) uint64 {
	if v, exists := c.Get(string(CtxKeyUserID)); exists {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}

// CurrentUsername extracts the authenticated username from the gin context.
func CurrentUsername(c *gin.Context) string {
	if v, exists := c.Get(string(CtxKeyUsername)); exists {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// RequireAuth is a middleware that aborts if no user is authenticated.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if CurrentUserID(c) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Authentication credentials were not provided."})
			return
		}
		c.Next()
	}
}

// RequireWriteEnabled aborts if the token does not allow writes.
func RequireWriteEnabled() gin.HandlerFunc {
	return func(c *gin.Context) {
		if v, exists := c.Get(string(CtxKeyTokenWrite)); exists {
			if write, ok := v.(bool); ok && !write {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"detail": "This token does not allow write operations."})
				return
			}
		}
		c.Next()
	}
}

// RequireSuperuser aborts if the user is not a superuser.
func RequireSuperuser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if v, exists := c.Get(string(CtxKeyIsSuperuser)); exists {
			if super, ok := v.(bool); ok && super {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"detail": "You do not have permission to perform this action."})
	}
}
