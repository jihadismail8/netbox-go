package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasCustomlinkLogicer = (*extrasCustomlinkHandler)(nil)

type extrasCustomlinkHandler struct {
	server netbox_goV1.ExtrasCustomlinkServer
}

// NewExtrasCustomlinkHandler create a handler
func NewExtrasCustomlinkHandler() netbox_goV1.ExtrasCustomlinkLogicer {
	return &extrasCustomlinkHandler{
		server: service.NewExtrasCustomlinkServer(),
	}
}

// Create a new extrasCustomlink
func (h *extrasCustomlinkHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasCustomlinkRequest) (*netbox_goV1.CreateExtrasCustomlinkReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasCustomlink by id
func (h *extrasCustomlinkHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasCustomlinkByIDRequest) (*netbox_goV1.DeleteExtrasCustomlinkByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasCustomlink by id
func (h *extrasCustomlinkHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasCustomlinkByIDRequest) (*netbox_goV1.UpdateExtrasCustomlinkByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasCustomlink by id
func (h *extrasCustomlinkHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasCustomlinkByIDRequest) (*netbox_goV1.GetExtrasCustomlinkByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasCustomlinks by custom conditions
func (h *extrasCustomlinkHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasCustomlinkRequest) (*netbox_goV1.ListExtrasCustomlinkReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasCustomlink by ids
func (h *extrasCustomlinkHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasCustomlinkByIDsRequest) (*netbox_goV1.DeleteExtrasCustomlinkByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasCustomlink by custom condition
func (h *extrasCustomlinkHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasCustomlinkByConditionRequest) (*netbox_goV1.GetExtrasCustomlinkByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasCustomlink by ids
func (h *extrasCustomlinkHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasCustomlinkByIDsRequest) (*netbox_goV1.ListExtrasCustomlinkByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasCustomlinks by last id
func (h *extrasCustomlinkHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasCustomlinkByLastIDRequest) (*netbox_goV1.ListExtrasCustomlinkByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
