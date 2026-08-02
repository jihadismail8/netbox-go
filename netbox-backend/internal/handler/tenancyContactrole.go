package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.TenancyContactroleLogicer = (*tenancyContactroleHandler)(nil)

type tenancyContactroleHandler struct {
	server netbox_goV1.TenancyContactroleServer
}

// NewTenancyContactroleHandler create a handler
func NewTenancyContactroleHandler() netbox_goV1.TenancyContactroleLogicer {
	return &tenancyContactroleHandler{
		server: service.NewTenancyContactroleServer(),
	}
}

// Create a new tenancyContactrole
func (h *tenancyContactroleHandler) Create(ctx context.Context, req *netbox_goV1.CreateTenancyContactroleRequest) (*netbox_goV1.CreateTenancyContactroleReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a tenancyContactrole by id
func (h *tenancyContactroleHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteTenancyContactroleByIDRequest) (*netbox_goV1.DeleteTenancyContactroleByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a tenancyContactrole by id
func (h *tenancyContactroleHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateTenancyContactroleByIDRequest) (*netbox_goV1.UpdateTenancyContactroleByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a tenancyContactrole by id
func (h *tenancyContactroleHandler) GetByID(ctx context.Context, req *netbox_goV1.GetTenancyContactroleByIDRequest) (*netbox_goV1.GetTenancyContactroleByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of tenancyContactroles by custom conditions
func (h *tenancyContactroleHandler) List(ctx context.Context, req *netbox_goV1.ListTenancyContactroleRequest) (*netbox_goV1.ListTenancyContactroleReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete tenancyContactrole by ids
func (h *tenancyContactroleHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteTenancyContactroleByIDsRequest) (*netbox_goV1.DeleteTenancyContactroleByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a tenancyContactrole by custom condition
func (h *tenancyContactroleHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetTenancyContactroleByConditionRequest) (*netbox_goV1.GetTenancyContactroleByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get tenancyContactrole by ids
func (h *tenancyContactroleHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListTenancyContactroleByIDsRequest) (*netbox_goV1.ListTenancyContactroleByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of tenancyContactroles by last id
func (h *tenancyContactroleHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListTenancyContactroleByLastIDRequest) (*netbox_goV1.ListTenancyContactroleByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
