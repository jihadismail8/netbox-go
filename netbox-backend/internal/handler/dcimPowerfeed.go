package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimPowerfeedLogicer = (*dcimPowerfeedHandler)(nil)

type dcimPowerfeedHandler struct {
	server netbox_goV1.DcimPowerfeedServer
}

// NewDcimPowerfeedHandler create a handler
func NewDcimPowerfeedHandler() netbox_goV1.DcimPowerfeedLogicer {
	return &dcimPowerfeedHandler{
		server: service.NewDcimPowerfeedServer(),
	}
}

// Create a new dcimPowerfeed
func (h *dcimPowerfeedHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimPowerfeedRequest) (*netbox_goV1.CreateDcimPowerfeedReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimPowerfeed by id
func (h *dcimPowerfeedHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimPowerfeedByIDRequest) (*netbox_goV1.DeleteDcimPowerfeedByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimPowerfeed by id
func (h *dcimPowerfeedHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimPowerfeedByIDRequest) (*netbox_goV1.UpdateDcimPowerfeedByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimPowerfeed by id
func (h *dcimPowerfeedHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimPowerfeedByIDRequest) (*netbox_goV1.GetDcimPowerfeedByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimPowerfeeds by custom conditions
func (h *dcimPowerfeedHandler) List(ctx context.Context, req *netbox_goV1.ListDcimPowerfeedRequest) (*netbox_goV1.ListDcimPowerfeedReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimPowerfeed by ids
func (h *dcimPowerfeedHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimPowerfeedByIDsRequest) (*netbox_goV1.DeleteDcimPowerfeedByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimPowerfeed by custom condition
func (h *dcimPowerfeedHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimPowerfeedByConditionRequest) (*netbox_goV1.GetDcimPowerfeedByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimPowerfeed by ids
func (h *dcimPowerfeedHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimPowerfeedByIDsRequest) (*netbox_goV1.ListDcimPowerfeedByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimPowerfeeds by last id
func (h *dcimPowerfeedHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimPowerfeedByLastIDRequest) (*netbox_goV1.ListDcimPowerfeedByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
