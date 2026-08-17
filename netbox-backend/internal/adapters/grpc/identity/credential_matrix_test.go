package identity

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
)

type grpcCredentialClock struct{ now time.Time }

func (clock grpcCredentialClock) Now() time.Time { return clock.now }

type grpcCredentialStore struct {
	application.Store
	record    application.TokenRecord
	user      domain.User
	lookupErr error
	touchErr  error
}

func (store *grpcCredentialStore) TokenByHash(context.Context, []byte) (application.TokenRecord, domain.User, error) {
	return store.record, store.user, store.lookupErr
}

func (store *grpcCredentialStore) TouchToken(context.Context, int64, time.Time) error {
	return store.touchErr
}

func TestGRPCTokenCredentialMatrixMappings(t *testing.T) {
	now := time.Date(2026, 8, 17, 7, 0, 0, 0, time.UTC)
	infrastructureFailure := errors.New("credential storage unavailable")
	current := now

	tests := []struct {
		name     string
		store    *grpcCredentialStore
		remote   net.IP
		method   string
		wantCode codes.Code
	}{
		{
			name: "lookup failure",
			store: &grpcCredentialStore{
				lookupErr: infrastructureFailure,
			},
			remote:   net.ParseIP("192.0.2.1"),
			method:   "/netbox.dcim.v1.DCIMService/GetSite",
			wantCode: codes.Internal,
		},
		{
			name: "touch failure",
			store: &grpcCredentialStore{
				record: application.TokenRecord{Token: domain.APIToken{
					ID: 17, UserID: 41, WriteEnabled: true,
				}},
				user:     domain.User{ID: 41, Username: "grpc-user", IsActive: true},
				touchErr: infrastructureFailure,
			},
			remote:   net.ParseIP("192.0.2.1"),
			method:   "/netbox.dcim.v1.DCIMService/GetSite",
			wantCode: codes.Internal,
		},
		{
			name: "source denial precedes write denial",
			store: &grpcCredentialStore{
				record: application.TokenRecord{Token: domain.APIToken{
					ID: 17, UserID: 41, WriteEnabled: false,
					AllowedIPs: []string{"192.0.2.0/24"}, LastUsed: &current,
				}},
				user: domain.User{ID: 41, Username: "grpc-user", IsActive: true},
			},
			remote:   net.ParseIP("198.51.100.1"),
			method:   "/netbox.identity.v1.IdentityService/CreateAPIToken",
			wantCode: codes.Unauthenticated,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := application.NewService(test.store, grpcCredentialClock{now: now})
			interceptor := UnaryAuthenticator(service)
			ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("authorization", "Bearer present"))
			ctx = peer.NewContext(ctx, &peer.Peer{Addr: &net.TCPAddr{IP: test.remote, Port: 443}})
			handlerCalled := false

			response, err := interceptor(ctx, struct{}{}, &grpc.UnaryServerInfo{
				FullMethod: test.method,
			}, func(context.Context, any) (any, error) {
				handlerCalled = true
				return struct{}{}, nil
			})

			require.Nil(t, response)
			require.Equal(t, test.wantCode, status.Code(err))
			if test.wantCode == codes.Internal {
				require.Equal(t, "An internal error occurred.", status.Convert(err).Message())
				require.NotContains(t, status.Convert(err).Message(), infrastructureFailure.Error())
			}
			require.False(t, handlerCalled)
		})
	}
}
