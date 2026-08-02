package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.UsersGroupLogicer = (*usersGroupHandler)(nil)

type usersGroupHandler struct {
	server netbox_goV1.UsersGroupServer
}

// NewUsersGroupHandler create a handler
func NewUsersGroupHandler() netbox_goV1.UsersGroupLogicer {
	return &usersGroupHandler{
		server: service.NewUsersGroupServer(),
	}
}

// Create a new usersGroup
func (h *usersGroupHandler) Create(ctx context.Context, req *netbox_goV1.CreateUsersGroupRequest) (*netbox_goV1.CreateUsersGroupReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a usersGroup by id
func (h *usersGroupHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteUsersGroupByIDRequest) (*netbox_goV1.DeleteUsersGroupByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a usersGroup by id
func (h *usersGroupHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateUsersGroupByIDRequest) (*netbox_goV1.UpdateUsersGroupByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a usersGroup by id
func (h *usersGroupHandler) GetByID(ctx context.Context, req *netbox_goV1.GetUsersGroupByIDRequest) (*netbox_goV1.GetUsersGroupByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of usersGroups by custom conditions
func (h *usersGroupHandler) List(ctx context.Context, req *netbox_goV1.ListUsersGroupRequest) (*netbox_goV1.ListUsersGroupReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete usersGroup by ids
func (h *usersGroupHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteUsersGroupByIDsRequest) (*netbox_goV1.DeleteUsersGroupByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a usersGroup by custom condition
func (h *usersGroupHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetUsersGroupByConditionRequest) (*netbox_goV1.GetUsersGroupByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get usersGroup by ids
func (h *usersGroupHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListUsersGroupByIDsRequest) (*netbox_goV1.ListUsersGroupByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of usersGroups by last id
func (h *usersGroupHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListUsersGroupByLastIDRequest) (*netbox_goV1.ListUsersGroupByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
