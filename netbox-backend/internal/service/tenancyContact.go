package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/go-dev-frame/sponge/pkg/copier"
	"github.com/go-dev-frame/sponge/pkg/grpc/interceptor"
	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/sgorm/query"

	netbox_goV1 "netbox-go/api/netbox_go/v1"
	"netbox-go/internal/cache"
	"netbox-go/internal/dao"
	"netbox-go/internal/database"
	"netbox-go/internal/ecode"
	"netbox-go/internal/model"
)

func init() {
	registerFns = append(registerFns, func(server *grpc.Server) {
		netbox_goV1.RegisterTenancyContactServer(server, NewTenancyContactServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.TenancyContactServer = (*tenancyContact)(nil)
var _ time.Time

type tenancyContact struct {
	netbox_goV1.UnimplementedTenancyContactServer

	iDao dao.TenancyContactDao
}

// NewTenancyContactServer create a new service
func NewTenancyContactServer() netbox_goV1.TenancyContactServer {
	return &tenancyContact{
		iDao: dao.NewTenancyContactDao(
			database.GetDB(), // db driver is postgresql
			cache.NewTenancyContactCache(database.GetCacheType()),
		),
	}
}

// Create a new tenancyContact
func (s *tenancyContact) Create(ctx context.Context, req *netbox_goV1.CreateTenancyContactRequest) (*netbox_goV1.CreateTenancyContactReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.TenancyContact{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateTenancyContact.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("tenancyContact", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateTenancyContactReply{Id: record.ID}, nil
}

// DeleteByID delete a tenancyContact by id
func (s *tenancyContact) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteTenancyContactByIDRequest) (*netbox_goV1.DeleteTenancyContactByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	err = s.iDao.DeleteByID(ctx, req.Id)
	if err != nil {
		logger.Error("DeleteByID error", logger.Err(err), logger.Any("id", req.Id), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.DeleteTenancyContactByIDReply{}, nil
}

// UpdateByID update a tenancyContact by id
func (s *tenancyContact) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateTenancyContactByIDRequest) (*netbox_goV1.UpdateTenancyContactByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.TenancyContact{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDTenancyContact.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("tenancyContact", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateTenancyContactByIDReply{}, nil
}

// GetByID get a tenancyContact by id
func (s *tenancyContact) GetByID(ctx context.Context, req *netbox_goV1.GetTenancyContactByIDRequest) (*netbox_goV1.GetTenancyContactByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record, err := s.iDao.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, database.ErrRecordNotFound) {
			logger.Warn("GetByID error", logger.Err(err), logger.Any("id", req.Id), interceptor.ServerCtxRequestIDField(ctx))
			return nil, ecode.StatusNotFound.Err()
		}
		logger.Error("GetByID error", logger.Err(err), logger.Any("id", req.Id), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	data, err := convertTenancyContact(record)
	if err != nil {
		logger.Warn("convertTenancyContact error", logger.Err(err), logger.Any("tenancyContact", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDTenancyContact.Err()
	}

	return &netbox_goV1.GetTenancyContactByIDReply{TenancyContact: data}, nil
}

// List get a paginated list of tenancyContacts by custom conditions
func (s *tenancyContact) List(ctx context.Context, req *netbox_goV1.ListTenancyContactRequest) (*netbox_goV1.ListTenancyContactReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListTenancyContact.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	records, total, err := s.iDao.GetByColumns(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "query params error:") {
			logger.Warn("GetByColumns error", logger.Err(err), logger.Any("params", params), interceptor.ServerCtxRequestIDField(ctx))
			return nil, ecode.StatusInvalidParams.Err()
		}
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("params", params), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	tenancyContacts := []*netbox_goV1.TenancyContact{}
	for _, record := range records {
		data, err := convertTenancyContact(record)
		if err != nil {
			logger.Warn("convertTenancyContact error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		tenancyContacts = append(tenancyContacts, data)
	}

	return &netbox_goV1.ListTenancyContactReply{
		Total:           total,
		TenancyContacts: tenancyContacts,
	}, nil
}

// DeleteByIDs batch delete tenancyContact by ids
func (s *tenancyContact) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteTenancyContactByIDsRequest) (*netbox_goV1.DeleteTenancyContactByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	err = s.iDao.DeleteByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("DeleteByID error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.DeleteTenancyContactByIDsReply{}, nil
}

// GetByCondition get a tenancyContact by custom condition
func (s *tenancyContact) GetByCondition(ctx context.Context, req *netbox_goV1.GetTenancyContactByConditionRequest) (*netbox_goV1.GetTenancyContactByConditionReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	conditions := &query.Conditions{}
	for _, v := range req.Conditions.GetColumns() {
		column := query.Column{}
		_ = copier.Copy(&column, v)
		conditions.Columns = append(conditions.Columns, column)
	}
	err = conditions.CheckValid()
	if err != nil {
		logger.Warn("Parameters error", logger.Err(err), logger.Any("conditions", conditions), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}

	record, err := s.iDao.GetByCondition(ctx, conditions)
	if err != nil {
		if errors.Is(err, database.ErrRecordNotFound) {
			logger.Warn("GetByCondition error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
			return nil, ecode.StatusNotFound.Err()
		}
		logger.Error("GetByCondition error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	data, err := convertTenancyContact(record)
	if err != nil {
		logger.Warn("convertTenancyContact error", logger.Err(err), logger.Any("tenancyContact", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionTenancyContact.Err()
	}

	return &netbox_goV1.GetTenancyContactByConditionReply{
		TenancyContact: data,
	}, nil
}

// ListByIDs batch get tenancyContact by ids
func (s *tenancyContact) ListByIDs(ctx context.Context, req *netbox_goV1.ListTenancyContactByIDsRequest) (*netbox_goV1.ListTenancyContactByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	tenancyContactMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	tenancyContacts := []*netbox_goV1.TenancyContact{}
	for _, id := range req.Ids {
		if v, ok := tenancyContactMap[id]; ok {
			record, err := convertTenancyContact(v)
			if err != nil {
				logger.Warn("convertTenancyContact error", logger.Err(err), logger.Any("tenancyContact", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			tenancyContacts = append(tenancyContacts, record)
		}
	}

	return &netbox_goV1.ListTenancyContactByIDsReply{TenancyContacts: tenancyContacts}, nil
}

// ListByLastID get a paginated list of tenancyContacts by last id
func (s *tenancyContact) ListByLastID(ctx context.Context, req *netbox_goV1.ListTenancyContactByLastIDRequest) (*netbox_goV1.ListTenancyContactByLastIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.CtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	if req.LastID == 0 {
		req.LastID = math.MaxInt32
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	records, err := s.iDao.GetByLastID(ctx, req.LastID, int(req.Limit), req.Sort)
	if err != nil {
		logger.Error("ListByLastID error", logger.Err(err), interceptor.CtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	tenancyContacts := []*netbox_goV1.TenancyContact{}
	for _, record := range records {
		data, err := convertTenancyContact(record)
		if err != nil {
			logger.Warn("convertTenancyContact error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		tenancyContacts = append(tenancyContacts, data)
	}

	return &netbox_goV1.ListTenancyContactByLastIDReply{
		TenancyContacts: tenancyContacts,
	}, nil
}

func convertTenancyContact(record *model.TenancyContact) (*netbox_goV1.TenancyContact, error) {
	value := &netbox_goV1.TenancyContact{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
