package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasCustomfieldchoicesetLogicer = (*extrasCustomfieldchoicesetHandler)(nil)

type extrasCustomfieldchoicesetHandler struct {
	server netbox_goV1.ExtrasCustomfieldchoicesetServer
}

// NewExtrasCustomfieldchoicesetHandler create a handler
func NewExtrasCustomfieldchoicesetHandler() netbox_goV1.ExtrasCustomfieldchoicesetLogicer {
	return &extrasCustomfieldchoicesetHandler{
		server: service.NewExtrasCustomfieldchoicesetServer(),
	}
}

// Create a new extrasCustomfieldchoiceset
func (h *extrasCustomfieldchoicesetHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasCustomfieldchoicesetRequest) (*netbox_goV1.CreateExtrasCustomfieldchoicesetReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasCustomfieldchoiceset by id
func (h *extrasCustomfieldchoicesetHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasCustomfieldchoicesetByIDRequest) (*netbox_goV1.DeleteExtrasCustomfieldchoicesetByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasCustomfieldchoiceset by id
func (h *extrasCustomfieldchoicesetHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasCustomfieldchoicesetByIDRequest) (*netbox_goV1.UpdateExtrasCustomfieldchoicesetByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasCustomfieldchoiceset by id
func (h *extrasCustomfieldchoicesetHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasCustomfieldchoicesetByIDRequest) (*netbox_goV1.GetExtrasCustomfieldchoicesetByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasCustomfieldchoicesets by custom conditions
func (h *extrasCustomfieldchoicesetHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasCustomfieldchoicesetRequest) (*netbox_goV1.ListExtrasCustomfieldchoicesetReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasCustomfieldchoiceset by ids
func (h *extrasCustomfieldchoicesetHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasCustomfieldchoicesetByIDsRequest) (*netbox_goV1.DeleteExtrasCustomfieldchoicesetByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasCustomfieldchoiceset by custom condition
func (h *extrasCustomfieldchoicesetHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasCustomfieldchoicesetByConditionRequest) (*netbox_goV1.GetExtrasCustomfieldchoicesetByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasCustomfieldchoiceset by ids
func (h *extrasCustomfieldchoicesetHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasCustomfieldchoicesetByIDsRequest) (*netbox_goV1.ListExtrasCustomfieldchoicesetByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasCustomfieldchoicesets by last id
func (h *extrasCustomfieldchoicesetHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasCustomfieldchoicesetByLastIDRequest) (*netbox_goV1.ListExtrasCustomfieldchoicesetByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
