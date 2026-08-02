package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimConsoleserverportLogicer = (*dcimConsoleserverportHandler)(nil)

type dcimConsoleserverportHandler struct {
	server netbox_goV1.DcimConsoleserverportServer
}

// NewDcimConsoleserverportHandler create a handler
func NewDcimConsoleserverportHandler() netbox_goV1.DcimConsoleserverportLogicer {
	return &dcimConsoleserverportHandler{
		server: service.NewDcimConsoleserverportServer(),
	}
}

// Create a new dcimConsoleserverport
func (h *dcimConsoleserverportHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimConsoleserverportRequest) (*netbox_goV1.CreateDcimConsoleserverportReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimConsoleserverport by id
func (h *dcimConsoleserverportHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimConsoleserverportByIDRequest) (*netbox_goV1.DeleteDcimConsoleserverportByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimConsoleserverport by id
func (h *dcimConsoleserverportHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimConsoleserverportByIDRequest) (*netbox_goV1.UpdateDcimConsoleserverportByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimConsoleserverport by id
func (h *dcimConsoleserverportHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimConsoleserverportByIDRequest) (*netbox_goV1.GetDcimConsoleserverportByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimConsoleserverports by custom conditions
func (h *dcimConsoleserverportHandler) List(ctx context.Context, req *netbox_goV1.ListDcimConsoleserverportRequest) (*netbox_goV1.ListDcimConsoleserverportReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimConsoleserverport by ids
func (h *dcimConsoleserverportHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimConsoleserverportByIDsRequest) (*netbox_goV1.DeleteDcimConsoleserverportByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimConsoleserverport by custom condition
func (h *dcimConsoleserverportHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimConsoleserverportByConditionRequest) (*netbox_goV1.GetDcimConsoleserverportByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimConsoleserverport by ids
func (h *dcimConsoleserverportHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimConsoleserverportByIDsRequest) (*netbox_goV1.ListDcimConsoleserverportByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimConsoleserverports by last id
func (h *dcimConsoleserverportHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimConsoleserverportByLastIDRequest) (*netbox_goV1.ListDcimConsoleserverportByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
