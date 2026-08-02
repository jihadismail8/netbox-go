package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CircuitsProvideraccountLogicer = (*circuitsProvideraccountHandler)(nil)

type circuitsProvideraccountHandler struct {
	server netbox_goV1.CircuitsProvideraccountServer
}

// NewCircuitsProvideraccountHandler create a handler
func NewCircuitsProvideraccountHandler() netbox_goV1.CircuitsProvideraccountLogicer {
	return &circuitsProvideraccountHandler{
		server: service.NewCircuitsProvideraccountServer(),
	}
}

// Create a new circuitsProvideraccount
func (h *circuitsProvideraccountHandler) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsProvideraccountRequest) (*netbox_goV1.CreateCircuitsProvideraccountReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a circuitsProvideraccount by id
func (h *circuitsProvideraccountHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsProvideraccountByIDRequest) (*netbox_goV1.DeleteCircuitsProvideraccountByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a circuitsProvideraccount by id
func (h *circuitsProvideraccountHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsProvideraccountByIDRequest) (*netbox_goV1.UpdateCircuitsProvideraccountByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a circuitsProvideraccount by id
func (h *circuitsProvideraccountHandler) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsProvideraccountByIDRequest) (*netbox_goV1.GetCircuitsProvideraccountByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of circuitsProvideraccounts by custom conditions
func (h *circuitsProvideraccountHandler) List(ctx context.Context, req *netbox_goV1.ListCircuitsProvideraccountRequest) (*netbox_goV1.ListCircuitsProvideraccountReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete circuitsProvideraccount by ids
func (h *circuitsProvideraccountHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsProvideraccountByIDsRequest) (*netbox_goV1.DeleteCircuitsProvideraccountByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a circuitsProvideraccount by custom condition
func (h *circuitsProvideraccountHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsProvideraccountByConditionRequest) (*netbox_goV1.GetCircuitsProvideraccountByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get circuitsProvideraccount by ids
func (h *circuitsProvideraccountHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsProvideraccountByIDsRequest) (*netbox_goV1.ListCircuitsProvideraccountByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of circuitsProvideraccounts by last id
func (h *circuitsProvideraccountHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsProvideraccountByLastIDRequest) (*netbox_goV1.ListCircuitsProvideraccountByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
