package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimDevicebaytemplateLogicer = (*dcimDevicebaytemplateHandler)(nil)

type dcimDevicebaytemplateHandler struct {
	server netbox_goV1.DcimDevicebaytemplateServer
}

// NewDcimDevicebaytemplateHandler create a handler
func NewDcimDevicebaytemplateHandler() netbox_goV1.DcimDevicebaytemplateLogicer {
	return &dcimDevicebaytemplateHandler{
		server: service.NewDcimDevicebaytemplateServer(),
	}
}

// Create a new dcimDevicebaytemplate
func (h *dcimDevicebaytemplateHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimDevicebaytemplateRequest) (*netbox_goV1.CreateDcimDevicebaytemplateReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimDevicebaytemplate by id
func (h *dcimDevicebaytemplateHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimDevicebaytemplateByIDRequest) (*netbox_goV1.DeleteDcimDevicebaytemplateByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimDevicebaytemplate by id
func (h *dcimDevicebaytemplateHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimDevicebaytemplateByIDRequest) (*netbox_goV1.UpdateDcimDevicebaytemplateByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimDevicebaytemplate by id
func (h *dcimDevicebaytemplateHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimDevicebaytemplateByIDRequest) (*netbox_goV1.GetDcimDevicebaytemplateByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimDevicebaytemplates by custom conditions
func (h *dcimDevicebaytemplateHandler) List(ctx context.Context, req *netbox_goV1.ListDcimDevicebaytemplateRequest) (*netbox_goV1.ListDcimDevicebaytemplateReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimDevicebaytemplate by ids
func (h *dcimDevicebaytemplateHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimDevicebaytemplateByIDsRequest) (*netbox_goV1.DeleteDcimDevicebaytemplateByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimDevicebaytemplate by custom condition
func (h *dcimDevicebaytemplateHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimDevicebaytemplateByConditionRequest) (*netbox_goV1.GetDcimDevicebaytemplateByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimDevicebaytemplate by ids
func (h *dcimDevicebaytemplateHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimDevicebaytemplateByIDsRequest) (*netbox_goV1.ListDcimDevicebaytemplateByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimDevicebaytemplates by last id
func (h *dcimDevicebaytemplateHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimDevicebaytemplateByLastIDRequest) (*netbox_goV1.ListDcimDevicebaytemplateByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
