package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.IpamVlantranslationruleLogicer = (*ipamVlantranslationruleHandler)(nil)

type ipamVlantranslationruleHandler struct {
	server netbox_goV1.IpamVlantranslationruleServer
}

// NewIpamVlantranslationruleHandler create a handler
func NewIpamVlantranslationruleHandler() netbox_goV1.IpamVlantranslationruleLogicer {
	return &ipamVlantranslationruleHandler{
		server: service.NewIpamVlantranslationruleServer(),
	}
}

// Create a new ipamVlantranslationrule
func (h *ipamVlantranslationruleHandler) Create(ctx context.Context, req *netbox_goV1.CreateIpamVlantranslationruleRequest) (*netbox_goV1.CreateIpamVlantranslationruleReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a ipamVlantranslationrule by id
func (h *ipamVlantranslationruleHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamVlantranslationruleByIDRequest) (*netbox_goV1.DeleteIpamVlantranslationruleByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a ipamVlantranslationrule by id
func (h *ipamVlantranslationruleHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamVlantranslationruleByIDRequest) (*netbox_goV1.UpdateIpamVlantranslationruleByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a ipamVlantranslationrule by id
func (h *ipamVlantranslationruleHandler) GetByID(ctx context.Context, req *netbox_goV1.GetIpamVlantranslationruleByIDRequest) (*netbox_goV1.GetIpamVlantranslationruleByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of ipamVlantranslationrules by custom conditions
func (h *ipamVlantranslationruleHandler) List(ctx context.Context, req *netbox_goV1.ListIpamVlantranslationruleRequest) (*netbox_goV1.ListIpamVlantranslationruleReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete ipamVlantranslationrule by ids
func (h *ipamVlantranslationruleHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamVlantranslationruleByIDsRequest) (*netbox_goV1.DeleteIpamVlantranslationruleByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a ipamVlantranslationrule by custom condition
func (h *ipamVlantranslationruleHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamVlantranslationruleByConditionRequest) (*netbox_goV1.GetIpamVlantranslationruleByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get ipamVlantranslationrule by ids
func (h *ipamVlantranslationruleHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamVlantranslationruleByIDsRequest) (*netbox_goV1.ListIpamVlantranslationruleByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of ipamVlantranslationrules by last id
func (h *ipamVlantranslationruleHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamVlantranslationruleByLastIDRequest) (*netbox_goV1.ListIpamVlantranslationruleByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
