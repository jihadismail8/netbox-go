package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.IpamVlanLogicer = (*ipamVlanHandler)(nil)

type ipamVlanHandler struct {
	server netbox_goV1.IpamVlanServer
}

// NewIpamVlanHandler create a handler
func NewIpamVlanHandler() netbox_goV1.IpamVlanLogicer {
	return &ipamVlanHandler{
		server: service.NewIpamVlanServer(),
	}
}

// Create a new ipamVlan
func (h *ipamVlanHandler) Create(ctx context.Context, req *netbox_goV1.CreateIpamVlanRequest) (*netbox_goV1.CreateIpamVlanReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a ipamVlan by id
func (h *ipamVlanHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamVlanByIDRequest) (*netbox_goV1.DeleteIpamVlanByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a ipamVlan by id
func (h *ipamVlanHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamVlanByIDRequest) (*netbox_goV1.UpdateIpamVlanByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a ipamVlan by id
func (h *ipamVlanHandler) GetByID(ctx context.Context, req *netbox_goV1.GetIpamVlanByIDRequest) (*netbox_goV1.GetIpamVlanByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of ipamVlans by custom conditions
func (h *ipamVlanHandler) List(ctx context.Context, req *netbox_goV1.ListIpamVlanRequest) (*netbox_goV1.ListIpamVlanReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete ipamVlan by ids
func (h *ipamVlanHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamVlanByIDsRequest) (*netbox_goV1.DeleteIpamVlanByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a ipamVlan by custom condition
func (h *ipamVlanHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamVlanByConditionRequest) (*netbox_goV1.GetIpamVlanByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get ipamVlan by ids
func (h *ipamVlanHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamVlanByIDsRequest) (*netbox_goV1.ListIpamVlanByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of ipamVlans by last id
func (h *ipamVlanHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamVlanByLastIDRequest) (*netbox_goV1.ListIpamVlanByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
