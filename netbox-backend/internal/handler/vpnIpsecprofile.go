package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VpnIpsecprofileLogicer = (*vpnIpsecprofileHandler)(nil)

type vpnIpsecprofileHandler struct {
	server netbox_goV1.VpnIpsecprofileServer
}

// NewVpnIpsecprofileHandler create a handler
func NewVpnIpsecprofileHandler() netbox_goV1.VpnIpsecprofileLogicer {
	return &vpnIpsecprofileHandler{
		server: service.NewVpnIpsecprofileServer(),
	}
}

// Create a new vpnIpsecprofile
func (h *vpnIpsecprofileHandler) Create(ctx context.Context, req *netbox_goV1.CreateVpnIpsecprofileRequest) (*netbox_goV1.CreateVpnIpsecprofileReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a vpnIpsecprofile by id
func (h *vpnIpsecprofileHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnIpsecprofileByIDRequest) (*netbox_goV1.DeleteVpnIpsecprofileByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a vpnIpsecprofile by id
func (h *vpnIpsecprofileHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnIpsecprofileByIDRequest) (*netbox_goV1.UpdateVpnIpsecprofileByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a vpnIpsecprofile by id
func (h *vpnIpsecprofileHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVpnIpsecprofileByIDRequest) (*netbox_goV1.GetVpnIpsecprofileByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of vpnIpsecprofiles by custom conditions
func (h *vpnIpsecprofileHandler) List(ctx context.Context, req *netbox_goV1.ListVpnIpsecprofileRequest) (*netbox_goV1.ListVpnIpsecprofileReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete vpnIpsecprofile by ids
func (h *vpnIpsecprofileHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnIpsecprofileByIDsRequest) (*netbox_goV1.DeleteVpnIpsecprofileByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a vpnIpsecprofile by custom condition
func (h *vpnIpsecprofileHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnIpsecprofileByConditionRequest) (*netbox_goV1.GetVpnIpsecprofileByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get vpnIpsecprofile by ids
func (h *vpnIpsecprofileHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnIpsecprofileByIDsRequest) (*netbox_goV1.ListVpnIpsecprofileByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of vpnIpsecprofiles by last id
func (h *vpnIpsecprofileHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnIpsecprofileByLastIDRequest) (*netbox_goV1.ListVpnIpsecprofileByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
