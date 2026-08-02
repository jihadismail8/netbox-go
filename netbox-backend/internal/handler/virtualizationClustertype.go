package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VirtualizationClustertypeLogicer = (*virtualizationClustertypeHandler)(nil)

type virtualizationClustertypeHandler struct {
	server netbox_goV1.VirtualizationClustertypeServer
}

// NewVirtualizationClustertypeHandler create a handler
func NewVirtualizationClustertypeHandler() netbox_goV1.VirtualizationClustertypeLogicer {
	return &virtualizationClustertypeHandler{
		server: service.NewVirtualizationClustertypeServer(),
	}
}

// Create a new virtualizationClustertype
func (h *virtualizationClustertypeHandler) Create(ctx context.Context, req *netbox_goV1.CreateVirtualizationClustertypeRequest) (*netbox_goV1.CreateVirtualizationClustertypeReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a virtualizationClustertype by id
func (h *virtualizationClustertypeHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVirtualizationClustertypeByIDRequest) (*netbox_goV1.DeleteVirtualizationClustertypeByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a virtualizationClustertype by id
func (h *virtualizationClustertypeHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVirtualizationClustertypeByIDRequest) (*netbox_goV1.UpdateVirtualizationClustertypeByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a virtualizationClustertype by id
func (h *virtualizationClustertypeHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVirtualizationClustertypeByIDRequest) (*netbox_goV1.GetVirtualizationClustertypeByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of virtualizationClustertypes by custom conditions
func (h *virtualizationClustertypeHandler) List(ctx context.Context, req *netbox_goV1.ListVirtualizationClustertypeRequest) (*netbox_goV1.ListVirtualizationClustertypeReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete virtualizationClustertype by ids
func (h *virtualizationClustertypeHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVirtualizationClustertypeByIDsRequest) (*netbox_goV1.DeleteVirtualizationClustertypeByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a virtualizationClustertype by custom condition
func (h *virtualizationClustertypeHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVirtualizationClustertypeByConditionRequest) (*netbox_goV1.GetVirtualizationClustertypeByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get virtualizationClustertype by ids
func (h *virtualizationClustertypeHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVirtualizationClustertypeByIDsRequest) (*netbox_goV1.ListVirtualizationClustertypeByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of virtualizationClustertypes by last id
func (h *virtualizationClustertypeHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVirtualizationClustertypeByLastIDRequest) (*netbox_goV1.ListVirtualizationClustertypeByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
