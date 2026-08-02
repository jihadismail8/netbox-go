package workflow

import (
	"context"

	ipamv1 "netbox-go/gen/go/netbox/ipam/v1"
)

// IPAMServer is a transport-only dispatcher over the typed IPAM application
// services. The first-profile runtime has no generic map-shaped fallback.
type IPAMServer struct {
	ipamv1.UnimplementedIPAMServiceServer
	vrfs      *VRFRPCHandler
	prefixes  *PrefixRPCHandler
	addresses *IPAddressRPCHandler
}

var _ ipamv1.IPAMServiceServer = (*IPAMServer)(nil)

func NewIPAMServer(
	vrfs IPAMVRFService,
	prefixes IPAMPrefixService,
	addresses IPAMIPAddressService,
) *IPAMServer {
	if vrfs == nil || prefixes == nil || addresses == nil {
		panic("IPAM gRPC server requires typed VRF, Prefix, and IPAddress services")
	}
	return &IPAMServer{
		vrfs:      NewVRFRPCHandler(vrfs),
		prefixes:  NewPrefixRPCHandler(prefixes),
		addresses: NewIPAddressRPCHandler(addresses),
	}
}

func (server *IPAMServer) ListVRFs(
	ctx context.Context,
	request *ipamv1.ListVRFsRequest,
) (*ipamv1.ListVRFsResponse, error) {
	return server.vrfs.ListVRFs(ctx, request)
}

func (server *IPAMServer) GetVRF(
	ctx context.Context,
	request *ipamv1.GetVRFRequest,
) (*ipamv1.GetVRFResponse, error) {
	return server.vrfs.GetVRF(ctx, request)
}

func (server *IPAMServer) CreateVRF(
	ctx context.Context,
	request *ipamv1.CreateVRFRequest,
) (*ipamv1.CreateVRFResponse, error) {
	return server.vrfs.CreateVRF(ctx, request)
}

func (server *IPAMServer) ReplaceVRF(
	ctx context.Context,
	request *ipamv1.ReplaceVRFRequest,
) (*ipamv1.ReplaceVRFResponse, error) {
	return server.vrfs.ReplaceVRF(ctx, request)
}

func (server *IPAMServer) UpdateVRF(
	ctx context.Context,
	request *ipamv1.UpdateVRFRequest,
) (*ipamv1.UpdateVRFResponse, error) {
	return server.vrfs.UpdateVRF(ctx, request)
}

func (server *IPAMServer) DeleteVRF(
	ctx context.Context,
	request *ipamv1.DeleteVRFRequest,
) (*ipamv1.DeleteVRFResponse, error) {
	return server.vrfs.DeleteVRF(ctx, request)
}

func (server *IPAMServer) ListPrefixes(
	ctx context.Context,
	request *ipamv1.ListPrefixesRequest,
) (*ipamv1.ListPrefixesResponse, error) {
	return server.prefixes.ListPrefixes(ctx, request)
}

func (server *IPAMServer) GetPrefix(
	ctx context.Context,
	request *ipamv1.GetPrefixRequest,
) (*ipamv1.GetPrefixResponse, error) {
	return server.prefixes.GetPrefix(ctx, request)
}

func (server *IPAMServer) CreatePrefix(
	ctx context.Context,
	request *ipamv1.CreatePrefixRequest,
) (*ipamv1.CreatePrefixResponse, error) {
	return server.prefixes.CreatePrefix(ctx, request)
}

func (server *IPAMServer) ReplacePrefix(
	ctx context.Context,
	request *ipamv1.ReplacePrefixRequest,
) (*ipamv1.ReplacePrefixResponse, error) {
	return server.prefixes.ReplacePrefix(ctx, request)
}

func (server *IPAMServer) UpdatePrefix(
	ctx context.Context,
	request *ipamv1.UpdatePrefixRequest,
) (*ipamv1.UpdatePrefixResponse, error) {
	return server.prefixes.UpdatePrefix(ctx, request)
}

func (server *IPAMServer) DeletePrefix(
	ctx context.Context,
	request *ipamv1.DeletePrefixRequest,
) (*ipamv1.DeletePrefixResponse, error) {
	return server.prefixes.DeletePrefix(ctx, request)
}

func (server *IPAMServer) ListIPAddresses(
	ctx context.Context,
	request *ipamv1.ListIPAddressesRequest,
) (*ipamv1.ListIPAddressesResponse, error) {
	return server.addresses.ListIPAddresses(ctx, request)
}

func (server *IPAMServer) GetIPAddress(
	ctx context.Context,
	request *ipamv1.GetIPAddressRequest,
) (*ipamv1.GetIPAddressResponse, error) {
	return server.addresses.GetIPAddress(ctx, request)
}

func (server *IPAMServer) CreateIPAddress(
	ctx context.Context,
	request *ipamv1.CreateIPAddressRequest,
) (*ipamv1.CreateIPAddressResponse, error) {
	return server.addresses.CreateIPAddress(ctx, request)
}

func (server *IPAMServer) ReplaceIPAddress(
	ctx context.Context,
	request *ipamv1.ReplaceIPAddressRequest,
) (*ipamv1.ReplaceIPAddressResponse, error) {
	return server.addresses.ReplaceIPAddress(ctx, request)
}

func (server *IPAMServer) UpdateIPAddress(
	ctx context.Context,
	request *ipamv1.UpdateIPAddressRequest,
) (*ipamv1.UpdateIPAddressResponse, error) {
	return server.addresses.UpdateIPAddress(ctx, request)
}

func (server *IPAMServer) DeleteIPAddress(
	ctx context.Context,
	request *ipamv1.DeleteIPAddressRequest,
) (*ipamv1.DeleteIPAddressResponse, error) {
	return server.addresses.DeleteIPAddress(ctx, request)
}

func (server *IPAMServer) AssignIPAddress(
	ctx context.Context,
	request *ipamv1.AssignIPAddressRequest,
) (*ipamv1.AssignIPAddressResponse, error) {
	return server.addresses.AssignIPAddress(ctx, request)
}

func (server *IPAMServer) UnassignIPAddress(
	ctx context.Context,
	request *ipamv1.UnassignIPAddressRequest,
) (*ipamv1.UnassignIPAddressResponse, error) {
	return server.addresses.UnassignIPAddress(ctx, request)
}
