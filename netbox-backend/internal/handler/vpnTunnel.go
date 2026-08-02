package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VpnTunnelLogicer = (*vpnTunnelHandler)(nil)

type vpnTunnelHandler struct {
	server netbox_goV1.VpnTunnelServer
}

// NewVpnTunnelHandler create a handler
func NewVpnTunnelHandler() netbox_goV1.VpnTunnelLogicer {
	return &vpnTunnelHandler{
		server: service.NewVpnTunnelServer(),
	}
}

// Create a new vpnTunnel
func (h *vpnTunnelHandler) Create(ctx context.Context, req *netbox_goV1.CreateVpnTunnelRequest) (*netbox_goV1.CreateVpnTunnelReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a vpnTunnel by id
func (h *vpnTunnelHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnTunnelByIDRequest) (*netbox_goV1.DeleteVpnTunnelByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a vpnTunnel by id
func (h *vpnTunnelHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnTunnelByIDRequest) (*netbox_goV1.UpdateVpnTunnelByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a vpnTunnel by id
func (h *vpnTunnelHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVpnTunnelByIDRequest) (*netbox_goV1.GetVpnTunnelByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of vpnTunnels by custom conditions
func (h *vpnTunnelHandler) List(ctx context.Context, req *netbox_goV1.ListVpnTunnelRequest) (*netbox_goV1.ListVpnTunnelReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete vpnTunnel by ids
func (h *vpnTunnelHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnTunnelByIDsRequest) (*netbox_goV1.DeleteVpnTunnelByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a vpnTunnel by custom condition
func (h *vpnTunnelHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnTunnelByConditionRequest) (*netbox_goV1.GetVpnTunnelByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get vpnTunnel by ids
func (h *vpnTunnelHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnTunnelByIDsRequest) (*netbox_goV1.ListVpnTunnelByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of vpnTunnels by last id
func (h *vpnTunnelHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnTunnelByLastIDRequest) (*netbox_goV1.ListVpnTunnelByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
