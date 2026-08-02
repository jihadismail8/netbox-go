package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimModuleLogicer = (*dcimModuleHandler)(nil)

type dcimModuleHandler struct {
	server netbox_goV1.DcimModuleServer
}

// NewDcimModuleHandler create a handler
func NewDcimModuleHandler() netbox_goV1.DcimModuleLogicer {
	return &dcimModuleHandler{
		server: service.NewDcimModuleServer(),
	}
}

// Create a new dcimModule
func (h *dcimModuleHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimModuleRequest) (*netbox_goV1.CreateDcimModuleReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimModule by id
func (h *dcimModuleHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimModuleByIDRequest) (*netbox_goV1.DeleteDcimModuleByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimModule by id
func (h *dcimModuleHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimModuleByIDRequest) (*netbox_goV1.UpdateDcimModuleByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimModule by id
func (h *dcimModuleHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimModuleByIDRequest) (*netbox_goV1.GetDcimModuleByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimModules by custom conditions
func (h *dcimModuleHandler) List(ctx context.Context, req *netbox_goV1.ListDcimModuleRequest) (*netbox_goV1.ListDcimModuleReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimModule by ids
func (h *dcimModuleHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimModuleByIDsRequest) (*netbox_goV1.DeleteDcimModuleByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimModule by custom condition
func (h *dcimModuleHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimModuleByConditionRequest) (*netbox_goV1.GetDcimModuleByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimModule by ids
func (h *dcimModuleHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimModuleByIDsRequest) (*netbox_goV1.ListDcimModuleByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimModules by last id
func (h *dcimModuleHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimModuleByLastIDRequest) (*netbox_goV1.ListDcimModuleByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
