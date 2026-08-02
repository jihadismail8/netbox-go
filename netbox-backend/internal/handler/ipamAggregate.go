package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.IpamAggregateLogicer = (*ipamAggregateHandler)(nil)

type ipamAggregateHandler struct {
	server netbox_goV1.IpamAggregateServer
}

// NewIpamAggregateHandler create a handler
func NewIpamAggregateHandler() netbox_goV1.IpamAggregateLogicer {
	return &ipamAggregateHandler{
		server: service.NewIpamAggregateServer(),
	}
}

// Create a new ipamAggregate
func (h *ipamAggregateHandler) Create(ctx context.Context, req *netbox_goV1.CreateIpamAggregateRequest) (*netbox_goV1.CreateIpamAggregateReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a ipamAggregate by id
func (h *ipamAggregateHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamAggregateByIDRequest) (*netbox_goV1.DeleteIpamAggregateByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a ipamAggregate by id
func (h *ipamAggregateHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamAggregateByIDRequest) (*netbox_goV1.UpdateIpamAggregateByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a ipamAggregate by id
func (h *ipamAggregateHandler) GetByID(ctx context.Context, req *netbox_goV1.GetIpamAggregateByIDRequest) (*netbox_goV1.GetIpamAggregateByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of ipamAggregates by custom conditions
func (h *ipamAggregateHandler) List(ctx context.Context, req *netbox_goV1.ListIpamAggregateRequest) (*netbox_goV1.ListIpamAggregateReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete ipamAggregate by ids
func (h *ipamAggregateHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamAggregateByIDsRequest) (*netbox_goV1.DeleteIpamAggregateByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a ipamAggregate by custom condition
func (h *ipamAggregateHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamAggregateByConditionRequest) (*netbox_goV1.GetIpamAggregateByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get ipamAggregate by ids
func (h *ipamAggregateHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamAggregateByIDsRequest) (*netbox_goV1.ListIpamAggregateByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of ipamAggregates by last id
func (h *ipamAggregateHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamAggregateByLastIDRequest) (*netbox_goV1.ListIpamAggregateByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
