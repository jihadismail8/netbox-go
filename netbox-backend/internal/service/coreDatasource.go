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
		netbox_goV1.RegisterCoreDatasourceServer(server, NewCoreDatasourceServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CoreDatasourceServer = (*coreDatasource)(nil)
var _ time.Time

type coreDatasource struct {
	netbox_goV1.UnimplementedCoreDatasourceServer

	iDao dao.CoreDatasourceDao
}

// NewCoreDatasourceServer create a new service
func NewCoreDatasourceServer() netbox_goV1.CoreDatasourceServer {
	return &coreDatasource{
		iDao: dao.NewCoreDatasourceDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCoreDatasourceCache(database.GetCacheType()),
		),
	}
}

// Create a new coreDatasource
func (s *coreDatasource) Create(ctx context.Context, req *netbox_goV1.CreateCoreDatasourceRequest) (*netbox_goV1.CreateCoreDatasourceReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CoreDatasource{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCoreDatasource.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("coreDatasource", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCoreDatasourceReply{Id: record.ID}, nil
}

// DeleteByID delete a coreDatasource by id
func (s *coreDatasource) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCoreDatasourceByIDRequest) (*netbox_goV1.DeleteCoreDatasourceByIDReply, error) {
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

	return &netbox_goV1.DeleteCoreDatasourceByIDReply{}, nil
}

// UpdateByID update a coreDatasource by id
func (s *coreDatasource) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCoreDatasourceByIDRequest) (*netbox_goV1.UpdateCoreDatasourceByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CoreDatasource{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDCoreDatasource.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("coreDatasource", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCoreDatasourceByIDReply{}, nil
}

// GetByID get a coreDatasource by id
func (s *coreDatasource) GetByID(ctx context.Context, req *netbox_goV1.GetCoreDatasourceByIDRequest) (*netbox_goV1.GetCoreDatasourceByIDReply, error) {
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

	data, err := convertCoreDatasource(record)
	if err != nil {
		logger.Warn("convertCoreDatasource error", logger.Err(err), logger.Any("coreDatasource", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDCoreDatasource.Err()
	}

	return &netbox_goV1.GetCoreDatasourceByIDReply{CoreDatasource: data}, nil
}

// List get a paginated list of coreDatasources by custom conditions
func (s *coreDatasource) List(ctx context.Context, req *netbox_goV1.ListCoreDatasourceRequest) (*netbox_goV1.ListCoreDatasourceReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCoreDatasource.Err()
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

	coreDatasources := []*netbox_goV1.CoreDatasource{}
	for _, record := range records {
		data, err := convertCoreDatasource(record)
		if err != nil {
			logger.Warn("convertCoreDatasource error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		coreDatasources = append(coreDatasources, data)
	}

	return &netbox_goV1.ListCoreDatasourceReply{
		Total:           total,
		CoreDatasources: coreDatasources,
	}, nil
}

// DeleteByIDs batch delete coreDatasource by ids
func (s *coreDatasource) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCoreDatasourceByIDsRequest) (*netbox_goV1.DeleteCoreDatasourceByIDsReply, error) {
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

	return &netbox_goV1.DeleteCoreDatasourceByIDsReply{}, nil
}

// GetByCondition get a coreDatasource by custom condition
func (s *coreDatasource) GetByCondition(ctx context.Context, req *netbox_goV1.GetCoreDatasourceByConditionRequest) (*netbox_goV1.GetCoreDatasourceByConditionReply, error) {
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

	data, err := convertCoreDatasource(record)
	if err != nil {
		logger.Warn("convertCoreDatasource error", logger.Err(err), logger.Any("coreDatasource", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCoreDatasource.Err()
	}

	return &netbox_goV1.GetCoreDatasourceByConditionReply{
		CoreDatasource: data,
	}, nil
}

// ListByIDs batch get coreDatasource by ids
func (s *coreDatasource) ListByIDs(ctx context.Context, req *netbox_goV1.ListCoreDatasourceByIDsRequest) (*netbox_goV1.ListCoreDatasourceByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	coreDatasourceMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	coreDatasources := []*netbox_goV1.CoreDatasource{}
	for _, id := range req.Ids {
		if v, ok := coreDatasourceMap[id]; ok {
			record, err := convertCoreDatasource(v)
			if err != nil {
				logger.Warn("convertCoreDatasource error", logger.Err(err), logger.Any("coreDatasource", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			coreDatasources = append(coreDatasources, record)
		}
	}

	return &netbox_goV1.ListCoreDatasourceByIDsReply{CoreDatasources: coreDatasources}, nil
}

// ListByLastID get a paginated list of coreDatasources by last id
func (s *coreDatasource) ListByLastID(ctx context.Context, req *netbox_goV1.ListCoreDatasourceByLastIDRequest) (*netbox_goV1.ListCoreDatasourceByLastIDReply, error) {
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

	coreDatasources := []*netbox_goV1.CoreDatasource{}
	for _, record := range records {
		data, err := convertCoreDatasource(record)
		if err != nil {
			logger.Warn("convertCoreDatasource error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		coreDatasources = append(coreDatasources, data)
	}

	return &netbox_goV1.ListCoreDatasourceByLastIDReply{
		CoreDatasources: coreDatasources,
	}, nil
}

func convertCoreDatasource(record *model.CoreDatasource) (*netbox_goV1.CoreDatasource, error) {
	value := &netbox_goV1.CoreDatasource{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
