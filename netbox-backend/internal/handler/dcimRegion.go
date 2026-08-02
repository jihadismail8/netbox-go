package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimRegionLogicer = (*dcimRegionHandler)(nil)

type dcimRegionHandler struct {
	server netbox_goV1.DcimRegionServer
}

// NewDcimRegionHandler create a handler
func NewDcimRegionHandler() netbox_goV1.DcimRegionLogicer {
	return &dcimRegionHandler{
		server: service.NewDcimRegionServer(),
	}
}

// Create a new dcimRegion
func (h *dcimRegionHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimRegionRequest) (*netbox_goV1.CreateDcimRegionReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimRegion by id
func (h *dcimRegionHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimRegionByIDRequest) (*netbox_goV1.DeleteDcimRegionByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimRegion by id
func (h *dcimRegionHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimRegionByIDRequest) (*netbox_goV1.UpdateDcimRegionByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimRegion by id
func (h *dcimRegionHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimRegionByIDRequest) (*netbox_goV1.GetDcimRegionByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimRegions by custom conditions
func (h *dcimRegionHandler) List(ctx context.Context, req *netbox_goV1.ListDcimRegionRequest) (*netbox_goV1.ListDcimRegionReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimRegion by ids
func (h *dcimRegionHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimRegionByIDsRequest) (*netbox_goV1.DeleteDcimRegionByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimRegion by custom condition
func (h *dcimRegionHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimRegionByConditionRequest) (*netbox_goV1.GetDcimRegionByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimRegion by ids
func (h *dcimRegionHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimRegionByIDsRequest) (*netbox_goV1.ListDcimRegionByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimRegions by last id
func (h *dcimRegionHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimRegionByLastIDRequest) (*netbox_goV1.ListDcimRegionByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
