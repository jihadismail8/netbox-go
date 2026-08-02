package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.IpamIprangeLogicer = (*ipamIprangeHandler)(nil)

type ipamIprangeHandler struct {
	server netbox_goV1.IpamIprangeServer
}

// NewIpamIprangeHandler create a handler
func NewIpamIprangeHandler() netbox_goV1.IpamIprangeLogicer {
	return &ipamIprangeHandler{
		server: service.NewIpamIprangeServer(),
	}
}

// Create a new ipamIprange
func (h *ipamIprangeHandler) Create(ctx context.Context, req *netbox_goV1.CreateIpamIprangeRequest) (*netbox_goV1.CreateIpamIprangeReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a ipamIprange by id
func (h *ipamIprangeHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamIprangeByIDRequest) (*netbox_goV1.DeleteIpamIprangeByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a ipamIprange by id
func (h *ipamIprangeHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamIprangeByIDRequest) (*netbox_goV1.UpdateIpamIprangeByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a ipamIprange by id
func (h *ipamIprangeHandler) GetByID(ctx context.Context, req *netbox_goV1.GetIpamIprangeByIDRequest) (*netbox_goV1.GetIpamIprangeByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of ipamIpranges by custom conditions
func (h *ipamIprangeHandler) List(ctx context.Context, req *netbox_goV1.ListIpamIprangeRequest) (*netbox_goV1.ListIpamIprangeReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete ipamIprange by ids
func (h *ipamIprangeHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamIprangeByIDsRequest) (*netbox_goV1.DeleteIpamIprangeByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a ipamIprange by custom condition
func (h *ipamIprangeHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamIprangeByConditionRequest) (*netbox_goV1.GetIpamIprangeByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get ipamIprange by ids
func (h *ipamIprangeHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamIprangeByIDsRequest) (*netbox_goV1.ListIpamIprangeByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of ipamIpranges by last id
func (h *ipamIprangeHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamIprangeByLastIDRequest) (*netbox_goV1.ListIpamIprangeByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
