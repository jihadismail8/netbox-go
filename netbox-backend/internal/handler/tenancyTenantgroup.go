package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.TenancyTenantgroupLogicer = (*tenancyTenantgroupHandler)(nil)

type tenancyTenantgroupHandler struct {
	server netbox_goV1.TenancyTenantgroupServer
}

// NewTenancyTenantgroupHandler create a handler
func NewTenancyTenantgroupHandler() netbox_goV1.TenancyTenantgroupLogicer {
	return &tenancyTenantgroupHandler{
		server: service.NewTenancyTenantgroupServer(),
	}
}

// Create a new tenancyTenantgroup
func (h *tenancyTenantgroupHandler) Create(ctx context.Context, req *netbox_goV1.CreateTenancyTenantgroupRequest) (*netbox_goV1.CreateTenancyTenantgroupReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a tenancyTenantgroup by id
func (h *tenancyTenantgroupHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteTenancyTenantgroupByIDRequest) (*netbox_goV1.DeleteTenancyTenantgroupByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a tenancyTenantgroup by id
func (h *tenancyTenantgroupHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateTenancyTenantgroupByIDRequest) (*netbox_goV1.UpdateTenancyTenantgroupByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a tenancyTenantgroup by id
func (h *tenancyTenantgroupHandler) GetByID(ctx context.Context, req *netbox_goV1.GetTenancyTenantgroupByIDRequest) (*netbox_goV1.GetTenancyTenantgroupByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of tenancyTenantgroups by custom conditions
func (h *tenancyTenantgroupHandler) List(ctx context.Context, req *netbox_goV1.ListTenancyTenantgroupRequest) (*netbox_goV1.ListTenancyTenantgroupReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete tenancyTenantgroup by ids
func (h *tenancyTenantgroupHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteTenancyTenantgroupByIDsRequest) (*netbox_goV1.DeleteTenancyTenantgroupByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a tenancyTenantgroup by custom condition
func (h *tenancyTenantgroupHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetTenancyTenantgroupByConditionRequest) (*netbox_goV1.GetTenancyTenantgroupByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get tenancyTenantgroup by ids
func (h *tenancyTenantgroupHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListTenancyTenantgroupByIDsRequest) (*netbox_goV1.ListTenancyTenantgroupByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of tenancyTenantgroups by last id
func (h *tenancyTenantgroupHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListTenancyTenantgroupByLastIDRequest) (*netbox_goV1.ListTenancyTenantgroupByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
