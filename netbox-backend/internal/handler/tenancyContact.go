package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.TenancyContactLogicer = (*tenancyContactHandler)(nil)

type tenancyContactHandler struct {
	server netbox_goV1.TenancyContactServer
}

// NewTenancyContactHandler create a handler
func NewTenancyContactHandler() netbox_goV1.TenancyContactLogicer {
	return &tenancyContactHandler{
		server: service.NewTenancyContactServer(),
	}
}

// Create a new tenancyContact
func (h *tenancyContactHandler) Create(ctx context.Context, req *netbox_goV1.CreateTenancyContactRequest) (*netbox_goV1.CreateTenancyContactReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a tenancyContact by id
func (h *tenancyContactHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteTenancyContactByIDRequest) (*netbox_goV1.DeleteTenancyContactByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a tenancyContact by id
func (h *tenancyContactHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateTenancyContactByIDRequest) (*netbox_goV1.UpdateTenancyContactByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a tenancyContact by id
func (h *tenancyContactHandler) GetByID(ctx context.Context, req *netbox_goV1.GetTenancyContactByIDRequest) (*netbox_goV1.GetTenancyContactByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of tenancyContacts by custom conditions
func (h *tenancyContactHandler) List(ctx context.Context, req *netbox_goV1.ListTenancyContactRequest) (*netbox_goV1.ListTenancyContactReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete tenancyContact by ids
func (h *tenancyContactHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteTenancyContactByIDsRequest) (*netbox_goV1.DeleteTenancyContactByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a tenancyContact by custom condition
func (h *tenancyContactHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetTenancyContactByConditionRequest) (*netbox_goV1.GetTenancyContactByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get tenancyContact by ids
func (h *tenancyContactHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListTenancyContactByIDsRequest) (*netbox_goV1.ListTenancyContactByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of tenancyContacts by last id
func (h *tenancyContactHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListTenancyContactByLastIDRequest) (*netbox_goV1.ListTenancyContactByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
