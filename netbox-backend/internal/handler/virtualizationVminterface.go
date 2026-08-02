package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VirtualizationVminterfaceLogicer = (*virtualizationVminterfaceHandler)(nil)

type virtualizationVminterfaceHandler struct {
	server netbox_goV1.VirtualizationVminterfaceServer
}

// NewVirtualizationVminterfaceHandler create a handler
func NewVirtualizationVminterfaceHandler() netbox_goV1.VirtualizationVminterfaceLogicer {
	return &virtualizationVminterfaceHandler{
		server: service.NewVirtualizationVminterfaceServer(),
	}
}

// Create a new virtualizationVminterface
func (h *virtualizationVminterfaceHandler) Create(ctx context.Context, req *netbox_goV1.CreateVirtualizationVminterfaceRequest) (*netbox_goV1.CreateVirtualizationVminterfaceReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a virtualizationVminterface by id
func (h *virtualizationVminterfaceHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVirtualizationVminterfaceByIDRequest) (*netbox_goV1.DeleteVirtualizationVminterfaceByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a virtualizationVminterface by id
func (h *virtualizationVminterfaceHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVirtualizationVminterfaceByIDRequest) (*netbox_goV1.UpdateVirtualizationVminterfaceByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a virtualizationVminterface by id
func (h *virtualizationVminterfaceHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVirtualizationVminterfaceByIDRequest) (*netbox_goV1.GetVirtualizationVminterfaceByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of virtualizationVminterfaces by custom conditions
func (h *virtualizationVminterfaceHandler) List(ctx context.Context, req *netbox_goV1.ListVirtualizationVminterfaceRequest) (*netbox_goV1.ListVirtualizationVminterfaceReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete virtualizationVminterface by ids
func (h *virtualizationVminterfaceHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVirtualizationVminterfaceByIDsRequest) (*netbox_goV1.DeleteVirtualizationVminterfaceByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a virtualizationVminterface by custom condition
func (h *virtualizationVminterfaceHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVirtualizationVminterfaceByConditionRequest) (*netbox_goV1.GetVirtualizationVminterfaceByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get virtualizationVminterface by ids
func (h *virtualizationVminterfaceHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVirtualizationVminterfaceByIDsRequest) (*netbox_goV1.ListVirtualizationVminterfaceByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of virtualizationVminterfaces by last id
func (h *virtualizationVminterfaceHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVirtualizationVminterfaceByLastIDRequest) (*netbox_goV1.ListVirtualizationVminterfaceByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
