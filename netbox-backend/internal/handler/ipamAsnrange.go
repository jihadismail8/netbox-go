package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.IpamAsnrangeLogicer = (*ipamAsnrangeHandler)(nil)

type ipamAsnrangeHandler struct {
	server netbox_goV1.IpamAsnrangeServer
}

// NewIpamAsnrangeHandler create a handler
func NewIpamAsnrangeHandler() netbox_goV1.IpamAsnrangeLogicer {
	return &ipamAsnrangeHandler{
		server: service.NewIpamAsnrangeServer(),
	}
}

// Create a new ipamAsnrange
func (h *ipamAsnrangeHandler) Create(ctx context.Context, req *netbox_goV1.CreateIpamAsnrangeRequest) (*netbox_goV1.CreateIpamAsnrangeReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a ipamAsnrange by id
func (h *ipamAsnrangeHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamAsnrangeByIDRequest) (*netbox_goV1.DeleteIpamAsnrangeByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a ipamAsnrange by id
func (h *ipamAsnrangeHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamAsnrangeByIDRequest) (*netbox_goV1.UpdateIpamAsnrangeByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a ipamAsnrange by id
func (h *ipamAsnrangeHandler) GetByID(ctx context.Context, req *netbox_goV1.GetIpamAsnrangeByIDRequest) (*netbox_goV1.GetIpamAsnrangeByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of ipamAsnranges by custom conditions
func (h *ipamAsnrangeHandler) List(ctx context.Context, req *netbox_goV1.ListIpamAsnrangeRequest) (*netbox_goV1.ListIpamAsnrangeReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete ipamAsnrange by ids
func (h *ipamAsnrangeHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamAsnrangeByIDsRequest) (*netbox_goV1.DeleteIpamAsnrangeByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a ipamAsnrange by custom condition
func (h *ipamAsnrangeHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamAsnrangeByConditionRequest) (*netbox_goV1.GetIpamAsnrangeByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get ipamAsnrange by ids
func (h *ipamAsnrangeHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamAsnrangeByIDsRequest) (*netbox_goV1.ListIpamAsnrangeByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of ipamAsnranges by last id
func (h *ipamAsnrangeHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamAsnrangeByLastIDRequest) (*netbox_goV1.ListIpamAsnrangeByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
