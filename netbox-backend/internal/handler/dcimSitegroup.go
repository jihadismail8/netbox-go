package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.DcimSitegroupLogicer = (*dcimSitegroupHandler)(nil)

type dcimSitegroupHandler struct {
	server netbox_goV1.DcimSitegroupServer
}

// NewDcimSitegroupHandler create a handler
func NewDcimSitegroupHandler() netbox_goV1.DcimSitegroupLogicer {
	return &dcimSitegroupHandler{
		server: service.NewDcimSitegroupServer(),
	}
}

// Create a new dcimSitegroup
func (h *dcimSitegroupHandler) Create(ctx context.Context, req *netbox_goV1.CreateDcimSitegroupRequest) (*netbox_goV1.CreateDcimSitegroupReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a dcimSitegroup by id
func (h *dcimSitegroupHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimSitegroupByIDRequest) (*netbox_goV1.DeleteDcimSitegroupByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a dcimSitegroup by id
func (h *dcimSitegroupHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimSitegroupByIDRequest) (*netbox_goV1.UpdateDcimSitegroupByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a dcimSitegroup by id
func (h *dcimSitegroupHandler) GetByID(ctx context.Context, req *netbox_goV1.GetDcimSitegroupByIDRequest) (*netbox_goV1.GetDcimSitegroupByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of dcimSitegroups by custom conditions
func (h *dcimSitegroupHandler) List(ctx context.Context, req *netbox_goV1.ListDcimSitegroupRequest) (*netbox_goV1.ListDcimSitegroupReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete dcimSitegroup by ids
func (h *dcimSitegroupHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimSitegroupByIDsRequest) (*netbox_goV1.DeleteDcimSitegroupByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a dcimSitegroup by custom condition
func (h *dcimSitegroupHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimSitegroupByConditionRequest) (*netbox_goV1.GetDcimSitegroupByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get dcimSitegroup by ids
func (h *dcimSitegroupHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimSitegroupByIDsRequest) (*netbox_goV1.ListDcimSitegroupByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of dcimSitegroups by last id
func (h *dcimSitegroupHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimSitegroupByLastIDRequest) (*netbox_goV1.ListDcimSitegroupByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
