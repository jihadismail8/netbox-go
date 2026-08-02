package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimModulebaytemplateLogicer = (*dcimModulebaytemplateHandler)(nil)

type dcimModulebaytemplateHandler struct {
	server netbox_goV1.DcimModulebaytemplateServer
}

// NewDcimModulebaytemplateHandler create a handler
func NewDcimModulebaytemplateHandler() netbox_goV1.DcimModulebaytemplateLogicer {
	return &dcimModulebaytemplateHandler{
		server: service.NewDcimModulebaytemplateServer(),
	}
}

// Create a new dcimModulebaytemplate
func (h *dcimModulebaytemplateHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimModulebaytemplateRequest) (*netbox_goV1.CreateDcimModulebaytemplateReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimModulebaytemplate by id
func (h *dcimModulebaytemplateHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimModulebaytemplateByIDRequest) (*netbox_goV1.DeleteDcimModulebaytemplateByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimModulebaytemplate by id
func (h *dcimModulebaytemplateHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimModulebaytemplateByIDRequest) (*netbox_goV1.UpdateDcimModulebaytemplateByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimModulebaytemplate by id
func (h *dcimModulebaytemplateHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimModulebaytemplateByIDRequest) (*netbox_goV1.GetDcimModulebaytemplateByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimModulebaytemplates by custom conditions
func (h *dcimModulebaytemplateHandler) List(ctx context.Context, req *netbox_goV1.ListDcimModulebaytemplateRequest) (*netbox_goV1.ListDcimModulebaytemplateReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimModulebaytemplate by ids
func (h *dcimModulebaytemplateHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimModulebaytemplateByIDsRequest) (*netbox_goV1.DeleteDcimModulebaytemplateByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimModulebaytemplate by custom condition
func (h *dcimModulebaytemplateHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimModulebaytemplateByConditionRequest) (*netbox_goV1.GetDcimModulebaytemplateByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimModulebaytemplate by ids
func (h *dcimModulebaytemplateHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimModulebaytemplateByIDsRequest) (*netbox_goV1.ListDcimModulebaytemplateByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimModulebaytemplates by last id
func (h *dcimModulebaytemplateHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimModulebaytemplateByLastIDRequest) (*netbox_goV1.ListDcimModulebaytemplateByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
