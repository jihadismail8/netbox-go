package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasSavedfilterLogicer = (*extrasSavedfilterHandler)(nil)

type extrasSavedfilterHandler struct {
	server netbox_goV1.ExtrasSavedfilterServer
}

// NewExtrasSavedfilterHandler create a handler
func NewExtrasSavedfilterHandler() netbox_goV1.ExtrasSavedfilterLogicer {
	return &extrasSavedfilterHandler{
		server: service.NewExtrasSavedfilterServer(),
	}
}

// Create a new extrasSavedfilter
func (h *extrasSavedfilterHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasSavedfilterRequest) (*netbox_goV1.CreateExtrasSavedfilterReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasSavedfilter by id
func (h *extrasSavedfilterHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasSavedfilterByIDRequest) (*netbox_goV1.DeleteExtrasSavedfilterByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasSavedfilter by id
func (h *extrasSavedfilterHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasSavedfilterByIDRequest) (*netbox_goV1.UpdateExtrasSavedfilterByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasSavedfilter by id
func (h *extrasSavedfilterHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasSavedfilterByIDRequest) (*netbox_goV1.GetExtrasSavedfilterByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasSavedfilters by custom conditions
func (h *extrasSavedfilterHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasSavedfilterRequest) (*netbox_goV1.ListExtrasSavedfilterReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasSavedfilter by ids
func (h *extrasSavedfilterHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasSavedfilterByIDsRequest) (*netbox_goV1.DeleteExtrasSavedfilterByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasSavedfilter by custom condition
func (h *extrasSavedfilterHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasSavedfilterByConditionRequest) (*netbox_goV1.GetExtrasSavedfilterByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasSavedfilter by ids
func (h *extrasSavedfilterHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasSavedfilterByIDsRequest) (*netbox_goV1.ListExtrasSavedfilterByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasSavedfilters by last id
func (h *extrasSavedfilterHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasSavedfilterByLastIDRequest) (*netbox_goV1.ListExtrasSavedfilterByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
