package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.WirelessWirelesslanLogicer = (*wirelessWirelesslanHandler)(nil)

type wirelessWirelesslanHandler struct {
	server netbox_goV1.WirelessWirelesslanServer
}

// NewWirelessWirelesslanHandler create a handler
func NewWirelessWirelesslanHandler() netbox_goV1.WirelessWirelesslanLogicer {
	return &wirelessWirelesslanHandler{
		server: service.NewWirelessWirelesslanServer(),
	}
}

// Create a new wirelessWirelesslan
func (h *wirelessWirelesslanHandler) Create(ctx context.Context, req *netbox_goV1.CreateWirelessWirelesslanRequest) (*netbox_goV1.CreateWirelessWirelesslanReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a wirelessWirelesslan by id
func (h *wirelessWirelesslanHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteWirelessWirelesslanByIDRequest) (*netbox_goV1.DeleteWirelessWirelesslanByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a wirelessWirelesslan by id
func (h *wirelessWirelesslanHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateWirelessWirelesslanByIDRequest) (*netbox_goV1.UpdateWirelessWirelesslanByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a wirelessWirelesslan by id
func (h *wirelessWirelesslanHandler) GetByID(ctx context.Context, req *netbox_goV1.GetWirelessWirelesslanByIDRequest) (*netbox_goV1.GetWirelessWirelesslanByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of wirelessWirelesslans by custom conditions
func (h *wirelessWirelesslanHandler) List(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslanRequest) (*netbox_goV1.ListWirelessWirelesslanReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete wirelessWirelesslan by ids
func (h *wirelessWirelesslanHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteWirelessWirelesslanByIDsRequest) (*netbox_goV1.DeleteWirelessWirelesslanByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a wirelessWirelesslan by custom condition
func (h *wirelessWirelesslanHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetWirelessWirelesslanByConditionRequest) (*netbox_goV1.GetWirelessWirelesslanByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get wirelessWirelesslan by ids
func (h *wirelessWirelesslanHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslanByIDsRequest) (*netbox_goV1.ListWirelessWirelesslanByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of wirelessWirelesslans by last id
func (h *wirelessWirelesslanHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslanByLastIDRequest) (*netbox_goV1.ListWirelessWirelesslanByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
