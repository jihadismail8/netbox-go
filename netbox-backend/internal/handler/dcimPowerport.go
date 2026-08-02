package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimPowerportLogicer = (*dcimPowerportHandler)(nil)

type dcimPowerportHandler struct {
	server netbox_goV1.DcimPowerportServer
}

// NewDcimPowerportHandler create a handler
func NewDcimPowerportHandler() netbox_goV1.DcimPowerportLogicer {
	return &dcimPowerportHandler{
		server: service.NewDcimPowerportServer(),
	}
}

// Create a new dcimPowerport
func (h *dcimPowerportHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimPowerportRequest) (*netbox_goV1.CreateDcimPowerportReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimPowerport by id
func (h *dcimPowerportHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimPowerportByIDRequest) (*netbox_goV1.DeleteDcimPowerportByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimPowerport by id
func (h *dcimPowerportHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimPowerportByIDRequest) (*netbox_goV1.UpdateDcimPowerportByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimPowerport by id
func (h *dcimPowerportHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimPowerportByIDRequest) (*netbox_goV1.GetDcimPowerportByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimPowerports by custom conditions
func (h *dcimPowerportHandler) List(ctx context.Context, req *netbox_goV1.ListDcimPowerportRequest) (*netbox_goV1.ListDcimPowerportReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimPowerport by ids
func (h *dcimPowerportHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimPowerportByIDsRequest) (*netbox_goV1.DeleteDcimPowerportByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimPowerport by custom condition
func (h *dcimPowerportHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimPowerportByConditionRequest) (*netbox_goV1.GetDcimPowerportByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimPowerport by ids
func (h *dcimPowerportHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimPowerportByIDsRequest) (*netbox_goV1.ListDcimPowerportByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimPowerports by last id
func (h *dcimPowerportHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimPowerportByLastIDRequest) (*netbox_goV1.ListDcimPowerportByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
