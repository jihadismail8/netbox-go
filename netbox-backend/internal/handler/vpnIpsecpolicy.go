package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VpnIpsecpolicyLogicer = (*vpnIpsecpolicyHandler)(nil)

type vpnIpsecpolicyHandler struct {
	server netbox_goV1.VpnIpsecpolicyServer
}

// NewVpnIpsecpolicyHandler create a handler
func NewVpnIpsecpolicyHandler() netbox_goV1.VpnIpsecpolicyLogicer {
	return &vpnIpsecpolicyHandler{
		server: service.NewVpnIpsecpolicyServer(),
	}
}

// Create a new vpnIpsecpolicy
func (h *vpnIpsecpolicyHandler) Create(ctx context.Context, req *netbox_goV1.CreateVpnIpsecpolicyRequest) (*netbox_goV1.CreateVpnIpsecpolicyReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a vpnIpsecpolicy by id
func (h *vpnIpsecpolicyHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnIpsecpolicyByIDRequest) (*netbox_goV1.DeleteVpnIpsecpolicyByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a vpnIpsecpolicy by id
func (h *vpnIpsecpolicyHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnIpsecpolicyByIDRequest) (*netbox_goV1.UpdateVpnIpsecpolicyByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a vpnIpsecpolicy by id
func (h *vpnIpsecpolicyHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVpnIpsecpolicyByIDRequest) (*netbox_goV1.GetVpnIpsecpolicyByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of vpnIpsecpolicys by custom conditions
func (h *vpnIpsecpolicyHandler) List(ctx context.Context, req *netbox_goV1.ListVpnIpsecpolicyRequest) (*netbox_goV1.ListVpnIpsecpolicyReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete vpnIpsecpolicy by ids
func (h *vpnIpsecpolicyHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnIpsecpolicyByIDsRequest) (*netbox_goV1.DeleteVpnIpsecpolicyByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a vpnIpsecpolicy by custom condition
func (h *vpnIpsecpolicyHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnIpsecpolicyByConditionRequest) (*netbox_goV1.GetVpnIpsecpolicyByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get vpnIpsecpolicy by ids
func (h *vpnIpsecpolicyHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnIpsecpolicyByIDsRequest) (*netbox_goV1.ListVpnIpsecpolicyByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of vpnIpsecpolicys by last id
func (h *vpnIpsecpolicyHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnIpsecpolicyByLastIDRequest) (*netbox_goV1.ListVpnIpsecpolicyByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
