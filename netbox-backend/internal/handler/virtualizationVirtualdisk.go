package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VirtualizationVirtualdiskLogicer = (*virtualizationVirtualdiskHandler)(nil)

type virtualizationVirtualdiskHandler struct {
	server netbox_goV1.VirtualizationVirtualdiskServer
}

// NewVirtualizationVirtualdiskHandler create a handler
func NewVirtualizationVirtualdiskHandler() netbox_goV1.VirtualizationVirtualdiskLogicer {
	return &virtualizationVirtualdiskHandler{
		server: service.NewVirtualizationVirtualdiskServer(),
	}
}

// Create a new virtualizationVirtualdisk
func (h *virtualizationVirtualdiskHandler) Create(ctx context.Context, req *netbox_goV1.CreateVirtualizationVirtualdiskRequest) (*netbox_goV1.CreateVirtualizationVirtualdiskReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a virtualizationVirtualdisk by id
func (h *virtualizationVirtualdiskHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVirtualizationVirtualdiskByIDRequest) (*netbox_goV1.DeleteVirtualizationVirtualdiskByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a virtualizationVirtualdisk by id
func (h *virtualizationVirtualdiskHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVirtualizationVirtualdiskByIDRequest) (*netbox_goV1.UpdateVirtualizationVirtualdiskByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a virtualizationVirtualdisk by id
func (h *virtualizationVirtualdiskHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVirtualizationVirtualdiskByIDRequest) (*netbox_goV1.GetVirtualizationVirtualdiskByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of virtualizationVirtualdisks by custom conditions
func (h *virtualizationVirtualdiskHandler) List(ctx context.Context, req *netbox_goV1.ListVirtualizationVirtualdiskRequest) (*netbox_goV1.ListVirtualizationVirtualdiskReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete virtualizationVirtualdisk by ids
func (h *virtualizationVirtualdiskHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVirtualizationVirtualdiskByIDsRequest) (*netbox_goV1.DeleteVirtualizationVirtualdiskByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a virtualizationVirtualdisk by custom condition
func (h *virtualizationVirtualdiskHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVirtualizationVirtualdiskByConditionRequest) (*netbox_goV1.GetVirtualizationVirtualdiskByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get virtualizationVirtualdisk by ids
func (h *virtualizationVirtualdiskHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVirtualizationVirtualdiskByIDsRequest) (*netbox_goV1.ListVirtualizationVirtualdiskByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of virtualizationVirtualdisks by last id
func (h *virtualizationVirtualdiskHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVirtualizationVirtualdiskByLastIDRequest) (*netbox_goV1.ListVirtualizationVirtualdiskByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
