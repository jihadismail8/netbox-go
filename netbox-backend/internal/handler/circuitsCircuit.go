package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CircuitsCircuitLogicer = (*circuitsCircuitHandler)(nil)

type circuitsCircuitHandler struct {
	server netbox_goV1.CircuitsCircuitServer
}

// NewCircuitsCircuitHandler create a handler
func NewCircuitsCircuitHandler() netbox_goV1.CircuitsCircuitLogicer {
	return &circuitsCircuitHandler{
		server: service.NewCircuitsCircuitServer(),
	}
}

// Create a new circuitsCircuit
func (h *circuitsCircuitHandler) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsCircuitRequest) (*netbox_goV1.CreateCircuitsCircuitReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a circuitsCircuit by id
func (h *circuitsCircuitHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitByIDRequest) (*netbox_goV1.DeleteCircuitsCircuitByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a circuitsCircuit by id
func (h *circuitsCircuitHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsCircuitByIDRequest) (*netbox_goV1.UpdateCircuitsCircuitByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a circuitsCircuit by id
func (h *circuitsCircuitHandler) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitByIDRequest) (*netbox_goV1.GetCircuitsCircuitByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of circuitsCircuits by custom conditions
func (h *circuitsCircuitHandler) List(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitRequest) (*netbox_goV1.ListCircuitsCircuitReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete circuitsCircuit by ids
func (h *circuitsCircuitHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitByIDsRequest) (*netbox_goV1.DeleteCircuitsCircuitByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a circuitsCircuit by custom condition
func (h *circuitsCircuitHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitByConditionRequest) (*netbox_goV1.GetCircuitsCircuitByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get circuitsCircuit by ids
func (h *circuitsCircuitHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitByIDsRequest) (*netbox_goV1.ListCircuitsCircuitByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of circuitsCircuits by last id
func (h *circuitsCircuitHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitByLastIDRequest) (*netbox_goV1.ListCircuitsCircuitByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
