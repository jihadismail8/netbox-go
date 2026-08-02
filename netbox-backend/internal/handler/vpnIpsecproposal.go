package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VpnIpsecproposalLogicer = (*vpnIpsecproposalHandler)(nil)

type vpnIpsecproposalHandler struct {
	server netbox_goV1.VpnIpsecproposalServer
}

// NewVpnIpsecproposalHandler create a handler
func NewVpnIpsecproposalHandler() netbox_goV1.VpnIpsecproposalLogicer {
	return &vpnIpsecproposalHandler{
		server: service.NewVpnIpsecproposalServer(),
	}
}

// Create a new vpnIpsecproposal
func (h *vpnIpsecproposalHandler) Create(ctx context.Context, req *netbox_goV1.CreateVpnIpsecproposalRequest) (*netbox_goV1.CreateVpnIpsecproposalReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a vpnIpsecproposal by id
func (h *vpnIpsecproposalHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnIpsecproposalByIDRequest) (*netbox_goV1.DeleteVpnIpsecproposalByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a vpnIpsecproposal by id
func (h *vpnIpsecproposalHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnIpsecproposalByIDRequest) (*netbox_goV1.UpdateVpnIpsecproposalByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a vpnIpsecproposal by id
func (h *vpnIpsecproposalHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVpnIpsecproposalByIDRequest) (*netbox_goV1.GetVpnIpsecproposalByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of vpnIpsecproposals by custom conditions
func (h *vpnIpsecproposalHandler) List(ctx context.Context, req *netbox_goV1.ListVpnIpsecproposalRequest) (*netbox_goV1.ListVpnIpsecproposalReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete vpnIpsecproposal by ids
func (h *vpnIpsecproposalHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnIpsecproposalByIDsRequest) (*netbox_goV1.DeleteVpnIpsecproposalByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a vpnIpsecproposal by custom condition
func (h *vpnIpsecproposalHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnIpsecproposalByConditionRequest) (*netbox_goV1.GetVpnIpsecproposalByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get vpnIpsecproposal by ids
func (h *vpnIpsecproposalHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnIpsecproposalByIDsRequest) (*netbox_goV1.ListVpnIpsecproposalByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of vpnIpsecproposals by last id
func (h *vpnIpsecproposalHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnIpsecproposalByLastIDRequest) (*netbox_goV1.ListVpnIpsecproposalByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
