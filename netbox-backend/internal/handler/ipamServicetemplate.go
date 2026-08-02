package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.IpamServicetemplateLogicer = (*ipamServicetemplateHandler)(nil)

type ipamServicetemplateHandler struct {
	server netbox_goV1.IpamServicetemplateServer
}

// NewIpamServicetemplateHandler create a handler
func NewIpamServicetemplateHandler() netbox_goV1.IpamServicetemplateLogicer {
	return &ipamServicetemplateHandler{
		server: service.NewIpamServicetemplateServer(),
	}
}

// Create a new ipamServicetemplate
func (h *ipamServicetemplateHandler) Create(ctx context.Context, req *netbox_goV1.CreateIpamServicetemplateRequest) (*netbox_goV1.CreateIpamServicetemplateReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a ipamServicetemplate by id
func (h *ipamServicetemplateHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamServicetemplateByIDRequest) (*netbox_goV1.DeleteIpamServicetemplateByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a ipamServicetemplate by id
func (h *ipamServicetemplateHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamServicetemplateByIDRequest) (*netbox_goV1.UpdateIpamServicetemplateByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a ipamServicetemplate by id
func (h *ipamServicetemplateHandler) GetByID(ctx context.Context, req *netbox_goV1.GetIpamServicetemplateByIDRequest) (*netbox_goV1.GetIpamServicetemplateByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of ipamServicetemplates by custom conditions
func (h *ipamServicetemplateHandler) List(ctx context.Context, req *netbox_goV1.ListIpamServicetemplateRequest) (*netbox_goV1.ListIpamServicetemplateReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete ipamServicetemplate by ids
func (h *ipamServicetemplateHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamServicetemplateByIDsRequest) (*netbox_goV1.DeleteIpamServicetemplateByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a ipamServicetemplate by custom condition
func (h *ipamServicetemplateHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamServicetemplateByConditionRequest) (*netbox_goV1.GetIpamServicetemplateByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get ipamServicetemplate by ids
func (h *ipamServicetemplateHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamServicetemplateByIDsRequest) (*netbox_goV1.ListIpamServicetemplateByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of ipamServicetemplates by last id
func (h *ipamServicetemplateHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamServicetemplateByLastIDRequest) (*netbox_goV1.ListIpamServicetemplateByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
