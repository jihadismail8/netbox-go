package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasNotificationLogicer = (*extrasNotificationHandler)(nil)

type extrasNotificationHandler struct {
	server netbox_goV1.ExtrasNotificationServer
}

// NewExtrasNotificationHandler create a handler
func NewExtrasNotificationHandler() netbox_goV1.ExtrasNotificationLogicer {
	return &extrasNotificationHandler{
		server: service.NewExtrasNotificationServer(),
	}
}

// Create a new extrasNotification
func (h *extrasNotificationHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasNotificationRequest) (*netbox_goV1.CreateExtrasNotificationReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasNotification by id
func (h *extrasNotificationHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasNotificationByIDRequest) (*netbox_goV1.DeleteExtrasNotificationByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasNotification by id
func (h *extrasNotificationHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasNotificationByIDRequest) (*netbox_goV1.UpdateExtrasNotificationByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasNotification by id
func (h *extrasNotificationHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasNotificationByIDRequest) (*netbox_goV1.GetExtrasNotificationByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasNotifications by custom conditions
func (h *extrasNotificationHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasNotificationRequest) (*netbox_goV1.ListExtrasNotificationReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasNotification by ids
func (h *extrasNotificationHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasNotificationByIDsRequest) (*netbox_goV1.DeleteExtrasNotificationByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasNotification by custom condition
func (h *extrasNotificationHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasNotificationByConditionRequest) (*netbox_goV1.GetExtrasNotificationByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasNotification by ids
func (h *extrasNotificationHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasNotificationByIDsRequest) (*netbox_goV1.ListExtrasNotificationByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasNotifications by last id
func (h *extrasNotificationHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasNotificationByLastIDRequest) (*netbox_goV1.ListExtrasNotificationByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
