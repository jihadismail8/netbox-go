package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.WirelessWirelesslinkLogicer = (*wirelessWirelesslinkHandler)(nil)

type wirelessWirelesslinkHandler struct {
	server netbox_goV1.WirelessWirelesslinkServer
}

// NewWirelessWirelesslinkHandler create a handler
func NewWirelessWirelesslinkHandler() netbox_goV1.WirelessWirelesslinkLogicer {
	return &wirelessWirelesslinkHandler{
		server: service.NewWirelessWirelesslinkServer(),
	}
}

// Create a new wirelessWirelesslink
func (h *wirelessWirelesslinkHandler) Create(ctx context.Context, req *netbox_goV1.CreateWirelessWirelesslinkRequest) (*netbox_goV1.CreateWirelessWirelesslinkReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a wirelessWirelesslink by id
func (h *wirelessWirelesslinkHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteWirelessWirelesslinkByIDRequest) (*netbox_goV1.DeleteWirelessWirelesslinkByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a wirelessWirelesslink by id
func (h *wirelessWirelesslinkHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateWirelessWirelesslinkByIDRequest) (*netbox_goV1.UpdateWirelessWirelesslinkByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a wirelessWirelesslink by id
func (h *wirelessWirelesslinkHandler) GetByID(ctx context.Context, req *netbox_goV1.GetWirelessWirelesslinkByIDRequest) (*netbox_goV1.GetWirelessWirelesslinkByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of wirelessWirelesslinks by custom conditions
func (h *wirelessWirelesslinkHandler) List(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslinkRequest) (*netbox_goV1.ListWirelessWirelesslinkReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete wirelessWirelesslink by ids
func (h *wirelessWirelesslinkHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteWirelessWirelesslinkByIDsRequest) (*netbox_goV1.DeleteWirelessWirelesslinkByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a wirelessWirelesslink by custom condition
func (h *wirelessWirelesslinkHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetWirelessWirelesslinkByConditionRequest) (*netbox_goV1.GetWirelessWirelesslinkByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get wirelessWirelesslink by ids
func (h *wirelessWirelesslinkHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslinkByIDsRequest) (*netbox_goV1.ListWirelessWirelesslinkByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of wirelessWirelesslinks by last id
func (h *wirelessWirelesslinkHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslinkByLastIDRequest) (*netbox_goV1.ListWirelessWirelesslinkByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
