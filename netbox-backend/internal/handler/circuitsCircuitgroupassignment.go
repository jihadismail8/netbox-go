package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CircuitsCircuitgroupassignmentLogicer = (*circuitsCircuitgroupassignmentHandler)(nil)

type circuitsCircuitgroupassignmentHandler struct {
	server netbox_goV1.CircuitsCircuitgroupassignmentServer
}

// NewCircuitsCircuitgroupassignmentHandler create a handler
func NewCircuitsCircuitgroupassignmentHandler() netbox_goV1.CircuitsCircuitgroupassignmentLogicer {
	return &circuitsCircuitgroupassignmentHandler{
		server: service.NewCircuitsCircuitgroupassignmentServer(),
	}
}

// Create a new circuitsCircuitgroupassignment
func (h *circuitsCircuitgroupassignmentHandler) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsCircuitgroupassignmentRequest) (*netbox_goV1.CreateCircuitsCircuitgroupassignmentReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a circuitsCircuitgroupassignment by id
func (h *circuitsCircuitgroupassignmentHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitgroupassignmentByIDRequest) (*netbox_goV1.DeleteCircuitsCircuitgroupassignmentByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a circuitsCircuitgroupassignment by id
func (h *circuitsCircuitgroupassignmentHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsCircuitgroupassignmentByIDRequest) (*netbox_goV1.UpdateCircuitsCircuitgroupassignmentByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a circuitsCircuitgroupassignment by id
func (h *circuitsCircuitgroupassignmentHandler) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitgroupassignmentByIDRequest) (*netbox_goV1.GetCircuitsCircuitgroupassignmentByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of circuitsCircuitgroupassignments by custom conditions
func (h *circuitsCircuitgroupassignmentHandler) List(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitgroupassignmentRequest) (*netbox_goV1.ListCircuitsCircuitgroupassignmentReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete circuitsCircuitgroupassignment by ids
func (h *circuitsCircuitgroupassignmentHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitgroupassignmentByIDsRequest) (*netbox_goV1.DeleteCircuitsCircuitgroupassignmentByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a circuitsCircuitgroupassignment by custom condition
func (h *circuitsCircuitgroupassignmentHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitgroupassignmentByConditionRequest) (*netbox_goV1.GetCircuitsCircuitgroupassignmentByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get circuitsCircuitgroupassignment by ids
func (h *circuitsCircuitgroupassignmentHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitgroupassignmentByIDsRequest) (*netbox_goV1.ListCircuitsCircuitgroupassignmentByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of circuitsCircuitgroupassignments by last id
func (h *circuitsCircuitgroupassignmentHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitgroupassignmentByLastIDRequest) (*netbox_goV1.ListCircuitsCircuitgroupassignmentByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
