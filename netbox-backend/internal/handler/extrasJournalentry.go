package handler

import (
	"context"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/service"
)

var _ netbox_goV1.ExtrasJournalentryLogicer = (*extrasJournalentryHandler)(nil)

type extrasJournalentryHandler struct {
	server netbox_goV1.ExtrasJournalentryServer
}

// NewExtrasJournalentryHandler create a handler
func NewExtrasJournalentryHandler() netbox_goV1.ExtrasJournalentryLogicer {
	return &extrasJournalentryHandler{
		server: service.NewExtrasJournalentryServer(),
	}
}

// Create a new extrasJournalentry
func (h *extrasJournalentryHandler) Create(ctx context.Context, req *netbox_goV1.CreateExtrasJournalentryRequest) (*netbox_goV1.CreateExtrasJournalentryReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a extrasJournalentry by id
func (h *extrasJournalentryHandler) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasJournalentryByIDRequest) (*netbox_goV1.DeleteExtrasJournalentryByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a extrasJournalentry by id
func (h *extrasJournalentryHandler) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasJournalentryByIDRequest) (*netbox_goV1.UpdateExtrasJournalentryByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

// GetByID get a extrasJournalentry by id
func (h *extrasJournalentryHandler) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasJournalentryByIDRequest) (*netbox_goV1.GetExtrasJournalentryByIDReply, error) {
	return h.server.GetByID(ctx, req)
}

// List get a paginated list of extrasJournalentrys by custom conditions
func (h *extrasJournalentryHandler) List(ctx context.Context, req *netbox_goV1.ListExtrasJournalentryRequest) (*netbox_goV1.ListExtrasJournalentryReply, error) {
	return h.server.List(ctx, req)
}

// DeleteByIDs batch delete extrasJournalentry by ids
func (h *extrasJournalentryHandler) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasJournalentryByIDsRequest) (*netbox_goV1.DeleteExtrasJournalentryByIDsReply, error) {
	return h.server.DeleteByIDs(ctx, req)
}

// GetByCondition get a extrasJournalentry by custom condition
func (h *extrasJournalentryHandler) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasJournalentryByConditionRequest) (*netbox_goV1.GetExtrasJournalentryByConditionReply, error) {
	return h.server.GetByCondition(ctx, req)
}

// ListByIDs batch get extrasJournalentry by ids
func (h *extrasJournalentryHandler) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasJournalentryByIDsRequest) (*netbox_goV1.ListExtrasJournalentryByIDsReply, error) {
	return h.server.ListByIDs(ctx, req)
}

// ListByLastID get a paginated list of extrasJournalentrys by last id
func (h *extrasJournalentryHandler) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasJournalentryByLastIDRequest) (*netbox_goV1.ListExtrasJournalentryByLastIDReply, error) {
	return h.server.ListByLastID(ctx, req)
}
