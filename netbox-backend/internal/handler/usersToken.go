package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.UsersTokenLogicer = (*usersTokenHandler)(nil)

type usersTokenHandler struct {
	server netbox_goV1.UsersTokenServer
}

// NewUsersTokenHandler create a handler
func NewUsersTokenHandler() netbox_goV1.UsersTokenLogicer {
	return &usersTokenHandler{
		server: service.NewUsersTokenServer(),
	}
}

// Create a new usersToken
func (h *usersTokenHandler) Create(ctx context.Context, req *netbox_goV1.CreateUsersTokenRequest) (*netbox_goV1.CreateUsersTokenReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a usersToken by id
func (h *usersTokenHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteUsersTokenByIDRequest) (*netbox_goV1.DeleteUsersTokenByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a usersToken by id
func (h *usersTokenHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateUsersTokenByIDRequest) (*netbox_goV1.UpdateUsersTokenByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a usersToken by id
func (h *usersTokenHandler) GetByID(ctx context.Context, req *netbox_goV1.GetUsersTokenByIDRequest) (*netbox_goV1.GetUsersTokenByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of usersTokens by custom conditions
func (h *usersTokenHandler) List(ctx context.Context, req *netbox_goV1.ListUsersTokenRequest) (*netbox_goV1.ListUsersTokenReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete usersToken by ids
func (h *usersTokenHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteUsersTokenByIDsRequest) (*netbox_goV1.DeleteUsersTokenByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a usersToken by custom condition
func (h *usersTokenHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetUsersTokenByConditionRequest) (*netbox_goV1.GetUsersTokenByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get usersToken by ids
func (h *usersTokenHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListUsersTokenByIDsRequest) (*netbox_goV1.ListUsersTokenByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of usersTokens by last id
func (h *usersTokenHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListUsersTokenByLastIDRequest) (*netbox_goV1.ListUsersTokenByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
