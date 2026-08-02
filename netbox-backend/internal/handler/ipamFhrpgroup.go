package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.IpamFhrpgroupLogicer = (*ipamFhrpgroupHandler)(nil)

type ipamFhrpgroupHandler struct {
	server netbox_goV1.IpamFhrpgroupServer
}

// NewIpamFhrpgroupHandler create a handler
func NewIpamFhrpgroupHandler() netbox_goV1.IpamFhrpgroupLogicer {
	return &ipamFhrpgroupHandler{
		server: service.NewIpamFhrpgroupServer(),
	}
}

// Create a new ipamFhrpgroup
func (h *ipamFhrpgroupHandler) Create(ctx context.Context, req *netbox_goV1.CreateIpamFhrpgroupRequest) (*netbox_goV1.CreateIpamFhrpgroupReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a ipamFhrpgroup by id
func (h *ipamFhrpgroupHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamFhrpgroupByIDRequest) (*netbox_goV1.DeleteIpamFhrpgroupByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a ipamFhrpgroup by id
func (h *ipamFhrpgroupHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamFhrpgroupByIDRequest) (*netbox_goV1.UpdateIpamFhrpgroupByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a ipamFhrpgroup by id
func (h *ipamFhrpgroupHandler) GetByID(ctx context.Context, req *netbox_goV1.GetIpamFhrpgroupByIDRequest) (*netbox_goV1.GetIpamFhrpgroupByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of ipamFhrpgroups by custom conditions
func (h *ipamFhrpgroupHandler) List(ctx context.Context, req *netbox_goV1.ListIpamFhrpgroupRequest) (*netbox_goV1.ListIpamFhrpgroupReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete ipamFhrpgroup by ids
func (h *ipamFhrpgroupHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamFhrpgroupByIDsRequest) (*netbox_goV1.DeleteIpamFhrpgroupByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a ipamFhrpgroup by custom condition
func (h *ipamFhrpgroupHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamFhrpgroupByConditionRequest) (*netbox_goV1.GetIpamFhrpgroupByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get ipamFhrpgroup by ids
func (h *ipamFhrpgroupHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamFhrpgroupByIDsRequest) (*netbox_goV1.ListIpamFhrpgroupByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of ipamFhrpgroups by last id
func (h *ipamFhrpgroupHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamFhrpgroupByLastIDRequest) (*netbox_goV1.ListIpamFhrpgroupByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
