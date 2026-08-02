package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimModulebayLogicer = (*dcimModulebayHandler)(nil)

type dcimModulebayHandler struct {
	server netbox_goV1.DcimModulebayServer
}

// NewDcimModulebayHandler create a handler
func NewDcimModulebayHandler() netbox_goV1.DcimModulebayLogicer {
	return &dcimModulebayHandler{
		server: service.NewDcimModulebayServer(),
	}
}

// Create a new dcimModulebay
func (h *dcimModulebayHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimModulebayRequest) (*netbox_goV1.CreateDcimModulebayReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimModulebay by id
func (h *dcimModulebayHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimModulebayByIDRequest) (*netbox_goV1.DeleteDcimModulebayByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimModulebay by id
func (h *dcimModulebayHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimModulebayByIDRequest) (*netbox_goV1.UpdateDcimModulebayByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimModulebay by id
func (h *dcimModulebayHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimModulebayByIDRequest) (*netbox_goV1.GetDcimModulebayByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimModulebays by custom conditions
func (h *dcimModulebayHandler) List(ctx context.Context, req *netbox_goV1.ListDcimModulebayRequest) (*netbox_goV1.ListDcimModulebayReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimModulebay by ids
func (h *dcimModulebayHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimModulebayByIDsRequest) (*netbox_goV1.DeleteDcimModulebayByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimModulebay by custom condition
func (h *dcimModulebayHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimModulebayByConditionRequest) (*netbox_goV1.GetDcimModulebayByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimModulebay by ids
func (h *dcimModulebayHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimModulebayByIDsRequest) (*netbox_goV1.ListDcimModulebayByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimModulebays by last id
func (h *dcimModulebayHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimModulebayByLastIDRequest) (*netbox_goV1.ListDcimModulebayByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
