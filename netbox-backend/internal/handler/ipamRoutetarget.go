package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.IpamRoutetargetLogicer = (*ipamRoutetargetHandler)(nil)

type ipamRoutetargetHandler struct {
	server netbox_goV1.IpamRoutetargetServer
}

// NewIpamRoutetargetHandler create a handler
func NewIpamRoutetargetHandler() netbox_goV1.IpamRoutetargetLogicer {
	return &ipamRoutetargetHandler{
		server: service.NewIpamRoutetargetServer(),
	}
}

// Create a new ipamRoutetarget
func (h *ipamRoutetargetHandler) Create(ctx context.Context, req *netbox_goV1.CreateIpamRoutetargetRequest) (*netbox_goV1.CreateIpamRoutetargetReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a ipamRoutetarget by id
func (h *ipamRoutetargetHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamRoutetargetByIDRequest) (*netbox_goV1.DeleteIpamRoutetargetByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a ipamRoutetarget by id
func (h *ipamRoutetargetHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamRoutetargetByIDRequest) (*netbox_goV1.UpdateIpamRoutetargetByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a ipamRoutetarget by id
func (h *ipamRoutetargetHandler) GetByID(ctx context.Context, req *netbox_goV1.GetIpamRoutetargetByIDRequest) (*netbox_goV1.GetIpamRoutetargetByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of ipamRoutetargets by custom conditions
func (h *ipamRoutetargetHandler) List(ctx context.Context, req *netbox_goV1.ListIpamRoutetargetRequest) (*netbox_goV1.ListIpamRoutetargetReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete ipamRoutetarget by ids
func (h *ipamRoutetargetHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamRoutetargetByIDsRequest) (*netbox_goV1.DeleteIpamRoutetargetByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a ipamRoutetarget by custom condition
func (h *ipamRoutetargetHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamRoutetargetByConditionRequest) (*netbox_goV1.GetIpamRoutetargetByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get ipamRoutetarget by ids
func (h *ipamRoutetargetHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamRoutetargetByIDsRequest) (*netbox_goV1.ListIpamRoutetargetByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of ipamRoutetargets by last id
func (h *ipamRoutetargetHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamRoutetargetByLastIDRequest) (*netbox_goV1.ListIpamRoutetargetByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
