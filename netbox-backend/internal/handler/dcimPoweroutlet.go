package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimPoweroutletLogicer = (*dcimPoweroutletHandler)(nil)

type dcimPoweroutletHandler struct {
	server netbox_goV1.DcimPoweroutletServer
}

// NewDcimPoweroutletHandler create a handler
func NewDcimPoweroutletHandler() netbox_goV1.DcimPoweroutletLogicer {
	return &dcimPoweroutletHandler{
		server: service.NewDcimPoweroutletServer(),
	}
}

// Create a new dcimPoweroutlet
func (h *dcimPoweroutletHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimPoweroutletRequest) (*netbox_goV1.CreateDcimPoweroutletReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimPoweroutlet by id
func (h *dcimPoweroutletHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimPoweroutletByIDRequest) (*netbox_goV1.DeleteDcimPoweroutletByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimPoweroutlet by id
func (h *dcimPoweroutletHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimPoweroutletByIDRequest) (*netbox_goV1.UpdateDcimPoweroutletByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimPoweroutlet by id
func (h *dcimPoweroutletHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimPoweroutletByIDRequest) (*netbox_goV1.GetDcimPoweroutletByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimPoweroutlets by custom conditions
func (h *dcimPoweroutletHandler) List(ctx context.Context, req *netbox_goV1.ListDcimPoweroutletRequest) (*netbox_goV1.ListDcimPoweroutletReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimPoweroutlet by ids
func (h *dcimPoweroutletHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimPoweroutletByIDsRequest) (*netbox_goV1.DeleteDcimPoweroutletByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimPoweroutlet by custom condition
func (h *dcimPoweroutletHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimPoweroutletByConditionRequest) (*netbox_goV1.GetDcimPoweroutletByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimPoweroutlet by ids
func (h *dcimPoweroutletHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimPoweroutletByIDsRequest) (*netbox_goV1.ListDcimPoweroutletByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimPoweroutlets by last id
func (h *dcimPoweroutletHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimPoweroutletByLastIDRequest) (*netbox_goV1.ListDcimPoweroutletByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
