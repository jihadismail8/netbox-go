package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.IpamRoleLogicer = (*ipamRoleHandler)(nil)

type ipamRoleHandler struct {
	server netbox_goV1.IpamRoleServer
}

// NewIpamRoleHandler create a handler
func NewIpamRoleHandler() netbox_goV1.IpamRoleLogicer {
	return &ipamRoleHandler{
		server: service.NewIpamRoleServer(),
	}
}

// Create a new ipamRole
func (h *ipamRoleHandler) Create(ctx context.Context, req *netbox_goV1.CreateIpamRoleRequest) (*netbox_goV1.CreateIpamRoleReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a ipamRole by id
func (h *ipamRoleHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamRoleByIDRequest) (*netbox_goV1.DeleteIpamRoleByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a ipamRole by id
func (h *ipamRoleHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamRoleByIDRequest) (*netbox_goV1.UpdateIpamRoleByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a ipamRole by id
func (h *ipamRoleHandler) GetByID(ctx context.Context, req *netbox_goV1.GetIpamRoleByIDRequest) (*netbox_goV1.GetIpamRoleByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of ipamRoles by custom conditions
func (h *ipamRoleHandler) List(ctx context.Context, req *netbox_goV1.ListIpamRoleRequest) (*netbox_goV1.ListIpamRoleReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete ipamRole by ids
func (h *ipamRoleHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamRoleByIDsRequest) (*netbox_goV1.DeleteIpamRoleByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a ipamRole by custom condition
func (h *ipamRoleHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamRoleByConditionRequest) (*netbox_goV1.GetIpamRoleByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get ipamRole by ids
func (h *ipamRoleHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamRoleByIDsRequest) (*netbox_goV1.ListIpamRoleByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of ipamRoles by last id
func (h *ipamRoleHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamRoleByLastIDRequest) (*netbox_goV1.ListIpamRoleByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
