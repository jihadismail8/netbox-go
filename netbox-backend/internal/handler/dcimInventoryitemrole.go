package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimInventoryitemroleLogicer = (*dcimInventoryitemroleHandler)(nil)

type dcimInventoryitemroleHandler struct {
	server netbox_goV1.DcimInventoryitemroleServer
}

// NewDcimInventoryitemroleHandler create a handler
func NewDcimInventoryitemroleHandler() netbox_goV1.DcimInventoryitemroleLogicer {
	return &dcimInventoryitemroleHandler{
		server: service.NewDcimInventoryitemroleServer(),
	}
}

// Create a new dcimInventoryitemrole
func (h *dcimInventoryitemroleHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimInventoryitemroleRequest) (*netbox_goV1.CreateDcimInventoryitemroleReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimInventoryitemrole by id
func (h *dcimInventoryitemroleHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimInventoryitemroleByIDRequest) (*netbox_goV1.DeleteDcimInventoryitemroleByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimInventoryitemrole by id
func (h *dcimInventoryitemroleHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimInventoryitemroleByIDRequest) (*netbox_goV1.UpdateDcimInventoryitemroleByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimInventoryitemrole by id
func (h *dcimInventoryitemroleHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimInventoryitemroleByIDRequest) (*netbox_goV1.GetDcimInventoryitemroleByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimInventoryitemroles by custom conditions
func (h *dcimInventoryitemroleHandler) List(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemroleRequest) (*netbox_goV1.ListDcimInventoryitemroleReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimInventoryitemrole by ids
func (h *dcimInventoryitemroleHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimInventoryitemroleByIDsRequest) (*netbox_goV1.DeleteDcimInventoryitemroleByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimInventoryitemrole by custom condition
func (h *dcimInventoryitemroleHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimInventoryitemroleByConditionRequest) (*netbox_goV1.GetDcimInventoryitemroleByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimInventoryitemrole by ids
func (h *dcimInventoryitemroleHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemroleByIDsRequest) (*netbox_goV1.ListDcimInventoryitemroleByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimInventoryitemroles by last id
func (h *dcimInventoryitemroleHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemroleByLastIDRequest) (*netbox_goV1.ListDcimInventoryitemroleByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
