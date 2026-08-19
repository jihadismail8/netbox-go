package identity_test

import (
	"context"
	"crypto/subtle"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	postgres "netbox-go/internal/adapters/postgres/identity"
	application "netbox-go/internal/application/identity"
	"netbox-go/internal/domain/shared"
)

type testClock struct{ now time.Time }

func (clock *testClock) Now() time.Time { return clock.now }

func TestStandaloneIdentityLifecycle(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:identity_lifecycle?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(postgres.Models()...))
	service := application.NewService(postgres.NewStore(db), application.RealClock{})
	ctx := context.Background()
	admin, err := service.BootstrapAdministrator(ctx, "admin", "admin@example.test", "Correct-Horse-2026!")
	require.NoError(t, err)
	require.True(t, admin.IsSuperuser)
	_, err = service.BootstrapAdministrator(ctx, "other", "", "Correct-Horse-2026!")
	var appErr *shared.Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.ErrorReasonConflict, appErr.Reason)

	session, err := service.Login(ctx, "admin", "Correct-Horse-2026!")
	require.NoError(t, err)
	require.NotEmpty(t, session.Secret)
	require.NotEmpty(t, session.CSRFToken)
	user, err := service.AuthenticateSession(ctx, session.Secret)
	require.NoError(t, err)
	principal := user.Principal()
	created, err := service.CreateToken(ctx, principal, application.CreateTokenInput{Description: "automation", WriteEnabled: true, AllowedIPs: []string{"127.0.0.0/8"}})
	require.NoError(t, err)
	require.NotEmpty(t, created.Secret)
	tokens, count, err := service.ListTokens(ctx, principal, 50, 0)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
	require.Equal(t, "automation", tokens[0].Description)
	if strings.Contains(tokens[0].Display, created.Secret) {
		t.Error("token display exposed one-time credential material")
	}
	_, err = service.AuthenticateToken(ctx, created.Secret, "127.0.0.1:1234", true)
	require.NoError(t, err)
	_, err = service.AuthenticateToken(ctx, created.Secret, "192.0.2.1:1234", false)
	require.Error(t, err)
	require.NoError(t, service.RevokeToken(ctx, principal, created.Token.ID))
	_, err = service.AuthenticateToken(ctx, created.Secret, "127.0.0.1:1234", false)
	require.Error(t, err)

	require.NoError(t, service.ChangePassword(ctx, principal, "Correct-Horse-2026!", "Different-Horse-2026!"))
	_, err = service.AuthenticateSession(ctx, session.Secret)
	require.Error(t, err)
	_, err = service.Login(ctx, "admin", "Correct-Horse-2026!")
	require.Error(t, err)
	_, err = service.Login(ctx, "admin", "Different-Horse-2026!")
	require.NoError(t, err)
}

func TestTokenExpiryAndReadOnlyAreFailClosed(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:identity_edges?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(postgres.Models()...))
	service := application.NewService(postgres.NewStore(db), application.RealClock{})
	ctx := context.Background()
	admin, err := service.BootstrapAdministrator(ctx, "admin", "", "Correct-Horse-2026!")
	require.NoError(t, err)
	expires := time.Now().Add(time.Hour)
	created, err := service.CreateToken(ctx, admin.Principal(), application.CreateTokenInput{Expires: &expires, WriteEnabled: false})
	require.NoError(t, err)
	_, err = service.AuthenticateToken(ctx, created.Secret, "127.0.0.1", false)
	require.NoError(t, err)
	_, err = service.AuthenticateToken(ctx, created.Secret, "127.0.0.1", true)
	require.Error(t, err)
}

func TestRecognizedTokenTouchOrderingAndRateLimit(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:identity_token_order?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(postgres.Models()...))
	clock := &testClock{now: time.Date(2026, 7, 18, 9, 0, 0, 0, time.UTC)}
	service := application.NewService(postgres.NewStore(db), clock)
	admin, err := service.BootstrapAdministrator(t.Context(), "admin", "", "Correct-Horse-2026!")
	require.NoError(t, err)
	expires := clock.now.Add(2 * time.Hour)
	created, err := service.CreateToken(t.Context(), admin.Principal(), application.CreateTokenInput{
		Expires: expiresPtr(expires), AllowedIPs: []string{"192.0.2.0/24"},
	})
	require.NoError(t, err)

	// A recognized token is touched before an allowed-IP rejection, but never
	// more than once per minute.
	_, err = service.AuthenticateToken(t.Context(), created.Secret, "198.51.100.10:443", false)
	require.Error(t, err)
	firstTouch := tokenLastUsed(t, db, created.Token.ID)
	require.Equal(t, clock.now, firstTouch)
	clock.now = clock.now.Add(30 * time.Second)
	_, err = service.AuthenticateToken(t.Context(), created.Secret, "198.51.100.10:443", false)
	require.Error(t, err)
	require.Equal(t, firstTouch, tokenLastUsed(t, db, created.Token.ID))
	clock.now = clock.now.Add(31 * time.Second)
	_, err = service.AuthenticateToken(t.Context(), created.Secret, "198.51.100.10:443", false)
	require.Error(t, err)
	require.Equal(t, clock.now, tokenLastUsed(t, db, created.Token.ID))

	// Expiry is also checked after the recognized-key touch. Unknown material
	// cannot identify or mutate any token row.
	clock.now = expires.Add(time.Minute)
	_, err = service.AuthenticateToken(t.Context(), created.Secret, "192.0.2.1", false)
	require.Error(t, err)
	expiryTouch := tokenLastUsed(t, db, created.Token.ID)
	require.Equal(t, clock.now, expiryTouch)
	clock.now = clock.now.Add(2 * time.Minute)
	_, err = service.AuthenticateToken(t.Context(), "unknown-secret", "192.0.2.1", false)
	require.Error(t, err)
	require.Equal(t, expiryTouch, tokenLastUsed(t, db, created.Token.ID))

	// Token creation is bounded identically for all transports because the
	// final limiter lives in the shared application service.
	for index := 0; index < 10; index++ {
		_, err = service.CreateToken(t.Context(), admin.Principal(), application.CreateTokenInput{})
		require.NoError(t, err, "creation %d", index+1)
	}
	_, err = service.CreateToken(t.Context(), admin.Principal(), application.CreateTokenInput{})
	var appErr *shared.Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.ErrorReasonRateLimited, appErr.Reason)
}

func TestBrowserSessionExpiresAtBoundary(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:identity_session_expiry?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(postgres.Models()...))
	clock := &testClock{now: time.Date(2026, 7, 18, 9, 0, 0, 0, time.UTC)}
	service := application.NewService(postgres.NewStore(db), clock)
	_, err = service.BootstrapAdministrator(t.Context(), "admin", "", "Correct-Horse-2026!")
	require.NoError(t, err)
	session, err := service.Login(t.Context(), "admin", "Correct-Horse-2026!")
	require.NoError(t, err)
	clock.now = session.Expires
	_, err = service.AuthenticateSession(t.Context(), session.Secret)
	require.Error(t, err)
	require.Error(t, service.VerifyCSRF(t.Context(), session.Secret, session.CSRFToken))
}

func TestSessionRotationAndAdministratorResetInvalidateExistingSessions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:identity_session_rotation?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(postgres.Models()...))
	service := application.NewService(postgres.NewStore(db), application.RealClock{})
	_, err = service.BootstrapAdministrator(t.Context(), "admin", "", "Correct-Horse-2026!")
	require.NoError(t, err)

	first, err := service.Login(t.Context(), "admin", "Correct-Horse-2026!")
	require.NoError(t, err)
	rotated, err := service.LoginReplacing(t.Context(), "admin", "Correct-Horse-2026!", first.Secret)
	require.NoError(t, err)
	if subtle.ConstantTimeCompare([]byte(first.Secret), []byte(rotated.Secret)) == 1 {
		t.Error("session rotation reused the prior credential")
	}
	if subtle.ConstantTimeCompare([]byte(first.CSRFToken), []byte(rotated.CSRFToken)) == 1 {
		t.Error("session rotation reused the prior CSRF value")
	}
	_, err = service.AuthenticateSession(t.Context(), first.Secret)
	require.Error(t, err, "the session presented during re-authentication must be invalidated")
	_, err = service.AuthenticateSession(t.Context(), rotated.Secret)
	require.NoError(t, err)

	second, err := service.Login(t.Context(), "admin", "Correct-Horse-2026!")
	require.NoError(t, err)
	require.NoError(t, service.ResetAdministratorPassword(t.Context(), "admin", "Reset-Horse-2026!"))
	for _, secret := range []string{rotated.Secret, second.Secret} {
		_, err = service.AuthenticateSession(t.Context(), secret)
		require.Error(t, err, "administrator reset must invalidate every browser session")
	}
	_, err = service.Login(t.Context(), "admin", "Correct-Horse-2026!")
	require.Error(t, err)
	_, err = service.Login(t.Context(), "admin", "Reset-Horse-2026!")
	require.NoError(t, err)
}

func expiresPtr(value time.Time) *time.Time { return &value }

func tokenLastUsed(t *testing.T, db *gorm.DB, id int64) time.Time {
	t.Helper()
	var row postgres.TokenRow
	require.NoError(t, db.First(&row, id).Error)
	require.NotNil(t, row.LastUsed)
	return row.LastUsed.UTC()
}
