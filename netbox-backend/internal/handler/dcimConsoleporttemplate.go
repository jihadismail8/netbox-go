package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimConsoleporttemplateLogicer = (*dcimConsoleporttemplateHandler)(nil)

type dcimConsoleporttemplateHandler struct {
	server netbox_goV1.DcimConsoleporttemplateServer
}

// NewDcimConsoleporttemplateHandler create a handler
func NewDcimConsoleporttemplateHandler() netbox_goV1.DcimConsoleporttemplateLogicer {
	return &dcimConsoleporttemplateHandler{
		server: service.NewDcimConsoleporttemplateServer(),
	}
}

// Create a new dcimConsoleporttemplate
func (h *dcimConsoleporttemplateHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimConsoleporttemplateRequest) (*netbox_goV1.CreateDcimConsoleporttemplateReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimConsoleporttemplate by id
func (h *dcimConsoleporttemplateHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimConsoleporttemplateByIDRequest) (*netbox_goV1.DeleteDcimConsoleporttemplateByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimConsoleporttemplate by id
func (h *dcimConsoleporttemplateHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimConsoleporttemplateByIDRequest) (*netbox_goV1.UpdateDcimConsoleporttemplateByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimConsoleporttemplate by id
func (h *dcimConsoleporttemplateHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimConsoleporttemplateByIDRequest) (*netbox_goV1.GetDcimConsoleporttemplateByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimConsoleporttemplates by custom conditions
func (h *dcimConsoleporttemplateHandler) List(ctx context.Context, req *netbox_goV1.ListDcimConsoleporttemplateRequest) (*netbox_goV1.ListDcimConsoleporttemplateReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimConsoleporttemplate by ids
func (h *dcimConsoleporttemplateHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimConsoleporttemplateByIDsRequest) (*netbox_goV1.DeleteDcimConsoleporttemplateByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimConsoleporttemplate by custom condition
func (h *dcimConsoleporttemplateHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimConsoleporttemplateByConditionRequest) (*netbox_goV1.GetDcimConsoleporttemplateByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimConsoleporttemplate by ids
func (h *dcimConsoleporttemplateHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimConsoleporttemplateByIDsRequest) (*netbox_goV1.ListDcimConsoleporttemplateByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimConsoleporttemplates by last id
func (h *dcimConsoleporttemplateHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimConsoleporttemplateByLastIDRequest) (*netbox_goV1.ListDcimConsoleporttemplateByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
