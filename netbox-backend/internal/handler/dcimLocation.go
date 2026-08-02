package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimLocationLogicer = (*dcimLocationHandler)(nil)

type dcimLocationHandler struct {
	server netbox_goV1.DcimLocationServer
}

// NewDcimLocationHandler create a handler
func NewDcimLocationHandler() netbox_goV1.DcimLocationLogicer {
	return &dcimLocationHandler{
		server: service.NewDcimLocationServer(),
	}
}

// Create a new dcimLocation
func (h *dcimLocationHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimLocationRequest) (*netbox_goV1.CreateDcimLocationReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimLocation by id
func (h *dcimLocationHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimLocationByIDRequest) (*netbox_goV1.DeleteDcimLocationByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimLocation by id
func (h *dcimLocationHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimLocationByIDRequest) (*netbox_goV1.UpdateDcimLocationByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimLocation by id
func (h *dcimLocationHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimLocationByIDRequest) (*netbox_goV1.GetDcimLocationByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimLocations by custom conditions
func (h *dcimLocationHandler) List(ctx context.Context, req *netbox_goV1.ListDcimLocationRequest) (*netbox_goV1.ListDcimLocationReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimLocation by ids
func (h *dcimLocationHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimLocationByIDsRequest) (*netbox_goV1.DeleteDcimLocationByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimLocation by custom condition
func (h *dcimLocationHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimLocationByConditionRequest) (*netbox_goV1.GetDcimLocationByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimLocation by ids
func (h *dcimLocationHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimLocationByIDsRequest) (*netbox_goV1.ListDcimLocationByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimLocations by last id
func (h *dcimLocationHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimLocationByLastIDRequest) (*netbox_goV1.ListDcimLocationByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
