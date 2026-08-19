package parity

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	identityv1 "netbox-go/gen/go/netbox/identity/v1"
	grpcidentity "netbox-go/internal/adapters/grpc/identity"
	postgresidentity "netbox-go/internal/adapters/postgres/identity"
	restidentity "netbox-go/internal/adapters/rest/netbox/identity"
	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

var (
	passwordParityDatabaseSequence atomic.Uint64
	errPasswordParityRollback      = errors.New("injected password-change rollback")
)

type passwordParityStore struct {
	application.Store
	failAfterSessionDelete atomic.Bool
	tokenLookups           atomic.Int64
	tokenTouches           atomic.Int64
	authenticatedUserID    atomic.Int64
	sessionDeletes         atomic.Int64
}

func (store *passwordParityStore) Transaction(
	ctx context.Context,
	apply func(application.Store) error,
) error {
	return store.Store.Transaction(ctx, func(tx application.Store) error {
		return apply(&passwordParityTransaction{Store: tx, owner: store})
	})
}

func (store *passwordParityStore) TokenByHash(
	ctx context.Context,
	hash []byte,
) (application.TokenRecord, domain.User, error) {
	store.tokenLookups.Add(1)
	record, user, err := store.Store.TokenByHash(ctx, hash)
	if err == nil {
		store.authenticatedUserID.Store(user.ID)
	}
	return record, user, err
}

func (store *passwordParityStore) TouchToken(
	ctx context.Context,
	id int64,
	at time.Time,
) error {
	store.tokenTouches.Add(1)
	return store.Store.TouchToken(ctx, id, at)
}

type passwordParityTransaction struct {
	application.Store
	owner *passwordParityStore
}

func (tx *passwordParityTransaction) DeleteSessionsForUser(
	ctx context.Context,
	userID int64,
) error {
	if err := tx.Store.DeleteSessionsForUser(ctx, userID); err != nil {
		return err
	}
	tx.owner.sessionDeletes.Add(1)
	if tx.owner.failAfterSessionDelete.Load() {
		return errPasswordParityRollback
	}
	return nil
}

type passwordParityFixture struct {
	db           *gorm.DB
	service      *application.Service
	store        *passwordParityStore
	user         domain.User
	token        application.CreatedToken
	sessions     []domain.BrowserSession
	oldPassword  string
	newPassword  string
	sessionCount int64
}

type passwordParityTransportObservation struct {
	code                codes.Code
	grpcMessage         string
	httpStatus          int
	body                string
	setCookieCount      int
	tokenLookups        int64
	tokenTouches        int64
	authenticatedUserID int64
	responsePresent     bool
}

type passwordParityApplicationObservation struct {
	reason         shared.ErrorReason
	resultPresent  bool
	browserPresent bool
}

func TestPasswordChangeRESTGRPCParity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		rest := newPasswordParityFixture(t)
		grpc := newPasswordParityFixture(t)
		applicationFixture := newPasswordParityFixture(t)

		restObservation := observeRESTPasswordChange(
			t,
			rest,
			rest.oldPassword,
			rest.newPassword,
		)
		grpcObservation := observeGRPCPasswordChange(
			t,
			grpc,
			grpc.oldPassword,
			grpc.newPassword,
		)
		applicationObservation := observeApplicationPasswordChange(
			t,
			applicationFixture,
			applicationFixture.oldPassword,
			applicationFixture.newPassword,
		)

		assertPasswordParitySuccessTransport(
			t,
			rest,
			grpc,
			restObservation,
			grpcObservation,
			applicationObservation,
		)
		assertPasswordParityCommittedState(t, rest)
		assertPasswordParityCommittedState(t, grpc)
		assertPasswordParityCommittedState(t, applicationFixture)
	})

	t.Run("current password validation", func(t *testing.T) {
		rest := newPasswordParityFixture(t)
		grpc := newPasswordParityFixture(t)
		applicationFixture := newPasswordParityFixture(t)
		wrongCurrent := "Incorrect-Current-2026!"

		restObservation := observeRESTPasswordChange(
			t,
			rest,
			wrongCurrent,
			rest.newPassword,
		)
		grpcObservation := observeGRPCPasswordChange(
			t,
			grpc,
			wrongCurrent,
			grpc.newPassword,
		)
		applicationObservation := observeApplicationPasswordChange(
			t,
			applicationFixture,
			wrongCurrent,
			applicationFixture.newPassword,
		)

		assertPasswordParityValidationTransport(
			t,
			rest,
			grpc,
			restObservation,
			grpcObservation,
			applicationObservation,
			`{"current_password":["Current password is incorrect."]}`,
		)
		assertPasswordParityUnchangedState(t, rest)
		assertPasswordParityUnchangedState(t, grpc)
		assertPasswordParityUnchangedState(t, applicationFixture)
	})

	t.Run("new password validation", func(t *testing.T) {
		rest := newPasswordParityFixture(t)
		grpc := newPasswordParityFixture(t)
		applicationFixture := newPasswordParityFixture(t)
		invalidNext := "short"

		restObservation := observeRESTPasswordChange(
			t,
			rest,
			rest.oldPassword,
			invalidNext,
		)
		grpcObservation := observeGRPCPasswordChange(
			t,
			grpc,
			grpc.oldPassword,
			invalidNext,
		)
		applicationObservation := observeApplicationPasswordChange(
			t,
			applicationFixture,
			applicationFixture.oldPassword,
			invalidNext,
		)

		assertPasswordParityValidationTransport(
			t,
			rest,
			grpc,
			restObservation,
			grpcObservation,
			applicationObservation,
			`{"new_password":["Password must contain at least 12 characters."]}`,
		)
		assertPasswordParityUnchangedState(t, rest)
		assertPasswordParityUnchangedState(t, grpc)
		assertPasswordParityUnchangedState(t, applicationFixture)
	})

	t.Run("post delete rollback", func(t *testing.T) {
		rest := newPasswordParityFixture(t)
		grpc := newPasswordParityFixture(t)
		applicationFixture := newPasswordParityFixture(t)
		rest.store.failAfterSessionDelete.Store(true)
		grpc.store.failAfterSessionDelete.Store(true)
		applicationFixture.store.failAfterSessionDelete.Store(true)

		restObservation := observeRESTPasswordChange(
			t,
			rest,
			rest.oldPassword,
			rest.newPassword,
		)
		grpcObservation := observeGRPCPasswordChange(
			t,
			grpc,
			grpc.oldPassword,
			grpc.newPassword,
		)
		applicationObservation := observeApplicationPasswordChange(
			t,
			applicationFixture,
			applicationFixture.oldPassword,
			applicationFixture.newPassword,
		)

		if restObservation.httpStatus != http.StatusInternalServerError {
			t.Fatal("REST password-change rollback did not map to Internal")
		}
		if restObservation.body != `{"detail":"An internal error occurred."}` {
			t.Fatal("REST password-change rollback returned an unexpected safe body")
		}
		if grpcObservation.code != codes.Internal {
			t.Fatal("gRPC password-change rollback did not map to Internal")
		}
		if grpcObservation.grpcMessage != "An internal error occurred." {
			t.Fatal("gRPC password-change rollback returned an unexpected safe reason")
		}
		if applicationObservation.reason != shared.ErrorReasonInternal ||
			applicationObservation.resultPresent {
			t.Fatal("shared application rollback did not retain Internal with no result")
		}
		if restObservation.setCookieCount != 0 {
			t.Fatal("REST Token password-change rollback emitted a cookie")
		}
		if rest.store.sessionDeletes.Load() != 1 || grpc.store.sessionDeletes.Load() != 1 ||
			applicationFixture.store.sessionDeletes.Load() != 1 {
			t.Fatal("password-change rollback did not execute the real session deletion")
		}
		assertPasswordParityCredentialBoundary(t, rest, grpc, restObservation, grpcObservation)
		assertPasswordParityUnchangedState(t, rest)
		assertPasswordParityUnchangedState(t, grpc)
		assertPasswordParityUnchangedState(t, applicationFixture)
	})

	t.Run("browser cookie mechanics are REST only", func(t *testing.T) {
		fixture := newPasswordParityFixture(t)
		payload, err := json.Marshal(map[string]string{
			"current_password": fixture.oldPassword,
			"new_password":     fixture.newPassword,
		})
		if err != nil {
			t.Fatal("could not encode browser password-change parity request")
		}
		handler := restidentity.NewHandler(fixture.service, false)
		router := gin.New()
		handler.Register(router)
		request := httptest.NewRequest(
			http.MethodPost,
			"/api/auth/password/change/",
			bytes.NewReader(payload),
		)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Token invalid-ambient-material")
		request.Header.Set("X-CSRFToken", fixture.sessions[0].CSRFToken)
		request.AddCookie(&http.Cookie{Name: "sessionid", Value: fixture.sessions[0].Secret})
		request.AddCookie(&http.Cookie{Name: "csrftoken", Value: fixture.sessions[0].CSRFToken})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		if response.Code != http.StatusNoContent || response.Body.Len() != 0 {
			t.Fatal("browser-origin REST password change did not preserve empty 204")
		}
		cookies := response.Result().Cookies()
		if len(cookies) != 2 {
			t.Fatal("browser-origin REST password change did not issue exactly two cookies")
		}
		var replacement domain.BrowserSession
		for _, cookie := range cookies {
			switch cookie.Name {
			case "sessionid":
				replacement.Secret = cookie.Value
				replacement.Expires = cookie.Expires
			case "csrftoken":
				replacement.CSRFToken = cookie.Value
			default:
				t.Fatal("browser-origin REST password change emitted an undeclared cookie")
			}
			if cookie.Value == "" || cookie.MaxAge <= 0 {
				t.Fatal("browser-origin REST password change emitted a tombstone")
			}
		}
		if passwordParitySessionCount(t, fixture.db, fixture.user.ID) != 1 {
			t.Fatal("browser-origin REST password change left an unexpected session set")
		}
		for _, oldSession := range fixture.sessions {
			if _, err := fixture.service.AuthenticateSession(t.Context(), oldSession.Secret); err == nil {
				t.Fatal("browser-origin REST password change retained an old session")
			}
		}
		if _, err := fixture.service.AuthenticateSession(t.Context(), replacement.Secret); err != nil {
			t.Fatal("browser-origin REST replacement did not authenticate")
		}
		if err := fixture.service.VerifyCSRF(
			t.Context(),
			replacement.Secret,
			replacement.CSRFToken,
		); err != nil {
			t.Fatal("browser-origin REST replacement rejected its CSRF value")
		}
		if fixture.store.tokenLookups.Load() != 0 || fixture.store.tokenTouches.Load() != 0 {
			t.Fatal("browser-origin REST password change consulted ambient Authorization")
		}
		assertPasswordParityTokenRetained(t, fixture)
	})
}

func newPasswordParityFixture(t *testing.T) *passwordParityFixture {
	t.Helper()
	databaseID := passwordParityDatabaseSequence.Add(1)
	db, err := gorm.Open(
		sqlite.Open("file:identity_password_change_parity_"+
			strconv.FormatUint(databaseID, 10)+"?mode=memory&cache=shared"),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		t.Fatal("could not create password-change parity database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal("could not own password-change parity database")
	}
	t.Cleanup(func() {
		if sqlDB.Close() != nil {
			t.Error("could not close password-change parity database")
		}
	})
	if err := db.AutoMigrate(postgresidentity.Models()...); err != nil {
		t.Fatal("could not prepare password-change parity schema")
	}

	now := time.Date(2026, 8, 19, 11, 0, 0, 0, time.UTC)
	oldPassword := "Parity-Current-2026!"
	newPassword := "Parity-Replacement-2026!"
	hash, err := application.HashPassword(oldPassword)
	if err != nil {
		t.Fatal("could not prepare password-change parity verifier")
	}
	baseStore := postgresidentity.NewStore(db)
	user, err := baseStore.CreateUser(t.Context(), domain.User{
		Username: "password-parity-user",
		IsActive: true,
		Created:  now,
		Updated:  now,
	}, hash)
	if err != nil {
		t.Fatal("could not prepare password-change parity user")
	}
	store := &passwordParityStore{Store: baseStore}
	service := application.NewService(store, identityFixedClock{now: now.Add(time.Hour)})
	sessions := make([]domain.BrowserSession, 0, 2)
	for range 2 {
		session, err := service.Login(t.Context(), user.Username, oldPassword)
		if err != nil {
			t.Fatal("could not prepare password-change parity session")
		}
		sessions = append(sessions, session)
	}
	token, err := service.CreateToken(t.Context(), user.Principal(), application.CreateTokenInput{
		Description:  "password parity",
		WriteEnabled: true,
	})
	if err != nil {
		t.Fatal("could not prepare password-change parity token")
	}
	if token.Secret == "" {
		t.Fatal("password-change parity token fixture has no credential")
	}
	store.tokenLookups.Store(0)
	store.tokenTouches.Store(0)
	store.authenticatedUserID.Store(0)
	store.sessionDeletes.Store(0)

	return &passwordParityFixture{
		db:           db,
		service:      service,
		store:        store,
		user:         user,
		token:        token,
		sessions:     sessions,
		oldPassword:  oldPassword,
		newPassword:  newPassword,
		sessionCount: 2,
	}
}

func observeRESTPasswordChange(
	t *testing.T,
	fixture *passwordParityFixture,
	currentPassword string,
	newPassword string,
) passwordParityTransportObservation {
	t.Helper()
	payload, err := json.Marshal(map[string]string{
		"current_password": currentPassword,
		"new_password":     newPassword,
	})
	if err != nil {
		t.Fatal("could not encode password-change parity request")
	}
	handler := restidentity.NewHandler(fixture.service, false)
	router := gin.New()
	handler.Register(router)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/password/change/",
		bytes.NewReader(payload),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Token "+fixture.token.Secret)
	request.RemoteAddr = "127.0.0.1:443"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	return passwordParityTransportObservation{
		httpStatus:          response.Code,
		body:                response.Body.String(),
		setCookieCount:      len(response.Header().Values("Set-Cookie")),
		tokenLookups:        fixture.store.tokenLookups.Load(),
		tokenTouches:        fixture.store.tokenTouches.Load(),
		authenticatedUserID: fixture.store.authenticatedUserID.Load(),
	}
}

func observeGRPCPasswordChange(
	t *testing.T,
	fixture *passwordParityFixture,
	currentPassword string,
	newPassword string,
) passwordParityTransportObservation {
	t.Helper()
	request := &identityv1.ChangePasswordRequest{
		CurrentPassword: currentPassword,
		NewPassword:     newPassword,
	}
	ctx := metadata.NewIncomingContext(
		t.Context(),
		metadata.Pairs("authorization", "Bearer "+fixture.token.Secret),
	)
	ctx = peer.NewContext(ctx, &peer.Peer{
		Addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 443},
	})
	server := grpcidentity.NewServer(fixture.service)
	response, err := grpcidentity.UnaryAuthenticator(fixture.service)(
		ctx,
		request,
		&grpc.UnaryServerInfo{FullMethod: identityv1.IdentityService_ChangePassword_FullMethodName},
		func(ctx context.Context, request any) (any, error) {
			return server.ChangePassword(ctx, request.(*identityv1.ChangePasswordRequest))
		},
	)

	return passwordParityTransportObservation{
		code:                status.Code(err),
		grpcMessage:         passwordParityGRPCMessage(err),
		tokenLookups:        fixture.store.tokenLookups.Load(),
		tokenTouches:        fixture.store.tokenTouches.Load(),
		authenticatedUserID: fixture.store.authenticatedUserID.Load(),
		responsePresent:     response != nil,
	}
}

func observeApplicationPasswordChange(
	t *testing.T,
	fixture *passwordParityFixture,
	currentPassword string,
	newPassword string,
) passwordParityApplicationObservation {
	t.Helper()
	result, err := fixture.service.ChangePassword(
		t.Context(),
		fixture.user.Principal(),
		application.NewPasswordChangeInput(
			currentPassword,
			newPassword,
			application.APITokenPasswordChangeCredential(),
		),
	)
	observation := passwordParityApplicationObservation{
		resultPresent: result != nil,
	}
	if err != nil {
		observation.reason = shared.ReasonOf(err)
	}
	if result != nil {
		_, observation.browserPresent = result.BrowserSession()
	}
	return observation
}

func assertPasswordParitySuccessTransport(
	t *testing.T,
	rest *passwordParityFixture,
	grpc *passwordParityFixture,
	restObservation passwordParityTransportObservation,
	grpcObservation passwordParityTransportObservation,
	applicationObservation passwordParityApplicationObservation,
) {
	t.Helper()
	if restObservation.httpStatus != http.StatusNoContent || restObservation.body != "" {
		t.Fatal("REST Token password change did not preserve its empty 204 response")
	}
	if restObservation.setCookieCount != 0 {
		t.Fatal("REST Token password change emitted browser cookie mechanics")
	}
	if grpcObservation.code != codes.OK || !grpcObservation.responsePresent {
		t.Fatal("bearer-gRPC password change did not return its empty success response")
	}
	if grpcObservation.grpcMessage != "" {
		t.Fatal("successful bearer-gRPC password change returned an error message")
	}
	if applicationObservation.reason != "" || !applicationObservation.resultPresent ||
		applicationObservation.browserPresent {
		t.Fatal("shared application success did not return the Token-origin result")
	}
	if rest.store.sessionDeletes.Load() != 1 || grpc.store.sessionDeletes.Load() != 1 {
		t.Fatal("successful Token password change did not revoke sessions once")
	}
	assertPasswordParityCredentialBoundary(t, rest, grpc, restObservation, grpcObservation)
}

func assertPasswordParityValidationTransport(
	t *testing.T,
	rest *passwordParityFixture,
	grpc *passwordParityFixture,
	restObservation passwordParityTransportObservation,
	grpcObservation passwordParityTransportObservation,
	applicationObservation passwordParityApplicationObservation,
	expectedRESTBody string,
) {
	t.Helper()
	if restObservation.httpStatus != http.StatusBadRequest {
		t.Fatal("REST Token password validation did not map to HTTP 400")
	}
	if restObservation.body != expectedRESTBody {
		t.Fatal("REST Token password validation returned an unexpected field result")
	}
	if restObservation.setCookieCount != 0 {
		t.Fatal("REST Token password validation emitted a cookie")
	}
	if grpcObservation.code != codes.InvalidArgument {
		t.Fatal("bearer-gRPC password validation did not map to InvalidArgument")
	}
	if grpcObservation.grpcMessage != "Invalid input." {
		t.Fatal("bearer-gRPC password validation returned an unexpected safe reason")
	}
	if applicationObservation.reason != shared.ErrorReasonValidation ||
		applicationObservation.resultPresent {
		t.Fatal("shared application validation reason/result drifted from both transports")
	}
	if rest.store.sessionDeletes.Load() != 0 || grpc.store.sessionDeletes.Load() != 0 {
		t.Fatal("password validation reached session mutation")
	}
	assertPasswordParityCredentialBoundary(t, rest, grpc, restObservation, grpcObservation)
}

func assertPasswordParityCredentialBoundary(
	t *testing.T,
	rest *passwordParityFixture,
	grpc *passwordParityFixture,
	restObservation passwordParityTransportObservation,
	grpcObservation passwordParityTransportObservation,
) {
	t.Helper()
	if restObservation.authenticatedUserID != rest.user.ID ||
		grpcObservation.authenticatedUserID != grpc.user.ID ||
		restObservation.authenticatedUserID != grpcObservation.authenticatedUserID {
		t.Fatal("REST Token and bearer-gRPC did not resolve the same Principal")
	}
	if restObservation.tokenLookups != 1 || grpcObservation.tokenLookups != 1 {
		t.Fatal("password change performed an unexpected API-token lookup")
	}
	if restObservation.tokenTouches != 1 || grpcObservation.tokenTouches != 1 {
		t.Fatal("password change performed an unexpected API-token touch")
	}
}

func assertPasswordParityCommittedState(t *testing.T, fixture *passwordParityFixture) {
	t.Helper()
	assertPasswordParityPassword(t, fixture, fixture.oldPassword, false)
	assertPasswordParityPassword(t, fixture, fixture.newPassword, true)
	if passwordParitySessionCount(t, fixture.db, fixture.user.ID) != 0 {
		t.Fatal("Token-origin password change left a browser session")
	}
	assertPasswordParityTokenRetained(t, fixture)
}

func assertPasswordParityUnchangedState(t *testing.T, fixture *passwordParityFixture) {
	t.Helper()
	assertPasswordParityPassword(t, fixture, fixture.oldPassword, true)
	assertPasswordParityPassword(t, fixture, fixture.newPassword, false)
	if passwordParitySessionCount(t, fixture.db, fixture.user.ID) != fixture.sessionCount {
		t.Fatal("failed password change mutated the browser-session set")
	}
	assertPasswordParityTokenRetained(t, fixture)
}

func assertPasswordParityPassword(
	t *testing.T,
	fixture *passwordParityFixture,
	password string,
	wantAccepted bool,
) {
	t.Helper()
	_, err := fixture.service.AuthenticatePassword(t.Context(), fixture.user.Username, password)
	if (err == nil) != wantAccepted {
		t.Fatal("password-change parity state has an unexpected password outcome")
	}
}

func assertPasswordParityTokenRetained(t *testing.T, fixture *passwordParityFixture) {
	t.Helper()
	var count int64
	if err := fixture.db.Model(&postgresidentity.TokenRow{}).
		Where("id = ? AND user_id = ? AND revoked_at IS NULL", fixture.token.Token.ID, fixture.user.ID).
		Count(&count).Error; err != nil {
		t.Fatal("could not inspect password-change parity token state")
	}
	if count != 1 {
		t.Fatal("password change mutated the independent API token")
	}
	if _, err := fixture.service.AuthenticateToken(
		t.Context(),
		fixture.token.Secret,
		"127.0.0.1:443",
		true,
	); err != nil {
		t.Fatal("password change made the independent API token unusable")
	}
}

func passwordParitySessionCount(t *testing.T, db *gorm.DB, userID int64) int64 {
	t.Helper()
	var count int64
	if err := db.Model(&postgresidentity.SessionRow{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		t.Fatal("could not inspect password-change parity session state")
	}
	return count
}

func passwordParityGRPCMessage(err error) string {
	if err == nil {
		return ""
	}
	return status.Convert(err).Message()
}
