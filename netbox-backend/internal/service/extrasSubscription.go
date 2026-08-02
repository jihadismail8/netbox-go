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
		netbox_goV1.RegisterExtrasSubscriptionServer(server, NewExtrasSubscriptionServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasSubscriptionServer = (*extrasSubscription)(nil)
var _ time.Time

type extrasSubscription struct {
	netbox_goV1.UnimplementedExtrasSubscriptionServer

	iDao dao.ExtrasSubscriptionDao
}

// NewExtrasSubscriptionServer create a new service
func NewExtrasSubscriptionServer() netbox_goV1.ExtrasSubscriptionServer {
	return &extrasSubscription{
		iDao: dao.NewExtrasSubscriptionDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasSubscriptionCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasSubscription
func (s *extrasSubscription) Create(ctx context.Context, req *netbox_goV1.CreateExtrasSubscriptionRequest) (*netbox_goV1.CreateExtrasSubscriptionReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasSubscription{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasSubscription.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasSubscription", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasSubscriptionReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasSubscription by id
func (s *extrasSubscription) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasSubscriptionByIDRequest) (*netbox_goV1.DeleteExtrasSubscriptionByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasSubscriptionByIDReply{}, nil
}

// UpdateByID update a extrasSubscription by id
func (s *extrasSubscription) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasSubscriptionByIDRequest) (*netbox_goV1.UpdateExtrasSubscriptionByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasSubscription{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasSubscription.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasSubscription", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasSubscriptionByIDReply{}, nil
}

// GetByID get a extrasSubscription by id
func (s *extrasSubscription) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasSubscriptionByIDRequest) (*netbox_goV1.GetExtrasSubscriptionByIDReply, error) {
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

	data, err := convertExtrasSubscription(record)
	if err != nil {
		logger.Warn("convertExtrasSubscription error", logger.Err(err), logger.Any("extrasSubscription", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasSubscription.Err()
	}

	return &netbox_goV1.GetExtrasSubscriptionByIDReply{ExtrasSubscription: data}, nil
}

// List get a paginated list of extrasSubscriptions by custom conditions
func (s *extrasSubscription) List(ctx context.Context, req *netbox_goV1.ListExtrasSubscriptionRequest) (*netbox_goV1.ListExtrasSubscriptionReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasSubscription.Err()
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

	extrasSubscriptions := []*netbox_goV1.ExtrasSubscription{}
	for _, record := range records {
		data, err := convertExtrasSubscription(record)
		if err != nil {
			logger.Warn("convertExtrasSubscription error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasSubscriptions = append(extrasSubscriptions, data)
	}

	return &netbox_goV1.ListExtrasSubscriptionReply{
		Total:               total,
		ExtrasSubscriptions: extrasSubscriptions,
	}, nil
}

// DeleteByIDs batch delete extrasSubscription by ids
func (s *extrasSubscription) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasSubscriptionByIDsRequest) (*netbox_goV1.DeleteExtrasSubscriptionByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasSubscriptionByIDsReply{}, nil
}

// GetByCondition get a extrasSubscription by custom condition
func (s *extrasSubscription) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasSubscriptionByConditionRequest) (*netbox_goV1.GetExtrasSubscriptionByConditionReply, error) {
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

	data, err := convertExtrasSubscription(record)
	if err != nil {
		logger.Warn("convertExtrasSubscription error", logger.Err(err), logger.Any("extrasSubscription", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasSubscription.Err()
	}

	return &netbox_goV1.GetExtrasSubscriptionByConditionReply{
		ExtrasSubscription: data,
	}, nil
}

// ListByIDs batch get extrasSubscription by ids
func (s *extrasSubscription) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasSubscriptionByIDsRequest) (*netbox_goV1.ListExtrasSubscriptionByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasSubscriptionMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasSubscriptions := []*netbox_goV1.ExtrasSubscription{}
	for _, id := range req.Ids {
		if v, ok := extrasSubscriptionMap[id]; ok {
			record, err := convertExtrasSubscription(v)
			if err != nil {
				logger.Warn("convertExtrasSubscription error", logger.Err(err), logger.Any("extrasSubscription", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasSubscriptions = append(extrasSubscriptions, record)
		}
	}

	return &netbox_goV1.ListExtrasSubscriptionByIDsReply{ExtrasSubscriptions: extrasSubscriptions}, nil
}

// ListByLastID get a paginated list of extrasSubscriptions by last id
func (s *extrasSubscription) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasSubscriptionByLastIDRequest) (*netbox_goV1.ListExtrasSubscriptionByLastIDReply, error) {
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

	extrasSubscriptions := []*netbox_goV1.ExtrasSubscription{}
	for _, record := range records {
		data, err := convertExtrasSubscription(record)
		if err != nil {
			logger.Warn("convertExtrasSubscription error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasSubscriptions = append(extrasSubscriptions, data)
	}

	return &netbox_goV1.ListExtrasSubscriptionByLastIDReply{
		ExtrasSubscriptions: extrasSubscriptions,
	}, nil
}

func convertExtrasSubscription(record *model.ExtrasSubscription) (*netbox_goV1.ExtrasSubscription, error) {
	value := &netbox_goV1.ExtrasSubscription{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
