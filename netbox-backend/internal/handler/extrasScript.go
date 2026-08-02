package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasScriptLogicer = (*extrasScriptHandler)(nil)

type extrasScriptHandler struct {
	server netbox_goV1.ExtrasScriptServer
}

// NewExtrasScriptHandler create a handler
func NewExtrasScriptHandler() netbox_goV1.ExtrasScriptLogicer {
	return &extrasScriptHandler{
		server: service.NewExtrasScriptServer(),
	}
}

// Create a new extrasScript
func (h *extrasScriptHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasScriptRequest) (*netbox_goV1.CreateExtrasScriptReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasScript by id
func (h *extrasScriptHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasScriptByIDRequest) (*netbox_goV1.DeleteExtrasScriptByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasScript by id
func (h *extrasScriptHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasScriptByIDRequest) (*netbox_goV1.UpdateExtrasScriptByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasScript by id
func (h *extrasScriptHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasScriptByIDRequest) (*netbox_goV1.GetExtrasScriptByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasScripts by custom conditions
func (h *extrasScriptHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasScriptRequest) (*netbox_goV1.ListExtrasScriptReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasScript by ids
func (h *extrasScriptHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasScriptByIDsRequest) (*netbox_goV1.DeleteExtrasScriptByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasScript by custom condition
func (h *extrasScriptHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasScriptByConditionRequest) (*netbox_goV1.GetExtrasScriptByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasScript by ids
func (h *extrasScriptHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasScriptByIDsRequest) (*netbox_goV1.ListExtrasScriptByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasScripts by last id
func (h *extrasScriptHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasScriptByLastIDRequest) (*netbox_goV1.ListExtrasScriptByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
