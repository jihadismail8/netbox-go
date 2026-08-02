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
		netbox_goV1.RegisterExtrasJournalentryServer(server, NewExtrasJournalentryServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasJournalentryServer = (*extrasJournalentry)(nil)
var _ time.Time

type extrasJournalentry struct {
	netbox_goV1.UnimplementedExtrasJournalentryServer

	iDao dao.ExtrasJournalentryDao
}

// NewExtrasJournalentryServer create a new service
func NewExtrasJournalentryServer() netbox_goV1.ExtrasJournalentryServer {
	return &extrasJournalentry{
		iDao: dao.NewExtrasJournalentryDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasJournalentryCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasJournalentry
func (s *extrasJournalentry) Create(ctx context.Context, req *netbox_goV1.CreateExtrasJournalentryRequest) (*netbox_goV1.CreateExtrasJournalentryReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasJournalentry{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasJournalentry.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasJournalentry", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasJournalentryReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasJournalentry by id
func (s *extrasJournalentry) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasJournalentryByIDRequest) (*netbox_goV1.DeleteExtrasJournalentryByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasJournalentryByIDReply{}, nil
}

// UpdateByID update a extrasJournalentry by id
func (s *extrasJournalentry) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasJournalentryByIDRequest) (*netbox_goV1.UpdateExtrasJournalentryByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasJournalentry{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasJournalentry.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasJournalentry", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasJournalentryByIDReply{}, nil
}

// GetByID get a extrasJournalentry by id
func (s *extrasJournalentry) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasJournalentryByIDRequest) (*netbox_goV1.GetExtrasJournalentryByIDReply, error) {
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

	data, err := convertExtrasJournalentry(record)
	if err != nil {
		logger.Warn("convertExtrasJournalentry error", logger.Err(err), logger.Any("extrasJournalentry", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasJournalentry.Err()
	}

	return &netbox_goV1.GetExtrasJournalentryByIDReply{ExtrasJournalentry: data}, nil
}

// List get a paginated list of extrasJournalentrys by custom conditions
func (s *extrasJournalentry) List(ctx context.Context, req *netbox_goV1.ListExtrasJournalentryRequest) (*netbox_goV1.ListExtrasJournalentryReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasJournalentry.Err()
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

	extrasJournalentrys := []*netbox_goV1.ExtrasJournalentry{}
	for _, record := range records {
		data, err := convertExtrasJournalentry(record)
		if err != nil {
			logger.Warn("convertExtrasJournalentry error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasJournalentrys = append(extrasJournalentrys, data)
	}

	return &netbox_goV1.ListExtrasJournalentryReply{
		Total:               total,
		ExtrasJournalentrys: extrasJournalentrys,
	}, nil
}

// DeleteByIDs batch delete extrasJournalentry by ids
func (s *extrasJournalentry) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasJournalentryByIDsRequest) (*netbox_goV1.DeleteExtrasJournalentryByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasJournalentryByIDsReply{}, nil
}

// GetByCondition get a extrasJournalentry by custom condition
func (s *extrasJournalentry) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasJournalentryByConditionRequest) (*netbox_goV1.GetExtrasJournalentryByConditionReply, error) {
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

	data, err := convertExtrasJournalentry(record)
	if err != nil {
		logger.Warn("convertExtrasJournalentry error", logger.Err(err), logger.Any("extrasJournalentry", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasJournalentry.Err()
	}

	return &netbox_goV1.GetExtrasJournalentryByConditionReply{
		ExtrasJournalentry: data,
	}, nil
}

// ListByIDs batch get extrasJournalentry by ids
func (s *extrasJournalentry) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasJournalentryByIDsRequest) (*netbox_goV1.ListExtrasJournalentryByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasJournalentryMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasJournalentrys := []*netbox_goV1.ExtrasJournalentry{}
	for _, id := range req.Ids {
		if v, ok := extrasJournalentryMap[id]; ok {
			record, err := convertExtrasJournalentry(v)
			if err != nil {
				logger.Warn("convertExtrasJournalentry error", logger.Err(err), logger.Any("extrasJournalentry", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasJournalentrys = append(extrasJournalentrys, record)
		}
	}

	return &netbox_goV1.ListExtrasJournalentryByIDsReply{ExtrasJournalentrys: extrasJournalentrys}, nil
}

// ListByLastID get a paginated list of extrasJournalentrys by last id
func (s *extrasJournalentry) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasJournalentryByLastIDRequest) (*netbox_goV1.ListExtrasJournalentryByLastIDReply, error) {
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

	extrasJournalentrys := []*netbox_goV1.ExtrasJournalentry{}
	for _, record := range records {
		data, err := convertExtrasJournalentry(record)
		if err != nil {
			logger.Warn("convertExtrasJournalentry error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasJournalentrys = append(extrasJournalentrys, data)
	}

	return &netbox_goV1.ListExtrasJournalentryByLastIDReply{
		ExtrasJournalentrys: extrasJournalentrys,
	}, nil
}

func convertExtrasJournalentry(record *model.ExtrasJournalentry) (*netbox_goV1.ExtrasJournalentry, error) {
	value := &netbox_goV1.ExtrasJournalentry{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
