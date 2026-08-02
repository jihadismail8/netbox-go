package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CircuitsCircuitterminationLogicer = (*circuitsCircuitterminationHandler)(nil)

type circuitsCircuitterminationHandler struct {
	server netbox_goV1.CircuitsCircuitterminationServer
}

// NewCircuitsCircuitterminationHandler create a handler
func NewCircuitsCircuitterminationHandler() netbox_goV1.CircuitsCircuitterminationLogicer {
	return &circuitsCircuitterminationHandler{
		server: service.NewCircuitsCircuitterminationServer(),
	}
}

// Create a new circuitsCircuittermination
func (h *circuitsCircuitterminationHandler) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsCircuitterminationRequest) (*netbox_goV1.CreateCircuitsCircuitterminationReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a circuitsCircuittermination by id
func (h *circuitsCircuitterminationHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitterminationByIDRequest) (*netbox_goV1.DeleteCircuitsCircuitterminationByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a circuitsCircuittermination by id
func (h *circuitsCircuitterminationHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsCircuitterminationByIDRequest) (*netbox_goV1.UpdateCircuitsCircuitterminationByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a circuitsCircuittermination by id
func (h *circuitsCircuitterminationHandler) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitterminationByIDRequest) (*netbox_goV1.GetCircuitsCircuitterminationByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of circuitsCircuitterminations by custom conditions
func (h *circuitsCircuitterminationHandler) List(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitterminationRequest) (*netbox_goV1.ListCircuitsCircuitterminationReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete circuitsCircuittermination by ids
func (h *circuitsCircuitterminationHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitterminationByIDsRequest) (*netbox_goV1.DeleteCircuitsCircuitterminationByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a circuitsCircuittermination by custom condition
func (h *circuitsCircuitterminationHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitterminationByConditionRequest) (*netbox_goV1.GetCircuitsCircuitterminationByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get circuitsCircuittermination by ids
func (h *circuitsCircuitterminationHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitterminationByIDsRequest) (*netbox_goV1.ListCircuitsCircuitterminationByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of circuitsCircuitterminations by last id
func (h *circuitsCircuitterminationHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitterminationByLastIDRequest) (*netbox_goV1.ListCircuitsCircuitterminationByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
