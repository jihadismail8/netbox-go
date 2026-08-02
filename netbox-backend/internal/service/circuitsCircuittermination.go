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
		netbox_goV1.RegisterCircuitsCircuitterminationServer(server, NewCircuitsCircuitterminationServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CircuitsCircuitterminationServer = (*circuitsCircuittermination)(nil)
var _ time.Time

type circuitsCircuittermination struct {
	netbox_goV1.UnimplementedCircuitsCircuitterminationServer

	iDao dao.CircuitsCircuitterminationDao
}

// NewCircuitsCircuitterminationServer create a new service
func NewCircuitsCircuitterminationServer() netbox_goV1.CircuitsCircuitterminationServer {
	return &circuitsCircuittermination{
		iDao: dao.NewCircuitsCircuitterminationDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCircuitsCircuitterminationCache(database.GetCacheType()),
		),
	}
}

// Create a new circuitsCircuittermination
func (s *circuitsCircuittermination) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsCircuitterminationRequest) (*netbox_goV1.CreateCircuitsCircuitterminationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsCircuittermination{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCircuitsCircuittermination.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("circuitsCircuittermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCircuitsCircuitterminationReply{Id: record.ID}, nil
}

// DeleteByID delete a circuitsCircuittermination by id
func (s *circuitsCircuittermination) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitterminationByIDRequest) (*netbox_goV1.DeleteCircuitsCircuitterminationByIDReply, error) {
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

	return &netbox_goV1.DeleteCircuitsCircuitterminationByIDReply{}, nil
}

// UpdateByID update a circuitsCircuittermination by id
func (s *circuitsCircuittermination) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsCircuitterminationByIDRequest) (*netbox_goV1.UpdateCircuitsCircuitterminationByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsCircuittermination{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDCircuitsCircuittermination.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("circuitsCircuittermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCircuitsCircuitterminationByIDReply{}, nil
}

// GetByID get a circuitsCircuittermination by id
func (s *circuitsCircuittermination) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitterminationByIDRequest) (*netbox_goV1.GetCircuitsCircuitterminationByIDReply, error) {
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

	data, err := convertCircuitsCircuittermination(record)
	if err != nil {
		logger.Warn("convertCircuitsCircuittermination error", logger.Err(err), logger.Any("circuitsCircuittermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDCircuitsCircuittermination.Err()
	}

	return &netbox_goV1.GetCircuitsCircuitterminationByIDReply{CircuitsCircuittermination: data}, nil
}

// List get a paginated list of circuitsCircuitterminations by custom conditions
func (s *circuitsCircuittermination) List(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitterminationRequest) (*netbox_goV1.ListCircuitsCircuitterminationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCircuitsCircuittermination.Err()
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

	circuitsCircuitterminations := []*netbox_goV1.CircuitsCircuittermination{}
	for _, record := range records {
		data, err := convertCircuitsCircuittermination(record)
		if err != nil {
			logger.Warn("convertCircuitsCircuittermination error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsCircuitterminations = append(circuitsCircuitterminations, data)
	}

	return &netbox_goV1.ListCircuitsCircuitterminationReply{
		Total:                       total,
		CircuitsCircuitterminations: circuitsCircuitterminations,
	}, nil
}

// DeleteByIDs batch delete circuitsCircuittermination by ids
func (s *circuitsCircuittermination) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitterminationByIDsRequest) (*netbox_goV1.DeleteCircuitsCircuitterminationByIDsReply, error) {
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

	return &netbox_goV1.DeleteCircuitsCircuitterminationByIDsReply{}, nil
}

// GetByCondition get a circuitsCircuittermination by custom condition
func (s *circuitsCircuittermination) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitterminationByConditionRequest) (*netbox_goV1.GetCircuitsCircuitterminationByConditionReply, error) {
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

	data, err := convertCircuitsCircuittermination(record)
	if err != nil {
		logger.Warn("convertCircuitsCircuittermination error", logger.Err(err), logger.Any("circuitsCircuittermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCircuitsCircuittermination.Err()
	}

	return &netbox_goV1.GetCircuitsCircuitterminationByConditionReply{
		CircuitsCircuittermination: data,
	}, nil
}

// ListByIDs batch get circuitsCircuittermination by ids
func (s *circuitsCircuittermination) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitterminationByIDsRequest) (*netbox_goV1.ListCircuitsCircuitterminationByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	circuitsCircuitterminationMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	circuitsCircuitterminations := []*netbox_goV1.CircuitsCircuittermination{}
	for _, id := range req.Ids {
		if v, ok := circuitsCircuitterminationMap[id]; ok {
			record, err := convertCircuitsCircuittermination(v)
			if err != nil {
				logger.Warn("convertCircuitsCircuittermination error", logger.Err(err), logger.Any("circuitsCircuittermination", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			circuitsCircuitterminations = append(circuitsCircuitterminations, record)
		}
	}

	return &netbox_goV1.ListCircuitsCircuitterminationByIDsReply{CircuitsCircuitterminations: circuitsCircuitterminations}, nil
}

// ListByLastID get a paginated list of circuitsCircuitterminations by last id
func (s *circuitsCircuittermination) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitterminationByLastIDRequest) (*netbox_goV1.ListCircuitsCircuitterminationByLastIDReply, error) {
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

	circuitsCircuitterminations := []*netbox_goV1.CircuitsCircuittermination{}
	for _, record := range records {
		data, err := convertCircuitsCircuittermination(record)
		if err != nil {
			logger.Warn("convertCircuitsCircuittermination error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsCircuitterminations = append(circuitsCircuitterminations, data)
	}

	return &netbox_goV1.ListCircuitsCircuitterminationByLastIDReply{
		CircuitsCircuitterminations: circuitsCircuitterminations,
	}, nil
}

func convertCircuitsCircuittermination(record *model.CircuitsCircuittermination) (*netbox_goV1.CircuitsCircuittermination, error) {
	value := &netbox_goV1.CircuitsCircuittermination{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
