package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.TenancyContactgroupLogicer = (*tenancyContactgroupHandler)(nil)

type tenancyContactgroupHandler struct {
	server netbox_goV1.TenancyContactgroupServer
}

// NewTenancyContactgroupHandler create a handler
func NewTenancyContactgroupHandler() netbox_goV1.TenancyContactgroupLogicer {
	return &tenancyContactgroupHandler{
		server: service.NewTenancyContactgroupServer(),
	}
}

// Create a new tenancyContactgroup
func (h *tenancyContactgroupHandler) Create(ctx context.Context, req *netbox_goV1.CreateTenancyContactgroupRequest) (*netbox_goV1.CreateTenancyContactgroupReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a tenancyContactgroup by id
func (h *tenancyContactgroupHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteTenancyContactgroupByIDRequest) (*netbox_goV1.DeleteTenancyContactgroupByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a tenancyContactgroup by id
func (h *tenancyContactgroupHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateTenancyContactgroupByIDRequest) (*netbox_goV1.UpdateTenancyContactgroupByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a tenancyContactgroup by id
func (h *tenancyContactgroupHandler) GetByID(ctx context.Context, req *netbox_goV1.GetTenancyContactgroupByIDRequest) (*netbox_goV1.GetTenancyContactgroupByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of tenancyContactgroups by custom conditions
func (h *tenancyContactgroupHandler) List(ctx context.Context, req *netbox_goV1.ListTenancyContactgroupRequest) (*netbox_goV1.ListTenancyContactgroupReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete tenancyContactgroup by ids
func (h *tenancyContactgroupHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteTenancyContactgroupByIDsRequest) (*netbox_goV1.DeleteTenancyContactgroupByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a tenancyContactgroup by custom condition
func (h *tenancyContactgroupHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetTenancyContactgroupByConditionRequest) (*netbox_goV1.GetTenancyContactgroupByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get tenancyContactgroup by ids
func (h *tenancyContactgroupHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListTenancyContactgroupByIDsRequest) (*netbox_goV1.ListTenancyContactgroupByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of tenancyContactgroups by last id
func (h *tenancyContactgroupHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListTenancyContactgroupByLastIDRequest) (*netbox_goV1.ListTenancyContactgroupByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
