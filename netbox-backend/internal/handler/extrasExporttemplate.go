package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasExporttemplateLogicer = (*extrasExporttemplateHandler)(nil)

type extrasExporttemplateHandler struct {
	server netbox_goV1.ExtrasExporttemplateServer
}

// NewExtrasExporttemplateHandler create a handler
func NewExtrasExporttemplateHandler() netbox_goV1.ExtrasExporttemplateLogicer {
	return &extrasExporttemplateHandler{
		server: service.NewExtrasExporttemplateServer(),
	}
}

// Create a new extrasExporttemplate
func (h *extrasExporttemplateHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasExporttemplateRequest) (*netbox_goV1.CreateExtrasExporttemplateReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasExporttemplate by id
func (h *extrasExporttemplateHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasExporttemplateByIDRequest) (*netbox_goV1.DeleteExtrasExporttemplateByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasExporttemplate by id
func (h *extrasExporttemplateHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasExporttemplateByIDRequest) (*netbox_goV1.UpdateExtrasExporttemplateByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasExporttemplate by id
func (h *extrasExporttemplateHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasExporttemplateByIDRequest) (*netbox_goV1.GetExtrasExporttemplateByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasExporttemplates by custom conditions
func (h *extrasExporttemplateHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasExporttemplateRequest) (*netbox_goV1.ListExtrasExporttemplateReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasExporttemplate by ids
func (h *extrasExporttemplateHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasExporttemplateByIDsRequest) (*netbox_goV1.DeleteExtrasExporttemplateByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasExporttemplate by custom condition
func (h *extrasExporttemplateHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasExporttemplateByConditionRequest) (*netbox_goV1.GetExtrasExporttemplateByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasExporttemplate by ids
func (h *extrasExporttemplateHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasExporttemplateByIDsRequest) (*netbox_goV1.ListExtrasExporttemplateByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasExporttemplates by last id
func (h *extrasExporttemplateHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasExporttemplateByLastIDRequest) (*netbox_goV1.ListExtrasExporttemplateByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
