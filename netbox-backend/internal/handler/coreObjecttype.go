package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.CoreObjecttypeLogicer = (*coreObjecttypeHandler)(nil)

type coreObjecttypeHandler struct {
	server netbox_goV1.CoreObjecttypeServer
}

// NewCoreObjecttypeHandler create a handler
func NewCoreObjecttypeHandler() netbox_goV1.CoreObjecttypeLogicer {
	return &coreObjecttypeHandler{
		server: service.NewCoreObjecttypeServer(),
	}
}

// Create a new coreObjecttype
func (h *coreObjecttypeHandler) Create(ctx context.Context, req *netbox_goV1.CreateCoreObjecttypeRequest) (*netbox_goV1.CreateCoreObjecttypeReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByContenttypePtrID delete a coreObjecttype by contenttypePtrID
func (h *coreObjecttypeHandler) DeleteByContenttypePtrID(ctx context.Context, req *netbox_goV1.DeleteCoreObjecttypeByContenttypePtrIDRequest) (*netbox_goV1.DeleteCoreObjecttypeByContenttypePtrIDReply, error) {
	return h.server.DeleteByContenttypePtrID(ctx, req)
}

// UpdateByContenttypePtrID update a coreObjecttype by contenttypePtrID
func (h *coreObjecttypeHandler) UpdateByContenttypePtrID(ctx context.Context, req *netbox_goV1.UpdateCoreObjecttypeByContenttypePtrIDRequest) (*netbox_goV1.UpdateCoreObjecttypeByContenttypePtrIDReply, error) {
	return h.server.UpdateByContenttypePtrID(ctx, req)
}

// GetByContenttypePtrID get a coreObjecttype by contenttypePtrID
func (h *coreObjecttypeHandler) GetByContenttypePtrID(ctx context.Context, req *netbox_goV1.GetCoreObjecttypeByContenttypePtrIDRequest) (*netbox_goV1.GetCoreObjecttypeByContenttypePtrIDReply, error) {
	return h.server.GetByContenttypePtrID(ctx, req)
}

// List get a paginated list of coreObjecttypes by custom conditions
func (h *coreObjecttypeHandler) List(ctx context.Context, req *netbox_goV1.ListCoreObjecttypeRequest) (*netbox_goV1.ListCoreObjecttypeReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByContenttypePtrIDs batch delete coreObjecttypes by contenttypePtrIDs
func (h *coreObjecttypeHandler) DeleteByContenttypePtrIDs(ctx context.Context, req *netbox_goV1.DeleteCoreObjecttypeByContenttypePtrIDsRequest) (*netbox_goV1.DeleteCoreObjecttypeByContenttypePtrIDsReply, error) {
	return h.server.DeleteByContenttypePtrIDs(ctx, req)
}

// GetByCondition get a coreObjecttype by custom condition
func (h *coreObjecttypeHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetCoreObjecttypeByConditionRequest) (*netbox_goV1.GetCoreObjecttypeByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByContenttypePtrIDs batch get coreObjecttypes by contenttypePtrIDs
func (h *coreObjecttypeHandler) ListByContenttypePtrIDs(ctx context.Context, req *netbox_goV1.ListCoreObjecttypeByContenttypePtrIDsRequest) (*netbox_goV1.ListCoreObjecttypeByContenttypePtrIDsReply, error) {
	return h.server.ListByContenttypePtrIDs(ctx, req)
}

// ListByLastContenttypePtrID get a paginated list of coreObjecttypes by last contenttypePtrID
func (h *coreObjecttypeHandler) ListByLastContenttypePtrID(ctx context.Context, req *netbox_goV1.ListCoreObjecttypeByLastContenttypePtrIDRequest) (*netbox_goV1.ListCoreObjecttypeByLastContenttypePtrIDReply, error) {
	return h.server.ListByLastContenttypePtrID(ctx, req)
}
