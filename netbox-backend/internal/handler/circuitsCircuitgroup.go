package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CircuitsCircuitgroupLogicer = (*circuitsCircuitgroupHandler)(nil)

type circuitsCircuitgroupHandler struct {
	server netbox_goV1.CircuitsCircuitgroupServer
}

// NewCircuitsCircuitgroupHandler create a handler
func NewCircuitsCircuitgroupHandler() netbox_goV1.CircuitsCircuitgroupLogicer {
	return &circuitsCircuitgroupHandler{
		server: service.NewCircuitsCircuitgroupServer(),
	}
}

// Create a new circuitsCircuitgroup
func (h *circuitsCircuitgroupHandler) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsCircuitgroupRequest) (*netbox_goV1.CreateCircuitsCircuitgroupReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a circuitsCircuitgroup by id
func (h *circuitsCircuitgroupHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitgroupByIDRequest) (*netbox_goV1.DeleteCircuitsCircuitgroupByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a circuitsCircuitgroup by id
func (h *circuitsCircuitgroupHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsCircuitgroupByIDRequest) (*netbox_goV1.UpdateCircuitsCircuitgroupByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a circuitsCircuitgroup by id
func (h *circuitsCircuitgroupHandler) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitgroupByIDRequest) (*netbox_goV1.GetCircuitsCircuitgroupByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of circuitsCircuitgroups by custom conditions
func (h *circuitsCircuitgroupHandler) List(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitgroupRequest) (*netbox_goV1.ListCircuitsCircuitgroupReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete circuitsCircuitgroup by ids
func (h *circuitsCircuitgroupHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitgroupByIDsRequest) (*netbox_goV1.DeleteCircuitsCircuitgroupByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a circuitsCircuitgroup by custom condition
func (h *circuitsCircuitgroupHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitgroupByConditionRequest) (*netbox_goV1.GetCircuitsCircuitgroupByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get circuitsCircuitgroup by ids
func (h *circuitsCircuitgroupHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitgroupByIDsRequest) (*netbox_goV1.ListCircuitsCircuitgroupByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of circuitsCircuitgroups by last id
func (h *circuitsCircuitgroupHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitgroupByLastIDRequest) (*netbox_goV1.ListCircuitsCircuitgroupByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
