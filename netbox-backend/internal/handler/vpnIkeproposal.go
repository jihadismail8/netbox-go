package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VpnIkeproposalLogicer = (*vpnIkeproposalHandler)(nil)

type vpnIkeproposalHandler struct {
	server netbox_goV1.VpnIkeproposalServer
}

// NewVpnIkeproposalHandler create a handler
func NewVpnIkeproposalHandler() netbox_goV1.VpnIkeproposalLogicer {
	return &vpnIkeproposalHandler{
		server: service.NewVpnIkeproposalServer(),
	}
}

// Create a new vpnIkeproposal
func (h *vpnIkeproposalHandler) Create(ctx context.Context, req *netbox_goV1.CreateVpnIkeproposalRequest) (*netbox_goV1.CreateVpnIkeproposalReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a vpnIkeproposal by id
func (h *vpnIkeproposalHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnIkeproposalByIDRequest) (*netbox_goV1.DeleteVpnIkeproposalByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a vpnIkeproposal by id
func (h *vpnIkeproposalHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnIkeproposalByIDRequest) (*netbox_goV1.UpdateVpnIkeproposalByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a vpnIkeproposal by id
func (h *vpnIkeproposalHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVpnIkeproposalByIDRequest) (*netbox_goV1.GetVpnIkeproposalByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of vpnIkeproposals by custom conditions
func (h *vpnIkeproposalHandler) List(ctx context.Context, req *netbox_goV1.ListVpnIkeproposalRequest) (*netbox_goV1.ListVpnIkeproposalReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete vpnIkeproposal by ids
func (h *vpnIkeproposalHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnIkeproposalByIDsRequest) (*netbox_goV1.DeleteVpnIkeproposalByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a vpnIkeproposal by custom condition
func (h *vpnIkeproposalHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnIkeproposalByConditionRequest) (*netbox_goV1.GetVpnIkeproposalByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get vpnIkeproposal by ids
func (h *vpnIkeproposalHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnIkeproposalByIDsRequest) (*netbox_goV1.ListVpnIkeproposalByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of vpnIkeproposals by last id
func (h *vpnIkeproposalHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnIkeproposalByLastIDRequest) (*netbox_goV1.ListVpnIkeproposalByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
