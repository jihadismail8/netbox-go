package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasImageattachmentLogicer = (*extrasImageattachmentHandler)(nil)

type extrasImageattachmentHandler struct {
	server netbox_goV1.ExtrasImageattachmentServer
}

// NewExtrasImageattachmentHandler create a handler
func NewExtrasImageattachmentHandler() netbox_goV1.ExtrasImageattachmentLogicer {
	return &extrasImageattachmentHandler{
		server: service.NewExtrasImageattachmentServer(),
	}
}

// Create a new extrasImageattachment
func (h *extrasImageattachmentHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasImageattachmentRequest) (*netbox_goV1.CreateExtrasImageattachmentReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasImageattachment by id
func (h *extrasImageattachmentHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasImageattachmentByIDRequest) (*netbox_goV1.DeleteExtrasImageattachmentByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasImageattachment by id
func (h *extrasImageattachmentHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasImageattachmentByIDRequest) (*netbox_goV1.UpdateExtrasImageattachmentByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasImageattachment by id
func (h *extrasImageattachmentHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasImageattachmentByIDRequest) (*netbox_goV1.GetExtrasImageattachmentByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasImageattachments by custom conditions
func (h *extrasImageattachmentHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasImageattachmentRequest) (*netbox_goV1.ListExtrasImageattachmentReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasImageattachment by ids
func (h *extrasImageattachmentHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasImageattachmentByIDsRequest) (*netbox_goV1.DeleteExtrasImageattachmentByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasImageattachment by custom condition
func (h *extrasImageattachmentHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasImageattachmentByConditionRequest) (*netbox_goV1.GetExtrasImageattachmentByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasImageattachment by ids
func (h *extrasImageattachmentHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasImageattachmentByIDsRequest) (*netbox_goV1.ListExtrasImageattachmentByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasImageattachments by last id
func (h *extrasImageattachmentHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasImageattachmentByLastIDRequest) (*netbox_goV1.ListExtrasImageattachmentByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
