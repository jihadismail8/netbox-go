package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VpnL2VpnterminationLogicer = (*vpnL2VpnterminationHandler)(nil)

type vpnL2VpnterminationHandler struct {
	server netbox_goV1.VpnL2VpnterminationServer
}

// NewVpnL2VpnterminationHandler create a handler
func NewVpnL2VpnterminationHandler() netbox_goV1.VpnL2VpnterminationLogicer {
	return &vpnL2VpnterminationHandler{
		server: service.NewVpnL2VpnterminationServer(),
	}
}

// Create a new vpnL2Vpntermination
func (h *vpnL2VpnterminationHandler) Create(ctx context.Context, req *netbox_goV1.CreateVpnL2VpnterminationRequest) (*netbox_goV1.CreateVpnL2VpnterminationReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a vpnL2Vpntermination by id
func (h *vpnL2VpnterminationHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnL2VpnterminationByIDRequest) (*netbox_goV1.DeleteVpnL2VpnterminationByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a vpnL2Vpntermination by id
func (h *vpnL2VpnterminationHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnL2VpnterminationByIDRequest) (*netbox_goV1.UpdateVpnL2VpnterminationByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a vpnL2Vpntermination by id
func (h *vpnL2VpnterminationHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVpnL2VpnterminationByIDRequest) (*netbox_goV1.GetVpnL2VpnterminationByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of vpnL2Vpnterminations by custom conditions
func (h *vpnL2VpnterminationHandler) List(ctx context.Context, req *netbox_goV1.ListVpnL2VpnterminationRequest) (*netbox_goV1.ListVpnL2VpnterminationReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete vpnL2Vpntermination by ids
func (h *vpnL2VpnterminationHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnL2VpnterminationByIDsRequest) (*netbox_goV1.DeleteVpnL2VpnterminationByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a vpnL2Vpntermination by custom condition
func (h *vpnL2VpnterminationHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnL2VpnterminationByConditionRequest) (*netbox_goV1.GetVpnL2VpnterminationByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get vpnL2Vpntermination by ids
func (h *vpnL2VpnterminationHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnL2VpnterminationByIDsRequest) (*netbox_goV1.ListVpnL2VpnterminationByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of vpnL2Vpnterminations by last id
func (h *vpnL2VpnterminationHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnL2VpnterminationByLastIDRequest) (*netbox_goV1.ListVpnL2VpnterminationByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
