package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CircuitsVirtualcircuitLogicer = (*circuitsVirtualcircuitHandler)(nil)

type circuitsVirtualcircuitHandler struct {
	server netbox_goV1.CircuitsVirtualcircuitServer
}

// NewCircuitsVirtualcircuitHandler create a handler
func NewCircuitsVirtualcircuitHandler() netbox_goV1.CircuitsVirtualcircuitLogicer {
	return &circuitsVirtualcircuitHandler{
		server: service.NewCircuitsVirtualcircuitServer(),
	}
}

// Create a new circuitsVirtualcircuit
func (h *circuitsVirtualcircuitHandler) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsVirtualcircuitRequest) (*netbox_goV1.CreateCircuitsVirtualcircuitReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a circuitsVirtualcircuit by id
func (h *circuitsVirtualcircuitHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsVirtualcircuitByIDRequest) (*netbox_goV1.DeleteCircuitsVirtualcircuitByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a circuitsVirtualcircuit by id
func (h *circuitsVirtualcircuitHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsVirtualcircuitByIDRequest) (*netbox_goV1.UpdateCircuitsVirtualcircuitByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a circuitsVirtualcircuit by id
func (h *circuitsVirtualcircuitHandler) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsVirtualcircuitByIDRequest) (*netbox_goV1.GetCircuitsVirtualcircuitByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of circuitsVirtualcircuits by custom conditions
func (h *circuitsVirtualcircuitHandler) List(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuitRequest) (*netbox_goV1.ListCircuitsVirtualcircuitReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete circuitsVirtualcircuit by ids
func (h *circuitsVirtualcircuitHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsVirtualcircuitByIDsRequest) (*netbox_goV1.DeleteCircuitsVirtualcircuitByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a circuitsVirtualcircuit by custom condition
func (h *circuitsVirtualcircuitHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsVirtualcircuitByConditionRequest) (*netbox_goV1.GetCircuitsVirtualcircuitByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get circuitsVirtualcircuit by ids
func (h *circuitsVirtualcircuitHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuitByIDsRequest) (*netbox_goV1.ListCircuitsVirtualcircuitByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of circuitsVirtualcircuits by last id
func (h *circuitsVirtualcircuitHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuitByLastIDRequest) (*netbox_goV1.ListCircuitsVirtualcircuitByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
