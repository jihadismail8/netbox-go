package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimDevicebayLogicer = (*dcimDevicebayHandler)(nil)

type dcimDevicebayHandler struct {
	server netbox_goV1.DcimDevicebayServer
}

// NewDcimDevicebayHandler create a handler
func NewDcimDevicebayHandler() netbox_goV1.DcimDevicebayLogicer {
	return &dcimDevicebayHandler{
		server: service.NewDcimDevicebayServer(),
	}
}

// Create a new dcimDevicebay
func (h *dcimDevicebayHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimDevicebayRequest) (*netbox_goV1.CreateDcimDevicebayReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimDevicebay by id
func (h *dcimDevicebayHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimDevicebayByIDRequest) (*netbox_goV1.DeleteDcimDevicebayByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimDevicebay by id
func (h *dcimDevicebayHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimDevicebayByIDRequest) (*netbox_goV1.UpdateDcimDevicebayByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimDevicebay by id
func (h *dcimDevicebayHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimDevicebayByIDRequest) (*netbox_goV1.GetDcimDevicebayByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimDevicebays by custom conditions
func (h *dcimDevicebayHandler) List(ctx context.Context, req *netbox_goV1.ListDcimDevicebayRequest) (*netbox_goV1.ListDcimDevicebayReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimDevicebay by ids
func (h *dcimDevicebayHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimDevicebayByIDsRequest) (*netbox_goV1.DeleteDcimDevicebayByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimDevicebay by custom condition
func (h *dcimDevicebayHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimDevicebayByConditionRequest) (*netbox_goV1.GetDcimDevicebayByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimDevicebay by ids
func (h *dcimDevicebayHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimDevicebayByIDsRequest) (*netbox_goV1.ListDcimDevicebayByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimDevicebays by last id
func (h *dcimDevicebayHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimDevicebayByLastIDRequest) (*netbox_goV1.ListDcimDevicebayByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
