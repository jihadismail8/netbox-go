package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimMacaddressLogicer = (*dcimMacaddressHandler)(nil)

type dcimMacaddressHandler struct {
	server netbox_goV1.DcimMacaddressServer
}

// NewDcimMacaddressHandler create a handler
func NewDcimMacaddressHandler() netbox_goV1.DcimMacaddressLogicer {
	return &dcimMacaddressHandler{
		server: service.NewDcimMacaddressServer(),
	}
}

// Create a new dcimMacaddress
func (h *dcimMacaddressHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimMacaddressRequest) (*netbox_goV1.CreateDcimMacaddressReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimMacaddress by id
func (h *dcimMacaddressHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimMacaddressByIDRequest) (*netbox_goV1.DeleteDcimMacaddressByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimMacaddress by id
func (h *dcimMacaddressHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimMacaddressByIDRequest) (*netbox_goV1.UpdateDcimMacaddressByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimMacaddress by id
func (h *dcimMacaddressHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimMacaddressByIDRequest) (*netbox_goV1.GetDcimMacaddressByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimMacaddresss by custom conditions
func (h *dcimMacaddressHandler) List(ctx context.Context, req *netbox_goV1.ListDcimMacaddressRequest) (*netbox_goV1.ListDcimMacaddressReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimMacaddress by ids
func (h *dcimMacaddressHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimMacaddressByIDsRequest) (*netbox_goV1.DeleteDcimMacaddressByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimMacaddress by custom condition
func (h *dcimMacaddressHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimMacaddressByConditionRequest) (*netbox_goV1.GetDcimMacaddressByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimMacaddress by ids
func (h *dcimMacaddressHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimMacaddressByIDsRequest) (*netbox_goV1.ListDcimMacaddressByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimMacaddresss by last id
func (h *dcimMacaddressHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimMacaddressByLastIDRequest) (*netbox_goV1.ListDcimMacaddressByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
