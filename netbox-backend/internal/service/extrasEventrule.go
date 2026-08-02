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
		netbox_goV1.RegisterExtrasEventruleServer(server, NewExtrasEventruleServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasEventruleServer = (*extrasEventrule)(nil)
var _ time.Time

type extrasEventrule struct {
	netbox_goV1.UnimplementedExtrasEventruleServer

	iDao dao.ExtrasEventruleDao
}

// NewExtrasEventruleServer create a new service
func NewExtrasEventruleServer() netbox_goV1.ExtrasEventruleServer {
	return &extrasEventrule{
		iDao: dao.NewExtrasEventruleDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasEventruleCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasEventrule
func (s *extrasEventrule) Create(ctx context.Context, req *netbox_goV1.CreateExtrasEventruleRequest) (*netbox_goV1.CreateExtrasEventruleReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasEventrule{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasEventrule.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasEventrule", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasEventruleReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasEventrule by id
func (s *extrasEventrule) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasEventruleByIDRequest) (*netbox_goV1.DeleteExtrasEventruleByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasEventruleByIDReply{}, nil
}

// UpdateByID update a extrasEventrule by id
func (s *extrasEventrule) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasEventruleByIDRequest) (*netbox_goV1.UpdateExtrasEventruleByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasEventrule{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasEventrule.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasEventrule", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasEventruleByIDReply{}, nil
}

// GetByID get a extrasEventrule by id
func (s *extrasEventrule) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasEventruleByIDRequest) (*netbox_goV1.GetExtrasEventruleByIDReply, error) {
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

	data, err := convertExtrasEventrule(record)
	if err != nil {
		logger.Warn("convertExtrasEventrule error", logger.Err(err), logger.Any("extrasEventrule", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasEventrule.Err()
	}

	return &netbox_goV1.GetExtrasEventruleByIDReply{ExtrasEventrule: data}, nil
}

// List get a paginated list of extrasEventrules by custom conditions
func (s *extrasEventrule) List(ctx context.Context, req *netbox_goV1.ListExtrasEventruleRequest) (*netbox_goV1.ListExtrasEventruleReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasEventrule.Err()
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

	extrasEventrules := []*netbox_goV1.ExtrasEventrule{}
	for _, record := range records {
		data, err := convertExtrasEventrule(record)
		if err != nil {
			logger.Warn("convertExtrasEventrule error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasEventrules = append(extrasEventrules, data)
	}

	return &netbox_goV1.ListExtrasEventruleReply{
		Total:            total,
		ExtrasEventrules: extrasEventrules,
	}, nil
}

// DeleteByIDs batch delete extrasEventrule by ids
func (s *extrasEventrule) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasEventruleByIDsRequest) (*netbox_goV1.DeleteExtrasEventruleByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasEventruleByIDsReply{}, nil
}

// GetByCondition get a extrasEventrule by custom condition
func (s *extrasEventrule) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasEventruleByConditionRequest) (*netbox_goV1.GetExtrasEventruleByConditionReply, error) {
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

	data, err := convertExtrasEventrule(record)
	if err != nil {
		logger.Warn("convertExtrasEventrule error", logger.Err(err), logger.Any("extrasEventrule", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasEventrule.Err()
	}

	return &netbox_goV1.GetExtrasEventruleByConditionReply{
		ExtrasEventrule: data,
	}, nil
}

// ListByIDs batch get extrasEventrule by ids
func (s *extrasEventrule) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasEventruleByIDsRequest) (*netbox_goV1.ListExtrasEventruleByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasEventruleMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasEventrules := []*netbox_goV1.ExtrasEventrule{}
	for _, id := range req.Ids {
		if v, ok := extrasEventruleMap[id]; ok {
			record, err := convertExtrasEventrule(v)
			if err != nil {
				logger.Warn("convertExtrasEventrule error", logger.Err(err), logger.Any("extrasEventrule", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasEventrules = append(extrasEventrules, record)
		}
	}

	return &netbox_goV1.ListExtrasEventruleByIDsReply{ExtrasEventrules: extrasEventrules}, nil
}

// ListByLastID get a paginated list of extrasEventrules by last id
func (s *extrasEventrule) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasEventruleByLastIDRequest) (*netbox_goV1.ListExtrasEventruleByLastIDReply, error) {
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

	extrasEventrules := []*netbox_goV1.ExtrasEventrule{}
	for _, record := range records {
		data, err := convertExtrasEventrule(record)
		if err != nil {
			logger.Warn("convertExtrasEventrule error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasEventrules = append(extrasEventrules, data)
	}

	return &netbox_goV1.ListExtrasEventruleByLastIDReply{
		ExtrasEventrules: extrasEventrules,
	}, nil
}

func convertExtrasEventrule(record *model.ExtrasEventrule) (*netbox_goV1.ExtrasEventrule, error) {
	value := &netbox_goV1.ExtrasEventrule{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
