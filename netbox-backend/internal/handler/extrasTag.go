package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasTagLogicer = (*extrasTagHandler)(nil)

type extrasTagHandler struct {
	server netbox_goV1.ExtrasTagServer
}

// NewExtrasTagHandler create a handler
func NewExtrasTagHandler() netbox_goV1.ExtrasTagLogicer {
	return &extrasTagHandler{
		server: service.NewExtrasTagServer(),
	}
}

// Create a new extrasTag
func (h *extrasTagHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasTagRequest) (*netbox_goV1.CreateExtrasTagReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasTag by id
func (h *extrasTagHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasTagByIDRequest) (*netbox_goV1.DeleteExtrasTagByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasTag by id
func (h *extrasTagHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasTagByIDRequest) (*netbox_goV1.UpdateExtrasTagByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasTag by id
func (h *extrasTagHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasTagByIDRequest) (*netbox_goV1.GetExtrasTagByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasTags by custom conditions
func (h *extrasTagHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasTagRequest) (*netbox_goV1.ListExtrasTagReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasTag by ids
func (h *extrasTagHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasTagByIDsRequest) (*netbox_goV1.DeleteExtrasTagByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasTag by custom condition
func (h *extrasTagHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasTagByConditionRequest) (*netbox_goV1.GetExtrasTagByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasTag by ids
func (h *extrasTagHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasTagByIDsRequest) (*netbox_goV1.ListExtrasTagByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasTags by last id
func (h *extrasTagHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasTagByLastIDRequest) (*netbox_goV1.ListExtrasTagByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
