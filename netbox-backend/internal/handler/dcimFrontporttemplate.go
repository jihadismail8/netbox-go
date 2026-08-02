package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimFrontporttemplateLogicer = (*dcimFrontporttemplateHandler)(nil)

type dcimFrontporttemplateHandler struct {
	server netbox_goV1.DcimFrontporttemplateServer
}

// NewDcimFrontporttemplateHandler create a handler
func NewDcimFrontporttemplateHandler() netbox_goV1.DcimFrontporttemplateLogicer {
	return &dcimFrontporttemplateHandler{
		server: service.NewDcimFrontporttemplateServer(),
	}
}

// Create a new dcimFrontporttemplate
func (h *dcimFrontporttemplateHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimFrontporttemplateRequest) (*netbox_goV1.CreateDcimFrontporttemplateReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimFrontporttemplate by id
func (h *dcimFrontporttemplateHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimFrontporttemplateByIDRequest) (*netbox_goV1.DeleteDcimFrontporttemplateByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimFrontporttemplate by id
func (h *dcimFrontporttemplateHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimFrontporttemplateByIDRequest) (*netbox_goV1.UpdateDcimFrontporttemplateByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimFrontporttemplate by id
func (h *dcimFrontporttemplateHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimFrontporttemplateByIDRequest) (*netbox_goV1.GetDcimFrontporttemplateByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimFrontporttemplates by custom conditions
func (h *dcimFrontporttemplateHandler) List(ctx context.Context, req *netbox_goV1.ListDcimFrontporttemplateRequest) (*netbox_goV1.ListDcimFrontporttemplateReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimFrontporttemplate by ids
func (h *dcimFrontporttemplateHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimFrontporttemplateByIDsRequest) (*netbox_goV1.DeleteDcimFrontporttemplateByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimFrontporttemplate by custom condition
func (h *dcimFrontporttemplateHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimFrontporttemplateByConditionRequest) (*netbox_goV1.GetDcimFrontporttemplateByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimFrontporttemplate by ids
func (h *dcimFrontporttemplateHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimFrontporttemplateByIDsRequest) (*netbox_goV1.ListDcimFrontporttemplateByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimFrontporttemplates by last id
func (h *dcimFrontporttemplateHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimFrontporttemplateByLastIDRequest) (*netbox_goV1.ListDcimFrontporttemplateByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
