package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasSubscriptionLogicer = (*extrasSubscriptionHandler)(nil)

type extrasSubscriptionHandler struct {
	server netbox_goV1.ExtrasSubscriptionServer
}

// NewExtrasSubscriptionHandler create a handler
func NewExtrasSubscriptionHandler() netbox_goV1.ExtrasSubscriptionLogicer {
	return &extrasSubscriptionHandler{
		server: service.NewExtrasSubscriptionServer(),
	}
}

// Create a new extrasSubscription
func (h *extrasSubscriptionHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasSubscriptionRequest) (*netbox_goV1.CreateExtrasSubscriptionReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasSubscription by id
func (h *extrasSubscriptionHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasSubscriptionByIDRequest) (*netbox_goV1.DeleteExtrasSubscriptionByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasSubscription by id
func (h *extrasSubscriptionHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasSubscriptionByIDRequest) (*netbox_goV1.UpdateExtrasSubscriptionByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasSubscription by id
func (h *extrasSubscriptionHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasSubscriptionByIDRequest) (*netbox_goV1.GetExtrasSubscriptionByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasSubscriptions by custom conditions
func (h *extrasSubscriptionHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasSubscriptionRequest) (*netbox_goV1.ListExtrasSubscriptionReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasSubscription by ids
func (h *extrasSubscriptionHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasSubscriptionByIDsRequest) (*netbox_goV1.DeleteExtrasSubscriptionByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasSubscription by custom condition
func (h *extrasSubscriptionHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasSubscriptionByConditionRequest) (*netbox_goV1.GetExtrasSubscriptionByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasSubscription by ids
func (h *extrasSubscriptionHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasSubscriptionByIDsRequest) (*netbox_goV1.ListExtrasSubscriptionByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasSubscriptions by last id
func (h *extrasSubscriptionHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasSubscriptionByLastIDRequest) (*netbox_goV1.ListExtrasSubscriptionByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
