package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasConfigtemplateLogicer = (*extrasConfigtemplateHandler)(nil)

type extrasConfigtemplateHandler struct {
	server netbox_goV1.ExtrasConfigtemplateServer
}

// NewExtrasConfigtemplateHandler create a handler
func NewExtrasConfigtemplateHandler() netbox_goV1.ExtrasConfigtemplateLogicer {
	return &extrasConfigtemplateHandler{
		server: service.NewExtrasConfigtemplateServer(),
	}
}

// Create a new extrasConfigtemplate
func (h *extrasConfigtemplateHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasConfigtemplateRequest) (*netbox_goV1.CreateExtrasConfigtemplateReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasConfigtemplate by id
func (h *extrasConfigtemplateHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasConfigtemplateByIDRequest) (*netbox_goV1.DeleteExtrasConfigtemplateByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasConfigtemplate by id
func (h *extrasConfigtemplateHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasConfigtemplateByIDRequest) (*netbox_goV1.UpdateExtrasConfigtemplateByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasConfigtemplate by id
func (h *extrasConfigtemplateHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasConfigtemplateByIDRequest) (*netbox_goV1.GetExtrasConfigtemplateByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasConfigtemplates by custom conditions
func (h *extrasConfigtemplateHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasConfigtemplateRequest) (*netbox_goV1.ListExtrasConfigtemplateReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasConfigtemplate by ids
func (h *extrasConfigtemplateHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasConfigtemplateByIDsRequest) (*netbox_goV1.DeleteExtrasConfigtemplateByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasConfigtemplate by custom condition
func (h *extrasConfigtemplateHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasConfigtemplateByConditionRequest) (*netbox_goV1.GetExtrasConfigtemplateByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasConfigtemplate by ids
func (h *extrasConfigtemplateHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasConfigtemplateByIDsRequest) (*netbox_goV1.ListExtrasConfigtemplateByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasConfigtemplates by last id
func (h *extrasConfigtemplateHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasConfigtemplateByLastIDRequest) (*netbox_goV1.ListExtrasConfigtemplateByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
