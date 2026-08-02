package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VpnTunnelterminationLogicer = (*vpnTunnelterminationHandler)(nil)

type vpnTunnelterminationHandler struct {
	server netbox_goV1.VpnTunnelterminationServer
}

// NewVpnTunnelterminationHandler create a handler
func NewVpnTunnelterminationHandler() netbox_goV1.VpnTunnelterminationLogicer {
	return &vpnTunnelterminationHandler{
		server: service.NewVpnTunnelterminationServer(),
	}
}

// Create a new vpnTunneltermination
func (h *vpnTunnelterminationHandler) Create(ctx context.Context, req *netbox_goV1.CreateVpnTunnelterminationRequest) (*netbox_goV1.CreateVpnTunnelterminationReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a vpnTunneltermination by id
func (h *vpnTunnelterminationHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnTunnelterminationByIDRequest) (*netbox_goV1.DeleteVpnTunnelterminationByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a vpnTunneltermination by id
func (h *vpnTunnelterminationHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnTunnelterminationByIDRequest) (*netbox_goV1.UpdateVpnTunnelterminationByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a vpnTunneltermination by id
func (h *vpnTunnelterminationHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVpnTunnelterminationByIDRequest) (*netbox_goV1.GetVpnTunnelterminationByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of vpnTunnelterminations by custom conditions
func (h *vpnTunnelterminationHandler) List(ctx context.Context, req *netbox_goV1.ListVpnTunnelterminationRequest) (*netbox_goV1.ListVpnTunnelterminationReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete vpnTunneltermination by ids
func (h *vpnTunnelterminationHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnTunnelterminationByIDsRequest) (*netbox_goV1.DeleteVpnTunnelterminationByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a vpnTunneltermination by custom condition
func (h *vpnTunnelterminationHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnTunnelterminationByConditionRequest) (*netbox_goV1.GetVpnTunnelterminationByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get vpnTunneltermination by ids
func (h *vpnTunnelterminationHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnTunnelterminationByIDsRequest) (*netbox_goV1.ListVpnTunnelterminationByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of vpnTunnelterminations by last id
func (h *vpnTunnelterminationHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnTunnelterminationByLastIDRequest) (*netbox_goV1.ListVpnTunnelterminationByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
