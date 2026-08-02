package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasConfigcontextLogicer = (*extrasConfigcontextHandler)(nil)

type extrasConfigcontextHandler struct {
	server netbox_goV1.ExtrasConfigcontextServer
}

// NewExtrasConfigcontextHandler create a handler
func NewExtrasConfigcontextHandler() netbox_goV1.ExtrasConfigcontextLogicer {
	return &extrasConfigcontextHandler{
		server: service.NewExtrasConfigcontextServer(),
	}
}

// Create a new extrasConfigcontext
func (h *extrasConfigcontextHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasConfigcontextRequest) (*netbox_goV1.CreateExtrasConfigcontextReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasConfigcontext by id
func (h *extrasConfigcontextHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasConfigcontextByIDRequest) (*netbox_goV1.DeleteExtrasConfigcontextByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasConfigcontext by id
func (h *extrasConfigcontextHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasConfigcontextByIDRequest) (*netbox_goV1.UpdateExtrasConfigcontextByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasConfigcontext by id
func (h *extrasConfigcontextHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasConfigcontextByIDRequest) (*netbox_goV1.GetExtrasConfigcontextByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasConfigcontexts by custom conditions
func (h *extrasConfigcontextHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasConfigcontextRequest) (*netbox_goV1.ListExtrasConfigcontextReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasConfigcontext by ids
func (h *extrasConfigcontextHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasConfigcontextByIDsRequest) (*netbox_goV1.DeleteExtrasConfigcontextByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasConfigcontext by custom condition
func (h *extrasConfigcontextHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasConfigcontextByConditionRequest) (*netbox_goV1.GetExtrasConfigcontextByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasConfigcontext by ids
func (h *extrasConfigcontextHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasConfigcontextByIDsRequest) (*netbox_goV1.ListExtrasConfigcontextByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasConfigcontexts by last id
func (h *extrasConfigcontextHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasConfigcontextByLastIDRequest) (*netbox_goV1.ListExtrasConfigcontextByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
