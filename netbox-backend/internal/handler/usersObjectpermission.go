package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.UsersObjectpermissionLogicer = (*usersObjectpermissionHandler)(nil)

type usersObjectpermissionHandler struct {
	server netbox_goV1.UsersObjectpermissionServer
}

// NewUsersObjectpermissionHandler create a handler
func NewUsersObjectpermissionHandler() netbox_goV1.UsersObjectpermissionLogicer {
	return &usersObjectpermissionHandler{
		server: service.NewUsersObjectpermissionServer(),
	}
}

// Create a new usersObjectpermission
func (h *usersObjectpermissionHandler) Create(ctx context.Context, req *netbox_goV1.CreateUsersObjectpermissionRequest) (*netbox_goV1.CreateUsersObjectpermissionReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a usersObjectpermission by id
func (h *usersObjectpermissionHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteUsersObjectpermissionByIDRequest) (*netbox_goV1.DeleteUsersObjectpermissionByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a usersObjectpermission by id
func (h *usersObjectpermissionHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateUsersObjectpermissionByIDRequest) (*netbox_goV1.UpdateUsersObjectpermissionByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a usersObjectpermission by id
func (h *usersObjectpermissionHandler) GetByID(ctx context.Context, req *netbox_goV1.GetUsersObjectpermissionByIDRequest) (*netbox_goV1.GetUsersObjectpermissionByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of usersObjectpermissions by custom conditions
func (h *usersObjectpermissionHandler) List(ctx context.Context, req *netbox_goV1.ListUsersObjectpermissionRequest) (*netbox_goV1.ListUsersObjectpermissionReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete usersObjectpermission by ids
func (h *usersObjectpermissionHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteUsersObjectpermissionByIDsRequest) (*netbox_goV1.DeleteUsersObjectpermissionByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a usersObjectpermission by custom condition
func (h *usersObjectpermissionHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetUsersObjectpermissionByConditionRequest) (*netbox_goV1.GetUsersObjectpermissionByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get usersObjectpermission by ids
func (h *usersObjectpermissionHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListUsersObjectpermissionByIDsRequest) (*netbox_goV1.ListUsersObjectpermissionByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of usersObjectpermissions by last id
func (h *usersObjectpermissionHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListUsersObjectpermissionByLastIDRequest) (*netbox_goV1.ListUsersObjectpermissionByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
