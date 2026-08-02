package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasDashboardLogicer = (*extrasDashboardHandler)(nil)

type extrasDashboardHandler struct {
	server netbox_goV1.ExtrasDashboardServer
}

// NewExtrasDashboardHandler create a handler
func NewExtrasDashboardHandler() netbox_goV1.ExtrasDashboardLogicer {
	return &extrasDashboardHandler{
		server: service.NewExtrasDashboardServer(),
	}
}

// Create a new extrasDashboard
func (h *extrasDashboardHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasDashboardRequest) (*netbox_goV1.CreateExtrasDashboardReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasDashboard by id
func (h *extrasDashboardHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasDashboardByIDRequest) (*netbox_goV1.DeleteExtrasDashboardByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasDashboard by id
func (h *extrasDashboardHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasDashboardByIDRequest) (*netbox_goV1.UpdateExtrasDashboardByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasDashboard by id
func (h *extrasDashboardHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasDashboardByIDRequest) (*netbox_goV1.GetExtrasDashboardByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasDashboards by custom conditions
func (h *extrasDashboardHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasDashboardRequest) (*netbox_goV1.ListExtrasDashboardReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasDashboard by ids
func (h *extrasDashboardHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasDashboardByIDsRequest) (*netbox_goV1.DeleteExtrasDashboardByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasDashboard by custom condition
func (h *extrasDashboardHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasDashboardByConditionRequest) (*netbox_goV1.GetExtrasDashboardByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasDashboard by ids
func (h *extrasDashboardHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasDashboardByIDsRequest) (*netbox_goV1.ListExtrasDashboardByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasDashboards by last id
func (h *extrasDashboardHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasDashboardByLastIDRequest) (*netbox_goV1.ListExtrasDashboardByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
