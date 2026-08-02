package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CircuitsVirtualcircuitterminationLogicer = (*circuitsVirtualcircuitterminationHandler)(nil)

type circuitsVirtualcircuitterminationHandler struct {
	server netbox_goV1.CircuitsVirtualcircuitterminationServer
}

// NewCircuitsVirtualcircuitterminationHandler create a handler
func NewCircuitsVirtualcircuitterminationHandler() netbox_goV1.CircuitsVirtualcircuitterminationLogicer {
	return &circuitsVirtualcircuitterminationHandler{
		server: service.NewCircuitsVirtualcircuitterminationServer(),
	}
}

// Create a new circuitsVirtualcircuittermination
func (h *circuitsVirtualcircuitterminationHandler) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsVirtualcircuitterminationRequest) (*netbox_goV1.CreateCircuitsVirtualcircuitterminationReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a circuitsVirtualcircuittermination by id
func (h *circuitsVirtualcircuitterminationHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsVirtualcircuitterminationByIDRequest) (*netbox_goV1.DeleteCircuitsVirtualcircuitterminationByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a circuitsVirtualcircuittermination by id
func (h *circuitsVirtualcircuitterminationHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsVirtualcircuitterminationByIDRequest) (*netbox_goV1.UpdateCircuitsVirtualcircuitterminationByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a circuitsVirtualcircuittermination by id
func (h *circuitsVirtualcircuitterminationHandler) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsVirtualcircuitterminationByIDRequest) (*netbox_goV1.GetCircuitsVirtualcircuitterminationByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of circuitsVirtualcircuitterminations by custom conditions
func (h *circuitsVirtualcircuitterminationHandler) List(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuitterminationRequest) (*netbox_goV1.ListCircuitsVirtualcircuitterminationReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete circuitsVirtualcircuittermination by ids
func (h *circuitsVirtualcircuitterminationHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsVirtualcircuitterminationByIDsRequest) (*netbox_goV1.DeleteCircuitsVirtualcircuitterminationByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a circuitsVirtualcircuittermination by custom condition
func (h *circuitsVirtualcircuitterminationHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsVirtualcircuitterminationByConditionRequest) (*netbox_goV1.GetCircuitsVirtualcircuitterminationByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get circuitsVirtualcircuittermination by ids
func (h *circuitsVirtualcircuitterminationHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuitterminationByIDsRequest) (*netbox_goV1.ListCircuitsVirtualcircuitterminationByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of circuitsVirtualcircuitterminations by last id
func (h *circuitsVirtualcircuitterminationHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuitterminationByLastIDRequest) (*netbox_goV1.ListCircuitsVirtualcircuitterminationByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
