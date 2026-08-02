package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimInventoryitemtemplateLogicer = (*dcimInventoryitemtemplateHandler)(nil)

type dcimInventoryitemtemplateHandler struct {
	server netbox_goV1.DcimInventoryitemtemplateServer
}

// NewDcimInventoryitemtemplateHandler create a handler
func NewDcimInventoryitemtemplateHandler() netbox_goV1.DcimInventoryitemtemplateLogicer {
	return &dcimInventoryitemtemplateHandler{
		server: service.NewDcimInventoryitemtemplateServer(),
	}
}

// Create a new dcimInventoryitemtemplate
func (h *dcimInventoryitemtemplateHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimInventoryitemtemplateRequest) (*netbox_goV1.CreateDcimInventoryitemtemplateReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimInventoryitemtemplate by id
func (h *dcimInventoryitemtemplateHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimInventoryitemtemplateByIDRequest) (*netbox_goV1.DeleteDcimInventoryitemtemplateByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimInventoryitemtemplate by id
func (h *dcimInventoryitemtemplateHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimInventoryitemtemplateByIDRequest) (*netbox_goV1.UpdateDcimInventoryitemtemplateByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimInventoryitemtemplate by id
func (h *dcimInventoryitemtemplateHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimInventoryitemtemplateByIDRequest) (*netbox_goV1.GetDcimInventoryitemtemplateByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimInventoryitemtemplates by custom conditions
func (h *dcimInventoryitemtemplateHandler) List(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemtemplateRequest) (*netbox_goV1.ListDcimInventoryitemtemplateReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimInventoryitemtemplate by ids
func (h *dcimInventoryitemtemplateHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimInventoryitemtemplateByIDsRequest) (*netbox_goV1.DeleteDcimInventoryitemtemplateByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimInventoryitemtemplate by custom condition
func (h *dcimInventoryitemtemplateHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimInventoryitemtemplateByConditionRequest) (*netbox_goV1.GetDcimInventoryitemtemplateByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimInventoryitemtemplate by ids
func (h *dcimInventoryitemtemplateHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemtemplateByIDsRequest) (*netbox_goV1.ListDcimInventoryitemtemplateByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimInventoryitemtemplates by last id
func (h *dcimInventoryitemtemplateHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemtemplateByLastIDRequest) (*netbox_goV1.ListDcimInventoryitemtemplateByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
