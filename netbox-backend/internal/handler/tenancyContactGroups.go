package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.TenancyContactGroupsLogicer = (*tenancyContactGroupsHandler)(nil)

type tenancyContactGroupsHandler struct {
	server netbox_goV1.TenancyContactGroupsServer
}

// NewTenancyContactGroupsHandler create a handler
func NewTenancyContactGroupsHandler() netbox_goV1.TenancyContactGroupsLogicer {
	return &tenancyContactGroupsHandler{
		server: service.NewTenancyContactGroupsServer(),
	}
}

// Create a new tenancyContactGroups
func (h *tenancyContactGroupsHandler) Create(ctx context.Context, req *netbox_goV1.CreateTenancyContactGroupsRequest) (*netbox_goV1.CreateTenancyContactGroupsReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a tenancyContactGroups by id
func (h *tenancyContactGroupsHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteTenancyContactGroupsByIDRequest) (*netbox_goV1.DeleteTenancyContactGroupsByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a tenancyContactGroups by id
func (h *tenancyContactGroupsHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateTenancyContactGroupsByIDRequest) (*netbox_goV1.UpdateTenancyContactGroupsByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a tenancyContactGroups by id
func (h *tenancyContactGroupsHandler) GetByID(ctx context.Context, req *netbox_goV1.GetTenancyContactGroupsByIDRequest) (*netbox_goV1.GetTenancyContactGroupsByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of tenancyContactGroupss by custom conditions
func (h *tenancyContactGroupsHandler) List(ctx context.Context, req *netbox_goV1.ListTenancyContactGroupsRequest) (*netbox_goV1.ListTenancyContactGroupsReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete tenancyContactGroups by ids
func (h *tenancyContactGroupsHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteTenancyContactGroupsByIDsRequest) (*netbox_goV1.DeleteTenancyContactGroupsByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a tenancyContactGroups by custom condition
func (h *tenancyContactGroupsHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetTenancyContactGroupsByConditionRequest) (*netbox_goV1.GetTenancyContactGroupsByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get tenancyContactGroups by ids
func (h *tenancyContactGroupsHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListTenancyContactGroupsByIDsRequest) (*netbox_goV1.ListTenancyContactGroupsByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of tenancyContactGroupss by last id
func (h *tenancyContactGroupsHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListTenancyContactGroupsByLastIDRequest) (*netbox_goV1.ListTenancyContactGroupsByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
