package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CircuitsCircuittypeLogicer = (*circuitsCircuittypeHandler)(nil)

type circuitsCircuittypeHandler struct {
	server netbox_goV1.CircuitsCircuittypeServer
}

// NewCircuitsCircuittypeHandler create a handler
func NewCircuitsCircuittypeHandler() netbox_goV1.CircuitsCircuittypeLogicer {
	return &circuitsCircuittypeHandler{
		server: service.NewCircuitsCircuittypeServer(),
	}
}

// Create a new circuitsCircuittype
func (h *circuitsCircuittypeHandler) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsCircuittypeRequest) (*netbox_goV1.CreateCircuitsCircuittypeReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a circuitsCircuittype by id
func (h *circuitsCircuittypeHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuittypeByIDRequest) (*netbox_goV1.DeleteCircuitsCircuittypeByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a circuitsCircuittype by id
func (h *circuitsCircuittypeHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsCircuittypeByIDRequest) (*netbox_goV1.UpdateCircuitsCircuittypeByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a circuitsCircuittype by id
func (h *circuitsCircuittypeHandler) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsCircuittypeByIDRequest) (*netbox_goV1.GetCircuitsCircuittypeByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of circuitsCircuittypes by custom conditions
func (h *circuitsCircuittypeHandler) List(ctx context.Context, req *netbox_goV1.ListCircuitsCircuittypeRequest) (*netbox_goV1.ListCircuitsCircuittypeReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete circuitsCircuittype by ids
func (h *circuitsCircuittypeHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuittypeByIDsRequest) (*netbox_goV1.DeleteCircuitsCircuittypeByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a circuitsCircuittype by custom condition
func (h *circuitsCircuittypeHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsCircuittypeByConditionRequest) (*netbox_goV1.GetCircuitsCircuittypeByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get circuitsCircuittype by ids
func (h *circuitsCircuittypeHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsCircuittypeByIDsRequest) (*netbox_goV1.ListCircuitsCircuittypeByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of circuitsCircuittypes by last id
func (h *circuitsCircuittypeHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsCircuittypeByLastIDRequest) (*netbox_goV1.ListCircuitsCircuittypeByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
