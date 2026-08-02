package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimFrontportLogicer = (*dcimFrontportHandler)(nil)

type dcimFrontportHandler struct {
	server netbox_goV1.DcimFrontportServer
}

// NewDcimFrontportHandler create a handler
func NewDcimFrontportHandler() netbox_goV1.DcimFrontportLogicer {
	return &dcimFrontportHandler{
		server: service.NewDcimFrontportServer(),
	}
}

// Create a new dcimFrontport
func (h *dcimFrontportHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimFrontportRequest) (*netbox_goV1.CreateDcimFrontportReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimFrontport by id
func (h *dcimFrontportHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimFrontportByIDRequest) (*netbox_goV1.DeleteDcimFrontportByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimFrontport by id
func (h *dcimFrontportHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimFrontportByIDRequest) (*netbox_goV1.UpdateDcimFrontportByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimFrontport by id
func (h *dcimFrontportHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimFrontportByIDRequest) (*netbox_goV1.GetDcimFrontportByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimFrontports by custom conditions
func (h *dcimFrontportHandler) List(ctx context.Context, req *netbox_goV1.ListDcimFrontportRequest) (*netbox_goV1.ListDcimFrontportReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimFrontport by ids
func (h *dcimFrontportHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimFrontportByIDsRequest) (*netbox_goV1.DeleteDcimFrontportByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimFrontport by custom condition
func (h *dcimFrontportHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimFrontportByConditionRequest) (*netbox_goV1.GetDcimFrontportByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimFrontport by ids
func (h *dcimFrontportHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimFrontportByIDsRequest) (*netbox_goV1.ListDcimFrontportByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimFrontports by last id
func (h *dcimFrontportHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimFrontportByLastIDRequest) (*netbox_goV1.ListDcimFrontportByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
