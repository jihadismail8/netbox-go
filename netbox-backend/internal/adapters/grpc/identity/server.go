package identity

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	identityv1 "netbox-go/gen/go/netbox/identity/v1"
	typesv1 "netbox-go/gen/go/netbox/types/v1"
	"netbox-go/internal/adapters/grpc/statusmap"
	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
)

type Server struct {
	identityv1.UnimplementedIdentityServiceServer
	service *application.Service
}

func NewServer(service *application.Service) *Server {
	if service == nil {
		panic("identity gRPC server requires service")
	}
	return &Server{service: service}
}
func (s *Server) GetCurrentUser(ctx context.Context, _ *identityv1.GetCurrentUserRequest) (*identityv1.GetCurrentUserResponse, error) {
	principal, ok := domain.PrincipalFromContext(ctx)
	if !ok {
		return nil, statusmap.Error(unauthenticatedError())
	}
	user, err := s.service.CurrentUser(ctx, principal)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &identityv1.GetCurrentUserResponse{User: userProto(user)}, nil
}
func (s *Server) ListAPITokens(ctx context.Context, request *identityv1.ListAPITokensRequest) (*identityv1.ListAPITokensResponse, error) {
	principal, ok := domain.PrincipalFromContext(ctx)
	if !ok {
		return nil, statusmap.Error(unauthenticatedError())
	}
	limit, offset := page(request.GetPage())
	tokens, count, err := s.service.ListTokens(ctx, principal, limit, offset)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	results := make([]*identityv1.APIToken, 0, len(tokens))
	for _, token := range tokens {
		results = append(results, tokenProto(token))
	}
	return &identityv1.ListAPITokensResponse{Page: &typesv1.PageInfo{Count: uint64(count)}, Results: results}, nil
}
func (s *Server) CreateAPIToken(ctx context.Context, request *identityv1.CreateAPITokenRequest) (*identityv1.CreateAPITokenResponse, error) {
	principal, ok := domain.PrincipalFromContext(ctx)
	if !ok {
		return nil, statusmap.Error(unauthenticatedError())
	}
	var expires *time.Time
	if request.Expires != nil {
		value := request.Expires.AsTime()
		expires = &value
	}
	created, err := s.service.CreateToken(ctx, principal, application.CreateTokenInput{Description: request.Description, WriteEnabled: request.WriteEnabled, Expires: expires, AllowedIPs: request.AllowedIps})
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &identityv1.CreateAPITokenResponse{Token: tokenProto(created.Token), Secret: created.Secret}, nil
}
func (s *Server) RevokeAPIToken(ctx context.Context, request *identityv1.RevokeAPITokenRequest) (*identityv1.RevokeAPITokenResponse, error) {
	principal, ok := domain.PrincipalFromContext(ctx)
	if !ok {
		return nil, statusmap.Error(unauthenticatedError())
	}
	if err := s.service.RevokeToken(ctx, principal, request.Id); err != nil {
		return nil, statusmap.Error(err)
	}
	return &identityv1.RevokeAPITokenResponse{}, nil
}
func (s *Server) ChangePassword(ctx context.Context, request *identityv1.ChangePasswordRequest) (*identityv1.ChangePasswordResponse, error) {
	principal, ok := domain.PrincipalFromContext(ctx)
	if !ok {
		return nil, statusmap.Error(unauthenticatedError())
	}
	_, err := s.service.ChangePassword(
		ctx,
		principal,
		application.NewPasswordChangeInput(
			request.CurrentPassword,
			request.NewPassword,
			application.APITokenPasswordChangeCredential(),
		),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &identityv1.ChangePasswordResponse{}, nil
}

func userProto(user domain.User) *identityv1.User {
	return &identityv1.User{Id: user.ID, Username: user.Username, Email: user.Email, FirstName: user.FirstName, LastName: user.LastName, IsStaff: user.IsStaff, IsSuperuser: user.IsSuperuser, Permissions: user.Permissions}
}
func tokenProto(token domain.APIToken) *identityv1.APIToken {
	result := &identityv1.APIToken{Id: token.ID, Display: token.Display, Description: token.Description, WriteEnabled: token.WriteEnabled, Created: timestamppb.New(token.Created), AllowedIps: token.AllowedIPs}
	if token.Expires != nil {
		result.Expires = timestamppb.New(*token.Expires)
	}
	if token.LastUsed != nil {
		result.LastUsed = timestamppb.New(*token.LastUsed)
	}
	return result
}
func page(request *typesv1.PageRequest) (int, int) {
	if request == nil {
		return 50, 0
	}
	limit := 50
	if request.Limit != nil {
		limit = int(*request.Limit)
	}
	offset := 0
	if request.Offset != nil {
		offset = int(*request.Offset)
	}
	return limit, offset
}
