package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasTableconfigLogicer = (*extrasTableconfigHandler)(nil)

type extrasTableconfigHandler struct {
	server netbox_goV1.ExtrasTableconfigServer
}

// NewExtrasTableconfigHandler create a handler
func NewExtrasTableconfigHandler() netbox_goV1.ExtrasTableconfigLogicer {
	return &extrasTableconfigHandler{
		server: service.NewExtrasTableconfigServer(),
	}
}

// Create a new extrasTableconfig
func (h *extrasTableconfigHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasTableconfigRequest) (*netbox_goV1.CreateExtrasTableconfigReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasTableconfig by id
func (h *extrasTableconfigHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasTableconfigByIDRequest) (*netbox_goV1.DeleteExtrasTableconfigByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasTableconfig by id
func (h *extrasTableconfigHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasTableconfigByIDRequest) (*netbox_goV1.UpdateExtrasTableconfigByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasTableconfig by id
func (h *extrasTableconfigHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasTableconfigByIDRequest) (*netbox_goV1.GetExtrasTableconfigByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasTableconfigs by custom conditions
func (h *extrasTableconfigHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasTableconfigRequest) (*netbox_goV1.ListExtrasTableconfigReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasTableconfig by ids
func (h *extrasTableconfigHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasTableconfigByIDsRequest) (*netbox_goV1.DeleteExtrasTableconfigByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasTableconfig by custom condition
func (h *extrasTableconfigHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasTableconfigByConditionRequest) (*netbox_goV1.GetExtrasTableconfigByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasTableconfig by ids
func (h *extrasTableconfigHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasTableconfigByIDsRequest) (*netbox_goV1.ListExtrasTableconfigByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasTableconfigs by last id
func (h *extrasTableconfigHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasTableconfigByLastIDRequest) (*netbox_goV1.ListExtrasTableconfigByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
