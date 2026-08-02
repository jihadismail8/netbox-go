package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CircuitsVirtualcircuittypeLogicer = (*circuitsVirtualcircuittypeHandler)(nil)

type circuitsVirtualcircuittypeHandler struct {
	server netbox_goV1.CircuitsVirtualcircuittypeServer
}

// NewCircuitsVirtualcircuittypeHandler create a handler
func NewCircuitsVirtualcircuittypeHandler() netbox_goV1.CircuitsVirtualcircuittypeLogicer {
	return &circuitsVirtualcircuittypeHandler{
		server: service.NewCircuitsVirtualcircuittypeServer(),
	}
}

// Create a new circuitsVirtualcircuittype
func (h *circuitsVirtualcircuittypeHandler) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsVirtualcircuittypeRequest) (*netbox_goV1.CreateCircuitsVirtualcircuittypeReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a circuitsVirtualcircuittype by id
func (h *circuitsVirtualcircuittypeHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsVirtualcircuittypeByIDRequest) (*netbox_goV1.DeleteCircuitsVirtualcircuittypeByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a circuitsVirtualcircuittype by id
func (h *circuitsVirtualcircuittypeHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsVirtualcircuittypeByIDRequest) (*netbox_goV1.UpdateCircuitsVirtualcircuittypeByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a circuitsVirtualcircuittype by id
func (h *circuitsVirtualcircuittypeHandler) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsVirtualcircuittypeByIDRequest) (*netbox_goV1.GetCircuitsVirtualcircuittypeByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of circuitsVirtualcircuittypes by custom conditions
func (h *circuitsVirtualcircuittypeHandler) List(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuittypeRequest) (*netbox_goV1.ListCircuitsVirtualcircuittypeReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete circuitsVirtualcircuittype by ids
func (h *circuitsVirtualcircuittypeHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsVirtualcircuittypeByIDsRequest) (*netbox_goV1.DeleteCircuitsVirtualcircuittypeByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a circuitsVirtualcircuittype by custom condition
func (h *circuitsVirtualcircuittypeHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsVirtualcircuittypeByConditionRequest) (*netbox_goV1.GetCircuitsVirtualcircuittypeByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get circuitsVirtualcircuittype by ids
func (h *circuitsVirtualcircuittypeHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuittypeByIDsRequest) (*netbox_goV1.ListCircuitsVirtualcircuittypeByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of circuitsVirtualcircuittypes by last id
func (h *circuitsVirtualcircuittypeHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuittypeByLastIDRequest) (*netbox_goV1.ListCircuitsVirtualcircuittypeByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
