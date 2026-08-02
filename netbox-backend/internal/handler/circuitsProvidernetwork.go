package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CircuitsProvidernetworkLogicer = (*circuitsProvidernetworkHandler)(nil)

type circuitsProvidernetworkHandler struct {
	server netbox_goV1.CircuitsProvidernetworkServer
}

// NewCircuitsProvidernetworkHandler create a handler
func NewCircuitsProvidernetworkHandler() netbox_goV1.CircuitsProvidernetworkLogicer {
	return &circuitsProvidernetworkHandler{
		server: service.NewCircuitsProvidernetworkServer(),
	}
}

// Create a new circuitsProvidernetwork
func (h *circuitsProvidernetworkHandler) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsProvidernetworkRequest) (*netbox_goV1.CreateCircuitsProvidernetworkReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a circuitsProvidernetwork by id
func (h *circuitsProvidernetworkHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsProvidernetworkByIDRequest) (*netbox_goV1.DeleteCircuitsProvidernetworkByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a circuitsProvidernetwork by id
func (h *circuitsProvidernetworkHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsProvidernetworkByIDRequest) (*netbox_goV1.UpdateCircuitsProvidernetworkByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a circuitsProvidernetwork by id
func (h *circuitsProvidernetworkHandler) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsProvidernetworkByIDRequest) (*netbox_goV1.GetCircuitsProvidernetworkByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of circuitsProvidernetworks by custom conditions
func (h *circuitsProvidernetworkHandler) List(ctx context.Context, req *netbox_goV1.ListCircuitsProvidernetworkRequest) (*netbox_goV1.ListCircuitsProvidernetworkReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete circuitsProvidernetwork by ids
func (h *circuitsProvidernetworkHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsProvidernetworkByIDsRequest) (*netbox_goV1.DeleteCircuitsProvidernetworkByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a circuitsProvidernetwork by custom condition
func (h *circuitsProvidernetworkHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsProvidernetworkByConditionRequest) (*netbox_goV1.GetCircuitsProvidernetworkByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get circuitsProvidernetwork by ids
func (h *circuitsProvidernetworkHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsProvidernetworkByIDsRequest) (*netbox_goV1.ListCircuitsProvidernetworkByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of circuitsProvidernetworks by last id
func (h *circuitsProvidernetworkHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsProvidernetworkByLastIDRequest) (*netbox_goV1.ListCircuitsProvidernetworkByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
