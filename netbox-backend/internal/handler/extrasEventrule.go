package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasEventruleLogicer = (*extrasEventruleHandler)(nil)

type extrasEventruleHandler struct {
	server netbox_goV1.ExtrasEventruleServer
}

// NewExtrasEventruleHandler create a handler
func NewExtrasEventruleHandler() netbox_goV1.ExtrasEventruleLogicer {
	return &extrasEventruleHandler{
		server: service.NewExtrasEventruleServer(),
	}
}

// Create a new extrasEventrule
func (h *extrasEventruleHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasEventruleRequest) (*netbox_goV1.CreateExtrasEventruleReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasEventrule by id
func (h *extrasEventruleHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasEventruleByIDRequest) (*netbox_goV1.DeleteExtrasEventruleByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasEventrule by id
func (h *extrasEventruleHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasEventruleByIDRequest) (*netbox_goV1.UpdateExtrasEventruleByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasEventrule by id
func (h *extrasEventruleHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasEventruleByIDRequest) (*netbox_goV1.GetExtrasEventruleByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasEventrules by custom conditions
func (h *extrasEventruleHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasEventruleRequest) (*netbox_goV1.ListExtrasEventruleReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasEventrule by ids
func (h *extrasEventruleHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasEventruleByIDsRequest) (*netbox_goV1.DeleteExtrasEventruleByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasEventrule by custom condition
func (h *extrasEventruleHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasEventruleByConditionRequest) (*netbox_goV1.GetExtrasEventruleByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasEventrule by ids
func (h *extrasEventruleHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasEventruleByIDsRequest) (*netbox_goV1.ListExtrasEventruleByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasEventrules by last id
func (h *extrasEventruleHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasEventruleByLastIDRequest) (*netbox_goV1.ListExtrasEventruleByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
