package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VpnIkepolicyLogicer = (*vpnIkepolicyHandler)(nil)

type vpnIkepolicyHandler struct {
	server netbox_goV1.VpnIkepolicyServer
}

// NewVpnIkepolicyHandler create a handler
func NewVpnIkepolicyHandler() netbox_goV1.VpnIkepolicyLogicer {
	return &vpnIkepolicyHandler{
		server: service.NewVpnIkepolicyServer(),
	}
}

// Create a new vpnIkepolicy
func (h *vpnIkepolicyHandler) Create(ctx context.Context, req *netbox_goV1.CreateVpnIkepolicyRequest) (*netbox_goV1.CreateVpnIkepolicyReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a vpnIkepolicy by id
func (h *vpnIkepolicyHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnIkepolicyByIDRequest) (*netbox_goV1.DeleteVpnIkepolicyByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a vpnIkepolicy by id
func (h *vpnIkepolicyHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnIkepolicyByIDRequest) (*netbox_goV1.UpdateVpnIkepolicyByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a vpnIkepolicy by id
func (h *vpnIkepolicyHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVpnIkepolicyByIDRequest) (*netbox_goV1.GetVpnIkepolicyByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of vpnIkepolicys by custom conditions
func (h *vpnIkepolicyHandler) List(ctx context.Context, req *netbox_goV1.ListVpnIkepolicyRequest) (*netbox_goV1.ListVpnIkepolicyReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete vpnIkepolicy by ids
func (h *vpnIkepolicyHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnIkepolicyByIDsRequest) (*netbox_goV1.DeleteVpnIkepolicyByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a vpnIkepolicy by custom condition
func (h *vpnIkepolicyHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnIkepolicyByConditionRequest) (*netbox_goV1.GetVpnIkepolicyByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get vpnIkepolicy by ids
func (h *vpnIkepolicyHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnIkepolicyByIDsRequest) (*netbox_goV1.ListVpnIkepolicyByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of vpnIkepolicys by last id
func (h *vpnIkepolicyHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnIkepolicyByLastIDRequest) (*netbox_goV1.ListVpnIkepolicyByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
