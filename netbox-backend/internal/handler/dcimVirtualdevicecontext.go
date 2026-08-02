package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimVirtualdevicecontextLogicer = (*dcimVirtualdevicecontextHandler)(nil)

type dcimVirtualdevicecontextHandler struct {
	server netbox_goV1.DcimVirtualdevicecontextServer
}

// NewDcimVirtualdevicecontextHandler create a handler
func NewDcimVirtualdevicecontextHandler() netbox_goV1.DcimVirtualdevicecontextLogicer {
	return &dcimVirtualdevicecontextHandler{
		server: service.NewDcimVirtualdevicecontextServer(),
	}
}

// Create a new dcimVirtualdevicecontext
func (h *dcimVirtualdevicecontextHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimVirtualdevicecontextRequest) (*netbox_goV1.CreateDcimVirtualdevicecontextReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimVirtualdevicecontext by id
func (h *dcimVirtualdevicecontextHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimVirtualdevicecontextByIDRequest) (*netbox_goV1.DeleteDcimVirtualdevicecontextByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimVirtualdevicecontext by id
func (h *dcimVirtualdevicecontextHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimVirtualdevicecontextByIDRequest) (*netbox_goV1.UpdateDcimVirtualdevicecontextByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimVirtualdevicecontext by id
func (h *dcimVirtualdevicecontextHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimVirtualdevicecontextByIDRequest) (*netbox_goV1.GetDcimVirtualdevicecontextByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimVirtualdevicecontexts by custom conditions
func (h *dcimVirtualdevicecontextHandler) List(ctx context.Context, req *netbox_goV1.ListDcimVirtualdevicecontextRequest) (*netbox_goV1.ListDcimVirtualdevicecontextReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimVirtualdevicecontext by ids
func (h *dcimVirtualdevicecontextHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimVirtualdevicecontextByIDsRequest) (*netbox_goV1.DeleteDcimVirtualdevicecontextByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimVirtualdevicecontext by custom condition
func (h *dcimVirtualdevicecontextHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimVirtualdevicecontextByConditionRequest) (*netbox_goV1.GetDcimVirtualdevicecontextByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimVirtualdevicecontext by ids
func (h *dcimVirtualdevicecontextHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimVirtualdevicecontextByIDsRequest) (*netbox_goV1.ListDcimVirtualdevicecontextByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimVirtualdevicecontexts by last id
func (h *dcimVirtualdevicecontextHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimVirtualdevicecontextByLastIDRequest) (*netbox_goV1.ListDcimVirtualdevicecontextByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
