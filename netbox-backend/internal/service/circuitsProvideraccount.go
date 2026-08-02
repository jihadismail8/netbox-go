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
		netbox_goV1.RegisterCircuitsProvideraccountServer(server, NewCircuitsProvideraccountServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CircuitsProvideraccountServer = (*circuitsProvideraccount)(nil)
var _ time.Time

type circuitsProvideraccount struct {
	netbox_goV1.UnimplementedCircuitsProvideraccountServer

	iDao dao.CircuitsProvideraccountDao
}

// NewCircuitsProvideraccountServer create a new service
func NewCircuitsProvideraccountServer() netbox_goV1.CircuitsProvideraccountServer {
	return &circuitsProvideraccount{
		iDao: dao.NewCircuitsProvideraccountDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCircuitsProvideraccountCache(database.GetCacheType()),
		),
	}
}

// Create a new circuitsProvideraccount
func (s *circuitsProvideraccount) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsProvideraccountRequest) (*netbox_goV1.CreateCircuitsProvideraccountReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsProvideraccount{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCircuitsProvideraccount.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("circuitsProvideraccount", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCircuitsProvideraccountReply{Id: record.ID}, nil
}

// DeleteByID delete a circuitsProvideraccount by id
func (s *circuitsProvideraccount) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsProvideraccountByIDRequest) (*netbox_goV1.DeleteCircuitsProvideraccountByIDReply, error) {
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

	return &netbox_goV1.DeleteCircuitsProvideraccountByIDReply{}, nil
}

// UpdateByID update a circuitsProvideraccount by id
func (s *circuitsProvideraccount) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsProvideraccountByIDRequest) (*netbox_goV1.UpdateCircuitsProvideraccountByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsProvideraccount{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDCircuitsProvideraccount.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("circuitsProvideraccount", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCircuitsProvideraccountByIDReply{}, nil
}

// GetByID get a circuitsProvideraccount by id
func (s *circuitsProvideraccount) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsProvideraccountByIDRequest) (*netbox_goV1.GetCircuitsProvideraccountByIDReply, error) {
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

	data, err := convertCircuitsProvideraccount(record)
	if err != nil {
		logger.Warn("convertCircuitsProvideraccount error", logger.Err(err), logger.Any("circuitsProvideraccount", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDCircuitsProvideraccount.Err()
	}

	return &netbox_goV1.GetCircuitsProvideraccountByIDReply{CircuitsProvideraccount: data}, nil
}

// List get a paginated list of circuitsProvideraccounts by custom conditions
func (s *circuitsProvideraccount) List(ctx context.Context, req *netbox_goV1.ListCircuitsProvideraccountRequest) (*netbox_goV1.ListCircuitsProvideraccountReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCircuitsProvideraccount.Err()
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

	circuitsProvideraccounts := []*netbox_goV1.CircuitsProvideraccount{}
	for _, record := range records {
		data, err := convertCircuitsProvideraccount(record)
		if err != nil {
			logger.Warn("convertCircuitsProvideraccount error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsProvideraccounts = append(circuitsProvideraccounts, data)
	}

	return &netbox_goV1.ListCircuitsProvideraccountReply{
		Total:                    total,
		CircuitsProvideraccounts: circuitsProvideraccounts,
	}, nil
}

// DeleteByIDs batch delete circuitsProvideraccount by ids
func (s *circuitsProvideraccount) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsProvideraccountByIDsRequest) (*netbox_goV1.DeleteCircuitsProvideraccountByIDsReply, error) {
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

	return &netbox_goV1.DeleteCircuitsProvideraccountByIDsReply{}, nil
}

// GetByCondition get a circuitsProvideraccount by custom condition
func (s *circuitsProvideraccount) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsProvideraccountByConditionRequest) (*netbox_goV1.GetCircuitsProvideraccountByConditionReply, error) {
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

	data, err := convertCircuitsProvideraccount(record)
	if err != nil {
		logger.Warn("convertCircuitsProvideraccount error", logger.Err(err), logger.Any("circuitsProvideraccount", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCircuitsProvideraccount.Err()
	}

	return &netbox_goV1.GetCircuitsProvideraccountByConditionReply{
		CircuitsProvideraccount: data,
	}, nil
}

// ListByIDs batch get circuitsProvideraccount by ids
func (s *circuitsProvideraccount) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsProvideraccountByIDsRequest) (*netbox_goV1.ListCircuitsProvideraccountByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	circuitsProvideraccountMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	circuitsProvideraccounts := []*netbox_goV1.CircuitsProvideraccount{}
	for _, id := range req.Ids {
		if v, ok := circuitsProvideraccountMap[id]; ok {
			record, err := convertCircuitsProvideraccount(v)
			if err != nil {
				logger.Warn("convertCircuitsProvideraccount error", logger.Err(err), logger.Any("circuitsProvideraccount", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			circuitsProvideraccounts = append(circuitsProvideraccounts, record)
		}
	}

	return &netbox_goV1.ListCircuitsProvideraccountByIDsReply{CircuitsProvideraccounts: circuitsProvideraccounts}, nil
}

// ListByLastID get a paginated list of circuitsProvideraccounts by last id
func (s *circuitsProvideraccount) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsProvideraccountByLastIDRequest) (*netbox_goV1.ListCircuitsProvideraccountByLastIDReply, error) {
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

	circuitsProvideraccounts := []*netbox_goV1.CircuitsProvideraccount{}
	for _, record := range records {
		data, err := convertCircuitsProvideraccount(record)
		if err != nil {
			logger.Warn("convertCircuitsProvideraccount error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsProvideraccounts = append(circuitsProvideraccounts, data)
	}

	return &netbox_goV1.ListCircuitsProvideraccountByLastIDReply{
		CircuitsProvideraccounts: circuitsProvideraccounts,
	}, nil
}

func convertCircuitsProvideraccount(record *model.CircuitsProvideraccount) (*netbox_goV1.CircuitsProvideraccount, error) {
	value := &netbox_goV1.CircuitsProvideraccount{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
