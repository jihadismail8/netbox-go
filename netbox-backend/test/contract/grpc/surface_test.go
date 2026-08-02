package grpccontract

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	identityv1 "netbox-go/gen/go/netbox/identity/v1"
	ipamv1 "netbox-go/gen/go/netbox/ipam/v1"
)

func TestCanonicalGRPCSurfaceIsBoundedByCapabilityServices(t *testing.T) {
	dcimResources := [][2]string{
		{"Site", "Sites"},
		{"Manufacturer", "Manufacturers"},
		{"RackRole", "RackRoles"},
		{"RackType", "RackTypes"},
		{"Rack", "Racks"},
		{"DeviceRole", "DeviceRoles"},
		{"DeviceType", "DeviceTypes"},
		{"InterfaceTemplate", "InterfaceTemplates"},
		{"Device", "Devices"},
		{"Interface", "Interfaces"},
	}
	ipamResources := [][2]string{{"VRF", "VRFs"}, {"Prefix", "Prefixes"}, {"IPAddress", "IPAddresses"}}

	expectedDCIM := resourceMethods(dcimResources)
	expectedIPAM := append(resourceMethods(ipamResources), "AssignIPAddress", "UnassignIPAddress")
	expectedIdentity := []string{"GetCurrentUser", "ListAPITokens", "CreateAPIToken", "RevokeAPIToken", "ChangePassword"}
	require.Equal(t, sorted(expectedDCIM), methodNames(&dcimv1.DCIMService_ServiceDesc))
	require.Equal(t, sorted(expectedIPAM), methodNames(&ipamv1.IPAMService_ServiceDesc))
	require.Equal(t, sorted(expectedIdentity), methodNames(&identityv1.IdentityService_ServiceDesc))
	for _, service := range []*grpc.ServiceDesc{&dcimv1.DCIMService_ServiceDesc, &ipamv1.IPAMService_ServiceDesc, &identityv1.IdentityService_ServiceDesc} {
		require.Empty(t, service.Streams, "the first profile exposes unary methods only")
	}
}

func resourceMethods(resources [][2]string) []string {
	methods := make([]string, 0, len(resources)*6)
	for _, resource := range resources {
		singular, plural := resource[0], resource[1]
		methods = append(methods, "List"+plural, "Get"+singular, "Create"+singular, "Replace"+singular, "Update"+singular, "Delete"+singular)
	}
	return methods
}

func methodNames(service *grpc.ServiceDesc) []string {
	methods := make([]string, 0, len(service.Methods))
	for _, method := range service.Methods {
		methods = append(methods, method.MethodName)
	}
	return sorted(methods)
}

func sorted(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
