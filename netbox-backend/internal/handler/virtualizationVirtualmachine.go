package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VirtualizationVirtualmachineLogicer = (*virtualizationVirtualmachineHandler)(nil)

type virtualizationVirtualmachineHandler struct {
	server netbox_goV1.VirtualizationVirtualmachineServer
}

// NewVirtualizationVirtualmachineHandler create a handler
func NewVirtualizationVirtualmachineHandler() netbox_goV1.VirtualizationVirtualmachineLogicer {
	return &virtualizationVirtualmachineHandler{
		server: service.NewVirtualizationVirtualmachineServer(),
	}
}

// Create a new virtualizationVirtualmachine
func (h *virtualizationVirtualmachineHandler) Create(ctx context.Context, req *netbox_goV1.CreateVirtualizationVirtualmachineRequest) (*netbox_goV1.CreateVirtualizationVirtualmachineReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a virtualizationVirtualmachine by id
func (h *virtualizationVirtualmachineHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVirtualizationVirtualmachineByIDRequest) (*netbox_goV1.DeleteVirtualizationVirtualmachineByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a virtualizationVirtualmachine by id
func (h *virtualizationVirtualmachineHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVirtualizationVirtualmachineByIDRequest) (*netbox_goV1.UpdateVirtualizationVirtualmachineByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a virtualizationVirtualmachine by id
func (h *virtualizationVirtualmachineHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVirtualizationVirtualmachineByIDRequest) (*netbox_goV1.GetVirtualizationVirtualmachineByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of virtualizationVirtualmachines by custom conditions
func (h *virtualizationVirtualmachineHandler) List(ctx context.Context, req *netbox_goV1.ListVirtualizationVirtualmachineRequest) (*netbox_goV1.ListVirtualizationVirtualmachineReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete virtualizationVirtualmachine by ids
func (h *virtualizationVirtualmachineHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVirtualizationVirtualmachineByIDsRequest) (*netbox_goV1.DeleteVirtualizationVirtualmachineByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a virtualizationVirtualmachine by custom condition
func (h *virtualizationVirtualmachineHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVirtualizationVirtualmachineByConditionRequest) (*netbox_goV1.GetVirtualizationVirtualmachineByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get virtualizationVirtualmachine by ids
func (h *virtualizationVirtualmachineHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVirtualizationVirtualmachineByIDsRequest) (*netbox_goV1.ListVirtualizationVirtualmachineByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of virtualizationVirtualmachines by last id
func (h *virtualizationVirtualmachineHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVirtualizationVirtualmachineByLastIDRequest) (*netbox_goV1.ListVirtualizationVirtualmachineByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
