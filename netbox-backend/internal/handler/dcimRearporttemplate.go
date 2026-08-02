package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimRearporttemplateLogicer = (*dcimRearporttemplateHandler)(nil)

type dcimRearporttemplateHandler struct {
	server netbox_goV1.DcimRearporttemplateServer
}

// NewDcimRearporttemplateHandler create a handler
func NewDcimRearporttemplateHandler() netbox_goV1.DcimRearporttemplateLogicer {
	return &dcimRearporttemplateHandler{
		server: service.NewDcimRearporttemplateServer(),
	}
}

// Create a new dcimRearporttemplate
func (h *dcimRearporttemplateHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimRearporttemplateRequest) (*netbox_goV1.CreateDcimRearporttemplateReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimRearporttemplate by id
func (h *dcimRearporttemplateHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimRearporttemplateByIDRequest) (*netbox_goV1.DeleteDcimRearporttemplateByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimRearporttemplate by id
func (h *dcimRearporttemplateHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimRearporttemplateByIDRequest) (*netbox_goV1.UpdateDcimRearporttemplateByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimRearporttemplate by id
func (h *dcimRearporttemplateHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimRearporttemplateByIDRequest) (*netbox_goV1.GetDcimRearporttemplateByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimRearporttemplates by custom conditions
func (h *dcimRearporttemplateHandler) List(ctx context.Context, req *netbox_goV1.ListDcimRearporttemplateRequest) (*netbox_goV1.ListDcimRearporttemplateReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimRearporttemplate by ids
func (h *dcimRearporttemplateHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimRearporttemplateByIDsRequest) (*netbox_goV1.DeleteDcimRearporttemplateByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimRearporttemplate by custom condition
func (h *dcimRearporttemplateHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimRearporttemplateByConditionRequest) (*netbox_goV1.GetDcimRearporttemplateByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimRearporttemplate by ids
func (h *dcimRearporttemplateHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimRearporttemplateByIDsRequest) (*netbox_goV1.ListDcimRearporttemplateByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimRearporttemplates by last id
func (h *dcimRearporttemplateHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimRearporttemplateByLastIDRequest) (*netbox_goV1.ListDcimRearporttemplateByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
