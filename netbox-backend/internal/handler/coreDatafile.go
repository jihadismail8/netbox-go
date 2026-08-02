package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CoreDatafileLogicer = (*coreDatafileHandler)(nil)

type coreDatafileHandler struct {
	server netbox_goV1.CoreDatafileServer
}

// NewCoreDatafileHandler create a handler
func NewCoreDatafileHandler() netbox_goV1.CoreDatafileLogicer {
	return &coreDatafileHandler{
		server: service.NewCoreDatafileServer(),
	}
}

// Create a new coreDatafile
func (h *coreDatafileHandler) Create(ctx context.Context, req *netbox_goV1.CreateCoreDatafileRequest) (*netbox_goV1.CreateCoreDatafileReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a coreDatafile by id
func (h *coreDatafileHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCoreDatafileByIDRequest) (*netbox_goV1.DeleteCoreDatafileByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a coreDatafile by id
func (h *coreDatafileHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCoreDatafileByIDRequest) (*netbox_goV1.UpdateCoreDatafileByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a coreDatafile by id
func (h *coreDatafileHandler) GetByID(ctx context.Context, req *netbox_goV1.GetCoreDatafileByIDRequest) (*netbox_goV1.GetCoreDatafileByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of coreDatafiles by custom conditions
func (h *coreDatafileHandler) List(ctx context.Context, req *netbox_goV1.ListCoreDatafileRequest) (*netbox_goV1.ListCoreDatafileReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete coreDatafile by ids
func (h *coreDatafileHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCoreDatafileByIDsRequest) (*netbox_goV1.DeleteCoreDatafileByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a coreDatafile by custom condition
func (h *coreDatafileHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCoreDatafileByConditionRequest) (*netbox_goV1.GetCoreDatafileByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get coreDatafile by ids
func (h *coreDatafileHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListCoreDatafileByIDsRequest) (*netbox_goV1.ListCoreDatafileByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of coreDatafiles by last id
func (h *coreDatafileHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListCoreDatafileByLastIDRequest) (*netbox_goV1.ListCoreDatafileByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
