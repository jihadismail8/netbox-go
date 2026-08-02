package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimConsoleserverporttemplateLogicer = (*dcimConsoleserverporttemplateHandler)(nil)

type dcimConsoleserverporttemplateHandler struct {
	server netbox_goV1.DcimConsoleserverporttemplateServer
}

// NewDcimConsoleserverporttemplateHandler create a handler
func NewDcimConsoleserverporttemplateHandler() netbox_goV1.DcimConsoleserverporttemplateLogicer {
	return &dcimConsoleserverporttemplateHandler{
		server: service.NewDcimConsoleserverporttemplateServer(),
	}
}

// Create a new dcimConsoleserverporttemplate
func (h *dcimConsoleserverporttemplateHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimConsoleserverporttemplateRequest) (*netbox_goV1.CreateDcimConsoleserverporttemplateReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimConsoleserverporttemplate by id
func (h *dcimConsoleserverporttemplateHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimConsoleserverporttemplateByIDRequest) (*netbox_goV1.DeleteDcimConsoleserverporttemplateByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimConsoleserverporttemplate by id
func (h *dcimConsoleserverporttemplateHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimConsoleserverporttemplateByIDRequest) (*netbox_goV1.UpdateDcimConsoleserverporttemplateByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimConsoleserverporttemplate by id
func (h *dcimConsoleserverporttemplateHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimConsoleserverporttemplateByIDRequest) (*netbox_goV1.GetDcimConsoleserverporttemplateByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimConsoleserverporttemplates by custom conditions
func (h *dcimConsoleserverporttemplateHandler) List(ctx context.Context, req *netbox_goV1.ListDcimConsoleserverporttemplateRequest) (*netbox_goV1.ListDcimConsoleserverporttemplateReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimConsoleserverporttemplate by ids
func (h *dcimConsoleserverporttemplateHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimConsoleserverporttemplateByIDsRequest) (*netbox_goV1.DeleteDcimConsoleserverporttemplateByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimConsoleserverporttemplate by custom condition
func (h *dcimConsoleserverporttemplateHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimConsoleserverporttemplateByConditionRequest) (*netbox_goV1.GetDcimConsoleserverporttemplateByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimConsoleserverporttemplate by ids
func (h *dcimConsoleserverporttemplateHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimConsoleserverporttemplateByIDsRequest) (*netbox_goV1.ListDcimConsoleserverporttemplateByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimConsoleserverporttemplates by last id
func (h *dcimConsoleserverporttemplateHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimConsoleserverporttemplateByLastIDRequest) (*netbox_goV1.ListDcimConsoleserverporttemplateByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
