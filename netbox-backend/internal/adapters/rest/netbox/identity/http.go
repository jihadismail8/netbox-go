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
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/logger"

	workflowhttp "netbox-go/internal/adapters/rest/netbox/workflow"
	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

const (
	sessionCookie                      = "sessionid"
	csrfCookie                         = "csrftoken"
	logoutCredentialsContextKey        = "netbox-go.identity.logout-credentials"
	passwordChangeCredentialContextKey = "netbox-go.identity.password-change-credential"
	cookieMaxAgeSeconds                = int(application.BrowserSessionLifetime / time.Second)
)

type logoutCredentials struct {
	sessionSecret string
	csrf          string
}

type passwordChangeCredentialKind uint8

const (
	passwordChangeCredentialBrowserSession passwordChangeCredentialKind = iota + 1
	passwordChangeCredentialAPIToken
)

type passwordChangeCredentials struct {
	kind          passwordChangeCredentialKind
	sessionSecret string
	csrf          string
}

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
	r.POST("/api/auth/logout/", h.logoutMiddleware(), h.logout)
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
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		sessionSecret, sessionPresent, duplicateSession := uniqueCookieValue(c.Request, sessionCookie)
		if duplicateSession {
			writeBaselineTokenDetail(c, "Authentication credentials were not provided.")
			c.Abort()
			return
		}
		if sessionPresent && sessionSecret != "" {
			user, err := h.service.AuthenticateSession(ctx, sessionSecret)
			if err == nil {
				if !sessionCSRFSafe(c.Request.Method) {
					csrf, accepted := exactCSRFPair(c.Request)
					if !accepted {
						writeCSRFRejected(c)
						c.Abort()
						return
					}
					if err := h.service.VerifyCSRF(ctx, sessionSecret, csrf); err != nil {
						writeErrorStatus(c, err, http.StatusForbidden)
						c.Abort()
						return
					}
				}
				setPrincipal(c, user)
				c.Next()
				return
			}
			if !application.SessionCredentialAllowsTokenFallback(err) {
				writeErrorStatus(c, err, http.StatusForbidden)
				c.Abort()
				return
			}
		}

		var user domain.User
		var err error
		authorizationValues := c.Request.Header.Values("Authorization")
		hasAuthorization := len(authorizationValues) > 1 ||
			(len(authorizationValues) == 1 && authorizationValues[0] != "")
		if hasAuthorization {
			secret, detail, accepted := parseBaselineTokenAuthorization(authorizationValues)
			if !accepted {
				writeBaselineTokenDetail(c, detail)
				c.Abort()
				return
			}
			user, err = h.service.AuthenticateToken(
				ctx,
				secret,
				c.Request.RemoteAddr,
				tokenWrite(c.Request.Method),
			)
		} else {
			err = unauthenticatedError()
		}

		if err != nil {
			writeBaselineTokenError(c, err)
			c.Abort()
			return
		}
		setPrincipal(c, user)
		c.Next()
	}
}

func parseBaselineTokenAuthorization(values []string) (string, string, bool) {
	if len(values) != 1 {
		return "", "Authentication credentials were not provided.", false
	}
	fields := splitPythonByteWhitespace(values[0])
	if len(fields) == 0 || !equalFoldASCII(fields[0], "Token") {
		return "", "Authentication credentials were not provided.", false
	}
	if len(fields) == 1 {
		return "", "Invalid token header. No credentials provided.", false
	}
	if len(fields) > 2 {
		return "", "Invalid token header. Token string should not contain spaces.", false
	}
	if !utf8.Valid(fields[1]) {
		return "", "Invalid token header. Token string should not contain invalid characters.", false
	}
	return string(fields[1]), "", true
}

func splitPythonByteWhitespace(value string) [][]byte {
	raw := []byte(value)
	fields := make([][]byte, 0, 2)
	for index := 0; index < len(raw); {
		for index < len(raw) && pythonByteWhitespace(raw[index]) {
			index++
		}
		start := index
		for index < len(raw) && !pythonByteWhitespace(raw[index]) {
			index++
		}
		if start < index {
			fields = append(fields, raw[start:index])
		}
	}
	return fields
}

func pythonByteWhitespace(value byte) bool {
	return value == ' ' || value >= '\t' && value <= '\r'
}

func equalFoldASCII(value []byte, expected string) bool {
	if len(value) != len(expected) {
		return false
	}
	for index, character := range value {
		if character >= 'A' && character <= 'Z' {
			character += 'a' - 'A'
		}
		expectedCharacter := expected[index]
		if expectedCharacter >= 'A' && expectedCharacter <= 'Z' {
			expectedCharacter += 'a' - 'A'
		}
		if character != expectedCharacter {
			return false
		}
	}
	return true
}

func writeBaselineTokenError(c *gin.Context, err error) {
	var failure *application.TokenCredentialFailure
	if shared.ReasonOf(err) == shared.ErrorReasonUnauthenticated && errors.As(err, &failure) {
		detail := map[application.TokenCredentialFailureKind]string{
			application.TokenCredentialFailureMissing:           "Authentication credentials were not provided.",
			application.TokenCredentialFailureUnknown:           "Invalid token",
			application.TokenCredentialFailureRevoked:           "Invalid token",
			application.TokenCredentialFailureExpired:           "Token expired",
			application.TokenCredentialFailureInactiveOwner:     "User inactive",
			application.TokenCredentialFailureSourceUnavailable: "Client IP address could not be determined for validation. Check that the HTTP server is correctly configured to pass the required header(s).",
			application.TokenCredentialFailureSourceDenied:      "Source IP " + failure.SourceIP + " is not permitted to authenticate using this token.",
		}[failure.Kind]
		if detail != "" {
			writeBaselineTokenDetail(c, detail)
			return
		}
	}
	writeErrorStatus(c, err, http.StatusForbidden)
}

func writeBaselineTokenDetail(c *gin.Context, detail string) {
	c.JSON(http.StatusForbidden, gin.H{"detail": detail})
}

func (h *Handler) middleware(unauthenticatedStatus int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		sessionSecret, sessionPresent, duplicateSession := uniqueCookieValue(c.Request, sessionCookie)
		if duplicateSession {
			writeErrorStatus(c, unauthenticatedError(), unauthenticatedStatus)
			c.Abort()
			return
		}
		if sessionPresent && sessionSecret != "" {
			user, err := h.service.AuthenticateSession(ctx, sessionSecret)
			if err == nil {
				var passwordCredentials passwordChangeCredentials
				if !sessionCSRFSafe(c.Request.Method) {
					csrf, accepted := exactCSRFPair(c.Request)
					if !accepted {
						writeCSRFRejected(c)
						c.Abort()
						return
					}
					if err := h.service.VerifyCSRF(ctx, sessionSecret, csrf); err != nil {
						writeErrorStatus(c, err, http.StatusForbidden)
						c.Abort()
						return
					}
					passwordCredentials = passwordChangeCredentials{
						kind:          passwordChangeCredentialBrowserSession,
						sessionSecret: sessionSecret,
						csrf:          csrf,
					}
				}
				if passwordCredentials.kind != 0 {
					c.Set(passwordChangeCredentialContextKey, passwordCredentials)
				}
				setPrincipal(c, user)
				c.Next()
				return
			}
			if !application.SessionCredentialAllowsTokenFallback(err) {
				writeErrorStatus(c, err, unauthenticatedStatus)
				c.Abort()
				return
			}
		}

		var user domain.User
		var err error
		if authorization := c.GetHeader("Authorization"); strings.HasPrefix(authorization, "Token ") {
			user, err = h.service.AuthenticateToken(ctx, strings.TrimSpace(strings.TrimPrefix(authorization, "Token ")), c.Request.RemoteAddr, tokenWrite(c.Request.Method))
		} else if authorization != "" {
			writeErrorStatus(c, unauthenticatedError(), unauthenticatedStatus)
			c.Abort()
			return
		} else {
			err = unauthenticatedError()
		}
		if err != nil {
			writeErrorStatus(c, err, unauthenticatedStatus)
			c.Abort()
			return
		}
		c.Set(passwordChangeCredentialContextKey, passwordChangeCredentials{
			kind: passwordChangeCredentialAPIToken,
		})
		setPrincipal(c, user)
		c.Next()
	}
}

func (h *Handler) csrf(c *gin.Context) {
	sessionSecret, sessionPresent, duplicateSession := uniqueCookieValue(c.Request, sessionCookie)
	if duplicateSession {
		writeCSRFRejected(c)
		return
	}
	if sessionPresent && sessionSecret != "" {
		token, err := h.service.CSRFForSession(c.Request.Context(), sessionSecret)
		if err == nil {
			h.writeCSRFBootstrap(c, token)
			return
		}
		if !application.SessionCredentialAllowsTokenFallback(err) {
			writeError(c, err)
			return
		}
	}

	token, err := randomToken()
	if err != nil {
		writeError(c, shared.NewError(
			shared.ErrorReasonInternal,
			"An internal error occurred.",
		))
		return
	}
	h.writeCSRFBootstrap(c, token)
}

func (h *Handler) writeCSRFBootstrap(c *gin.Context, token string) {
	h.setCSRFCookie(c, token)
	c.JSON(http.StatusOK, gin.H{"csrf_token": token})
}

func (h *Handler) login(c *gin.Context) {
	existingSession, sessionPresent, duplicateSession := uniqueCookieValue(c.Request, sessionCookie)
	if duplicateSession {
		writeCSRFRejected(c)
		return
	}
	if !sessionPresent {
		existingSession = ""
	}
	if _, accepted := exactCSRFPair(c.Request); !accepted {
		writeCSRFRejected(c)
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

func (h *Handler) logoutMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionSecret, sessionPresent, duplicateSession := uniqueCookieValue(c.Request, sessionCookie)
		if duplicateSession || !sessionPresent || sessionSecret == "" {
			writeErrorStatus(c, unauthenticatedError(), http.StatusUnauthorized)
			c.Abort()
			return
		}

		user, err := h.service.AuthenticateSession(c.Request.Context(), sessionSecret)
		if err != nil {
			writeErrorStatus(c, err, http.StatusUnauthorized)
			c.Abort()
			return
		}
		csrf, accepted := exactCSRFPair(c.Request)
		if !accepted {
			writeCSRFRejected(c)
			c.Abort()
			return
		}

		c.Set(logoutCredentialsContextKey, logoutCredentials{
			sessionSecret: sessionSecret,
			csrf:          csrf,
		})
		setPrincipal(c, user)
		c.Next()
	}
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
	value, exists := c.Get(logoutCredentialsContextKey)
	credentials, accepted := value.(logoutCredentials)
	if !exists || !accepted {
		writeErrorStatus(c, unauthenticatedError(), http.StatusUnauthorized)
		return
	}
	if err := h.service.Logout(c.Request.Context(), credentials.sessionSecret, credentials.csrf); err != nil {
		writeError(c, err)
		return
	}
	h.clearCookies(c)
	c.Status(http.StatusNoContent)
}
func (h *Handler) changePassword(c *gin.Context) {
	value, exists := c.Get(passwordChangeCredentialContextKey)
	credentials, accepted := value.(passwordChangeCredentials)
	if !exists || !accepted {
		writePasswordChangeAdapterError(c)
		return
	}

	var credential application.PasswordChangeCredential
	switch credentials.kind {
	case passwordChangeCredentialBrowserSession:
		if credentials.sessionSecret == "" || credentials.csrf == "" {
			writePasswordChangeAdapterError(c)
			return
		}
		credential = application.BrowserSessionPasswordChangeCredential(
			credentials.sessionSecret,
			credentials.csrf,
		)
	case passwordChangeCredentialAPIToken:
		if credentials.sessionSecret != "" || credentials.csrf != "" {
			writePasswordChangeAdapterError(c)
			return
		}
		credential = application.APITokenPasswordChangeCredential()
	default:
		writePasswordChangeAdapterError(c)
		return
	}

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
	result, err := h.service.ChangePassword(
		c.Request.Context(),
		principal,
		application.NewPasswordChangeInput(input.Current, input.Next, credential),
	)
	if err != nil {
		writeError(c, err)
		return
	}
	if result == nil {
		writePasswordChangeAdapterError(c)
		return
	}
	browserSession, hasBrowserSession := result.BrowserSession()
	switch credentials.kind {
	case passwordChangeCredentialBrowserSession:
		if !hasBrowserSession {
			writePasswordChangeAdapterError(c)
			return
		}
		h.setSessionCookie(c, browserSession.Secret, browserSession.Expires)
		h.setCSRFCookie(c, browserSession.CSRFToken)
	case passwordChangeCredentialAPIToken:
		if hasBrowserSession {
			writePasswordChangeAdapterError(c)
			return
		}
	}
	logger.Info("identity password changed", logger.Int64("userID", principal.ID))
	c.Status(http.StatusNoContent)
}

func writePasswordChangeAdapterError(c *gin.Context) {
	writeError(c, shared.NewError(
		shared.ErrorReasonInternal,
		"An internal error occurred.",
	))
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

func sessionCSRFSafe(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func tokenWrite(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

func uniqueCookieValue(request *http.Request, name string) (value string, present, duplicate bool) {
	for _, cookie := range request.Cookies() {
		if cookie.Name != name {
			continue
		}
		if present {
			return "", true, true
		}
		value = cookie.Value
		present = true
	}
	return value, present, false
}

func exactCSRFPair(request *http.Request) (string, bool) {
	cookie, present, duplicate := uniqueCookieValue(request, csrfCookie)
	headers := request.Header.Values("X-CSRFToken")
	if duplicate || !present || cookie == "" || len(headers) != 1 || headers[0] == "" {
		return "", false
	}
	if subtle.ConstantTimeCompare([]byte(cookie), []byte(headers[0])) != 1 {
		return "", false
	}
	return headers[0], true
}

func setPrincipal(c *gin.Context, user domain.User) {
	principal := user.Principal()
	workflowhttp.SetPrincipal(c, principal)
	c.Request = c.Request.WithContext(domain.WithPrincipal(c.Request.Context(), principal))
}

func writeCSRFRejected(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{
		"detail": "You do not have permission to perform this action.",
	})
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
func (h *Handler) setSessionCookie(c *gin.Context, value string, expires time.Time) {
	http.SetCookie(c.Writer, &http.Cookie{Name: sessionCookie, Value: value, Path: "/", HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, Expires: expires, MaxAge: cookieMaxAgeSeconds})
}
func (h *Handler) setCSRFCookie(c *gin.Context, value string) {
	http.SetCookie(c.Writer, &http.Cookie{Name: csrfCookie, Value: value, Path: "/", HttpOnly: false, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, MaxAge: cookieMaxAgeSeconds})
}
func (h *Handler) clearCookies(c *gin.Context) {
	for _, name := range []string{sessionCookie, csrfCookie} {
		http.SetCookie(c.Writer, &http.Cookie{Name: name, Value: "", Path: "/", HttpOnly: name == sessionCookie, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0)})
	}
}
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
