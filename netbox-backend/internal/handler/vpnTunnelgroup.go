package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VpnTunnelgroupLogicer = (*vpnTunnelgroupHandler)(nil)

type vpnTunnelgroupHandler struct {
	server netbox_goV1.VpnTunnelgroupServer
}

// NewVpnTunnelgroupHandler create a handler
func NewVpnTunnelgroupHandler() netbox_goV1.VpnTunnelgroupLogicer {
	return &vpnTunnelgroupHandler{
		server: service.NewVpnTunnelgroupServer(),
	}
}

// Create a new vpnTunnelgroup
func (h *vpnTunnelgroupHandler) Create(ctx context.Context, req *netbox_goV1.CreateVpnTunnelgroupRequest) (*netbox_goV1.CreateVpnTunnelgroupReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a vpnTunnelgroup by id
func (h *vpnTunnelgroupHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnTunnelgroupByIDRequest) (*netbox_goV1.DeleteVpnTunnelgroupByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a vpnTunnelgroup by id
func (h *vpnTunnelgroupHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnTunnelgroupByIDRequest) (*netbox_goV1.UpdateVpnTunnelgroupByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a vpnTunnelgroup by id
func (h *vpnTunnelgroupHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVpnTunnelgroupByIDRequest) (*netbox_goV1.GetVpnTunnelgroupByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of vpnTunnelgroups by custom conditions
func (h *vpnTunnelgroupHandler) List(ctx context.Context, req *netbox_goV1.ListVpnTunnelgroupRequest) (*netbox_goV1.ListVpnTunnelgroupReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete vpnTunnelgroup by ids
func (h *vpnTunnelgroupHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnTunnelgroupByIDsRequest) (*netbox_goV1.DeleteVpnTunnelgroupByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a vpnTunnelgroup by custom condition
func (h *vpnTunnelgroupHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnTunnelgroupByConditionRequest) (*netbox_goV1.GetVpnTunnelgroupByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get vpnTunnelgroup by ids
func (h *vpnTunnelgroupHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnTunnelgroupByIDsRequest) (*netbox_goV1.ListVpnTunnelgroupByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of vpnTunnelgroups by last id
func (h *vpnTunnelgroupHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnTunnelgroupByLastIDRequest) (*netbox_goV1.ListVpnTunnelgroupByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
