// Package identity implements the profile's browser session and API-token REST
// extension. Credentials resolve to the same domain Principal used by gRPC.
package identity

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/logger"

	workflowhttp "netbox-go/internal/adapters/rest/netbox/workflow"
	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

const (
	sessionCookie = "netbox_session"
	csrfCookie    = "csrftoken"
)

type Handler struct {
	service       *application.Service
	secureCookies bool
	loginLimiter  *attemptLimiter
	tokenLimiter  *attemptLimiter
}

func NewHandler(service *application.Service, secureCookies bool) *Handler {
	if service == nil {
		panic("identity REST handler requires service")
	}
	return &Handler{service: service, secureCookies: secureCookies, loginLimiter: newAttemptLimiter(5, time.Minute), tokenLimiter: newAttemptLimiter(10, time.Minute)}
}

func (h *Handler) Register(r gin.IRoutes) {
	r.GET("/api/auth/csrf/", h.csrf)
	r.POST("/api/auth/login/", h.login)
	auth := h.Middleware()
	r.GET("/api/auth/session/", auth, h.session)
	r.POST("/api/auth/logout/", auth, h.logout)
	r.POST("/api/auth/password/change/", auth, h.changePassword)
	r.GET("/api/auth/tokens/", auth, h.listTokens)
	r.POST("/api/auth/tokens/", auth, h.createToken)
	r.DELETE("/api/auth/tokens/:id/", auth, h.revokeToken)
}

// Middleware authenticates a browser session or NetBox Token credential. It
// fails closed and applies CSRF only to cookie-authenticated unsafe requests.
func (h *Handler) Middleware() gin.HandlerFunc {
	return h.middleware(http.StatusUnauthorized)
}

// BaselineMiddleware preserves NetBox/DRF's observable 403 response for a
// missing or rejected credential on baseline API resources. The Go-owned
// identity extension keeps conventional 401 responses through Middleware.
func (h *Handler) BaselineMiddleware() gin.HandlerFunc {
	return h.middleware(http.StatusForbidden)
}

func (h *Handler) middleware(unauthenticatedStatus int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var user domain.User
		var err error
		sessionAuthenticated := false
		if authorization := c.GetHeader("Authorization"); strings.HasPrefix(authorization, "Token ") {
			write := c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead && c.Request.Method != http.MethodOptions
			user, err = h.service.AuthenticateToken(ctx, strings.TrimSpace(strings.TrimPrefix(authorization, "Token ")), c.Request.RemoteAddr, write)
		} else if authorization != "" {
			writeErrorStatus(c, unauthenticatedError(), unauthenticatedStatus)
			c.Abort()
			return
		} else if secret, cookieErr := c.Cookie(sessionCookie); cookieErr == nil {
			user, err = h.service.AuthenticateSession(ctx, secret)
			sessionAuthenticated = err == nil
		} else {
			err = unauthenticatedError()
		}
		if err != nil {
			writeErrorStatus(c, err, unauthenticatedStatus)
			c.Abort()
			return
		}
		if sessionAuthenticated && unsafe(c.Request.Method) {
			secret, _ := c.Cookie(sessionCookie)
			csrf := c.GetHeader("X-CSRFToken")
			if err := h.service.VerifyCSRF(ctx, secret, csrf); err != nil {
				writeError(c, err)
				c.Abort()
				return
			}
		}
		workflowhttp.SetPrincipal(c, user.Principal())
		c.Request = c.Request.WithContext(domain.WithPrincipal(ctx, user.Principal()))
		c.Next()
	}
}

func (h *Handler) csrf(c *gin.Context) {
	token, err := randomToken()
	if err != nil {
		writeError(c, shared.NewError(
			shared.ErrorReasonInternal,
			"An internal error occurred.",
		))
		return
	}
	h.setCSRFCookie(c, token)
	c.JSON(http.StatusOK, gin.H{"csrf_token": token})
}
func (h *Handler) login(c *gin.Context) {
	cookie, cookieErr := c.Cookie(csrfCookie)
	header := c.GetHeader("X-CSRFToken")
	if cookieErr != nil || cookie == "" || header == "" || subtle.ConstantTimeCompare([]byte(cookie), []byte(header)) != 1 {
		writeError(c, shared.NewError(
			shared.ErrorReasonForbidden,
			"You do not have permission to perform this action.",
		))
		return
	}
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, validationError("non_field_errors", "Expected username and password."))
		return
	}
	limitKey := remoteHost(c.Request.RemoteAddr) + "|" + strings.ToLower(input.Username)
	if !h.loginLimiter.allow(limitKey, time.Now()) {
		c.JSON(http.StatusTooManyRequests, gin.H{"detail": "Too many login attempts. Try again later."})
		return
	}
	existingSession, _ := c.Cookie(sessionCookie)
	session, err := h.service.LoginReplacing(c.Request.Context(), input.Username, input.Password, existingSession)
	if err != nil {
		logger.Warn("identity login rejected", logger.String("remote", c.Request.RemoteAddr))
		writeError(c, err)
		return
	}
	h.loginLimiter.reset(limitKey)
	logger.Info("identity login succeeded", logger.Int64("userID", session.User.ID))
	h.setSessionCookie(c, session.Secret, session.Expires)
	h.setCSRFCookie(c, session.CSRFToken)
	c.JSON(http.StatusOK, sessionResponse(session.User))
}
func (h *Handler) session(c *gin.Context) {
	principal, _ := workflowhttp.Principal(c)
	user, err := h.service.CurrentUser(c.Request.Context(), principal)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, sessionResponse(user))
}
func (h *Handler) logout(c *gin.Context) {
	secret, _ := c.Cookie(sessionCookie)
	if err := h.service.Logout(c.Request.Context(), secret); err != nil {
		writeError(c, err)
		return
	}
	h.clearCookies(c)
	c.Status(http.StatusNoContent)
}
func (h *Handler) changePassword(c *gin.Context) {
	principal, _ := workflowhttp.Principal(c)
	var input struct {
		Current string `json:"current_password"`
		Next    string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, validationError(
			"non_field_errors",
			"Expected current_password and new_password.",
		))
		return
	}
	if err := h.service.ChangePassword(c.Request.Context(), principal, input.Current, input.Next); err != nil {
		writeError(c, err)
		return
	}
	logger.Info("identity password changed", logger.Int64("userID", principal.ID))
	h.clearCookies(c)
	c.Status(http.StatusNoContent)
}
func (h *Handler) listTokens(c *gin.Context) {
	principal, _ := workflowhttp.Principal(c)
	limit, offset := parsePage(c)
	tokens, count, err := h.service.ListTokens(c.Request.Context(), principal, limit, offset)
	if err != nil {
		writeError(c, err)
		return
	}
	results := make([]gin.H, 0, len(tokens))
	for _, token := range tokens {
		results = append(results, tokenDTO(token))
	}
	c.JSON(http.StatusOK, gin.H{"count": count, "next": nil, "previous": nil, "results": results})
}
func (h *Handler) createToken(c *gin.Context) {
	principal, _ := workflowhttp.Principal(c)
	if !h.tokenLimiter.allow(strconv.FormatInt(principal.ID, 10), time.Now()) {
		c.JSON(http.StatusTooManyRequests, gin.H{"detail": "Too many token creation attempts. Try again later."})
		return
	}
	var input struct {
		Description  string     `json:"description"`
		WriteEnabled bool       `json:"write_enabled"`
		Expires      *time.Time `json:"expires"`
		AllowedIPs   []string   `json:"allowed_ips"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, validationError("non_field_errors", "Invalid token request."))
		return
	}
	created, err := h.service.CreateToken(c.Request.Context(), principal, application.CreateTokenInput{Description: input.Description, WriteEnabled: input.WriteEnabled, Expires: input.Expires, AllowedIPs: input.AllowedIPs})
	if err != nil {
		writeError(c, err)
		return
	}
	logger.Info("identity API token created", logger.Int64("userID", principal.ID), logger.Int64("tokenID", created.Token.ID))
	response := tokenDTO(created.Token)
	response["secret"] = created.Secret
	c.JSON(http.StatusCreated, response)
}
func (h *Handler) revokeToken(c *gin.Context) {
	principal, _ := workflowhttp.Principal(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, validationError("id", "A positive integer is required."))
		return
	}
	if err := h.service.RevokeToken(c.Request.Context(), principal, id); err != nil {
		writeError(c, err)
		return
	}
	logger.Info("identity API token revoked", logger.Int64("userID", principal.ID), logger.Int64("tokenID", id))
	c.Status(http.StatusNoContent)
}

func sessionResponse(user domain.User) gin.H {
	return gin.H{"user": gin.H{"id": user.ID, "username": user.Username, "email": user.Email, "first_name": user.FirstName, "last_name": user.LastName, "is_staff": user.IsStaff, "is_superuser": user.IsSuperuser}, "permissions": user.Permissions}
}
func tokenDTO(token domain.APIToken) gin.H {
	return gin.H{"id": token.ID, "display": token.Display, "description": token.Description, "write_enabled": token.WriteEnabled, "created": token.Created, "expires": token.Expires, "last_used": token.LastUsed, "allowed_ips": token.AllowedIPs}
}
func parsePage(c *gin.Context) (int, int) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	return limit, offset
}
func unsafe(method string) bool {
	return method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions
}
func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
func (h *Handler) setSessionCookie(c *gin.Context, value string, expires time.Time) {
	http.SetCookie(c.Writer, &http.Cookie{Name: sessionCookie, Value: value, Path: "/", HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, Expires: expires, MaxAge: int(time.Until(expires).Seconds())})
}
func (h *Handler) setCSRFCookie(c *gin.Context, value string) {
	http.SetCookie(c.Writer, &http.Cookie{Name: csrfCookie, Value: value, Path: "/", HttpOnly: false, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, MaxAge: int(sessionTTL().Seconds())})
}
func (h *Handler) clearCookies(c *gin.Context) {
	for _, name := range []string{sessionCookie, csrfCookie} {
		http.SetCookie(c.Writer, &http.Cookie{Name: name, Value: "", Path: "/", HttpOnly: name == sessionCookie, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0)})
	}
}
func sessionTTL() time.Duration { return 12 * time.Hour }
func remoteHost(address string) string {
	if host, _, err := net.SplitHostPort(address); err == nil {
		return host
	}
	return address
}
func writeError(c *gin.Context, err error) {
	writeErrorStatus(c, err, http.StatusUnauthorized)
}

func writeErrorStatus(c *gin.Context, err error, unauthenticatedStatus int) {
	var appErr *shared.Error
	if !errors.As(err, &appErr) {
		appErr = shared.NewError(
			shared.ErrorReasonInternal,
			"An internal error occurred.",
		)
	}
	status := map[shared.ErrorReason]int{
		shared.ErrorReasonValidation:      http.StatusBadRequest,
		shared.ErrorReasonUnauthenticated: unauthenticatedStatus,
		shared.ErrorReasonForbidden:       http.StatusForbidden,
		shared.ErrorReasonNotFound:        http.StatusNotFound,
		shared.ErrorReasonConflict:        http.StatusBadRequest,
		shared.ErrorReasonProtected:       http.StatusConflict,
		shared.ErrorReasonRateLimited:     http.StatusTooManyRequests,
		shared.ErrorReasonInternal:        http.StatusInternalServerError,
	}[appErr.Reason]
	if status == 0 {
		status = http.StatusInternalServerError
	}
	if len(appErr.FieldViolations) > 0 {
		fields := make(map[string][]string)
		for _, violation := range appErr.FieldViolations {
			fields[violation.Field] = append(fields[violation.Field], violation.Description)
		}
		c.JSON(status, fields)
		return
	}
	c.JSON(status, gin.H{"detail": appErr.Message})
}

func unauthenticatedError() error {
	return shared.NewError(
		shared.ErrorReasonUnauthenticated,
		"Authentication credentials were not provided.",
	)
}

func validationError(field, description string) error {
	return shared.NewValidationError(shared.FieldViolation{
		Field: field, Reason: "invalid", Description: description,
	})
}
