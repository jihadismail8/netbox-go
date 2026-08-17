package identity

import (
	"context"

	"google.golang.org/grpc"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	identityv1 "netbox-go/gen/go/netbox/identity/v1"
	ipamv1 "netbox-go/gen/go/netbox/ipam/v1"
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
		if info.FullMethod == healthv1.Health_Check_FullMethodName {
			return handler(ctx, request)
		}
		values := metadata.ValueFromIncomingContext(ctx, "authorization")
		if len(values) != 1 {
			return nil, statusmap.Error(unauthenticatedError())
		}
		credential, ok := parseBearerCredential(values[0])
		if !ok {
			return nil, statusmap.Error(unauthenticatedError())
		}
		remote := ""
		if value, ok := peer.FromContext(ctx); ok && value != nil && value.Addr != nil {
			remote = value.Addr.String()
		}
		write := !isUnaryReadMethod(info.FullMethod)
		user, err := service.AuthenticateToken(ctx, credential, remote, write)
		if err != nil {
			return nil, statusmap.Error(err)
		}
		return handler(domain.WithPrincipal(ctx, user.Principal()), request)
	}
}

func parseBearerCredential(value string) (string, bool) {
	const scheme = "Bearer"
	if len(value) <= len(scheme) || !equalFoldASCII(value[:len(scheme)], scheme) {
		return "", false
	}

	credentialStart := len(scheme)
	if value[credentialStart] != ' ' {
		return "", false
	}
	for credentialStart < len(value) && value[credentialStart] == ' ' {
		credentialStart++
	}
	if credentialStart == len(value) {
		return "", false
	}
	for index := credentialStart; index < len(value); index++ {
		if value[index] < 0x21 || value[index] > 0x7e {
			return "", false
		}
	}
	return value[credentialStart:], true
}

func equalFoldASCII(value, expected string) bool {
	if len(value) != len(expected) {
		return false
	}
	for index := 0; index < len(value); index++ {
		left := value[index]
		right := expected[index]
		if left >= 'A' && left <= 'Z' {
			left += 'a' - 'A'
		}
		if right >= 'A' && right <= 'Z' {
			right += 'a' - 'A'
		}
		if left != right {
			return false
		}
	}
	return true
}

func isUnaryReadMethod(fullMethod string) bool {
	switch fullMethod {
	case identityv1.IdentityService_GetCurrentUser_FullMethodName,
		identityv1.IdentityService_ListAPITokens_FullMethodName,
		dcimv1.DCIMService_ListSites_FullMethodName,
		dcimv1.DCIMService_GetSite_FullMethodName,
		dcimv1.DCIMService_ListManufacturers_FullMethodName,
		dcimv1.DCIMService_GetManufacturer_FullMethodName,
		dcimv1.DCIMService_ListRackRoles_FullMethodName,
		dcimv1.DCIMService_GetRackRole_FullMethodName,
		dcimv1.DCIMService_ListRackTypes_FullMethodName,
		dcimv1.DCIMService_GetRackType_FullMethodName,
		dcimv1.DCIMService_ListRacks_FullMethodName,
		dcimv1.DCIMService_GetRack_FullMethodName,
		dcimv1.DCIMService_ListDeviceRoles_FullMethodName,
		dcimv1.DCIMService_GetDeviceRole_FullMethodName,
		dcimv1.DCIMService_ListDeviceTypes_FullMethodName,
		dcimv1.DCIMService_GetDeviceType_FullMethodName,
		dcimv1.DCIMService_ListInterfaceTemplates_FullMethodName,
		dcimv1.DCIMService_GetInterfaceTemplate_FullMethodName,
		dcimv1.DCIMService_ListDevices_FullMethodName,
		dcimv1.DCIMService_GetDevice_FullMethodName,
		dcimv1.DCIMService_ListInterfaces_FullMethodName,
		dcimv1.DCIMService_GetInterface_FullMethodName,
		ipamv1.IPAMService_ListVRFs_FullMethodName,
		ipamv1.IPAMService_GetVRF_FullMethodName,
		ipamv1.IPAMService_ListPrefixes_FullMethodName,
		ipamv1.IPAMService_GetPrefix_FullMethodName,
		ipamv1.IPAMService_ListIPAddresses_FullMethodName,
		ipamv1.IPAMService_GetIPAddress_FullMethodName:
		return true
	default:
		return false
	}
}

func unauthenticatedError() error {
	return shared.NewError(
		shared.ErrorReasonUnauthenticated,
		"Authentication credentials were not provided.",
	)
}
