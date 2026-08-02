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
		netbox_goV1.RegisterExtrasNotificationgroupServer(server, NewExtrasNotificationgroupServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasNotificationgroupServer = (*extrasNotificationgroup)(nil)
var _ time.Time

type extrasNotificationgroup struct {
	netbox_goV1.UnimplementedExtrasNotificationgroupServer

	iDao dao.ExtrasNotificationgroupDao
}

// NewExtrasNotificationgroupServer create a new service
func NewExtrasNotificationgroupServer() netbox_goV1.ExtrasNotificationgroupServer {
	return &extrasNotificationgroup{
		iDao: dao.NewExtrasNotificationgroupDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasNotificationgroupCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasNotificationgroup
func (s *extrasNotificationgroup) Create(ctx context.Context, req *netbox_goV1.CreateExtrasNotificationgroupRequest) (*netbox_goV1.CreateExtrasNotificationgroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasNotificationgroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasNotificationgroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasNotificationgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasNotificationgroupReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasNotificationgroup by id
func (s *extrasNotificationgroup) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasNotificationgroupByIDRequest) (*netbox_goV1.DeleteExtrasNotificationgroupByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasNotificationgroupByIDReply{}, nil
}

// UpdateByID update a extrasNotificationgroup by id
func (s *extrasNotificationgroup) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasNotificationgroupByIDRequest) (*netbox_goV1.UpdateExtrasNotificationgroupByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasNotificationgroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasNotificationgroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasNotificationgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasNotificationgroupByIDReply{}, nil
}

// GetByID get a extrasNotificationgroup by id
func (s *extrasNotificationgroup) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasNotificationgroupByIDRequest) (*netbox_goV1.GetExtrasNotificationgroupByIDReply, error) {
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

	data, err := convertExtrasNotificationgroup(record)
	if err != nil {
		logger.Warn("convertExtrasNotificationgroup error", logger.Err(err), logger.Any("extrasNotificationgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasNotificationgroup.Err()
	}

	return &netbox_goV1.GetExtrasNotificationgroupByIDReply{ExtrasNotificationgroup: data}, nil
}

// List get a paginated list of extrasNotificationgroups by custom conditions
func (s *extrasNotificationgroup) List(ctx context.Context, req *netbox_goV1.ListExtrasNotificationgroupRequest) (*netbox_goV1.ListExtrasNotificationgroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasNotificationgroup.Err()
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

	extrasNotificationgroups := []*netbox_goV1.ExtrasNotificationgroup{}
	for _, record := range records {
		data, err := convertExtrasNotificationgroup(record)
		if err != nil {
			logger.Warn("convertExtrasNotificationgroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasNotificationgroups = append(extrasNotificationgroups, data)
	}

	return &netbox_goV1.ListExtrasNotificationgroupReply{
		Total:                    total,
		ExtrasNotificationgroups: extrasNotificationgroups,
	}, nil
}

// DeleteByIDs batch delete extrasNotificationgroup by ids
func (s *extrasNotificationgroup) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasNotificationgroupByIDsRequest) (*netbox_goV1.DeleteExtrasNotificationgroupByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasNotificationgroupByIDsReply{}, nil
}

// GetByCondition get a extrasNotificationgroup by custom condition
func (s *extrasNotificationgroup) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasNotificationgroupByConditionRequest) (*netbox_goV1.GetExtrasNotificationgroupByConditionReply, error) {
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

	data, err := convertExtrasNotificationgroup(record)
	if err != nil {
		logger.Warn("convertExtrasNotificationgroup error", logger.Err(err), logger.Any("extrasNotificationgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasNotificationgroup.Err()
	}

	return &netbox_goV1.GetExtrasNotificationgroupByConditionReply{
		ExtrasNotificationgroup: data,
	}, nil
}

// ListByIDs batch get extrasNotificationgroup by ids
func (s *extrasNotificationgroup) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasNotificationgroupByIDsRequest) (*netbox_goV1.ListExtrasNotificationgroupByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasNotificationgroupMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasNotificationgroups := []*netbox_goV1.ExtrasNotificationgroup{}
	for _, id := range req.Ids {
		if v, ok := extrasNotificationgroupMap[id]; ok {
			record, err := convertExtrasNotificationgroup(v)
			if err != nil {
				logger.Warn("convertExtrasNotificationgroup error", logger.Err(err), logger.Any("extrasNotificationgroup", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasNotificationgroups = append(extrasNotificationgroups, record)
		}
	}

	return &netbox_goV1.ListExtrasNotificationgroupByIDsReply{ExtrasNotificationgroups: extrasNotificationgroups}, nil
}

// ListByLastID get a paginated list of extrasNotificationgroups by last id
func (s *extrasNotificationgroup) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasNotificationgroupByLastIDRequest) (*netbox_goV1.ListExtrasNotificationgroupByLastIDReply, error) {
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

	extrasNotificationgroups := []*netbox_goV1.ExtrasNotificationgroup{}
	for _, record := range records {
		data, err := convertExtrasNotificationgroup(record)
		if err != nil {
			logger.Warn("convertExtrasNotificationgroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasNotificationgroups = append(extrasNotificationgroups, data)
	}

	return &netbox_goV1.ListExtrasNotificationgroupByLastIDReply{
		ExtrasNotificationgroups: extrasNotificationgroups,
	}, nil
}

func convertExtrasNotificationgroup(record *model.ExtrasNotificationgroup) (*netbox_goV1.ExtrasNotificationgroup, error) {
	value := &netbox_goV1.ExtrasNotificationgroup{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
