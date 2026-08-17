package parity

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	identityv1 "netbox-go/gen/go/netbox/identity/v1"
	grpcidentity "netbox-go/internal/adapters/grpc/identity"
	postgresidentity "netbox-go/internal/adapters/postgres/identity"
	restidentity "netbox-go/internal/adapters/rest/netbox/identity"
	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
)

type identityFixedClock struct{ now time.Time }

func (clock identityFixedClock) Now() time.Time { return clock.now }

type parityCredentialStore struct {
	application.Store
	record      application.TokenRecord
	user        domain.User
	lookupErr   error
	touchErr    error
	lookupCalls int
	touchCalls  int
	events      []tokenCredentialEvent
}

func (store *parityCredentialStore) TokenByHash(context.Context, []byte) (application.TokenRecord, domain.User, error) {
	store.lookupCalls++
	store.events = append(store.events, tokenCredentialEventLookup)
	return store.record, store.user, store.lookupErr
}

func (store *parityCredentialStore) TouchToken(context.Context, int64, time.Time) error {
	store.touchCalls++
	store.events = append(store.events, tokenCredentialEventTouch)
	return store.touchErr
}

type tokenCredentialEvent string

const (
	tokenCredentialEventLookup tokenCredentialEvent = "lookup"
	tokenCredentialEventTouch  tokenCredentialEvent = "touch"
)

type tokenCredentialParityCase struct {
	name            string
	hasCredential   bool
	record          application.TokenRecord
	user            domain.User
	lookupErr       error
	touchErr        error
	remoteAddress   string
	write           bool
	wantRESTStatus  int
	wantRESTDetail  string
	wantGRPCCode    codes.Code
	wantGRPCMessage string
	wantHandler     bool
	wantLookups     int
	wantTouches     int
	wantKind        application.TokenCredentialFailureKind
}

type tokenTransportObservation struct {
	handlerCalled bool
	principal     domain.Principal
	lookupCalls   int
	touchCalls    int
	events        []tokenCredentialEvent
	application   tokenApplicationObservation
}

type tokenApplicationObservation struct {
	kind   application.TokenCredentialFailureKind
	events []tokenCredentialEvent
}

func observeApplicationTokenCredential(
	t *testing.T,
	now time.Time,
	test tokenCredentialParityCase,
) tokenApplicationObservation {
	t.Helper()

	store := &parityCredentialStore{
		record:    test.record,
		user:      test.user,
		lookupErr: test.lookupErr,
		touchErr:  test.touchErr,
	}
	service := application.NewService(store, identityFixedClock{now: now})
	secret := ""
	if test.hasCredential {
		secret = "present"
	}
	_, err := service.AuthenticateToken(t.Context(), secret, test.remoteAddress, test.write)

	var failure *application.TokenCredentialFailure
	observation := tokenApplicationObservation{
		events: append([]tokenCredentialEvent(nil), store.events...),
	}
	if errors.As(err, &failure) {
		observation.kind = failure.Kind
	}
	return observation
}

func expectedTokenCredentialEvents(lookups, touches int) []tokenCredentialEvent {
	var events []tokenCredentialEvent
	for range lookups {
		events = append(events, tokenCredentialEventLookup)
	}
	for range touches {
		events = append(events, tokenCredentialEventTouch)
	}
	return events
}

func observeRESTTokenCredential(
	t *testing.T,
	now time.Time,
	test tokenCredentialParityCase,
) (*httptest.ResponseRecorder, tokenTransportObservation) {
	t.Helper()

	store := &parityCredentialStore{
		record:    test.record,
		user:      test.user,
		lookupErr: test.lookupErr,
		touchErr:  test.touchErr,
	}
	service := application.NewService(store, identityFixedClock{now: now})
	handler := restidentity.NewHandler(service, false)
	method := http.MethodGet
	if test.write {
		method = http.MethodPost
	}

	observation := tokenTransportObservation{}
	router := gin.New()
	router.Handle(method, "/protected", handler.BaselineMiddleware(), func(c *gin.Context) {
		observation.handlerCalled = true
		observation.principal, _ = domain.PrincipalFromContext(c.Request.Context())
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(method, "/protected", nil)
	request.RemoteAddr = test.remoteAddress
	if test.hasCredential {
		request.Header.Set("Authorization", "Token present")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	observation.lookupCalls = store.lookupCalls
	observation.touchCalls = store.touchCalls
	observation.events = append([]tokenCredentialEvent(nil), store.events...)
	observation.application = observeApplicationTokenCredential(t, now, test)
	return response, observation
}

func observeGRPCTokenCredential(
	t *testing.T,
	now time.Time,
	test tokenCredentialParityCase,
) (codes.Code, string, tokenTransportObservation) {
	t.Helper()

	store := &parityCredentialStore{
		record:    test.record,
		user:      test.user,
		lookupErr: test.lookupErr,
		touchErr:  test.touchErr,
	}
	service := application.NewService(store, identityFixedClock{now: now})
	ctx := t.Context()
	if test.hasCredential {
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", "Bearer present"))
	}
	if test.remoteAddress != "" {
		host, _, err := net.SplitHostPort(test.remoteAddress)
		require.NoError(t, err)
		ip := net.ParseIP(host)
		require.NotNil(t, ip)
		ctx = peer.NewContext(ctx, &peer.Peer{Addr: &net.TCPAddr{IP: ip, Port: 443}})
	}

	fullMethod := dcimv1.DCIMService_GetSite_FullMethodName
	if test.write {
		fullMethod = dcimv1.DCIMService_CreateSite_FullMethodName
	}
	observation := tokenTransportObservation{}
	_, err := grpcidentity.UnaryAuthenticator(service)(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: fullMethod},
		func(ctx context.Context, _ any) (any, error) {
			observation.handlerCalled = true
			observation.principal, _ = domain.PrincipalFromContext(ctx)
			return struct{}{}, nil
		},
	)

	message := ""
	if err != nil {
		message = status.Convert(err).Message()
	}
	observation.lookupCalls = store.lookupCalls
	observation.touchCalls = store.touchCalls
	observation.events = append([]tokenCredentialEvent(nil), store.events...)
	observation.application = observeApplicationTokenCredential(t, now, test)
	return status.Code(err), message, observation
}

func TestIdentityGRPCLifecycleAndBearerAuthentication(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:identity_grpc_parity?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(postgresidentity.Models()...))

	now := time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC)
	store := postgresidentity.NewStore(db)
	passwordHash, err := application.HashPassword("Current-Password-2026!")
	require.NoError(t, err)
	user, err := store.CreateUser(t.Context(), domain.User{
		Username: "grpc-admin", Email: "grpc-admin@example.test", FirstName: "gRPC", LastName: "Admin",
		IsStaff: true, IsSuperuser: true, IsActive: true, Permissions: []string{"dcim.view_site"}, Created: now, Updated: now,
	}, passwordHash)
	require.NoError(t, err)

	service := application.NewService(store, identityFixedClock{now: now})
	server := grpcidentity.NewServer(service)
	principalContext := domain.WithPrincipal(t.Context(), user.Principal())

	_, err = server.GetCurrentUser(t.Context(), &identityv1.GetCurrentUserRequest{})
	require.Equal(t, codes.Unauthenticated, status.Code(err))

	current, err := server.GetCurrentUser(principalContext, &identityv1.GetCurrentUserRequest{})
	require.NoError(t, err)
	require.Equal(t, user.ID, current.User.Id)
	require.Equal(t, user.Username, current.User.Username)
	require.Equal(t, user.Email, current.User.Email)
	require.Equal(t, user.Permissions, current.User.Permissions)

	expires := now.Add(2 * time.Hour)
	created, err := server.CreateAPIToken(principalContext, &identityv1.CreateAPITokenRequest{
		Description: "gRPC automation", Expires: timestamppb.New(expires), WriteEnabled: true, AllowedIps: []string{"127.0.0.0/8"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.Secret)
	require.NotContains(t, created.Token.Display, created.Secret)
	require.Equal(t, "gRPC automation", created.Token.Description)
	require.Equal(t, []string{"127.0.0.0/8"}, created.Token.AllowedIps)
	require.True(t, created.Token.Expires.AsTime().Equal(expires))

	listed, err := server.ListAPITokens(principalContext, &identityv1.ListAPITokensRequest{})
	require.NoError(t, err)
	require.Equal(t, uint64(1), listed.Page.Count)
	require.Len(t, listed.Results, 1)
	require.Equal(t, created.Token.Id, listed.Results[0].Id)
	// The list contract has no secret field: the credential is returned exactly
	// once by CreateAPIToken and cannot be recovered through the service.
	require.NotContains(t, listed.Results[0].Display, created.Secret)

	authenticated, err := service.AuthenticateToken(t.Context(), created.Secret, "127.0.0.1:443", true)
	require.NoError(t, err)
	require.Equal(t, user.ID, authenticated.ID)
	_, err = service.AuthenticateToken(t.Context(), created.Secret, "192.0.2.10:443", false)
	require.Error(t, err)

	_, err = server.ChangePassword(principalContext, &identityv1.ChangePasswordRequest{CurrentPassword: "wrong", NewPassword: "Replacement-Password-2026!"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	session, err := service.Login(t.Context(), user.Username, "Current-Password-2026!")
	require.NoError(t, err)
	_, err = server.ChangePassword(principalContext, &identityv1.ChangePasswordRequest{CurrentPassword: "Current-Password-2026!", NewPassword: "Replacement-Password-2026!"})
	require.NoError(t, err)
	_, err = service.AuthenticateSession(t.Context(), session.Secret)
	require.Error(t, err, "password changes must revoke existing browser sessions")
	_, err = service.AuthenticatePassword(t.Context(), user.Username, "Current-Password-2026!")
	require.Error(t, err)
	_, err = service.AuthenticatePassword(t.Context(), user.Username, "Replacement-Password-2026!")
	require.NoError(t, err)

	_, err = server.RevokeAPIToken(principalContext, &identityv1.RevokeAPITokenRequest{Id: created.Token.Id})
	require.NoError(t, err)
	_, err = service.AuthenticateToken(t.Context(), created.Secret, "127.0.0.1:443", false)
	require.Error(t, err, "revoked API token must stop authenticating immediately")
}

func TestIdentityGRPCBearerInterceptorEnforcesCredentialAndWriteScope(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:identity_grpc_interceptor?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(postgresidentity.Models()...))

	now := time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC)
	store := postgresidentity.NewStore(db)
	hash, err := application.HashPassword("Interceptor-Password-2026!")
	require.NoError(t, err)
	user, err := store.CreateUser(t.Context(), domain.User{Username: "reader", IsActive: true, Created: now, Updated: now}, hash)
	require.NoError(t, err)
	service := application.NewService(store, identityFixedClock{now: now})
	created, err := service.CreateToken(t.Context(), user.Principal(), application.CreateTokenInput{Description: "read-only", WriteEnabled: false, AllowedIPs: []string{"127.0.0.0/8"}})
	require.NoError(t, err)

	interceptor := grpcidentity.UnaryAuthenticator(service)
	info := &grpc.UnaryServerInfo{FullMethod: "/netbox.identity.v1.IdentityService/GetCurrentUser"}
	handlerCalled := false
	handler := func(ctx context.Context, _ any) (any, error) {
		handlerCalled = true
		principal, ok := domain.PrincipalFromContext(ctx)
		require.True(t, ok)
		require.Equal(t, user.ID, principal.ID)
		return &identityv1.GetCurrentUserResponse{}, nil
	}

	_, err = interceptor(t.Context(), nil, info, handler)
	require.Equal(t, codes.Unauthenticated, status.Code(err))

	authenticatedContext := metadata.NewIncomingContext(t.Context(), metadata.Pairs("authorization", "Bearer "+created.Secret))
	authenticatedContext = peer.NewContext(authenticatedContext, &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 53000}})
	_, err = interceptor(authenticatedContext, nil, info, handler)
	require.NoError(t, err)
	require.True(t, handlerCalled)

	writeInfo := &grpc.UnaryServerInfo{FullMethod: "/netbox.identity.v1.IdentityService/CreateAPIToken"}
	_, err = interceptor(authenticatedContext, nil, writeInfo, handler)
	require.Equal(t, codes.PermissionDenied, status.Code(err), "read-only tokens must not reach write RPC handlers")

	outsideNetwork := metadata.NewIncomingContext(t.Context(), metadata.Pairs("authorization", "Bearer "+created.Secret))
	outsideNetwork = peer.NewContext(outsideNetwork, &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP("192.0.2.20"), Port: 53000}})
	_, err = interceptor(outsideNetwork, nil, info, handler)
	require.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestRESTAndGRPCResolveTheSameEffectivePrincipal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:identity_transport_principal?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(postgresidentity.Models()...))
	service := application.NewService(postgresidentity.NewStore(db), application.RealClock{})
	admin, err := service.BootstrapAdministrator(t.Context(), "admin", "", "Administrator-Password-2026!")
	require.NoError(t, err)
	operator, err := service.CreateLocalUser(t.Context(), admin.Principal(), application.CreateUserInput{
		Username: "transport-operator", Password: "Operator-Password-2026!",
	})
	require.NoError(t, err)
	group, err := service.CreateGroup(t.Context(), admin.Principal(), "transport operators")
	require.NoError(t, err)
	require.NoError(t, service.AddGroupMember(t.Context(), admin.Principal(), operator.ID, group.ID))
	objectID := int64(73)
	_, err = service.GrantPermissionToGroup(t.Context(), admin.Principal(), group.ID, application.PermissionGrantInput{
		AppLabel: "dcim", Action: "view", Model: "site", ObjectID: &objectID,
	})
	require.NoError(t, err)
	_, err = service.GrantPermissionToUser(t.Context(), admin.Principal(), operator.ID, application.PermissionGrantInput{
		AppLabel: "ipam", Action: "add", Model: "prefix",
	})
	require.NoError(t, err)
	token, err := service.CreateToken(t.Context(), operator.Principal(), application.CreateTokenInput{WriteEnabled: true})
	require.NoError(t, err)
	loaded, err := service.AuthenticatePassword(t.Context(), operator.Username, "Operator-Password-2026!")
	require.NoError(t, err)
	expected := loaded.Principal()

	var restPrincipal domain.Principal
	restHandler := restidentity.NewHandler(service, false)
	router := gin.New()
	router.GET("/capture", restHandler.Middleware(), func(c *gin.Context) {
		restPrincipal, _ = domain.PrincipalFromContext(c.Request.Context())
		c.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/capture", nil)
	request.RemoteAddr = "127.0.0.1:54000"
	request.Header.Set("Authorization", "Token "+token.Secret)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusNoContent, response.Code)

	var grpcPrincipal domain.Principal
	grpcContext := metadata.NewIncomingContext(t.Context(), metadata.Pairs("authorization", "Bearer "+token.Secret))
	grpcContext = peer.NewContext(grpcContext, &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 54001}})
	_, err = grpcidentity.UnaryAuthenticator(service)(grpcContext, nil, &grpc.UnaryServerInfo{FullMethod: "/netbox.dcim.v1.DCIMService/GetSite"}, func(ctx context.Context, _ any) (any, error) {
		grpcPrincipal, _ = domain.PrincipalFromContext(ctx)
		return struct{}{}, nil
	})
	require.NoError(t, err)
	require.Equal(t, expected, restPrincipal)
	require.Equal(t, expected, grpcPrincipal)
}

func TestRESTAndGRPCTokenCredentialParity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	stale := now.Add(-2 * time.Minute)
	expired := now
	revoked := now.Add(-time.Hour)

	activeUser := domain.User{
		ID:          41,
		Username:    "transport-user",
		IsActive:    true,
		Permissions: []string{"dcim.view_site"},
	}
	inactiveUser := activeUser
	inactiveUser.IsActive = false

	token := func() domain.APIToken {
		return domain.APIToken{
			ID:           17,
			UserID:       activeUser.ID,
			WriteEnabled: true,
			LastUsed:     &stale,
		}
	}
	withExpiry := func() domain.APIToken {
		value := token()
		value.Expires = &expired
		return value
	}
	withRestriction := func(writeEnabled bool) domain.APIToken {
		value := token()
		value.WriteEnabled = writeEnabled
		value.AllowedIPs = []string{"192.0.2.0/24"}
		return value
	}
	readOnly := func() domain.APIToken {
		value := token()
		value.WriteEnabled = false
		return value
	}

	lookupFailure := errors.New("credential lookup unavailable")
	touchFailure := errors.New("credential touch unavailable")

	const (
		missingDetail           = "Authentication credentials were not provided."
		invalidDetail           = "Invalid token"
		expiredDetail           = "Token expired"
		inactiveDetail          = "User inactive"
		sourceUnavailableDetail = "Client IP address could not be determined for validation. " +
			"Check that the HTTP server is correctly configured to pass the required header(s)."
		sourceDeniedDetail = "Source IP 198.51.100.1 is not permitted to authenticate using this token."
		permissionDetail   = "You do not have permission to perform this action."
		internalDetail     = "An internal error occurred."
	)

	tests := []tokenCredentialParityCase{
		{
			name:            "missing",
			remoteAddress:   "192.0.2.10:443",
			wantRESTStatus:  http.StatusForbidden,
			wantRESTDetail:  missingDetail,
			wantGRPCCode:    codes.Unauthenticated,
			wantGRPCMessage: missingDetail,
			wantKind:        application.TokenCredentialFailureMissing,
		},
		{
			name:           "valid",
			hasCredential:  true,
			record:         application.TokenRecord{Token: token()},
			user:           activeUser,
			remoteAddress:  "192.0.2.10:443",
			wantRESTStatus: http.StatusNoContent,
			wantGRPCCode:   codes.OK,
			wantHandler:    true,
			wantLookups:    1,
			wantTouches:    1,
		},
		{
			name:            "unknown",
			hasCredential:   true,
			lookupErr:       application.ErrNotFound,
			remoteAddress:   "192.0.2.10:443",
			wantRESTStatus:  http.StatusForbidden,
			wantRESTDetail:  invalidDetail,
			wantGRPCCode:    codes.Unauthenticated,
			wantGRPCMessage: missingDetail,
			wantLookups:     1,
			wantKind:        application.TokenCredentialFailureUnknown,
		},
		{
			name:          "revoked",
			hasCredential: true,
			record: application.TokenRecord{
				Token:     token(),
				RevokedAt: &revoked,
			},
			user:            activeUser,
			remoteAddress:   "192.0.2.10:443",
			wantRESTStatus:  http.StatusForbidden,
			wantRESTDetail:  invalidDetail,
			wantGRPCCode:    codes.Unauthenticated,
			wantGRPCMessage: missingDetail,
			wantLookups:     1,
			wantKind:        application.TokenCredentialFailureRevoked,
		},
		{
			name:            "expired",
			hasCredential:   true,
			record:          application.TokenRecord{Token: withExpiry()},
			user:            activeUser,
			remoteAddress:   "192.0.2.10:443",
			wantRESTStatus:  http.StatusForbidden,
			wantRESTDetail:  expiredDetail,
			wantGRPCCode:    codes.Unauthenticated,
			wantGRPCMessage: missingDetail,
			wantLookups:     1,
			wantTouches:     1,
			wantKind:        application.TokenCredentialFailureExpired,
		},
		{
			name:            "inactive owner",
			hasCredential:   true,
			record:          application.TokenRecord{Token: token()},
			user:            inactiveUser,
			remoteAddress:   "192.0.2.10:443",
			wantRESTStatus:  http.StatusForbidden,
			wantRESTDetail:  inactiveDetail,
			wantGRPCCode:    codes.Unauthenticated,
			wantGRPCMessage: missingDetail,
			wantLookups:     1,
			wantTouches:     1,
			wantKind:        application.TokenCredentialFailureInactiveOwner,
		},
		{
			name:            "source unavailable",
			hasCredential:   true,
			record:          application.TokenRecord{Token: withRestriction(true)},
			user:            activeUser,
			wantRESTStatus:  http.StatusForbidden,
			wantRESTDetail:  sourceUnavailableDetail,
			wantGRPCCode:    codes.Unauthenticated,
			wantGRPCMessage: missingDetail,
			wantLookups:     1,
			wantTouches:     1,
			wantKind:        application.TokenCredentialFailureSourceUnavailable,
		},
		{
			name:            "source denied",
			hasCredential:   true,
			record:          application.TokenRecord{Token: withRestriction(true)},
			user:            activeUser,
			remoteAddress:   "198.51.100.1:443",
			wantRESTStatus:  http.StatusForbidden,
			wantRESTDetail:  sourceDeniedDetail,
			wantGRPCCode:    codes.Unauthenticated,
			wantGRPCMessage: missingDetail,
			wantLookups:     1,
			wantTouches:     1,
			wantKind:        application.TokenCredentialFailureSourceDenied,
		},
		{
			name:            "source denial precedes write denial",
			hasCredential:   true,
			record:          application.TokenRecord{Token: withRestriction(false)},
			user:            activeUser,
			remoteAddress:   "198.51.100.1:443",
			write:           true,
			wantRESTStatus:  http.StatusForbidden,
			wantRESTDetail:  sourceDeniedDetail,
			wantGRPCCode:    codes.Unauthenticated,
			wantGRPCMessage: missingDetail,
			wantLookups:     1,
			wantTouches:     1,
			wantKind:        application.TokenCredentialFailureSourceDenied,
		},
		{
			name:            "write disabled",
			hasCredential:   true,
			record:          application.TokenRecord{Token: readOnly()},
			user:            activeUser,
			remoteAddress:   "192.0.2.10:443",
			write:           true,
			wantRESTStatus:  http.StatusForbidden,
			wantRESTDetail:  permissionDetail,
			wantGRPCCode:    codes.PermissionDenied,
			wantGRPCMessage: permissionDetail,
			wantLookups:     1,
			wantTouches:     1,
		},
		{
			name:            "lookup infrastructure failure",
			hasCredential:   true,
			lookupErr:       lookupFailure,
			remoteAddress:   "192.0.2.10:443",
			wantRESTStatus:  http.StatusInternalServerError,
			wantRESTDetail:  internalDetail,
			wantGRPCCode:    codes.Internal,
			wantGRPCMessage: internalDetail,
			wantLookups:     1,
		},
		{
			name:            "touch infrastructure failure",
			hasCredential:   true,
			record:          application.TokenRecord{Token: token()},
			user:            activeUser,
			touchErr:        touchFailure,
			remoteAddress:   "192.0.2.10:443",
			wantRESTStatus:  http.StatusInternalServerError,
			wantRESTDetail:  internalDetail,
			wantGRPCCode:    codes.Internal,
			wantGRPCMessage: internalDetail,
			wantLookups:     1,
			wantTouches:     1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			restResponse, restObservation := observeRESTTokenCredential(t, now, test)
			grpcCode, grpcMessage, grpcObservation := observeGRPCTokenCredential(t, now, test)

			require.Equal(t, test.wantRESTStatus, restResponse.Code)
			require.Empty(t, restResponse.Header().Get("WWW-Authenticate"))
			if test.wantRESTDetail == "" {
				require.Empty(t, restResponse.Body.String())
			} else {
				require.JSONEq(
					t,
					`{"detail":`+strconv.Quote(test.wantRESTDetail)+`}`,
					restResponse.Body.String(),
				)
			}

			require.Equal(t, test.wantGRPCCode, grpcCode)
			require.Equal(t, test.wantGRPCMessage, grpcMessage)
			require.Equal(t, test.wantHandler, restObservation.handlerCalled)
			require.Equal(t, test.wantHandler, grpcObservation.handlerCalled)
			require.Equal(t, test.wantLookups, restObservation.lookupCalls)
			require.Equal(t, test.wantLookups, grpcObservation.lookupCalls)
			require.Equal(t, test.wantTouches, restObservation.touchCalls)
			require.Equal(t, test.wantTouches, grpcObservation.touchCalls)
			require.Equal(t, restObservation.lookupCalls, grpcObservation.lookupCalls)
			require.Equal(t, restObservation.touchCalls, grpcObservation.touchCalls)
			wantEvents := expectedTokenCredentialEvents(test.wantLookups, test.wantTouches)
			require.Equal(t, wantEvents, restObservation.events)
			require.Equal(t, wantEvents, grpcObservation.events)
			require.Equal(t, wantEvents, restObservation.application.events)
			require.Equal(t, wantEvents, grpcObservation.application.events)
			require.Equal(t, test.wantKind, restObservation.application.kind)
			require.Equal(t, test.wantKind, grpcObservation.application.kind)
			require.Equal(t, restObservation.application.kind, grpcObservation.application.kind)

			if test.wantHandler {
				expectedPrincipal := test.user.Principal()
				require.Equal(t, expectedPrincipal, restObservation.principal)
				require.Equal(t, expectedPrincipal, grpcObservation.principal)
				return
			}
			require.Equal(t, domain.Principal{}, restObservation.principal)
			require.Equal(t, domain.Principal{}, grpcObservation.principal)
		})
	}
}
