package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CircuitsProviderLogicer = (*circuitsProviderHandler)(nil)

type circuitsProviderHandler struct {
	server netbox_goV1.CircuitsProviderServer
}

// NewCircuitsProviderHandler create a handler
func NewCircuitsProviderHandler() netbox_goV1.CircuitsProviderLogicer {
	return &circuitsProviderHandler{
		server: service.NewCircuitsProviderServer(),
	}
}

// Create a new circuitsProvider
func (h *circuitsProviderHandler) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsProviderRequest) (*netbox_goV1.CreateCircuitsProviderReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a circuitsProvider by id
func (h *circuitsProviderHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsProviderByIDRequest) (*netbox_goV1.DeleteCircuitsProviderByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a circuitsProvider by id
func (h *circuitsProviderHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsProviderByIDRequest) (*netbox_goV1.UpdateCircuitsProviderByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a circuitsProvider by id
func (h *circuitsProviderHandler) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsProviderByIDRequest) (*netbox_goV1.GetCircuitsProviderByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of circuitsProviders by custom conditions
func (h *circuitsProviderHandler) List(ctx context.Context, req *netbox_goV1.ListCircuitsProviderRequest) (*netbox_goV1.ListCircuitsProviderReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete circuitsProvider by ids
func (h *circuitsProviderHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsProviderByIDsRequest) (*netbox_goV1.DeleteCircuitsProviderByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a circuitsProvider by custom condition
func (h *circuitsProviderHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsProviderByConditionRequest) (*netbox_goV1.GetCircuitsProviderByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get circuitsProvider by ids
func (h *circuitsProviderHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsProviderByIDsRequest) (*netbox_goV1.ListCircuitsProviderByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of circuitsProviders by last id
func (h *circuitsProviderHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsProviderByLastIDRequest) (*netbox_goV1.ListCircuitsProviderByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
