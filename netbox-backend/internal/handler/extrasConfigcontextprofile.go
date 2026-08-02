package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasConfigcontextprofileLogicer = (*extrasConfigcontextprofileHandler)(nil)

type extrasConfigcontextprofileHandler struct {
	server netbox_goV1.ExtrasConfigcontextprofileServer
}

// NewExtrasConfigcontextprofileHandler create a handler
func NewExtrasConfigcontextprofileHandler() netbox_goV1.ExtrasConfigcontextprofileLogicer {
	return &extrasConfigcontextprofileHandler{
		server: service.NewExtrasConfigcontextprofileServer(),
	}
}

// Create a new extrasConfigcontextprofile
func (h *extrasConfigcontextprofileHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasConfigcontextprofileRequest) (*netbox_goV1.CreateExtrasConfigcontextprofileReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasConfigcontextprofile by id
func (h *extrasConfigcontextprofileHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasConfigcontextprofileByIDRequest) (*netbox_goV1.DeleteExtrasConfigcontextprofileByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasConfigcontextprofile by id
func (h *extrasConfigcontextprofileHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasConfigcontextprofileByIDRequest) (*netbox_goV1.UpdateExtrasConfigcontextprofileByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasConfigcontextprofile by id
func (h *extrasConfigcontextprofileHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasConfigcontextprofileByIDRequest) (*netbox_goV1.GetExtrasConfigcontextprofileByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasConfigcontextprofiles by custom conditions
func (h *extrasConfigcontextprofileHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasConfigcontextprofileRequest) (*netbox_goV1.ListExtrasConfigcontextprofileReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasConfigcontextprofile by ids
func (h *extrasConfigcontextprofileHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasConfigcontextprofileByIDsRequest) (*netbox_goV1.DeleteExtrasConfigcontextprofileByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasConfigcontextprofile by custom condition
func (h *extrasConfigcontextprofileHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasConfigcontextprofileByConditionRequest) (*netbox_goV1.GetExtrasConfigcontextprofileByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasConfigcontextprofile by ids
func (h *extrasConfigcontextprofileHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasConfigcontextprofileByIDsRequest) (*netbox_goV1.ListExtrasConfigcontextprofileByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasConfigcontextprofiles by last id
func (h *extrasConfigcontextprofileHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasConfigcontextprofileByLastIDRequest) (*netbox_goV1.ListExtrasConfigcontextprofileByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
