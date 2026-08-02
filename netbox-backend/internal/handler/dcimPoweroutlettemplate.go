package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimPoweroutlettemplateLogicer = (*dcimPoweroutlettemplateHandler)(nil)

type dcimPoweroutlettemplateHandler struct {
	server netbox_goV1.DcimPoweroutlettemplateServer
}

// NewDcimPoweroutlettemplateHandler create a handler
func NewDcimPoweroutlettemplateHandler() netbox_goV1.DcimPoweroutlettemplateLogicer {
	return &dcimPoweroutlettemplateHandler{
		server: service.NewDcimPoweroutlettemplateServer(),
	}
}

// Create a new dcimPoweroutlettemplate
func (h *dcimPoweroutlettemplateHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimPoweroutlettemplateRequest) (*netbox_goV1.CreateDcimPoweroutlettemplateReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimPoweroutlettemplate by id
func (h *dcimPoweroutlettemplateHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimPoweroutlettemplateByIDRequest) (*netbox_goV1.DeleteDcimPoweroutlettemplateByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimPoweroutlettemplate by id
func (h *dcimPoweroutlettemplateHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimPoweroutlettemplateByIDRequest) (*netbox_goV1.UpdateDcimPoweroutlettemplateByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimPoweroutlettemplate by id
func (h *dcimPoweroutlettemplateHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimPoweroutlettemplateByIDRequest) (*netbox_goV1.GetDcimPoweroutlettemplateByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimPoweroutlettemplates by custom conditions
func (h *dcimPoweroutlettemplateHandler) List(ctx context.Context, req *netbox_goV1.ListDcimPoweroutlettemplateRequest) (*netbox_goV1.ListDcimPoweroutlettemplateReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimPoweroutlettemplate by ids
func (h *dcimPoweroutlettemplateHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimPoweroutlettemplateByIDsRequest) (*netbox_goV1.DeleteDcimPoweroutlettemplateByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimPoweroutlettemplate by custom condition
func (h *dcimPoweroutlettemplateHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimPoweroutlettemplateByConditionRequest) (*netbox_goV1.GetDcimPoweroutlettemplateByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimPoweroutlettemplate by ids
func (h *dcimPoweroutlettemplateHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimPoweroutlettemplateByIDsRequest) (*netbox_goV1.ListDcimPoweroutlettemplateByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimPoweroutlettemplates by last id
func (h *dcimPoweroutlettemplateHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimPoweroutlettemplateByLastIDRequest) (*netbox_goV1.ListDcimPoweroutlettemplateByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
