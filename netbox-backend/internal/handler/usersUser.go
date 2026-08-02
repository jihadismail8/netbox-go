package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.UsersUserLogicer = (*usersUserHandler)(nil)

type usersUserHandler struct {
	server netbox_goV1.UsersUserServer
}

// NewUsersUserHandler create a handler
func NewUsersUserHandler() netbox_goV1.UsersUserLogicer {
	return &usersUserHandler{
		server: service.NewUsersUserServer(),
	}
}

// Create a new usersUser
func (h *usersUserHandler) Create(ctx context.Context, req *netbox_goV1.CreateUsersUserRequest) (*netbox_goV1.CreateUsersUserReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a usersUser by id
func (h *usersUserHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteUsersUserByIDRequest) (*netbox_goV1.DeleteUsersUserByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a usersUser by id
func (h *usersUserHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateUsersUserByIDRequest) (*netbox_goV1.UpdateUsersUserByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a usersUser by id
func (h *usersUserHandler) GetByID(ctx context.Context, req *netbox_goV1.GetUsersUserByIDRequest) (*netbox_goV1.GetUsersUserByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of usersUsers by custom conditions
func (h *usersUserHandler) List(ctx context.Context, req *netbox_goV1.ListUsersUserRequest) (*netbox_goV1.ListUsersUserReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete usersUser by ids
func (h *usersUserHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteUsersUserByIDsRequest) (*netbox_goV1.DeleteUsersUserByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a usersUser by custom condition
func (h *usersUserHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetUsersUserByConditionRequest) (*netbox_goV1.GetUsersUserByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get usersUser by ids
func (h *usersUserHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListUsersUserByIDsRequest) (*netbox_goV1.ListUsersUserByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of usersUsers by last id
func (h *usersUserHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListUsersUserByLastIDRequest) (*netbox_goV1.ListUsersUserByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
