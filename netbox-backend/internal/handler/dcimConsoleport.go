package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimConsoleportLogicer = (*dcimConsoleportHandler)(nil)

type dcimConsoleportHandler struct {
	server netbox_goV1.DcimConsoleportServer
}

// NewDcimConsoleportHandler create a handler
func NewDcimConsoleportHandler() netbox_goV1.DcimConsoleportLogicer {
	return &dcimConsoleportHandler{
		server: service.NewDcimConsoleportServer(),
	}
}

// Create a new dcimConsoleport
func (h *dcimConsoleportHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimConsoleportRequest) (*netbox_goV1.CreateDcimConsoleportReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimConsoleport by id
func (h *dcimConsoleportHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimConsoleportByIDRequest) (*netbox_goV1.DeleteDcimConsoleportByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimConsoleport by id
func (h *dcimConsoleportHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimConsoleportByIDRequest) (*netbox_goV1.UpdateDcimConsoleportByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimConsoleport by id
func (h *dcimConsoleportHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimConsoleportByIDRequest) (*netbox_goV1.GetDcimConsoleportByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimConsoleports by custom conditions
func (h *dcimConsoleportHandler) List(ctx context.Context, req *netbox_goV1.ListDcimConsoleportRequest) (*netbox_goV1.ListDcimConsoleportReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimConsoleport by ids
func (h *dcimConsoleportHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimConsoleportByIDsRequest) (*netbox_goV1.DeleteDcimConsoleportByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimConsoleport by custom condition
func (h *dcimConsoleportHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimConsoleportByConditionRequest) (*netbox_goV1.GetDcimConsoleportByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimConsoleport by ids
func (h *dcimConsoleportHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimConsoleportByIDsRequest) (*netbox_goV1.ListDcimConsoleportByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimConsoleports by last id
func (h *dcimConsoleportHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimConsoleportByLastIDRequest) (*netbox_goV1.ListDcimConsoleportByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
