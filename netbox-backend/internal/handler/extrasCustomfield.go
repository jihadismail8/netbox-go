package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasCustomfieldLogicer = (*extrasCustomfieldHandler)(nil)

type extrasCustomfieldHandler struct {
	server netbox_goV1.ExtrasCustomfieldServer
}

// NewExtrasCustomfieldHandler create a handler
func NewExtrasCustomfieldHandler() netbox_goV1.ExtrasCustomfieldLogicer {
	return &extrasCustomfieldHandler{
		server: service.NewExtrasCustomfieldServer(),
	}
}

// Create a new extrasCustomfield
func (h *extrasCustomfieldHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasCustomfieldRequest) (*netbox_goV1.CreateExtrasCustomfieldReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasCustomfield by id
func (h *extrasCustomfieldHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasCustomfieldByIDRequest) (*netbox_goV1.DeleteExtrasCustomfieldByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasCustomfield by id
func (h *extrasCustomfieldHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasCustomfieldByIDRequest) (*netbox_goV1.UpdateExtrasCustomfieldByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasCustomfield by id
func (h *extrasCustomfieldHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasCustomfieldByIDRequest) (*netbox_goV1.GetExtrasCustomfieldByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasCustomfields by custom conditions
func (h *extrasCustomfieldHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasCustomfieldRequest) (*netbox_goV1.ListExtrasCustomfieldReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasCustomfield by ids
func (h *extrasCustomfieldHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasCustomfieldByIDsRequest) (*netbox_goV1.DeleteExtrasCustomfieldByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasCustomfield by custom condition
func (h *extrasCustomfieldHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasCustomfieldByConditionRequest) (*netbox_goV1.GetExtrasCustomfieldByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasCustomfield by ids
func (h *extrasCustomfieldHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasCustomfieldByIDsRequest) (*netbox_goV1.ListExtrasCustomfieldByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasCustomfields by last id
func (h *extrasCustomfieldHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasCustomfieldByLastIDRequest) (*netbox_goV1.ListExtrasCustomfieldByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
