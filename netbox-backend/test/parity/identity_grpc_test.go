package parity

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
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

	identityv1 "netbox-go/gen/go/netbox/identity/v1"
	grpcidentity "netbox-go/internal/adapters/grpc/identity"
	postgresidentity "netbox-go/internal/adapters/postgres/identity"
	restidentity "netbox-go/internal/adapters/rest/netbox/identity"
	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
)

type identityFixedClock struct{ now time.Time }

func (clock identityFixedClock) Now() time.Time { return clock.now }

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
