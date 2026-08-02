package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasWebhookLogicer = (*extrasWebhookHandler)(nil)

type extrasWebhookHandler struct {
	server netbox_goV1.ExtrasWebhookServer
}

// NewExtrasWebhookHandler create a handler
func NewExtrasWebhookHandler() netbox_goV1.ExtrasWebhookLogicer {
	return &extrasWebhookHandler{
		server: service.NewExtrasWebhookServer(),
	}
}

// Create a new extrasWebhook
func (h *extrasWebhookHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasWebhookRequest) (*netbox_goV1.CreateExtrasWebhookReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasWebhook by id
func (h *extrasWebhookHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasWebhookByIDRequest) (*netbox_goV1.DeleteExtrasWebhookByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasWebhook by id
func (h *extrasWebhookHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasWebhookByIDRequest) (*netbox_goV1.UpdateExtrasWebhookByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasWebhook by id
func (h *extrasWebhookHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasWebhookByIDRequest) (*netbox_goV1.GetExtrasWebhookByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasWebhooks by custom conditions
func (h *extrasWebhookHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasWebhookRequest) (*netbox_goV1.ListExtrasWebhookReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasWebhook by ids
func (h *extrasWebhookHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasWebhookByIDsRequest) (*netbox_goV1.DeleteExtrasWebhookByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasWebhook by custom condition
func (h *extrasWebhookHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasWebhookByConditionRequest) (*netbox_goV1.GetExtrasWebhookByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasWebhook by ids
func (h *extrasWebhookHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasWebhookByIDsRequest) (*netbox_goV1.ListExtrasWebhookByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasWebhooks by last id
func (h *extrasWebhookHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasWebhookByLastIDRequest) (*netbox_goV1.ListExtrasWebhookByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
