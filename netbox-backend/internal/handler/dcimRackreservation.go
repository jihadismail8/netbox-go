package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimRackreservationLogicer = (*dcimRackreservationHandler)(nil)

type dcimRackreservationHandler struct {
	server netbox_goV1.DcimRackreservationServer
}

// NewDcimRackreservationHandler create a handler
func NewDcimRackreservationHandler() netbox_goV1.DcimRackreservationLogicer {
	return &dcimRackreservationHandler{
		server: service.NewDcimRackreservationServer(),
	}
}

// Create a new dcimRackreservation
func (h *dcimRackreservationHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimRackreservationRequest) (*netbox_goV1.CreateDcimRackreservationReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimRackreservation by id
func (h *dcimRackreservationHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimRackreservationByIDRequest) (*netbox_goV1.DeleteDcimRackreservationByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimRackreservation by id
func (h *dcimRackreservationHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimRackreservationByIDRequest) (*netbox_goV1.UpdateDcimRackreservationByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimRackreservation by id
func (h *dcimRackreservationHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimRackreservationByIDRequest) (*netbox_goV1.GetDcimRackreservationByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimRackreservations by custom conditions
func (h *dcimRackreservationHandler) List(ctx context.Context, req *netbox_goV1.ListDcimRackreservationRequest) (*netbox_goV1.ListDcimRackreservationReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimRackreservation by ids
func (h *dcimRackreservationHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimRackreservationByIDsRequest) (*netbox_goV1.DeleteDcimRackreservationByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimRackreservation by custom condition
func (h *dcimRackreservationHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimRackreservationByConditionRequest) (*netbox_goV1.GetDcimRackreservationByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimRackreservation by ids
func (h *dcimRackreservationHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimRackreservationByIDsRequest) (*netbox_goV1.ListDcimRackreservationByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimRackreservations by last id
func (h *dcimRackreservationHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimRackreservationByLastIDRequest) (*netbox_goV1.ListDcimRackreservationByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
