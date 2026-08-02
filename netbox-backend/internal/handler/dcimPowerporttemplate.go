package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimPowerporttemplateLogicer = (*dcimPowerporttemplateHandler)(nil)

type dcimPowerporttemplateHandler struct {
	server netbox_goV1.DcimPowerporttemplateServer
}

// NewDcimPowerporttemplateHandler create a handler
func NewDcimPowerporttemplateHandler() netbox_goV1.DcimPowerporttemplateLogicer {
	return &dcimPowerporttemplateHandler{
		server: service.NewDcimPowerporttemplateServer(),
	}
}

// Create a new dcimPowerporttemplate
func (h *dcimPowerporttemplateHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimPowerporttemplateRequest) (*netbox_goV1.CreateDcimPowerporttemplateReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimPowerporttemplate by id
func (h *dcimPowerporttemplateHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimPowerporttemplateByIDRequest) (*netbox_goV1.DeleteDcimPowerporttemplateByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimPowerporttemplate by id
func (h *dcimPowerporttemplateHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimPowerporttemplateByIDRequest) (*netbox_goV1.UpdateDcimPowerporttemplateByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimPowerporttemplate by id
func (h *dcimPowerporttemplateHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimPowerporttemplateByIDRequest) (*netbox_goV1.GetDcimPowerporttemplateByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimPowerporttemplates by custom conditions
func (h *dcimPowerporttemplateHandler) List(ctx context.Context, req *netbox_goV1.ListDcimPowerporttemplateRequest) (*netbox_goV1.ListDcimPowerporttemplateReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimPowerporttemplate by ids
func (h *dcimPowerporttemplateHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimPowerporttemplateByIDsRequest) (*netbox_goV1.DeleteDcimPowerporttemplateByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimPowerporttemplate by custom condition
func (h *dcimPowerporttemplateHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimPowerporttemplateByConditionRequest) (*netbox_goV1.GetDcimPowerporttemplateByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimPowerporttemplate by ids
func (h *dcimPowerporttemplateHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimPowerporttemplateByIDsRequest) (*netbox_goV1.ListDcimPowerporttemplateByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimPowerporttemplates by last id
func (h *dcimPowerporttemplateHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimPowerporttemplateByLastIDRequest) (*netbox_goV1.ListDcimPowerporttemplateByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
