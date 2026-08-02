package identity

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"netbox-go/internal/adapters/grpc/statusmap"
	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func UnaryAuthenticator(service *application.Service) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Health is an operational readiness contract rather than a Managed
		// Object capability. It must remain callable by orchestrators before
		// they have application credentials.
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, request)
		}
		values := metadata.ValueFromIncomingContext(ctx, "authorization")
		if len(values) != 1 || !strings.HasPrefix(values[0], "Bearer ") {
			return nil, statusmap.Error(unauthenticatedError())
		}
		remote := ""
		if value, ok := peer.FromContext(ctx); ok {
			remote = value.Addr.String()
		}
		write := !strings.Contains(info.FullMethod, "/List") && !strings.Contains(info.FullMethod, "/Get")
		user, err := service.AuthenticateToken(ctx, strings.TrimSpace(strings.TrimPrefix(values[0], "Bearer ")), remote, write)
		if err != nil {
			return nil, statusmap.Error(err)
		}
		return handler(domain.WithPrincipal(ctx, user.Principal()), request)
	}
}

func unauthenticatedError() error {
	return shared.NewError(
		shared.ErrorReasonUnauthenticated,
		"Authentication credentials were not provided.",
	)
}
