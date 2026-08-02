package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimCableterminationLogicer = (*dcimCableterminationHandler)(nil)

type dcimCableterminationHandler struct {
	server netbox_goV1.DcimCableterminationServer
}

// NewDcimCableterminationHandler create a handler
func NewDcimCableterminationHandler() netbox_goV1.DcimCableterminationLogicer {
	return &dcimCableterminationHandler{
		server: service.NewDcimCableterminationServer(),
	}
}

// Create a new dcimCabletermination
func (h *dcimCableterminationHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimCableterminationRequest) (*netbox_goV1.CreateDcimCableterminationReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimCabletermination by id
func (h *dcimCableterminationHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimCableterminationByIDRequest) (*netbox_goV1.DeleteDcimCableterminationByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimCabletermination by id
func (h *dcimCableterminationHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimCableterminationByIDRequest) (*netbox_goV1.UpdateDcimCableterminationByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimCabletermination by id
func (h *dcimCableterminationHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimCableterminationByIDRequest) (*netbox_goV1.GetDcimCableterminationByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimCableterminations by custom conditions
func (h *dcimCableterminationHandler) List(ctx context.Context, req *netbox_goV1.ListDcimCableterminationRequest) (*netbox_goV1.ListDcimCableterminationReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimCabletermination by ids
func (h *dcimCableterminationHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimCableterminationByIDsRequest) (*netbox_goV1.DeleteDcimCableterminationByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimCabletermination by custom condition
func (h *dcimCableterminationHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimCableterminationByConditionRequest) (*netbox_goV1.GetDcimCableterminationByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimCabletermination by ids
func (h *dcimCableterminationHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimCableterminationByIDsRequest) (*netbox_goV1.ListDcimCableterminationByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimCableterminations by last id
func (h *dcimCableterminationHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimCableterminationByLastIDRequest) (*netbox_goV1.ListDcimCableterminationByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
