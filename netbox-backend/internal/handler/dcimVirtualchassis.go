package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimVirtualchassisLogicer = (*dcimVirtualchassisHandler)(nil)

type dcimVirtualchassisHandler struct {
	server netbox_goV1.DcimVirtualchassisServer
}

// NewDcimVirtualchassisHandler create a handler
func NewDcimVirtualchassisHandler() netbox_goV1.DcimVirtualchassisLogicer {
	return &dcimVirtualchassisHandler{
		server: service.NewDcimVirtualchassisServer(),
	}
}

// Create a new dcimVirtualchassis
func (h *dcimVirtualchassisHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimVirtualchassisRequest) (*netbox_goV1.CreateDcimVirtualchassisReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimVirtualchassis by id
func (h *dcimVirtualchassisHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimVirtualchassisByIDRequest) (*netbox_goV1.DeleteDcimVirtualchassisByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimVirtualchassis by id
func (h *dcimVirtualchassisHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimVirtualchassisByIDRequest) (*netbox_goV1.UpdateDcimVirtualchassisByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimVirtualchassis by id
func (h *dcimVirtualchassisHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimVirtualchassisByIDRequest) (*netbox_goV1.GetDcimVirtualchassisByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimVirtualchassiss by custom conditions
func (h *dcimVirtualchassisHandler) List(ctx context.Context, req *netbox_goV1.ListDcimVirtualchassisRequest) (*netbox_goV1.ListDcimVirtualchassisReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimVirtualchassis by ids
func (h *dcimVirtualchassisHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimVirtualchassisByIDsRequest) (*netbox_goV1.DeleteDcimVirtualchassisByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimVirtualchassis by custom condition
func (h *dcimVirtualchassisHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimVirtualchassisByConditionRequest) (*netbox_goV1.GetDcimVirtualchassisByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimVirtualchassis by ids
func (h *dcimVirtualchassisHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimVirtualchassisByIDsRequest) (*netbox_goV1.ListDcimVirtualchassisByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimVirtualchassiss by last id
func (h *dcimVirtualchassisHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimVirtualchassisByLastIDRequest) (*netbox_goV1.ListDcimVirtualchassisByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
