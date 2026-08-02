package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CoreDatasourceLogicer = (*coreDatasourceHandler)(nil)

type coreDatasourceHandler struct {
	server netbox_goV1.CoreDatasourceServer
}

// NewCoreDatasourceHandler create a handler
func NewCoreDatasourceHandler() netbox_goV1.CoreDatasourceLogicer {
	return &coreDatasourceHandler{
		server: service.NewCoreDatasourceServer(),
	}
}

// Create a new coreDatasource
func (h *coreDatasourceHandler) Create(ctx context.Context, req *netbox_goV1.CreateCoreDatasourceRequest) (*netbox_goV1.CreateCoreDatasourceReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a coreDatasource by id
func (h *coreDatasourceHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCoreDatasourceByIDRequest) (*netbox_goV1.DeleteCoreDatasourceByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a coreDatasource by id
func (h *coreDatasourceHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCoreDatasourceByIDRequest) (*netbox_goV1.UpdateCoreDatasourceByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a coreDatasource by id
func (h *coreDatasourceHandler) GetByID(ctx context.Context, req *netbox_goV1.GetCoreDatasourceByIDRequest) (*netbox_goV1.GetCoreDatasourceByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of coreDatasources by custom conditions
func (h *coreDatasourceHandler) List(ctx context.Context, req *netbox_goV1.ListCoreDatasourceRequest) (*netbox_goV1.ListCoreDatasourceReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete coreDatasource by ids
func (h *coreDatasourceHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCoreDatasourceByIDsRequest) (*netbox_goV1.DeleteCoreDatasourceByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a coreDatasource by custom condition
func (h *coreDatasourceHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCoreDatasourceByConditionRequest) (*netbox_goV1.GetCoreDatasourceByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get coreDatasource by ids
func (h *coreDatasourceHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListCoreDatasourceByIDsRequest) (*netbox_goV1.ListCoreDatasourceByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of coreDatasources by last id
func (h *coreDatasourceHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListCoreDatasourceByLastIDRequest) (*netbox_goV1.ListCoreDatasourceByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
