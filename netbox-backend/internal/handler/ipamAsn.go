package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.IpamAsnLogicer = (*ipamAsnHandler)(nil)

type ipamAsnHandler struct {
	server netbox_goV1.IpamAsnServer
}

// NewIpamAsnHandler create a handler
func NewIpamAsnHandler() netbox_goV1.IpamAsnLogicer {
	return &ipamAsnHandler{
		server: service.NewIpamAsnServer(),
	}
}

// Create a new ipamAsn
func (h *ipamAsnHandler) Create(ctx context.Context, req *netbox_goV1.CreateIpamAsnRequest) (*netbox_goV1.CreateIpamAsnReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a ipamAsn by id
func (h *ipamAsnHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamAsnByIDRequest) (*netbox_goV1.DeleteIpamAsnByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a ipamAsn by id
func (h *ipamAsnHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamAsnByIDRequest) (*netbox_goV1.UpdateIpamAsnByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a ipamAsn by id
func (h *ipamAsnHandler) GetByID(ctx context.Context, req *netbox_goV1.GetIpamAsnByIDRequest) (*netbox_goV1.GetIpamAsnByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of ipamAsns by custom conditions
func (h *ipamAsnHandler) List(ctx context.Context, req *netbox_goV1.ListIpamAsnRequest) (*netbox_goV1.ListIpamAsnReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete ipamAsn by ids
func (h *ipamAsnHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamAsnByIDsRequest) (*netbox_goV1.DeleteIpamAsnByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a ipamAsn by custom condition
func (h *ipamAsnHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamAsnByConditionRequest) (*netbox_goV1.GetIpamAsnByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get ipamAsn by ids
func (h *ipamAsnHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamAsnByIDsRequest) (*netbox_goV1.ListIpamAsnByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of ipamAsns by last id
func (h *ipamAsnHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamAsnByLastIDRequest) (*netbox_goV1.ListIpamAsnByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
