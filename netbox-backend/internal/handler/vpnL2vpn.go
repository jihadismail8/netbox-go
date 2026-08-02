package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VpnL2VpnLogicer = (*vpnL2VpnHandler)(nil)

type vpnL2VpnHandler struct {
	server netbox_goV1.VpnL2VpnServer
}

// NewVpnL2VpnHandler create a handler
func NewVpnL2VpnHandler() netbox_goV1.VpnL2VpnLogicer {
	return &vpnL2VpnHandler{
		server: service.NewVpnL2VpnServer(),
	}
}

// Create a new vpnL2Vpn
func (h *vpnL2VpnHandler) Create(ctx context.Context, req *netbox_goV1.CreateVpnL2VpnRequest) (*netbox_goV1.CreateVpnL2VpnReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a vpnL2Vpn by id
func (h *vpnL2VpnHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnL2VpnByIDRequest) (*netbox_goV1.DeleteVpnL2VpnByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a vpnL2Vpn by id
func (h *vpnL2VpnHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnL2VpnByIDRequest) (*netbox_goV1.UpdateVpnL2VpnByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a vpnL2Vpn by id
func (h *vpnL2VpnHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVpnL2VpnByIDRequest) (*netbox_goV1.GetVpnL2VpnByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of vpnL2Vpns by custom conditions
func (h *vpnL2VpnHandler) List(ctx context.Context, req *netbox_goV1.ListVpnL2VpnRequest) (*netbox_goV1.ListVpnL2VpnReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete vpnL2Vpn by ids
func (h *vpnL2VpnHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnL2VpnByIDsRequest) (*netbox_goV1.DeleteVpnL2VpnByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a vpnL2Vpn by custom condition
func (h *vpnL2VpnHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnL2VpnByConditionRequest) (*netbox_goV1.GetVpnL2VpnByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get vpnL2Vpn by ids
func (h *vpnL2VpnHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnL2VpnByIDsRequest) (*netbox_goV1.ListVpnL2VpnByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of vpnL2Vpns by last id
func (h *vpnL2VpnHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnL2VpnByLastIDRequest) (*netbox_goV1.ListVpnL2VpnByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
