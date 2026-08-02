package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.UsersObjectpermissionObjectTypesLogicer = (*usersObjectpermissionObjectTypesHandler)(nil)

type usersObjectpermissionObjectTypesHandler struct {
	server netbox_goV1.UsersObjectpermissionObjectTypesServer
}

// NewUsersObjectpermissionObjectTypesHandler create a handler
func NewUsersObjectpermissionObjectTypesHandler() netbox_goV1.UsersObjectpermissionObjectTypesLogicer {
	return &usersObjectpermissionObjectTypesHandler{
		server: service.NewUsersObjectpermissionObjectTypesServer(),
	}
}

// Create a new usersObjectpermissionObjectTypes
func (h *usersObjectpermissionObjectTypesHandler) Create(ctx context.Context, req *netbox_goV1.CreateUsersObjectpermissionObjectTypesRequest) (*netbox_goV1.CreateUsersObjectpermissionObjectTypesReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a usersObjectpermissionObjectTypes by id
func (h *usersObjectpermissionObjectTypesHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteUsersObjectpermissionObjectTypesByIDRequest) (*netbox_goV1.DeleteUsersObjectpermissionObjectTypesByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a usersObjectpermissionObjectTypes by id
func (h *usersObjectpermissionObjectTypesHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateUsersObjectpermissionObjectTypesByIDRequest) (*netbox_goV1.UpdateUsersObjectpermissionObjectTypesByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a usersObjectpermissionObjectTypes by id
func (h *usersObjectpermissionObjectTypesHandler) GetByID(ctx context.Context, req *netbox_goV1.GetUsersObjectpermissionObjectTypesByIDRequest) (*netbox_goV1.GetUsersObjectpermissionObjectTypesByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of usersObjectpermissionObjectTypess by custom conditions
func (h *usersObjectpermissionObjectTypesHandler) List(ctx context.Context, req *netbox_goV1.ListUsersObjectpermissionObjectTypesRequest) (*netbox_goV1.ListUsersObjectpermissionObjectTypesReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete usersObjectpermissionObjectTypes by ids
func (h *usersObjectpermissionObjectTypesHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteUsersObjectpermissionObjectTypesByIDsRequest) (*netbox_goV1.DeleteUsersObjectpermissionObjectTypesByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a usersObjectpermissionObjectTypes by custom condition
func (h *usersObjectpermissionObjectTypesHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetUsersObjectpermissionObjectTypesByConditionRequest) (*netbox_goV1.GetUsersObjectpermissionObjectTypesByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get usersObjectpermissionObjectTypes by ids
func (h *usersObjectpermissionObjectTypesHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListUsersObjectpermissionObjectTypesByIDsRequest) (*netbox_goV1.ListUsersObjectpermissionObjectTypesByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of usersObjectpermissionObjectTypess by last id
func (h *usersObjectpermissionObjectTypesHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListUsersObjectpermissionObjectTypesByLastIDRequest) (*netbox_goV1.ListUsersObjectpermissionObjectTypesByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
