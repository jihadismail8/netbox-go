package identity

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	identityv1 "netbox-go/gen/go/netbox/identity/v1"
	ipamv1 "netbox-go/gen/go/netbox/ipam/v1"
	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
)

type grpcTokenTransportClock struct{ now time.Time }

func (clock grpcTokenTransportClock) Now() time.Time { return clock.now }

type grpcTokenTransportTouch struct {
	id int64
	at time.Time
}

type grpcTokenTransportStore struct {
	application.Store
	record       application.TokenRecord
	user         domain.User
	lookupErr    error
	touchErr     error
	lookupCalls  int
	touches      []grpcTokenTransportTouch
	expectedHash []byte
	hashMatched  bool
}

func (store *grpcTokenTransportStore) TokenByHash(_ context.Context, hash []byte) (application.TokenRecord, domain.User, error) {
	store.lookupCalls++
	if store.expectedHash != nil {
		store.hashMatched = subtle.ConstantTimeCompare(store.expectedHash, hash) == 1
	}
	return store.record, store.user, store.lookupErr
}

func (store *grpcTokenTransportStore) TouchToken(_ context.Context, id int64, at time.Time) error {
	store.touches = append(store.touches, grpcTokenTransportTouch{id: id, at: at})
	return store.touchErr
}

type grpcTokenTransportAddress string

func (grpcTokenTransportAddress) Network() string { return "test" }

func (address grpcTokenTransportAddress) String() string { return string(address) }

func grpcTokenTransportContext(ctx context.Context, authorization []string, address net.Addr) context.Context {
	if authorization != nil {
		ctx = metadata.NewIncomingContext(ctx, metadata.MD{
			"authorization": append([]string(nil), authorization...),
		})
	}
	if address != nil {
		ctx = peer.NewContext(ctx, &peer.Peer{Addr: address})
	}
	return ctx
}

func grpcTokenTransportRecord(now time.Time, writeEnabled bool) application.TokenRecord {
	lastUsed := now
	return application.TokenRecord{Token: domain.APIToken{
		ID:           17,
		UserID:       41,
		WriteEnabled: writeEnabled,
		LastUsed:     &lastUsed,
	}}
}

func grpcTokenTransportUser() domain.User {
	return domain.User{ID: 41, Username: "grpc-user", IsActive: true}
}

func TestUnaryAuthenticatorBearerMetadataGrammar(t *testing.T) {
	now := time.Date(2026, 8, 17, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		authorization  []string
		wantCode       codes.Code
		wantLookups    int
		wantHandler    bool
		wantCredential string
	}{
		{name: "absent", wantCode: codes.Unauthenticated},
		{name: "empty field value", authorization: []string{""}, wantCode: codes.Unauthenticated},
		{name: "unsupported scheme", authorization: []string{"Token opaque"}, wantCode: codes.Unauthenticated},
		{name: "scheme without separator", authorization: []string{"Beareropaque"}, wantCode: codes.Unauthenticated},
		{name: "empty credential", authorization: []string{"Bearer "}, wantCode: codes.Unauthenticated},
		{name: "leading whitespace", authorization: []string{" Bearer opaque"}, wantCode: codes.Unauthenticated},
		{name: "trailing whitespace", authorization: []string{"Bearer opaque "}, wantCode: codes.Unauthenticated},
		{name: "tab separator", authorization: []string{"Bearer\topaque"}, wantCode: codes.Unauthenticated},
		{name: "control after separator", authorization: []string{"Bearer \topaque"}, wantCode: codes.Unauthenticated},
		{name: "embedded space", authorization: []string{"Bearer opaque extra"}, wantCode: codes.Unauthenticated},
		{name: "control in credential", authorization: []string{"Bearer opaque\x00"}, wantCode: codes.Unauthenticated},
		{name: "delete in credential", authorization: []string{"Bearer opaque\x7f"}, wantCode: codes.Unauthenticated},
		{name: "non ascii credential", authorization: []string{"Bearer opaque-\u00e9"}, wantCode: codes.Unauthenticated},
		{name: "duplicate values", authorization: []string{"Bearer opaque", "Bearer alternate"}, wantCode: codes.Unauthenticated},
		{name: "canonical", authorization: []string{"Bearer opaque"}, wantCode: codes.OK, wantLookups: 1, wantHandler: true, wantCredential: "opaque"},
		{name: "ascii case folded scheme", authorization: []string{"bEaReR opaque"}, wantCode: codes.OK, wantLookups: 1, wantHandler: true, wantCredential: "opaque"},
		{name: "multiple separator spaces", authorization: []string{"Bearer    opaque"}, wantCode: codes.OK, wantLookups: 1, wantHandler: true, wantCredential: "opaque"},
		{name: "visible opaque punctuation", authorization: []string{"Bearer opaque:~!"}, wantCode: codes.OK, wantLookups: 1, wantHandler: true, wantCredential: "opaque:~!"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wantHash := sha256.Sum256([]byte(test.wantCredential))
			store := &grpcTokenTransportStore{
				record: grpcTokenTransportRecord(now, true),
				user:   grpcTokenTransportUser(),
			}
			if test.wantCredential != "" {
				store.expectedHash = wantHash[:]
			}
			service := application.NewService(store, grpcTokenTransportClock{now: now})
			interceptor := UnaryAuthenticator(service)
			ctx := grpcTokenTransportContext(t.Context(), test.authorization, nil)
			handlerCalled := false
			responseMarker := &struct{}{}

			response, err := interceptor(
				ctx,
				struct{}{},
				&grpc.UnaryServerInfo{FullMethod: identityv1.IdentityService_GetCurrentUser_FullMethodName},
				func(handlerContext context.Context, _ any) (any, error) {
					handlerCalled = true
					principal, ok := domain.PrincipalFromContext(handlerContext)
					require.True(t, ok)
					require.Equal(t, int64(41), principal.ID)
					return responseMarker, nil
				},
			)

			require.Equal(t, test.wantLookups, store.lookupCalls)
			require.Empty(t, store.touches)
			require.Equal(t, test.wantHandler, handlerCalled)
			require.Equal(t, test.wantCode, status.Code(err))
			if test.wantCode == codes.OK {
				require.True(t, store.hashMatched, "application received an altered opaque credential")
				require.Same(t, responseMarker, response)
				return
			}
			require.Nil(t, response)
			require.Equal(t, "Authentication credentials were not provided.", status.Convert(err).Message())
		})
	}
}

func TestUnaryAuthenticatorRPCSafetyClassification(t *testing.T) {
	now := time.Date(2026, 8, 17, 14, 15, 0, 0, time.UTC)
	readMethods := map[string]struct{}{
		identityv1.IdentityService_GetCurrentUser_FullMethodName: {},
		identityv1.IdentityService_ListAPITokens_FullMethodName:  {},

		dcimv1.DCIMService_ListSites_FullMethodName:              {},
		dcimv1.DCIMService_GetSite_FullMethodName:                {},
		dcimv1.DCIMService_ListManufacturers_FullMethodName:      {},
		dcimv1.DCIMService_GetManufacturer_FullMethodName:        {},
		dcimv1.DCIMService_ListRackRoles_FullMethodName:          {},
		dcimv1.DCIMService_GetRackRole_FullMethodName:            {},
		dcimv1.DCIMService_ListRackTypes_FullMethodName:          {},
		dcimv1.DCIMService_GetRackType_FullMethodName:            {},
		dcimv1.DCIMService_ListRacks_FullMethodName:              {},
		dcimv1.DCIMService_GetRack_FullMethodName:                {},
		dcimv1.DCIMService_ListDeviceRoles_FullMethodName:        {},
		dcimv1.DCIMService_GetDeviceRole_FullMethodName:          {},
		dcimv1.DCIMService_ListDeviceTypes_FullMethodName:        {},
		dcimv1.DCIMService_GetDeviceType_FullMethodName:          {},
		dcimv1.DCIMService_ListInterfaceTemplates_FullMethodName: {},
		dcimv1.DCIMService_GetInterfaceTemplate_FullMethodName:   {},
		dcimv1.DCIMService_ListDevices_FullMethodName:            {},
		dcimv1.DCIMService_GetDevice_FullMethodName:              {},
		dcimv1.DCIMService_ListInterfaces_FullMethodName:         {},
		dcimv1.DCIMService_GetInterface_FullMethodName:           {},

		ipamv1.IPAMService_ListVRFs_FullMethodName:        {},
		ipamv1.IPAMService_GetVRF_FullMethodName:          {},
		ipamv1.IPAMService_ListPrefixes_FullMethodName:    {},
		ipamv1.IPAMService_GetPrefix_FullMethodName:       {},
		ipamv1.IPAMService_ListIPAddresses_FullMethodName: {},
		ipamv1.IPAMService_GetIPAddress_FullMethodName:    {},
	}
	services := []struct {
		name       string
		descriptor *grpc.ServiceDesc
		wantReads  int
		wantWrites int
	}{
		{name: "dcim", descriptor: &dcimv1.DCIMService_ServiceDesc, wantReads: 20, wantWrites: 40},
		{name: "ipam", descriptor: &ipamv1.IPAMService_ServiceDesc, wantReads: 6, wantWrites: 14},
		{name: "identity", descriptor: &identityv1.IdentityService_ServiceDesc, wantReads: 2, wantWrites: 3},
	}

	seenReadMethods := make(map[string]struct{}, len(readMethods))
	totalReads := 0
	totalWrites := 0
	totalProtected := 0
	for _, serviceCase := range services {
		reads := 0
		writes := 0
		require.Empty(t, serviceCase.descriptor.Streams)
		for _, method := range serviceCase.descriptor.Methods {
			fullMethod := "/" + serviceCase.descriptor.ServiceName + "/" + method.MethodName
			_, read := readMethods[fullMethod]
			if read {
				reads++
				seenReadMethods[fullMethod] = struct{}{}
			} else {
				writes++
			}
			totalProtected++

			t.Run(serviceCase.name+"/"+method.MethodName, func(t *testing.T) {
				store := &grpcTokenTransportStore{
					record: grpcTokenTransportRecord(now, false),
					user:   grpcTokenTransportUser(),
				}
				interceptor := UnaryAuthenticator(application.NewService(store, grpcTokenTransportClock{now: now}))
				handlerCalled := false
				handler := func(handlerContext context.Context, _ any) (any, error) {
					handlerCalled = true
					principal, ok := domain.PrincipalFromContext(handlerContext)
					require.True(t, ok)
					require.Equal(t, int64(41), principal.ID)
					return &struct{}{}, nil
				}

				response, err := interceptor(
					t.Context(),
					struct{}{},
					&grpc.UnaryServerInfo{FullMethod: fullMethod},
					handler,
				)
				require.Nil(t, response)
				require.Equal(t, 0, store.lookupCalls)
				require.False(t, handlerCalled)
				require.Equal(t, codes.Unauthenticated, status.Code(err))
				require.Equal(t, "Authentication credentials were not provided.", status.Convert(err).Message())

				handlerCalled = false
				ctx := grpcTokenTransportContext(t.Context(), []string{"Bearer opaque"}, nil)
				response, err = interceptor(
					ctx,
					struct{}{},
					&grpc.UnaryServerInfo{FullMethod: fullMethod},
					handler,
				)
				require.Equal(t, 1, store.lookupCalls)
				require.Empty(t, store.touches)
				if read {
					require.True(t, handlerCalled)
					require.Equal(t, codes.OK, status.Code(err))
					require.NotNil(t, response)
					return
				}
				require.False(t, handlerCalled)
				require.Nil(t, response)
				require.Equal(t, codes.PermissionDenied, status.Code(err))
				require.Equal(t, "You do not have permission to perform this action.", status.Convert(err).Message())
			})
		}

		require.Equal(t, serviceCase.wantReads, reads)
		require.Equal(t, serviceCase.wantWrites, writes)
		require.Equal(t, serviceCase.wantReads+serviceCase.wantWrites, len(serviceCase.descriptor.Methods))
		totalReads += reads
		totalWrites += writes
	}

	require.Equal(t, 28, totalReads)
	require.Equal(t, 57, totalWrites)
	require.Equal(t, 85, totalProtected)
	require.Len(t, seenReadMethods, len(readMethods))
	for method := range readMethods {
		require.Contains(t, seenReadMethods, method)
	}

	t.Run("health check is public", func(t *testing.T) {
		handlerCalled := false
		responseMarker := &struct{}{}
		response, err := UnaryAuthenticator(nil)(
			t.Context(),
			struct{}{},
			&grpc.UnaryServerInfo{FullMethod: healthv1.Health_Check_FullMethodName},
			func(context.Context, any) (any, error) {
				handlerCalled = true
				return responseMarker, nil
			},
		)
		require.NoError(t, err)
		require.True(t, handlerCalled)
		require.Same(t, responseMarker, response)
	})

	for _, fullMethod := range []string{
		"/netbox.unknown.v1.Unknown/Getaway",
		"/netbox.unknown.v1.Unknown/ListSomething",
	} {
		t.Run("unknown method defaults to write/"+fullMethod, func(t *testing.T) {
			store := &grpcTokenTransportStore{
				record: grpcTokenTransportRecord(now, false),
				user:   grpcTokenTransportUser(),
			}
			interceptor := UnaryAuthenticator(application.NewService(store, grpcTokenTransportClock{now: now}))
			handlerCalled := false
			handler := func(context.Context, any) (any, error) {
				handlerCalled = true
				return &struct{}{}, nil
			}

			response, err := interceptor(
				t.Context(),
				struct{}{},
				&grpc.UnaryServerInfo{FullMethod: fullMethod},
				handler,
			)
			require.Nil(t, response)
			require.Equal(t, 0, store.lookupCalls)
			require.False(t, handlerCalled)
			require.Equal(t, codes.Unauthenticated, status.Code(err))

			ctx := grpcTokenTransportContext(t.Context(), []string{"Bearer opaque"}, nil)
			response, err = interceptor(
				ctx,
				struct{}{},
				&grpc.UnaryServerInfo{FullMethod: fullMethod},
				handler,
			)
			require.Equal(t, 1, store.lookupCalls)
			require.False(t, handlerCalled)
			require.Nil(t, response)
			require.Equal(t, codes.PermissionDenied, status.Code(err))
			require.Equal(t, "You do not have permission to perform this action.", status.Convert(err).Message())
		})
	}
}

func TestUnaryAuthenticatorCredentialOutcomeMappings(t *testing.T) {
	now := time.Date(2026, 8, 17, 14, 30, 0, 0, time.UTC)
	stale := now.Add(-2 * time.Minute)
	revoked := now.Add(-time.Hour)
	backendFailure := errors.New("credential backend unavailable")

	tests := []struct {
		name        string
		method      string
		address     net.Addr
		configure   func(*grpcTokenTransportStore)
		wantCode    codes.Code
		wantMessage string
		wantLookups int
		wantTouches int
		wantHandler bool
		conceal     error
	}{
		{
			name: "unknown",
			configure: func(store *grpcTokenTransportStore) {
				store.lookupErr = application.ErrNotFound
			},
			wantCode: codes.Unauthenticated, wantMessage: "Authentication credentials were not provided.", wantLookups: 1,
		},
		{
			name: "revoked",
			configure: func(store *grpcTokenTransportStore) {
				store.record.RevokedAt = &revoked
			},
			wantCode: codes.Unauthenticated, wantMessage: "Authentication credentials were not provided.", wantLookups: 1,
		},
		{
			name: "expired",
			configure: func(store *grpcTokenTransportStore) {
				store.record.Token.LastUsed = &stale
				store.record.Token.Expires = &now
			},
			wantCode: codes.Unauthenticated, wantMessage: "Authentication credentials were not provided.", wantLookups: 1, wantTouches: 1,
		},
		{
			name: "inactive owner",
			configure: func(store *grpcTokenTransportStore) {
				store.record.Token.LastUsed = &stale
				store.user.IsActive = false
			},
			wantCode: codes.Unauthenticated, wantMessage: "Authentication credentials were not provided.", wantLookups: 1, wantTouches: 1,
		},
		{
			name: "source unavailable",
			configure: func(store *grpcTokenTransportStore) {
				store.record.Token.LastUsed = &stale
				store.record.Token.AllowedIPs = []string{"192.0.2.0/24"}
			},
			wantCode: codes.Unauthenticated, wantMessage: "Authentication credentials were not provided.", wantLookups: 1, wantTouches: 1,
		},
		{
			name:    "source malformed",
			address: grpcTokenTransportAddress("not-an-address"),
			configure: func(store *grpcTokenTransportStore) {
				store.record.Token.LastUsed = &stale
				store.record.Token.AllowedIPs = []string{"192.0.2.0/24"}
			},
			wantCode: codes.Unauthenticated, wantMessage: "Authentication credentials were not provided.", wantLookups: 1, wantTouches: 1,
		},
		{
			name:    "source denied",
			address: &net.TCPAddr{IP: net.ParseIP("198.51.100.1"), Port: 443},
			configure: func(store *grpcTokenTransportStore) {
				store.record.Token.LastUsed = &stale
				store.record.Token.AllowedIPs = []string{"192.0.2.0/24"}
			},
			wantCode: codes.Unauthenticated, wantMessage: "Authentication credentials were not provided.", wantLookups: 1, wantTouches: 1,
		},
		{
			name:   "write disabled mutation",
			method: identityv1.IdentityService_CreateAPIToken_FullMethodName,
			configure: func(store *grpcTokenTransportStore) {
				store.record.Token.LastUsed = &stale
				store.record.Token.WriteEnabled = false
			},
			wantCode: codes.PermissionDenied, wantMessage: "You do not have permission to perform this action.", wantLookups: 1, wantTouches: 1,
		},
		{
			name: "lookup infrastructure failure",
			configure: func(store *grpcTokenTransportStore) {
				store.lookupErr = backendFailure
			},
			wantCode: codes.Internal, wantMessage: "An internal error occurred.", wantLookups: 1, conceal: backendFailure,
		},
		{
			name: "touch infrastructure failure",
			configure: func(store *grpcTokenTransportStore) {
				store.record.Token.LastUsed = &stale
				store.touchErr = backendFailure
			},
			wantCode: codes.Internal, wantMessage: "An internal error occurred.", wantLookups: 1, wantTouches: 1, conceal: backendFailure,
		},
		{
			name:     "valid principal",
			wantCode: codes.OK, wantLookups: 1, wantHandler: true,
		},
		{
			name:     "valid write enabled mutation",
			method:   identityv1.IdentityService_CreateAPIToken_FullMethodName,
			wantCode: codes.OK, wantLookups: 1, wantHandler: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &grpcTokenTransportStore{
				record: grpcTokenTransportRecord(now, true),
				user:   grpcTokenTransportUser(),
			}
			if test.configure != nil {
				test.configure(store)
			}
			method := test.method
			if method == "" {
				method = identityv1.IdentityService_GetCurrentUser_FullMethodName
			}
			interceptor := UnaryAuthenticator(application.NewService(store, grpcTokenTransportClock{now: now}))
			ctx := grpcTokenTransportContext(t.Context(), []string{"Bearer opaque"}, test.address)
			handlerCalled := false
			responseMarker := &struct{}{}

			response, err := interceptor(
				ctx,
				struct{}{},
				&grpc.UnaryServerInfo{FullMethod: method},
				func(handlerContext context.Context, _ any) (any, error) {
					handlerCalled = true
					principal, ok := domain.PrincipalFromContext(handlerContext)
					require.True(t, ok)
					require.Equal(t, int64(41), principal.ID)
					return responseMarker, nil
				},
			)

			require.Equal(t, test.wantLookups, store.lookupCalls)
			require.Len(t, store.touches, test.wantTouches)
			if test.wantTouches == 1 {
				require.Equal(t, grpcTokenTransportTouch{id: 17, at: now}, store.touches[0])
			}
			require.Equal(t, test.wantHandler, handlerCalled)
			require.Equal(t, test.wantCode, status.Code(err))
			if test.wantCode == codes.OK {
				require.Same(t, responseMarker, response)
				return
			}
			require.Nil(t, response)
			require.Equal(t, test.wantMessage, status.Convert(err).Message())
			if test.conceal != nil {
				require.NotContains(t, status.Convert(err).Message(), test.conceal.Error())
			}
		})
	}
}
