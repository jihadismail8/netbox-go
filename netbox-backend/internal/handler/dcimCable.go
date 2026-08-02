package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimCableLogicer = (*dcimCableHandler)(nil)

type dcimCableHandler struct {
	server netbox_goV1.DcimCableServer
}

// NewDcimCableHandler create a handler
func NewDcimCableHandler() netbox_goV1.DcimCableLogicer {
	return &dcimCableHandler{
		server: service.NewDcimCableServer(),
	}
}

// Create a new dcimCable
func (h *dcimCableHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimCableRequest) (*netbox_goV1.CreateDcimCableReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimCable by id
func (h *dcimCableHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimCableByIDRequest) (*netbox_goV1.DeleteDcimCableByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimCable by id
func (h *dcimCableHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimCableByIDRequest) (*netbox_goV1.UpdateDcimCableByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimCable by id
func (h *dcimCableHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimCableByIDRequest) (*netbox_goV1.GetDcimCableByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimCables by custom conditions
func (h *dcimCableHandler) List(ctx context.Context, req *netbox_goV1.ListDcimCableRequest) (*netbox_goV1.ListDcimCableReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimCable by ids
func (h *dcimCableHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimCableByIDsRequest) (*netbox_goV1.DeleteDcimCableByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimCable by custom condition
func (h *dcimCableHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimCableByConditionRequest) (*netbox_goV1.GetDcimCableByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimCable by ids
func (h *dcimCableHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimCableByIDsRequest) (*netbox_goV1.ListDcimCableByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimCables by last id
func (h *dcimCableHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimCableByLastIDRequest) (*netbox_goV1.ListDcimCableByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
