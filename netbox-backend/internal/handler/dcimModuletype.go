package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimModuletypeLogicer = (*dcimModuletypeHandler)(nil)

type dcimModuletypeHandler struct {
	server netbox_goV1.DcimModuletypeServer
}

// NewDcimModuletypeHandler create a handler
func NewDcimModuletypeHandler() netbox_goV1.DcimModuletypeLogicer {
	return &dcimModuletypeHandler{
		server: service.NewDcimModuletypeServer(),
	}
}

// Create a new dcimModuletype
func (h *dcimModuletypeHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimModuletypeRequest) (*netbox_goV1.CreateDcimModuletypeReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimModuletype by id
func (h *dcimModuletypeHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimModuletypeByIDRequest) (*netbox_goV1.DeleteDcimModuletypeByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimModuletype by id
func (h *dcimModuletypeHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimModuletypeByIDRequest) (*netbox_goV1.UpdateDcimModuletypeByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimModuletype by id
func (h *dcimModuletypeHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimModuletypeByIDRequest) (*netbox_goV1.GetDcimModuletypeByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimModuletypes by custom conditions
func (h *dcimModuletypeHandler) List(ctx context.Context, req *netbox_goV1.ListDcimModuletypeRequest) (*netbox_goV1.ListDcimModuletypeReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimModuletype by ids
func (h *dcimModuletypeHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimModuletypeByIDsRequest) (*netbox_goV1.DeleteDcimModuletypeByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimModuletype by custom condition
func (h *dcimModuletypeHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimModuletypeByConditionRequest) (*netbox_goV1.GetDcimModuletypeByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimModuletype by ids
func (h *dcimModuletypeHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimModuletypeByIDsRequest) (*netbox_goV1.ListDcimModuletypeByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimModuletypes by last id
func (h *dcimModuletypeHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimModuletypeByLastIDRequest) (*netbox_goV1.ListDcimModuletypeByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
