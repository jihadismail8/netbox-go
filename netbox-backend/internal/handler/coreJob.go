package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CoreJobLogicer = (*coreJobHandler)(nil)

type coreJobHandler struct {
	server netbox_goV1.CoreJobServer
}

// NewCoreJobHandler create a handler
func NewCoreJobHandler() netbox_goV1.CoreJobLogicer {
	return &coreJobHandler{
		server: service.NewCoreJobServer(),
	}
}

// Create a new coreJob
func (h *coreJobHandler) Create(ctx context.Context, req *netbox_goV1.CreateCoreJobRequest) (*netbox_goV1.CreateCoreJobReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a coreJob by id
func (h *coreJobHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCoreJobByIDRequest) (*netbox_goV1.DeleteCoreJobByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a coreJob by id
func (h *coreJobHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCoreJobByIDRequest) (*netbox_goV1.UpdateCoreJobByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a coreJob by id
func (h *coreJobHandler) GetByID(ctx context.Context, req *netbox_goV1.GetCoreJobByIDRequest) (*netbox_goV1.GetCoreJobByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of coreJobs by custom conditions
func (h *coreJobHandler) List(ctx context.Context, req *netbox_goV1.ListCoreJobRequest) (*netbox_goV1.ListCoreJobReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete coreJob by ids
func (h *coreJobHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCoreJobByIDsRequest) (*netbox_goV1.DeleteCoreJobByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a coreJob by custom condition
func (h *coreJobHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCoreJobByConditionRequest) (*netbox_goV1.GetCoreJobByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get coreJob by ids
func (h *coreJobHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListCoreJobByIDsRequest) (*netbox_goV1.ListCoreJobByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of coreJobs by last id
func (h *coreJobHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListCoreJobByLastIDRequest) (*netbox_goV1.ListCoreJobByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
