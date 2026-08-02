package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.TenancyContactassignmentLogicer = (*tenancyContactassignmentHandler)(nil)

type tenancyContactassignmentHandler struct {
	server netbox_goV1.TenancyContactassignmentServer
}

// NewTenancyContactassignmentHandler create a handler
func NewTenancyContactassignmentHandler() netbox_goV1.TenancyContactassignmentLogicer {
	return &tenancyContactassignmentHandler{
		server: service.NewTenancyContactassignmentServer(),
	}
}

// Create a new tenancyContactassignment
func (h *tenancyContactassignmentHandler) Create(ctx context.Context, req *netbox_goV1.CreateTenancyContactassignmentRequest) (*netbox_goV1.CreateTenancyContactassignmentReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a tenancyContactassignment by id
func (h *tenancyContactassignmentHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteTenancyContactassignmentByIDRequest) (*netbox_goV1.DeleteTenancyContactassignmentByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a tenancyContactassignment by id
func (h *tenancyContactassignmentHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateTenancyContactassignmentByIDRequest) (*netbox_goV1.UpdateTenancyContactassignmentByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a tenancyContactassignment by id
func (h *tenancyContactassignmentHandler) GetByID(ctx context.Context, req *netbox_goV1.GetTenancyContactassignmentByIDRequest) (*netbox_goV1.GetTenancyContactassignmentByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of tenancyContactassignments by custom conditions
func (h *tenancyContactassignmentHandler) List(ctx context.Context, req *netbox_goV1.ListTenancyContactassignmentRequest) (*netbox_goV1.ListTenancyContactassignmentReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete tenancyContactassignment by ids
func (h *tenancyContactassignmentHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteTenancyContactassignmentByIDsRequest) (*netbox_goV1.DeleteTenancyContactassignmentByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a tenancyContactassignment by custom condition
func (h *tenancyContactassignmentHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetTenancyContactassignmentByConditionRequest) (*netbox_goV1.GetTenancyContactassignmentByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get tenancyContactassignment by ids
func (h *tenancyContactassignmentHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListTenancyContactassignmentByIDsRequest) (*netbox_goV1.ListTenancyContactassignmentByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of tenancyContactassignments by last id
func (h *tenancyContactassignmentHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListTenancyContactassignmentByLastIDRequest) (*netbox_goV1.ListTenancyContactassignmentByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
