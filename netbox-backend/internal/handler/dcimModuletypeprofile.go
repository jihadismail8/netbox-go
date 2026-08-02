package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimModuletypeprofileLogicer = (*dcimModuletypeprofileHandler)(nil)

type dcimModuletypeprofileHandler struct {
	server netbox_goV1.DcimModuletypeprofileServer
}

// NewDcimModuletypeprofileHandler create a handler
func NewDcimModuletypeprofileHandler() netbox_goV1.DcimModuletypeprofileLogicer {
	return &dcimModuletypeprofileHandler{
		server: service.NewDcimModuletypeprofileServer(),
	}
}

// Create a new dcimModuletypeprofile
func (h *dcimModuletypeprofileHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimModuletypeprofileRequest) (*netbox_goV1.CreateDcimModuletypeprofileReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimModuletypeprofile by id
func (h *dcimModuletypeprofileHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimModuletypeprofileByIDRequest) (*netbox_goV1.DeleteDcimModuletypeprofileByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimModuletypeprofile by id
func (h *dcimModuletypeprofileHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimModuletypeprofileByIDRequest) (*netbox_goV1.UpdateDcimModuletypeprofileByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimModuletypeprofile by id
func (h *dcimModuletypeprofileHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimModuletypeprofileByIDRequest) (*netbox_goV1.GetDcimModuletypeprofileByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimModuletypeprofiles by custom conditions
func (h *dcimModuletypeprofileHandler) List(ctx context.Context, req *netbox_goV1.ListDcimModuletypeprofileRequest) (*netbox_goV1.ListDcimModuletypeprofileReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimModuletypeprofile by ids
func (h *dcimModuletypeprofileHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimModuletypeprofileByIDsRequest) (*netbox_goV1.DeleteDcimModuletypeprofileByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimModuletypeprofile by custom condition
func (h *dcimModuletypeprofileHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimModuletypeprofileByConditionRequest) (*netbox_goV1.GetDcimModuletypeprofileByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimModuletypeprofile by ids
func (h *dcimModuletypeprofileHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimModuletypeprofileByIDsRequest) (*netbox_goV1.ListDcimModuletypeprofileByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimModuletypeprofiles by last id
func (h *dcimModuletypeprofileHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimModuletypeprofileByLastIDRequest) (*netbox_goV1.ListDcimModuletypeprofileByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
