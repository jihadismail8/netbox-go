package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimPlatformLogicer = (*dcimPlatformHandler)(nil)

type dcimPlatformHandler struct {
	server netbox_goV1.DcimPlatformServer
}

// NewDcimPlatformHandler create a handler
func NewDcimPlatformHandler() netbox_goV1.DcimPlatformLogicer {
	return &dcimPlatformHandler{
		server: service.NewDcimPlatformServer(),
	}
}

// Create a new dcimPlatform
func (h *dcimPlatformHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimPlatformRequest) (*netbox_goV1.CreateDcimPlatformReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimPlatform by id
func (h *dcimPlatformHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimPlatformByIDRequest) (*netbox_goV1.DeleteDcimPlatformByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimPlatform by id
func (h *dcimPlatformHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimPlatformByIDRequest) (*netbox_goV1.UpdateDcimPlatformByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimPlatform by id
func (h *dcimPlatformHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimPlatformByIDRequest) (*netbox_goV1.GetDcimPlatformByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimPlatforms by custom conditions
func (h *dcimPlatformHandler) List(ctx context.Context, req *netbox_goV1.ListDcimPlatformRequest) (*netbox_goV1.ListDcimPlatformReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimPlatform by ids
func (h *dcimPlatformHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimPlatformByIDsRequest) (*netbox_goV1.DeleteDcimPlatformByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimPlatform by custom condition
func (h *dcimPlatformHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimPlatformByConditionRequest) (*netbox_goV1.GetDcimPlatformByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimPlatform by ids
func (h *dcimPlatformHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimPlatformByIDsRequest) (*netbox_goV1.ListDcimPlatformByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimPlatforms by last id
func (h *dcimPlatformHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimPlatformByLastIDRequest) (*netbox_goV1.ListDcimPlatformByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
