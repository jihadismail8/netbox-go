// Package identity implements the shared identity lifecycle used by browser,
// REST-token, administrator CLI, and gRPC authentication.
package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	domain "netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

const (
	passwordCost = 12
	sessionTTL   = 12 * time.Hour
)

var ErrNotFound = errors.New("identity record not found")

type TokenRecord struct {
	Token      domain.APIToken
	SecretHash []byte
	RevokedAt  *time.Time
}
type SessionRecord struct {
	SecretHash, CSRFHash       []byte
	UserID                     int64
	Created, Expires, LastSeen time.Time
}

type Store interface {
	Transaction(context.Context, func(Store) error) error
	CreateUser(context.Context, domain.User, string) (domain.User, error)
	CreateGroup(context.Context, domain.Group) (domain.Group, error)
	AddGroupMember(context.Context, int64, int64) error
	CreatePermissionGrant(context.Context, domain.PermissionGrant) (domain.PermissionGrant, error)
	GrantPermissionToUser(context.Context, int64, int64) error
	GrantPermissionToGroup(context.Context, int64, int64) error
	UserByUsername(context.Context, string) (domain.User, string, error)
	UserByID(context.Context, int64) (domain.User, string, error)
	UpdatePassword(context.Context, int64, string) error
	CreateSession(context.Context, SessionRecord) error
	SessionByHash(context.Context, []byte) (SessionRecord, error)
	DeleteSession(context.Context, []byte) error
	DeleteSessionsForUser(context.Context, int64) error
	CreateToken(context.Context, TokenRecord) (domain.APIToken, error)
	TokenByHash(context.Context, []byte) (TokenRecord, domain.User, error)
	ListTokens(context.Context, int64, int, int) ([]domain.APIToken, int64, error)
	RevokeToken(context.Context, int64, int64, time.Time) error
	TouchToken(context.Context, int64, time.Time) error
	CountUsers(context.Context) (int64, error)
}

type Clock interface{ Now() time.Time }
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now().UTC() }

type Service struct {
	store         Store
	clock         Clock
	mu            sync.Mutex
	tokenAttempts map[int64][]time.Time
}

func NewService(store Store, clock Clock) *Service {
	if store == nil || clock == nil {
		panic("identity service requires store and clock")
	}
	return &Service{store: store, clock: clock, tokenAttempts: map[int64][]time.Time{}}
}

type CreateTokenInput struct {
	Description  string
	WriteEnabled bool
	Expires      *time.Time
	AllowedIPs   []string
}
type CreatedToken struct {
	Token  domain.APIToken
	Secret string
}

type CreateUserInput struct {
	Username, Email, FirstName, LastName, Password string
	IsStaff                                        bool
}

type PermissionGrantInput struct {
	Name, AppLabel, Action, Model string
	ObjectID                      *int64
}

func (s *Service) AuthenticatePassword(ctx context.Context, username, password string) (domain.User, error) {
	if username == "" || password == "" {
		return domain.User{}, unauthenticated()
	}
	user, hash, err := s.store.UserByUsername(ctx, username)
	if err != nil { // use a fixed hash to reduce username timing disclosure
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$12$8D3PZc7R7r6BfZtH1SgTSuU72fHoLXCB8xlQ9jJ3TQ0K9h3wCw0kW"), []byte(password))
		return domain.User{}, unauthenticated()
	}
	if !user.IsActive || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return domain.User{}, unauthenticated()
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (domain.BrowserSession, error) {
	return s.LoginReplacing(ctx, username, password, "")
}

// LoginReplacing creates fresh, server-generated session material and removes
// any session presented during re-authentication in the same transaction. It
// is the browser adapter's session-fixation boundary.
func (s *Service) LoginReplacing(ctx context.Context, username, password, existingSecret string) (domain.BrowserSession, error) {
	user, err := s.AuthenticatePassword(ctx, username, password)
	if err != nil {
		return domain.BrowserSession{}, err
	}
	secret, secretHash, err := newSecret()
	if err != nil {
		return domain.BrowserSession{}, internal(err)
	}
	csrf, csrfHash, err := newSecret()
	if err != nil {
		return domain.BrowserSession{}, internal(err)
	}
	now := s.clock.Now()
	session := SessionRecord{SecretHash: secretHash, CSRFHash: csrfHash, UserID: user.ID, Created: now, LastSeen: now, Expires: now.Add(sessionTTL)}
	persist := func(store Store) error {
		if existingSecret != "" {
			if deleteErr := store.DeleteSession(ctx, digest(existingSecret)); deleteErr != nil && !errors.Is(deleteErr, ErrNotFound) {
				return deleteErr
			}
		}
		return store.CreateSession(ctx, session)
	}
	if existingSecret == "" {
		err = persist(s.store)
	} else {
		err = s.store.Transaction(ctx, persist)
	}
	if err != nil {
		return domain.BrowserSession{}, internal(err)
	}
	return domain.BrowserSession{User: user, Secret: secret, CSRFToken: csrf, Expires: session.Expires}, nil
}

func (s *Service) AuthenticateSession(ctx context.Context, secret string) (domain.User, error) {
	if secret == "" {
		return domain.User{}, unauthenticated()
	}
	record, err := s.store.SessionByHash(ctx, digest(secret))
	if err != nil || !record.Expires.After(s.clock.Now()) {
		return domain.User{}, unauthenticated()
	}
	user, _, err := s.store.UserByID(ctx, record.UserID)
	if err != nil || !user.IsActive {
		return domain.User{}, unauthenticated()
	}
	return user, nil
}
func (s *Service) VerifyCSRF(ctx context.Context, sessionSecret, csrf string) error {
	record, err := s.store.SessionByHash(ctx, digest(sessionSecret))
	if err != nil || !record.Expires.After(s.clock.Now()) {
		return unauthenticated()
	}
	candidate := digest(csrf)
	if subtle.ConstantTimeCompare(candidate, record.CSRFHash) != 1 {
		return forbidden()
	}
	return nil
}
func (s *Service) Logout(ctx context.Context, secret string) error {
	if secret == "" {
		return nil
	}
	if err := s.store.DeleteSession(ctx, digest(secret)); err != nil && !errors.Is(err, ErrNotFound) {
		return internal(err)
	}
	return nil
}

func (s *Service) AuthenticateToken(ctx context.Context, secret, remoteAddress string, write bool) (domain.User, error) {
	if secret == "" {
		return domain.User{}, unauthenticated()
	}
	record, user, err := s.store.TokenByHash(ctx, digest(secret))
	if errors.Is(err, ErrNotFound) {
		return domain.User{}, unauthenticated()
	}
	if err != nil {
		return domain.User{}, internal(err)
	}
	// Baseline revocation deletes the token row. The Go-owned soft-revocation
	// extension must therefore behave like an unknown key and never mutate the
	// credential after revocation.
	if record.RevokedAt != nil {
		return domain.User{}, unauthenticated()
	}
	now := s.clock.Now()
	// Match the baseline ordering: a recognized key is touched at most once per
	// minute before expiry, active-user, and allowed-IP rejection.
	if record.Token.LastUsed == nil || now.Sub(*record.Token.LastUsed) > time.Minute {
		if err := s.store.TouchToken(ctx, record.Token.ID, now); err != nil {
			return domain.User{}, internal(err)
		}
	}
	if (record.Token.Expires != nil && !record.Token.Expires.After(now)) || !user.IsActive {
		return domain.User{}, unauthenticated()
	}
	if len(record.Token.AllowedIPs) > 0 && !allowedAddress(remoteAddress, record.Token.AllowedIPs) {
		return domain.User{}, unauthenticated()
	}
	if write && !record.Token.WriteEnabled {
		return domain.User{}, forbidden()
	}
	return user, nil
}

func (s *Service) CurrentUser(ctx context.Context, principal domain.Principal) (domain.User, error) {
	if !principal.Authenticated() {
		return domain.User{}, unauthenticated()
	}
	user, _, err := s.store.UserByID(ctx, principal.ID)
	if err != nil {
		return domain.User{}, unauthenticated()
	}
	return user, nil
}
func (s *Service) ListTokens(ctx context.Context, principal domain.Principal, limit, offset int) ([]domain.APIToken, int64, error) {
	if !principal.Authenticated() {
		return nil, 0, unauthenticated()
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 || offset < 0 {
		return nil, 0, invalid("page", "Invalid pagination.")
	}
	tokens, count, err := s.store.ListTokens(ctx, principal.ID, limit, offset)
	if err != nil {
		return nil, 0, internal(err)
	}
	return tokens, count, nil
}
func (s *Service) CreateToken(ctx context.Context, principal domain.Principal, input CreateTokenInput) (CreatedToken, error) {
	if !principal.Authenticated() {
		return CreatedToken{}, unauthenticated()
	}
	if !s.allowTokenCreation(principal.ID) {
		return CreatedToken{}, rateLimited()
	}
	if len(input.Description) > 200 {
		return CreatedToken{}, invalid("description", "Ensure this field has no more than 200 characters.")
	}
	for _, network := range input.AllowedIPs {
		if _, err := netip.ParsePrefix(network); err != nil {
			return CreatedToken{}, invalid("allowed_ips", "Enter valid CIDR networks.")
		}
	}
	if input.Expires != nil && !input.Expires.After(s.clock.Now()) {
		return CreatedToken{}, invalid("expires", "Expiry must be in the future.")
	}
	secret, hash, err := newSecret()
	if err != nil {
		return CreatedToken{}, internal(err)
	}
	now := s.clock.Now()
	token := domain.APIToken{UserID: principal.ID, Display: "nbx_" + secret[:8] + "…", Description: strings.TrimSpace(input.Description), WriteEnabled: input.WriteEnabled, AllowedIPs: append([]string(nil), input.AllowedIPs...), Created: now, Expires: input.Expires}
	token, err = s.store.CreateToken(ctx, TokenRecord{Token: token, SecretHash: hash})
	if err != nil {
		return CreatedToken{}, internal(err)
	}
	return CreatedToken{Token: token, Secret: secret}, nil
}

func (s *Service) allowTokenCreation(userID int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock.Now()
	cutoff := now.Add(-time.Minute)
	values := s.tokenAttempts[userID]
	kept := values[:0]
	for _, value := range values {
		if value.After(cutoff) {
			kept = append(kept, value)
		}
	}
	if len(kept) >= 10 {
		s.tokenAttempts[userID] = kept
		return false
	}
	s.tokenAttempts[userID] = append(kept, now)
	return true
}
func (s *Service) RevokeToken(ctx context.Context, principal domain.Principal, id int64) error {
	if !principal.Authenticated() {
		return unauthenticated()
	}
	if err := s.store.RevokeToken(ctx, principal.ID, id, s.clock.Now()); err != nil {
		if errors.Is(err, ErrNotFound) {
			return notFoundIdentity("api_token", id)
		}
		return internal(err)
	}
	return nil
}
func (s *Service) ChangePassword(ctx context.Context, principal domain.Principal, current, next string) error {
	if !principal.Authenticated() {
		return unauthenticated()
	}
	user, hash, err := s.store.UserByID(ctx, principal.ID)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(current)) != nil {
		return invalid("current_password", "Current password is incorrect.")
	}
	nextHash, err := HashPassword(next)
	if err != nil {
		return err
	}
	return s.store.Transaction(ctx, func(store Store) error {
		if err := store.UpdatePassword(ctx, user.ID, nextHash); err != nil {
			return err
		}
		return store.DeleteSessionsForUser(ctx, user.ID)
	})
}

func (s *Service) BootstrapAdministrator(ctx context.Context, username, email, password string) (domain.User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return domain.User{}, err
	}
	now := s.clock.Now()
	user := domain.User{Username: strings.TrimSpace(username), Email: strings.TrimSpace(email), IsStaff: true, IsSuperuser: true, IsActive: true, Created: now, Updated: now}
	if user.Username == "" {
		return domain.User{}, invalid("username", "This field is required.")
	}
	var created domain.User
	err = s.store.Transaction(ctx, func(store Store) error {
		count, countErr := store.CountUsers(ctx)
		if countErr != nil {
			return countErr
		}
		if count != 0 {
			return conflict("Administrator bootstrap is allowed only on an empty identity store.")
		}
		created, countErr = store.CreateUser(ctx, user, hash)
		return countErr
	})
	if err != nil {
		var appErr *shared.Error
		if errors.As(err, &appErr) {
			return domain.User{}, err
		}
		return domain.User{}, internal(err)
	}
	return created, nil
}
func (s *Service) ResetAdministratorPassword(ctx context.Context, username, password string) error {
	user, _, err := s.store.UserByUsername(ctx, username)
	if err != nil || !user.IsSuperuser || !user.IsActive {
		return invalid("username", "An active administrator is required.")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	return s.store.Transaction(ctx, func(store Store) error {
		if err := store.UpdatePassword(ctx, user.ID, hash); err != nil {
			return err
		}
		return store.DeleteSessionsForUser(ctx, user.ID)
	})
}

// CreateLocalUser creates a password-authenticated user through the internal
// administration boundary. No public transport route is exposed for it.
func (s *Service) CreateLocalUser(ctx context.Context, actor domain.Principal, input CreateUserInput) (domain.User, error) {
	if err := requireSuperuser(actor); err != nil {
		return domain.User{}, err
	}
	username := strings.TrimSpace(input.Username)
	if username == "" {
		return domain.User{}, invalid("username", "This field is required.")
	}
	if len(username) > 150 {
		return domain.User{}, invalid("username", "Ensure this field has no more than 150 characters.")
	}
	hash, err := HashPassword(input.Password)
	if err != nil {
		return domain.User{}, err
	}
	now := s.clock.Now()
	user := domain.User{
		Username: username, Email: strings.TrimSpace(input.Email), FirstName: strings.TrimSpace(input.FirstName),
		LastName: strings.TrimSpace(input.LastName), IsStaff: input.IsStaff, IsActive: true, Created: now, Updated: now,
	}
	created, err := s.store.CreateUser(ctx, user, hash)
	if err != nil {
		return domain.User{}, internal(err)
	}
	return created, nil
}

func (s *Service) CreateGroup(ctx context.Context, actor domain.Principal, name string) (domain.Group, error) {
	if err := requireSuperuser(actor); err != nil {
		return domain.Group{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Group{}, invalid("name", "This field is required.")
	}
	if len(name) > 150 {
		return domain.Group{}, invalid("name", "Ensure this field has no more than 150 characters.")
	}
	group, err := s.store.CreateGroup(ctx, domain.Group{Name: name})
	if err != nil {
		return domain.Group{}, internal(err)
	}
	return group, nil
}

func (s *Service) AddGroupMember(ctx context.Context, actor domain.Principal, userID, groupID int64) error {
	if err := requireSuperuser(actor); err != nil {
		return err
	}
	if userID <= 0 || groupID <= 0 {
		return invalid("membership", "User and group IDs must be positive integers.")
	}
	if err := s.store.AddGroupMember(ctx, userID, groupID); err != nil {
		return internal(err)
	}
	return nil
}

func (s *Service) GrantPermissionToUser(ctx context.Context, actor domain.Principal, userID int64, input PermissionGrantInput) (domain.PermissionGrant, error) {
	if err := requireSuperuser(actor); err != nil {
		return domain.PermissionGrant{}, err
	}
	if userID <= 0 {
		return domain.PermissionGrant{}, invalid("user_id", "A positive integer is required.")
	}
	return s.createPermissionGrant(ctx, input, func(store Store, permissionID int64) error {
		return store.GrantPermissionToUser(ctx, userID, permissionID)
	})
}

// GrantPermissionToUserByUsername resolves the target inside the identity
// application boundary. This keeps local administration callers from reaching
// through the service to its persistence adapter merely to discover a user ID.
func (s *Service) GrantPermissionToUserByUsername(ctx context.Context, actor domain.Principal, username string, input PermissionGrantInput) (domain.PermissionGrant, error) {
	if err := requireSuperuser(actor); err != nil {
		return domain.PermissionGrant{}, err
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return domain.PermissionGrant{}, invalid("username", "This field is required.")
	}
	target, _, err := s.store.UserByUsername(ctx, username)
	if errors.Is(err, ErrNotFound) {
		return domain.PermissionGrant{}, invalid("username", "A user with this username does not exist.")
	}
	if err != nil {
		return domain.PermissionGrant{}, internal(err)
	}
	return s.GrantPermissionToUser(ctx, actor, target.ID, input)
}

func (s *Service) GrantPermissionToGroup(ctx context.Context, actor domain.Principal, groupID int64, input PermissionGrantInput) (domain.PermissionGrant, error) {
	if err := requireSuperuser(actor); err != nil {
		return domain.PermissionGrant{}, err
	}
	if groupID <= 0 {
		return domain.PermissionGrant{}, invalid("group_id", "A positive integer is required.")
	}
	return s.createPermissionGrant(ctx, input, func(store Store, permissionID int64) error {
		return store.GrantPermissionToGroup(ctx, groupID, permissionID)
	})
}

func (s *Service) createPermissionGrant(ctx context.Context, input PermissionGrantInput, assign func(Store, int64) error) (domain.PermissionGrant, error) {
	permission, err := normalizePermissionGrant(input)
	if err != nil {
		return domain.PermissionGrant{}, err
	}
	var created domain.PermissionGrant
	err = s.store.Transaction(ctx, func(store Store) error {
		var createErr error
		created, createErr = store.CreatePermissionGrant(ctx, permission)
		if createErr != nil {
			return createErr
		}
		return assign(store, created.ID)
	})
	if err != nil {
		return domain.PermissionGrant{}, internal(err)
	}
	return created, nil
}

func normalizePermissionGrant(input PermissionGrantInput) (domain.PermissionGrant, error) {
	grant := domain.PermissionGrant{
		Name: strings.TrimSpace(input.Name), AppLabel: strings.ToLower(strings.TrimSpace(input.AppLabel)),
		Action: strings.ToLower(strings.TrimSpace(input.Action)), Model: strings.ToLower(strings.TrimSpace(input.Model)), ObjectID: input.ObjectID,
	}
	if !identifier(grant.AppLabel) {
		return domain.PermissionGrant{}, invalid("app_label", "Enter a lowercase application label.")
	}
	if !identifier(grant.Model) {
		return domain.PermissionGrant{}, invalid("model", "Enter a lowercase model name.")
	}
	switch grant.Action {
	case "view", "add", "change", "delete":
	default:
		return domain.PermissionGrant{}, invalid("action", "Use view, add, change, or delete.")
	}
	if grant.ObjectID != nil && *grant.ObjectID <= 0 {
		return domain.PermissionGrant{}, invalid("object_id", "A positive integer is required.")
	}
	if grant.ObjectID != nil && grant.Action == "add" {
		// This representation scopes grants by an existing object ID. Creation
		// has no object ID yet, so accepting an object-scoped add grant would be
		// interpreted as a global add permission by the authorizer.
		return domain.PermissionGrant{}, invalid("object_id", "Add permissions cannot be scoped to an existing object ID.")
	}
	if grant.Name == "" {
		grant.Name = grant.Codename()
	}
	if len(grant.Name) > 150 {
		return domain.PermissionGrant{}, invalid("name", "Ensure this field has no more than 150 characters.")
	}
	return grant, nil
}

func identifier(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}

func requireSuperuser(principal domain.Principal) error {
	if !principal.Authenticated() {
		return unauthenticated()
	}
	if !principal.IsSuperuser {
		return forbidden()
	}
	return nil
}

func HashPassword(password string) (string, error) {
	if len(password) < 12 {
		return "", invalid("new_password", "Password must contain at least 12 characters.")
	}
	if len(password) > 1024 {
		return "", invalid("new_password", "Password is too long.")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), passwordCost)
	if err != nil {
		return "", internal(err)
	}
	return string(hash), nil
}
func newSecret() (string, []byte, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", nil, err
	}
	secret := base64.RawURLEncoding.EncodeToString(bytes)
	return secret, digest(secret), nil
}
func digest(value string) []byte { sum := sha256.Sum256([]byte(value)); return sum[:] }
func allowedAddress(remote string, networks []string) bool {
	host := remote
	if parsedHost, _, err := net.SplitHostPort(remote); err == nil {
		host = parsedHost
	}
	host = strings.Trim(host, "[]")
	address, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	for _, raw := range networks {
		network, err := netip.ParsePrefix(raw)
		if err == nil && network.Contains(address) {
			return true
		}
	}
	return false
}

func invalid(field, description string) error {
	return shared.NewValidationError(shared.FieldViolation{
		Field: field, Reason: "invalid", Description: description,
	})
}

func unauthenticated() error {
	return shared.NewError(
		shared.ErrorReasonUnauthenticated,
		"Authentication credentials were not provided.",
	)
}

func forbidden() error {
	return shared.NewError(
		shared.ErrorReasonForbidden,
		"You do not have permission to perform this action.",
	)
}

func rateLimited() error {
	return shared.NewError(
		shared.ErrorReasonRateLimited,
		"Too many requests. Try again later.",
	)
}

func notFoundIdentity(resource string, id int64) error {
	return shared.NewError(
		shared.ErrorReasonNotFound,
		fmt.Sprintf("%s %d was not found", resource, id),
	)
}

func conflict(message string) error {
	return shared.NewError(shared.ErrorReasonConflict, message)
}

func internal(err error) error {
	return shared.WrapError(
		shared.ErrorReasonInternal,
		"An internal error occurred.",
		fmt.Errorf("identity: %w", err),
	)
}
