package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.IpamVlangroupLogicer = (*ipamVlangroupHandler)(nil)

type ipamVlangroupHandler struct {
	server netbox_goV1.IpamVlangroupServer
}

// NewIpamVlangroupHandler create a handler
func NewIpamVlangroupHandler() netbox_goV1.IpamVlangroupLogicer {
	return &ipamVlangroupHandler{
		server: service.NewIpamVlangroupServer(),
	}
}

// Create a new ipamVlangroup
func (h *ipamVlangroupHandler) Create(ctx context.Context, req *netbox_goV1.CreateIpamVlangroupRequest) (*netbox_goV1.CreateIpamVlangroupReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a ipamVlangroup by id
func (h *ipamVlangroupHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamVlangroupByIDRequest) (*netbox_goV1.DeleteIpamVlangroupByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a ipamVlangroup by id
func (h *ipamVlangroupHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamVlangroupByIDRequest) (*netbox_goV1.UpdateIpamVlangroupByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a ipamVlangroup by id
func (h *ipamVlangroupHandler) GetByID(ctx context.Context, req *netbox_goV1.GetIpamVlangroupByIDRequest) (*netbox_goV1.GetIpamVlangroupByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of ipamVlangroups by custom conditions
func (h *ipamVlangroupHandler) List(ctx context.Context, req *netbox_goV1.ListIpamVlangroupRequest) (*netbox_goV1.ListIpamVlangroupReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete ipamVlangroup by ids
func (h *ipamVlangroupHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamVlangroupByIDsRequest) (*netbox_goV1.DeleteIpamVlangroupByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a ipamVlangroup by custom condition
func (h *ipamVlangroupHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamVlangroupByConditionRequest) (*netbox_goV1.GetIpamVlangroupByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get ipamVlangroup by ids
func (h *ipamVlangroupHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamVlangroupByIDsRequest) (*netbox_goV1.ListIpamVlangroupByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of ipamVlangroups by last id
func (h *ipamVlangroupHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamVlangroupByLastIDRequest) (*netbox_goV1.ListIpamVlangroupByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
