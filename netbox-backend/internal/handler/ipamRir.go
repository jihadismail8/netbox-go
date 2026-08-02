package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.IpamRirLogicer = (*ipamRirHandler)(nil)

type ipamRirHandler struct {
	server netbox_goV1.IpamRirServer
}

// NewIpamRirHandler create a handler
func NewIpamRirHandler() netbox_goV1.IpamRirLogicer {
	return &ipamRirHandler{
		server: service.NewIpamRirServer(),
	}
}

// Create a new ipamRir
func (h *ipamRirHandler) Create(ctx context.Context, req *netbox_goV1.CreateIpamRirRequest) (*netbox_goV1.CreateIpamRirReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a ipamRir by id
func (h *ipamRirHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamRirByIDRequest) (*netbox_goV1.DeleteIpamRirByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a ipamRir by id
func (h *ipamRirHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamRirByIDRequest) (*netbox_goV1.UpdateIpamRirByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a ipamRir by id
func (h *ipamRirHandler) GetByID(ctx context.Context, req *netbox_goV1.GetIpamRirByIDRequest) (*netbox_goV1.GetIpamRirByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of ipamRirs by custom conditions
func (h *ipamRirHandler) List(ctx context.Context, req *netbox_goV1.ListIpamRirRequest) (*netbox_goV1.ListIpamRirReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete ipamRir by ids
func (h *ipamRirHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamRirByIDsRequest) (*netbox_goV1.DeleteIpamRirByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a ipamRir by custom condition
func (h *ipamRirHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamRirByConditionRequest) (*netbox_goV1.GetIpamRirByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get ipamRir by ids
func (h *ipamRirHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamRirByIDsRequest) (*netbox_goV1.ListIpamRirByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of ipamRirs by last id
func (h *ipamRirHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamRirByLastIDRequest) (*netbox_goV1.ListIpamRirByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
