package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.IpamVlantranslationpolicyLogicer = (*ipamVlantranslationpolicyHandler)(nil)

type ipamVlantranslationpolicyHandler struct {
	server netbox_goV1.IpamVlantranslationpolicyServer
}

// NewIpamVlantranslationpolicyHandler create a handler
func NewIpamVlantranslationpolicyHandler() netbox_goV1.IpamVlantranslationpolicyLogicer {
	return &ipamVlantranslationpolicyHandler{
		server: service.NewIpamVlantranslationpolicyServer(),
	}
}

// Create a new ipamVlantranslationpolicy
func (h *ipamVlantranslationpolicyHandler) Create(ctx context.Context, req *netbox_goV1.CreateIpamVlantranslationpolicyRequest) (*netbox_goV1.CreateIpamVlantranslationpolicyReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a ipamVlantranslationpolicy by id
func (h *ipamVlantranslationpolicyHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamVlantranslationpolicyByIDRequest) (*netbox_goV1.DeleteIpamVlantranslationpolicyByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a ipamVlantranslationpolicy by id
func (h *ipamVlantranslationpolicyHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamVlantranslationpolicyByIDRequest) (*netbox_goV1.UpdateIpamVlantranslationpolicyByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a ipamVlantranslationpolicy by id
func (h *ipamVlantranslationpolicyHandler) GetByID(ctx context.Context, req *netbox_goV1.GetIpamVlantranslationpolicyByIDRequest) (*netbox_goV1.GetIpamVlantranslationpolicyByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of ipamVlantranslationpolicys by custom conditions
func (h *ipamVlantranslationpolicyHandler) List(ctx context.Context, req *netbox_goV1.ListIpamVlantranslationpolicyRequest) (*netbox_goV1.ListIpamVlantranslationpolicyReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete ipamVlantranslationpolicy by ids
func (h *ipamVlantranslationpolicyHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamVlantranslationpolicyByIDsRequest) (*netbox_goV1.DeleteIpamVlantranslationpolicyByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a ipamVlantranslationpolicy by custom condition
func (h *ipamVlantranslationpolicyHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamVlantranslationpolicyByConditionRequest) (*netbox_goV1.GetIpamVlantranslationpolicyByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get ipamVlantranslationpolicy by ids
func (h *ipamVlantranslationpolicyHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamVlantranslationpolicyByIDsRequest) (*netbox_goV1.ListIpamVlantranslationpolicyByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of ipamVlantranslationpolicys by last id
func (h *ipamVlantranslationpolicyHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamVlantranslationpolicyByLastIDRequest) (*netbox_goV1.ListIpamVlantranslationpolicyByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
