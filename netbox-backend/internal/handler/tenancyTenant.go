package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.TenancyTenantLogicer = (*tenancyTenantHandler)(nil)

type tenancyTenantHandler struct {
	server netbox_goV1.TenancyTenantServer
}

// NewTenancyTenantHandler create a handler
func NewTenancyTenantHandler() netbox_goV1.TenancyTenantLogicer {
	return &tenancyTenantHandler{
		server: service.NewTenancyTenantServer(),
	}
}

// Create a new tenancyTenant
func (h *tenancyTenantHandler) Create(ctx context.Context, req *netbox_goV1.CreateTenancyTenantRequest) (*netbox_goV1.CreateTenancyTenantReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a tenancyTenant by id
func (h *tenancyTenantHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteTenancyTenantByIDRequest) (*netbox_goV1.DeleteTenancyTenantByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a tenancyTenant by id
func (h *tenancyTenantHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateTenancyTenantByIDRequest) (*netbox_goV1.UpdateTenancyTenantByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a tenancyTenant by id
func (h *tenancyTenantHandler) GetByID(ctx context.Context, req *netbox_goV1.GetTenancyTenantByIDRequest) (*netbox_goV1.GetTenancyTenantByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of tenancyTenants by custom conditions
func (h *tenancyTenantHandler) List(ctx context.Context, req *netbox_goV1.ListTenancyTenantRequest) (*netbox_goV1.ListTenancyTenantReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete tenancyTenant by ids
func (h *tenancyTenantHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteTenancyTenantByIDsRequest) (*netbox_goV1.DeleteTenancyTenantByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a tenancyTenant by custom condition
func (h *tenancyTenantHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetTenancyTenantByConditionRequest) (*netbox_goV1.GetTenancyTenantByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get tenancyTenant by ids
func (h *tenancyTenantHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListTenancyTenantByIDsRequest) (*netbox_goV1.ListTenancyTenantByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of tenancyTenants by last id
func (h *tenancyTenantHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListTenancyTenantByLastIDRequest) (*netbox_goV1.ListTenancyTenantByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
