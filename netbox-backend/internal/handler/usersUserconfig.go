package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.UsersUserconfigLogicer = (*usersUserconfigHandler)(nil)

type usersUserconfigHandler struct {
	server netbox_goV1.UsersUserconfigServer
}

// NewUsersUserconfigHandler create a handler
func NewUsersUserconfigHandler() netbox_goV1.UsersUserconfigLogicer {
	return &usersUserconfigHandler{
		server: service.NewUsersUserconfigServer(),
	}
}

// Create a new usersUserconfig
func (h *usersUserconfigHandler) Create(ctx context.Context, req *netbox_goV1.CreateUsersUserconfigRequest) (*netbox_goV1.CreateUsersUserconfigReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a usersUserconfig by id
func (h *usersUserconfigHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteUsersUserconfigByIDRequest) (*netbox_goV1.DeleteUsersUserconfigByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a usersUserconfig by id
func (h *usersUserconfigHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateUsersUserconfigByIDRequest) (*netbox_goV1.UpdateUsersUserconfigByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a usersUserconfig by id
func (h *usersUserconfigHandler) GetByID(ctx context.Context, req *netbox_goV1.GetUsersUserconfigByIDRequest) (*netbox_goV1.GetUsersUserconfigByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of usersUserconfigs by custom conditions
func (h *usersUserconfigHandler) List(ctx context.Context, req *netbox_goV1.ListUsersUserconfigRequest) (*netbox_goV1.ListUsersUserconfigReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete usersUserconfig by ids
func (h *usersUserconfigHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteUsersUserconfigByIDsRequest) (*netbox_goV1.DeleteUsersUserconfigByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a usersUserconfig by custom condition
func (h *usersUserconfigHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetUsersUserconfigByConditionRequest) (*netbox_goV1.GetUsersUserconfigByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get usersUserconfig by ids
func (h *usersUserconfigHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListUsersUserconfigByIDsRequest) (*netbox_goV1.ListUsersUserconfigByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of usersUserconfigs by last id
func (h *usersUserconfigHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListUsersUserconfigByLastIDRequest) (*netbox_goV1.ListUsersUserconfigByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
