package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CoreObjectchangeLogicer = (*coreObjectchangeHandler)(nil)

type coreObjectchangeHandler struct {
	server netbox_goV1.CoreObjectchangeServer
}

// NewCoreObjectchangeHandler create a handler
func NewCoreObjectchangeHandler() netbox_goV1.CoreObjectchangeLogicer {
	return &coreObjectchangeHandler{
		server: service.NewCoreObjectchangeServer(),
	}
}

// Create a new coreObjectchange
func (h *coreObjectchangeHandler) Create(ctx context.Context, req *netbox_goV1.CreateCoreObjectchangeRequest) (*netbox_goV1.CreateCoreObjectchangeReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a coreObjectchange by id
func (h *coreObjectchangeHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCoreObjectchangeByIDRequest) (*netbox_goV1.DeleteCoreObjectchangeByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a coreObjectchange by id
func (h *coreObjectchangeHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCoreObjectchangeByIDRequest) (*netbox_goV1.UpdateCoreObjectchangeByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a coreObjectchange by id
func (h *coreObjectchangeHandler) GetByID(ctx context.Context, req *netbox_goV1.GetCoreObjectchangeByIDRequest) (*netbox_goV1.GetCoreObjectchangeByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of coreObjectchanges by custom conditions
func (h *coreObjectchangeHandler) List(ctx context.Context, req *netbox_goV1.ListCoreObjectchangeRequest) (*netbox_goV1.ListCoreObjectchangeReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete coreObjectchange by ids
func (h *coreObjectchangeHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCoreObjectchangeByIDsRequest) (*netbox_goV1.DeleteCoreObjectchangeByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a coreObjectchange by custom condition
func (h *coreObjectchangeHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCoreObjectchangeByConditionRequest) (*netbox_goV1.GetCoreObjectchangeByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get coreObjectchange by ids
func (h *coreObjectchangeHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListCoreObjectchangeByIDsRequest) (*netbox_goV1.ListCoreObjectchangeByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of coreObjectchanges by last id
func (h *coreObjectchangeHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListCoreObjectchangeByLastIDRequest) (*netbox_goV1.ListCoreObjectchangeByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
