package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.WirelessWirelesslangroupLogicer = (*wirelessWirelesslangroupHandler)(nil)

type wirelessWirelesslangroupHandler struct {
	server netbox_goV1.WirelessWirelesslangroupServer
}

// NewWirelessWirelesslangroupHandler create a handler
func NewWirelessWirelesslangroupHandler() netbox_goV1.WirelessWirelesslangroupLogicer {
	return &wirelessWirelesslangroupHandler{
		server: service.NewWirelessWirelesslangroupServer(),
	}
}

// Create a new wirelessWirelesslangroup
func (h *wirelessWirelesslangroupHandler) Create(ctx context.Context, req *netbox_goV1.CreateWirelessWirelesslangroupRequest) (*netbox_goV1.CreateWirelessWirelesslangroupReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a wirelessWirelesslangroup by id
func (h *wirelessWirelesslangroupHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteWirelessWirelesslangroupByIDRequest) (*netbox_goV1.DeleteWirelessWirelesslangroupByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a wirelessWirelesslangroup by id
func (h *wirelessWirelesslangroupHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateWirelessWirelesslangroupByIDRequest) (*netbox_goV1.UpdateWirelessWirelesslangroupByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a wirelessWirelesslangroup by id
func (h *wirelessWirelesslangroupHandler) GetByID(ctx context.Context, req *netbox_goV1.GetWirelessWirelesslangroupByIDRequest) (*netbox_goV1.GetWirelessWirelesslangroupByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of wirelessWirelesslangroups by custom conditions
func (h *wirelessWirelesslangroupHandler) List(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslangroupRequest) (*netbox_goV1.ListWirelessWirelesslangroupReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete wirelessWirelesslangroup by ids
func (h *wirelessWirelesslangroupHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteWirelessWirelesslangroupByIDsRequest) (*netbox_goV1.DeleteWirelessWirelesslangroupByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a wirelessWirelesslangroup by custom condition
func (h *wirelessWirelesslangroupHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetWirelessWirelesslangroupByConditionRequest) (*netbox_goV1.GetWirelessWirelesslangroupByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get wirelessWirelesslangroup by ids
func (h *wirelessWirelesslangroupHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslangroupByIDsRequest) (*netbox_goV1.ListWirelessWirelesslangroupByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of wirelessWirelesslangroups by last id
func (h *wirelessWirelesslangroupHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslangroupByLastIDRequest) (*netbox_goV1.ListWirelessWirelesslangroupByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
