package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.IpamServiceLogicer = (*ipamServiceHandler)(nil)

type ipamServiceHandler struct {
	server netbox_goV1.IpamServiceServer
}

// NewIpamServiceHandler create a handler
func NewIpamServiceHandler() netbox_goV1.IpamServiceLogicer {
	return &ipamServiceHandler{
		server: service.NewIpamServiceServer(),
	}
}

// Create a new ipamService
func (h *ipamServiceHandler) Create(ctx context.Context, req *netbox_goV1.CreateIpamServiceRequest) (*netbox_goV1.CreateIpamServiceReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a ipamService by id
func (h *ipamServiceHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamServiceByIDRequest) (*netbox_goV1.DeleteIpamServiceByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a ipamService by id
func (h *ipamServiceHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamServiceByIDRequest) (*netbox_goV1.UpdateIpamServiceByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a ipamService by id
func (h *ipamServiceHandler) GetByID(ctx context.Context, req *netbox_goV1.GetIpamServiceByIDRequest) (*netbox_goV1.GetIpamServiceByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of ipamServices by custom conditions
func (h *ipamServiceHandler) List(ctx context.Context, req *netbox_goV1.ListIpamServiceRequest) (*netbox_goV1.ListIpamServiceReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete ipamService by ids
func (h *ipamServiceHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamServiceByIDsRequest) (*netbox_goV1.DeleteIpamServiceByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a ipamService by custom condition
func (h *ipamServiceHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamServiceByConditionRequest) (*netbox_goV1.GetIpamServiceByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get ipamService by ids
func (h *ipamServiceHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamServiceByIDsRequest) (*netbox_goV1.ListIpamServiceByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of ipamServices by last id
func (h *ipamServiceHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamServiceByLastIDRequest) (*netbox_goV1.ListIpamServiceByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
