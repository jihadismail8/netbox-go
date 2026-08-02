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
		netbox_goV1.RegisterCircuitsProviderServer(server, NewCircuitsProviderServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CircuitsProviderServer = (*circuitsProvider)(nil)
var _ time.Time

type circuitsProvider struct {
	netbox_goV1.UnimplementedCircuitsProviderServer

	iDao dao.CircuitsProviderDao
}

// NewCircuitsProviderServer create a new service
func NewCircuitsProviderServer() netbox_goV1.CircuitsProviderServer {
	return &circuitsProvider{
		iDao: dao.NewCircuitsProviderDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCircuitsProviderCache(database.GetCacheType()),
		),
	}
}

// Create a new circuitsProvider
func (s *circuitsProvider) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsProviderRequest) (*netbox_goV1.CreateCircuitsProviderReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsProvider{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCircuitsProvider.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("circuitsProvider", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCircuitsProviderReply{Id: record.ID}, nil
}

// DeleteByID delete a circuitsProvider by id
func (s *circuitsProvider) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsProviderByIDRequest) (*netbox_goV1.DeleteCircuitsProviderByIDReply, error) {
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

	return &netbox_goV1.DeleteCircuitsProviderByIDReply{}, nil
}

// UpdateByID update a circuitsProvider by id
func (s *circuitsProvider) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsProviderByIDRequest) (*netbox_goV1.UpdateCircuitsProviderByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsProvider{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDCircuitsProvider.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("circuitsProvider", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCircuitsProviderByIDReply{}, nil
}

// GetByID get a circuitsProvider by id
func (s *circuitsProvider) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsProviderByIDRequest) (*netbox_goV1.GetCircuitsProviderByIDReply, error) {
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

	data, err := convertCircuitsProvider(record)
	if err != nil {
		logger.Warn("convertCircuitsProvider error", logger.Err(err), logger.Any("circuitsProvider", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDCircuitsProvider.Err()
	}

	return &netbox_goV1.GetCircuitsProviderByIDReply{CircuitsProvider: data}, nil
}

// List get a paginated list of circuitsProviders by custom conditions
func (s *circuitsProvider) List(ctx context.Context, req *netbox_goV1.ListCircuitsProviderRequest) (*netbox_goV1.ListCircuitsProviderReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCircuitsProvider.Err()
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

	circuitsProviders := []*netbox_goV1.CircuitsProvider{}
	for _, record := range records {
		data, err := convertCircuitsProvider(record)
		if err != nil {
			logger.Warn("convertCircuitsProvider error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsProviders = append(circuitsProviders, data)
	}

	return &netbox_goV1.ListCircuitsProviderReply{
		Total:             total,
		CircuitsProviders: circuitsProviders,
	}, nil
}

// DeleteByIDs batch delete circuitsProvider by ids
func (s *circuitsProvider) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsProviderByIDsRequest) (*netbox_goV1.DeleteCircuitsProviderByIDsReply, error) {
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

	return &netbox_goV1.DeleteCircuitsProviderByIDsReply{}, nil
}

// GetByCondition get a circuitsProvider by custom condition
func (s *circuitsProvider) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsProviderByConditionRequest) (*netbox_goV1.GetCircuitsProviderByConditionReply, error) {
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

	data, err := convertCircuitsProvider(record)
	if err != nil {
		logger.Warn("convertCircuitsProvider error", logger.Err(err), logger.Any("circuitsProvider", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCircuitsProvider.Err()
	}

	return &netbox_goV1.GetCircuitsProviderByConditionReply{
		CircuitsProvider: data,
	}, nil
}

// ListByIDs batch get circuitsProvider by ids
func (s *circuitsProvider) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsProviderByIDsRequest) (*netbox_goV1.ListCircuitsProviderByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	circuitsProviderMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	circuitsProviders := []*netbox_goV1.CircuitsProvider{}
	for _, id := range req.Ids {
		if v, ok := circuitsProviderMap[id]; ok {
			record, err := convertCircuitsProvider(v)
			if err != nil {
				logger.Warn("convertCircuitsProvider error", logger.Err(err), logger.Any("circuitsProvider", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			circuitsProviders = append(circuitsProviders, record)
		}
	}

	return &netbox_goV1.ListCircuitsProviderByIDsReply{CircuitsProviders: circuitsProviders}, nil
}

// ListByLastID get a paginated list of circuitsProviders by last id
func (s *circuitsProvider) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsProviderByLastIDRequest) (*netbox_goV1.ListCircuitsProviderByLastIDReply, error) {
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

	circuitsProviders := []*netbox_goV1.CircuitsProvider{}
	for _, record := range records {
		data, err := convertCircuitsProvider(record)
		if err != nil {
			logger.Warn("convertCircuitsProvider error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsProviders = append(circuitsProviders, data)
	}

	return &netbox_goV1.ListCircuitsProviderByLastIDReply{
		CircuitsProviders: circuitsProviders,
	}, nil
}

func convertCircuitsProvider(record *model.CircuitsProvider) (*netbox_goV1.CircuitsProvider, error) {
	value := &netbox_goV1.CircuitsProvider{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
