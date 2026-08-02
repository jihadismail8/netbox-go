package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasNotificationgroupLogicer = (*extrasNotificationgroupHandler)(nil)

type extrasNotificationgroupHandler struct {
	server netbox_goV1.ExtrasNotificationgroupServer
}

// NewExtrasNotificationgroupHandler create a handler
func NewExtrasNotificationgroupHandler() netbox_goV1.ExtrasNotificationgroupLogicer {
	return &extrasNotificationgroupHandler{
		server: service.NewExtrasNotificationgroupServer(),
	}
}

// Create a new extrasNotificationgroup
func (h *extrasNotificationgroupHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasNotificationgroupRequest) (*netbox_goV1.CreateExtrasNotificationgroupReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasNotificationgroup by id
func (h *extrasNotificationgroupHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasNotificationgroupByIDRequest) (*netbox_goV1.DeleteExtrasNotificationgroupByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasNotificationgroup by id
func (h *extrasNotificationgroupHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasNotificationgroupByIDRequest) (*netbox_goV1.UpdateExtrasNotificationgroupByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasNotificationgroup by id
func (h *extrasNotificationgroupHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasNotificationgroupByIDRequest) (*netbox_goV1.GetExtrasNotificationgroupByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasNotificationgroups by custom conditions
func (h *extrasNotificationgroupHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasNotificationgroupRequest) (*netbox_goV1.ListExtrasNotificationgroupReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasNotificationgroup by ids
func (h *extrasNotificationgroupHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasNotificationgroupByIDsRequest) (*netbox_goV1.DeleteExtrasNotificationgroupByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasNotificationgroup by custom condition
func (h *extrasNotificationgroupHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasNotificationgroupByConditionRequest) (*netbox_goV1.GetExtrasNotificationgroupByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasNotificationgroup by ids
func (h *extrasNotificationgroupHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasNotificationgroupByIDsRequest) (*netbox_goV1.ListExtrasNotificationgroupByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasNotificationgroups by last id
func (h *extrasNotificationgroupHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasNotificationgroupByLastIDRequest) (*netbox_goV1.ListExtrasNotificationgroupByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
