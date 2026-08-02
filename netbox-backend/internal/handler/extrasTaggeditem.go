package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasTaggeditemLogicer = (*extrasTaggeditemHandler)(nil)

type extrasTaggeditemHandler struct {
	server netbox_goV1.ExtrasTaggeditemServer
}

// NewExtrasTaggeditemHandler create a handler
func NewExtrasTaggeditemHandler() netbox_goV1.ExtrasTaggeditemLogicer {
	return &extrasTaggeditemHandler{
		server: service.NewExtrasTaggeditemServer(),
	}
}

// Create a new extrasTaggeditem
func (h *extrasTaggeditemHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasTaggeditemRequest) (*netbox_goV1.CreateExtrasTaggeditemReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasTaggeditem by id
func (h *extrasTaggeditemHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasTaggeditemByIDRequest) (*netbox_goV1.DeleteExtrasTaggeditemByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasTaggeditem by id
func (h *extrasTaggeditemHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasTaggeditemByIDRequest) (*netbox_goV1.UpdateExtrasTaggeditemByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasTaggeditem by id
func (h *extrasTaggeditemHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasTaggeditemByIDRequest) (*netbox_goV1.GetExtrasTaggeditemByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasTaggeditems by custom conditions
func (h *extrasTaggeditemHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasTaggeditemRequest) (*netbox_goV1.ListExtrasTaggeditemReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasTaggeditem by ids
func (h *extrasTaggeditemHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasTaggeditemByIDsRequest) (*netbox_goV1.DeleteExtrasTaggeditemByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasTaggeditem by custom condition
func (h *extrasTaggeditemHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasTaggeditemByConditionRequest) (*netbox_goV1.GetExtrasTaggeditemByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasTaggeditem by ids
func (h *extrasTaggeditemHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasTaggeditemByIDsRequest) (*netbox_goV1.ListExtrasTaggeditemByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasTaggeditems by last id
func (h *extrasTaggeditemHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasTaggeditemByLastIDRequest) (*netbox_goV1.ListExtrasTaggeditemByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
