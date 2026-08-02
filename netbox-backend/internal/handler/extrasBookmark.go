package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasBookmarkLogicer = (*extrasBookmarkHandler)(nil)

type extrasBookmarkHandler struct {
	server netbox_goV1.ExtrasBookmarkServer
}

// NewExtrasBookmarkHandler create a handler
func NewExtrasBookmarkHandler() netbox_goV1.ExtrasBookmarkLogicer {
	return &extrasBookmarkHandler{
		server: service.NewExtrasBookmarkServer(),
	}
}

// Create a new extrasBookmark
func (h *extrasBookmarkHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasBookmarkRequest) (*netbox_goV1.CreateExtrasBookmarkReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasBookmark by id
func (h *extrasBookmarkHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasBookmarkByIDRequest) (*netbox_goV1.DeleteExtrasBookmarkByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasBookmark by id
func (h *extrasBookmarkHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasBookmarkByIDRequest) (*netbox_goV1.UpdateExtrasBookmarkByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasBookmark by id
func (h *extrasBookmarkHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasBookmarkByIDRequest) (*netbox_goV1.GetExtrasBookmarkByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasBookmarks by custom conditions
func (h *extrasBookmarkHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasBookmarkRequest) (*netbox_goV1.ListExtrasBookmarkReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasBookmark by ids
func (h *extrasBookmarkHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasBookmarkByIDsRequest) (*netbox_goV1.DeleteExtrasBookmarkByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasBookmark by custom condition
func (h *extrasBookmarkHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasBookmarkByConditionRequest) (*netbox_goV1.GetExtrasBookmarkByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasBookmark by ids
func (h *extrasBookmarkHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasBookmarkByIDsRequest) (*netbox_goV1.ListExtrasBookmarkByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasBookmarks by last id
func (h *extrasBookmarkHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasBookmarkByLastIDRequest) (*netbox_goV1.ListExtrasBookmarkByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
