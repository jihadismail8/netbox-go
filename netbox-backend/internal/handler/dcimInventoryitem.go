package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimInventoryitemLogicer = (*dcimInventoryitemHandler)(nil)

type dcimInventoryitemHandler struct {
	server netbox_goV1.DcimInventoryitemServer
}

// NewDcimInventoryitemHandler create a handler
func NewDcimInventoryitemHandler() netbox_goV1.DcimInventoryitemLogicer {
	return &dcimInventoryitemHandler{
		server: service.NewDcimInventoryitemServer(),
	}
}

// Create a new dcimInventoryitem
func (h *dcimInventoryitemHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimInventoryitemRequest) (*netbox_goV1.CreateDcimInventoryitemReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimInventoryitem by id
func (h *dcimInventoryitemHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimInventoryitemByIDRequest) (*netbox_goV1.DeleteDcimInventoryitemByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimInventoryitem by id
func (h *dcimInventoryitemHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimInventoryitemByIDRequest) (*netbox_goV1.UpdateDcimInventoryitemByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimInventoryitem by id
func (h *dcimInventoryitemHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimInventoryitemByIDRequest) (*netbox_goV1.GetDcimInventoryitemByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimInventoryitems by custom conditions
func (h *dcimInventoryitemHandler) List(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemRequest) (*netbox_goV1.ListDcimInventoryitemReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimInventoryitem by ids
func (h *dcimInventoryitemHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimInventoryitemByIDsRequest) (*netbox_goV1.DeleteDcimInventoryitemByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimInventoryitem by custom condition
func (h *dcimInventoryitemHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimInventoryitemByConditionRequest) (*netbox_goV1.GetDcimInventoryitemByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimInventoryitem by ids
func (h *dcimInventoryitemHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemByIDsRequest) (*netbox_goV1.ListDcimInventoryitemByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimInventoryitems by last id
func (h *dcimInventoryitemHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemByLastIDRequest) (*netbox_goV1.ListDcimInventoryitemByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
