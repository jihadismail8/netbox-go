package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.IpamFhrpgroupassignmentLogicer = (*ipamFhrpgroupassignmentHandler)(nil)

type ipamFhrpgroupassignmentHandler struct {
	server netbox_goV1.IpamFhrpgroupassignmentServer
}

// NewIpamFhrpgroupassignmentHandler create a handler
func NewIpamFhrpgroupassignmentHandler() netbox_goV1.IpamFhrpgroupassignmentLogicer {
	return &ipamFhrpgroupassignmentHandler{
		server: service.NewIpamFhrpgroupassignmentServer(),
	}
}

// Create a new ipamFhrpgroupassignment
func (h *ipamFhrpgroupassignmentHandler) Create(ctx context.Context, req *netbox_goV1.CreateIpamFhrpgroupassignmentRequest) (*netbox_goV1.CreateIpamFhrpgroupassignmentReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a ipamFhrpgroupassignment by id
func (h *ipamFhrpgroupassignmentHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamFhrpgroupassignmentByIDRequest) (*netbox_goV1.DeleteIpamFhrpgroupassignmentByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a ipamFhrpgroupassignment by id
func (h *ipamFhrpgroupassignmentHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamFhrpgroupassignmentByIDRequest) (*netbox_goV1.UpdateIpamFhrpgroupassignmentByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a ipamFhrpgroupassignment by id
func (h *ipamFhrpgroupassignmentHandler) GetByID(ctx context.Context, req *netbox_goV1.GetIpamFhrpgroupassignmentByIDRequest) (*netbox_goV1.GetIpamFhrpgroupassignmentByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of ipamFhrpgroupassignments by custom conditions
func (h *ipamFhrpgroupassignmentHandler) List(ctx context.Context, req *netbox_goV1.ListIpamFhrpgroupassignmentRequest) (*netbox_goV1.ListIpamFhrpgroupassignmentReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete ipamFhrpgroupassignment by ids
func (h *ipamFhrpgroupassignmentHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamFhrpgroupassignmentByIDsRequest) (*netbox_goV1.DeleteIpamFhrpgroupassignmentByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a ipamFhrpgroupassignment by custom condition
func (h *ipamFhrpgroupassignmentHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamFhrpgroupassignmentByConditionRequest) (*netbox_goV1.GetIpamFhrpgroupassignmentByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get ipamFhrpgroupassignment by ids
func (h *ipamFhrpgroupassignmentHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamFhrpgroupassignmentByIDsRequest) (*netbox_goV1.ListIpamFhrpgroupassignmentByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of ipamFhrpgroupassignments by last id
func (h *ipamFhrpgroupassignmentHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamFhrpgroupassignmentByLastIDRequest) (*netbox_goV1.ListIpamFhrpgroupassignmentByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
