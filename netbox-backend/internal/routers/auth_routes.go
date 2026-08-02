package routers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"golang.org/x/crypto/bcrypt"

	"netbox-go/internal/database"
	"netbox-go/internal/handler"
	"netbox-go/internal/model"
)

// registerAuthRoutes registers the authentication endpoints.
func registerAuthRoutes(r *gin.Engine, jwtCfg *handler.JWTConfig) {
	r.POST("/api/auth/login", makeLoginHandler(jwtCfg))
	r.POST("/api/auth/provision", makeProvisionHandler())
	r.GET("/api/users/tokens", makeTokenListHandler())
	r.POST("/api/users/tokens", makeTokenCreateHandler())
}

// makeLoginHandler authenticates a user with username/password and returns a JWT.
func makeLoginHandler(jwtCfg *handler.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		db := database.GetDB()
		var user model.UsersUser
		if err := db.Where("username = ?", body.Username).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid credentials."})
			return
		}

		// Check user is active
		if user.IsActive == nil || !bool(*user.IsActive) {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "User account is disabled."})
			return
		}

		// Verify password (Django uses bcrypt for password hashing)
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid credentials."})
			return
		}

		// If JWT is configured, return a JWT token
		if jwtCfg != nil {
			token, err := handler.GenerateJWT(&user, jwtCfg)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to generate token."})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"token":   token,
				"key":     token,
				"user_id": user.ID,
			})
			return
		}

		// Fallback: no JWT configured, just return user info
		c.JSON(http.StatusOK, gin.H{
			"user_id":  user.ID,
			"username": user.Username,
		})
	}
}

// makeProvisionHandler generates and returns a new API token for the authenticated user.
func makeProvisionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := handler.CurrentUserID(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Authentication required."})
			return
		}

		// Generate a 40-character token (like Django's token generation)
		tokenBytes := make([]byte, 20)
		if _, err := rand.Read(tokenBytes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to generate token."})
			return
		}
		tokenKey := hex.EncodeToString(tokenBytes)

		db := database.GetDB()
		now := time.Now()
		writeEnabled := sgorm.Bool(true)
		token := model.UsersToken{
			Created:      &now,
			Key:          tokenKey,
			WriteEnabled: &writeEnabled,
			Description:  "API token",
			UserID:       int64(userID),
		}
		if err := db.Create(&token).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create token."})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":          token.ID,
			"key":         tokenKey,
			"description": token.Description,
			"user_id":     userID,
		})
	}
}

// makeTokenListHandler returns tokens for the authenticated user.
func makeTokenListHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := handler.CurrentUserID(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Authentication required."})
			return
		}

		db := database.GetDB()
		var tokens []model.UsersToken
		db.Where("user_id = ?", userID).Find(&tokens)

		c.JSON(http.StatusOK, gin.H{
			"count":   len(tokens),
			"results": tokens,
		})
	}
}

// makeTokenCreateHandler creates a new token for the authenticated user.
func makeTokenCreateHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := handler.CurrentUserID(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Authentication required."})
			return
		}

		var body struct {
			Description string `json:"description"`
		}
		_ = c.ShouldBindJSON(&body)

		tokenBytes := make([]byte, 20)
		if _, err := rand.Read(tokenBytes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to generate token."})
			return
		}
		tokenKey := hex.EncodeToString(tokenBytes)

		db := database.GetDB()
		now := time.Now()
		writeEnabled := sgorm.Bool(true)
		token := model.UsersToken{
			Created:      &now,
			Key:          tokenKey,
			WriteEnabled: &writeEnabled,
			Description:  body.Description,
			UserID:       int64(userID),
		}
		if err := db.Create(&token).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create token."})
			return
		}

		c.JSON(http.StatusCreated, token)
	}
}
