package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimRearportLogicer = (*dcimRearportHandler)(nil)

type dcimRearportHandler struct {
	server netbox_goV1.DcimRearportServer
}

// NewDcimRearportHandler create a handler
func NewDcimRearportHandler() netbox_goV1.DcimRearportLogicer {
	return &dcimRearportHandler{
		server: service.NewDcimRearportServer(),
	}
}

// Create a new dcimRearport
func (h *dcimRearportHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimRearportRequest) (*netbox_goV1.CreateDcimRearportReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimRearport by id
func (h *dcimRearportHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimRearportByIDRequest) (*netbox_goV1.DeleteDcimRearportByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimRearport by id
func (h *dcimRearportHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimRearportByIDRequest) (*netbox_goV1.UpdateDcimRearportByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimRearport by id
func (h *dcimRearportHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimRearportByIDRequest) (*netbox_goV1.GetDcimRearportByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimRearports by custom conditions
func (h *dcimRearportHandler) List(ctx context.Context, req *netbox_goV1.ListDcimRearportRequest) (*netbox_goV1.ListDcimRearportReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimRearport by ids
func (h *dcimRearportHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimRearportByIDsRequest) (*netbox_goV1.DeleteDcimRearportByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimRearport by custom condition
func (h *dcimRearportHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimRearportByConditionRequest) (*netbox_goV1.GetDcimRearportByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimRearport by ids
func (h *dcimRearportHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimRearportByIDsRequest) (*netbox_goV1.ListDcimRearportByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimRearports by last id
func (h *dcimRearportHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimRearportByLastIDRequest) (*netbox_goV1.ListDcimRearportByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
