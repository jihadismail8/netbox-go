package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimPowerpanelLogicer = (*dcimPowerpanelHandler)(nil)

type dcimPowerpanelHandler struct {
	server netbox_goV1.DcimPowerpanelServer
}

// NewDcimPowerpanelHandler create a handler
func NewDcimPowerpanelHandler() netbox_goV1.DcimPowerpanelLogicer {
	return &dcimPowerpanelHandler{
		server: service.NewDcimPowerpanelServer(),
	}
}

// Create a new dcimPowerpanel
func (h *dcimPowerpanelHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimPowerpanelRequest) (*netbox_goV1.CreateDcimPowerpanelReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimPowerpanel by id
func (h *dcimPowerpanelHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimPowerpanelByIDRequest) (*netbox_goV1.DeleteDcimPowerpanelByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimPowerpanel by id
func (h *dcimPowerpanelHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimPowerpanelByIDRequest) (*netbox_goV1.UpdateDcimPowerpanelByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimPowerpanel by id
func (h *dcimPowerpanelHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimPowerpanelByIDRequest) (*netbox_goV1.GetDcimPowerpanelByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimPowerpanels by custom conditions
func (h *dcimPowerpanelHandler) List(ctx context.Context, req *netbox_goV1.ListDcimPowerpanelRequest) (*netbox_goV1.ListDcimPowerpanelReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimPowerpanel by ids
func (h *dcimPowerpanelHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimPowerpanelByIDsRequest) (*netbox_goV1.DeleteDcimPowerpanelByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimPowerpanel by custom condition
func (h *dcimPowerpanelHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimPowerpanelByConditionRequest) (*netbox_goV1.GetDcimPowerpanelByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimPowerpanel by ids
func (h *dcimPowerpanelHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimPowerpanelByIDsRequest) (*netbox_goV1.ListDcimPowerpanelByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimPowerpanels by last id
func (h *dcimPowerpanelHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimPowerpanelByLastIDRequest) (*netbox_goV1.ListDcimPowerpanelByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
