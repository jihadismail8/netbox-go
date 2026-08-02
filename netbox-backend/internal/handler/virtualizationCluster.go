package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.VirtualizationClusterLogicer = (*virtualizationClusterHandler)(nil)

type virtualizationClusterHandler struct {
	server netbox_goV1.VirtualizationClusterServer
}

// NewVirtualizationClusterHandler create a handler
func NewVirtualizationClusterHandler() netbox_goV1.VirtualizationClusterLogicer {
	return &virtualizationClusterHandler{
		server: service.NewVirtualizationClusterServer(),
	}
}

// Create a new virtualizationCluster
func (h *virtualizationClusterHandler) Create(ctx context.Context, req *netbox_goV1.CreateVirtualizationClusterRequest) (*netbox_goV1.CreateVirtualizationClusterReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a virtualizationCluster by id
func (h *virtualizationClusterHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVirtualizationClusterByIDRequest) (*netbox_goV1.DeleteVirtualizationClusterByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a virtualizationCluster by id
func (h *virtualizationClusterHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVirtualizationClusterByIDRequest) (*netbox_goV1.UpdateVirtualizationClusterByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a virtualizationCluster by id
func (h *virtualizationClusterHandler) GetByID(ctx context.Context, req *netbox_goV1.GetVirtualizationClusterByIDRequest) (*netbox_goV1.GetVirtualizationClusterByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of virtualizationClusters by custom conditions
func (h *virtualizationClusterHandler) List(ctx context.Context, req *netbox_goV1.ListVirtualizationClusterRequest) (*netbox_goV1.ListVirtualizationClusterReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete virtualizationCluster by ids
func (h *virtualizationClusterHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVirtualizationClusterByIDsRequest) (*netbox_goV1.DeleteVirtualizationClusterByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a virtualizationCluster by custom condition
func (h *virtualizationClusterHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetVirtualizationClusterByConditionRequest) (*netbox_goV1.GetVirtualizationClusterByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get virtualizationCluster by ids
func (h *virtualizationClusterHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListVirtualizationClusterByIDsRequest) (*netbox_goV1.ListVirtualizationClusterByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of virtualizationClusters by last id
func (h *virtualizationClusterHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListVirtualizationClusterByLastIDRequest) (*netbox_goV1.ListVirtualizationClusterByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
