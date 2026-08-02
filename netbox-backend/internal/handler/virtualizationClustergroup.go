package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VirtualizationClustergroupLogicer = (*virtualizationClustergroupHandler)(nil)

type virtualizationClustergroupHandler struct {
	server netbox_goV1.VirtualizationClustergroupServer
}

// NewVirtualizationClustergroupHandler create a handler
func NewVirtualizationClustergroupHandler() netbox_goV1.VirtualizationClustergroupLogicer {
	return &virtualizationClustergroupHandler{
		server: service.NewVirtualizationClustergroupServer(),
	}
}

// Create a new virtualizationClustergroup
func (h *virtualizationClustergroupHandler) Create(ctx context.Context, req *netbox_goV1.CreateVirtualizationClustergroupRequest) (*netbox_goV1.CreateVirtualizationClustergroupReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a virtualizationClustergroup by id
func (h *virtualizationClustergroupHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVirtualizationClustergroupByIDRequest) (*netbox_goV1.DeleteVirtualizationClustergroupByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a virtualizationClustergroup by id
func (h *virtualizationClustergroupHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVirtualizationClustergroupByIDRequest) (*netbox_goV1.UpdateVirtualizationClustergroupByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a virtualizationClustergroup by id
func (h *virtualizationClustergroupHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVirtualizationClustergroupByIDRequest) (*netbox_goV1.GetVirtualizationClustergroupByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of virtualizationClustergroups by custom conditions
func (h *virtualizationClustergroupHandler) List(ctx context.Context, req *netbox_goV1.ListVirtualizationClustergroupRequest) (*netbox_goV1.ListVirtualizationClustergroupReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete virtualizationClustergroup by ids
func (h *virtualizationClustergroupHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVirtualizationClustergroupByIDsRequest) (*netbox_goV1.DeleteVirtualizationClustergroupByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a virtualizationClustergroup by custom condition
func (h *virtualizationClustergroupHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVirtualizationClustergroupByConditionRequest) (*netbox_goV1.GetVirtualizationClustergroupByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get virtualizationClustergroup by ids
func (h *virtualizationClustergroupHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVirtualizationClustergroupByIDsRequest) (*netbox_goV1.ListVirtualizationClustergroupByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of virtualizationClustergroups by last id
func (h *virtualizationClustergroupHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVirtualizationClustergroupByLastIDRequest) (*netbox_goV1.ListVirtualizationClustergroupByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
