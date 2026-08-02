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
		netbox_goV1.RegisterCircuitsVirtualcircuitterminationServer(server, NewCircuitsVirtualcircuitterminationServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CircuitsVirtualcircuitterminationServer = (*circuitsVirtualcircuittermination)(nil)
var _ time.Time

type circuitsVirtualcircuittermination struct {
	netbox_goV1.UnimplementedCircuitsVirtualcircuitterminationServer

	iDao dao.CircuitsVirtualcircuitterminationDao
}

// NewCircuitsVirtualcircuitterminationServer create a new service
func NewCircuitsVirtualcircuitterminationServer() netbox_goV1.CircuitsVirtualcircuitterminationServer {
	return &circuitsVirtualcircuittermination{
		iDao: dao.NewCircuitsVirtualcircuitterminationDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCircuitsVirtualcircuitterminationCache(database.GetCacheType()),
		),
	}
}

// Create a new circuitsVirtualcircuittermination
func (s *circuitsVirtualcircuittermination) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsVirtualcircuitterminationRequest) (*netbox_goV1.CreateCircuitsVirtualcircuitterminationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsVirtualcircuittermination{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCircuitsVirtualcircuittermination.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("circuitsVirtualcircuittermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCircuitsVirtualcircuitterminationReply{Id: record.ID}, nil
}

// DeleteByID delete a circuitsVirtualcircuittermination by id
func (s *circuitsVirtualcircuittermination) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsVirtualcircuitterminationByIDRequest) (*netbox_goV1.DeleteCircuitsVirtualcircuitterminationByIDReply, error) {
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

	return &netbox_goV1.DeleteCircuitsVirtualcircuitterminationByIDReply{}, nil
}

// UpdateByID update a circuitsVirtualcircuittermination by id
func (s *circuitsVirtualcircuittermination) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsVirtualcircuitterminationByIDRequest) (*netbox_goV1.UpdateCircuitsVirtualcircuitterminationByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsVirtualcircuittermination{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDCircuitsVirtualcircuittermination.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("circuitsVirtualcircuittermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCircuitsVirtualcircuitterminationByIDReply{}, nil
}

// GetByID get a circuitsVirtualcircuittermination by id
func (s *circuitsVirtualcircuittermination) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsVirtualcircuitterminationByIDRequest) (*netbox_goV1.GetCircuitsVirtualcircuitterminationByIDReply, error) {
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

	data, err := convertCircuitsVirtualcircuittermination(record)
	if err != nil {
		logger.Warn("convertCircuitsVirtualcircuittermination error", logger.Err(err), logger.Any("circuitsVirtualcircuittermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDCircuitsVirtualcircuittermination.Err()
	}

	return &netbox_goV1.GetCircuitsVirtualcircuitterminationByIDReply{CircuitsVirtualcircuittermination: data}, nil
}

// List get a paginated list of circuitsVirtualcircuitterminations by custom conditions
func (s *circuitsVirtualcircuittermination) List(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuitterminationRequest) (*netbox_goV1.ListCircuitsVirtualcircuitterminationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCircuitsVirtualcircuittermination.Err()
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

	circuitsVirtualcircuitterminations := []*netbox_goV1.CircuitsVirtualcircuittermination{}
	for _, record := range records {
		data, err := convertCircuitsVirtualcircuittermination(record)
		if err != nil {
			logger.Warn("convertCircuitsVirtualcircuittermination error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsVirtualcircuitterminations = append(circuitsVirtualcircuitterminations, data)
	}

	return &netbox_goV1.ListCircuitsVirtualcircuitterminationReply{
		Total:                              total,
		CircuitsVirtualcircuitterminations: circuitsVirtualcircuitterminations,
	}, nil
}

// DeleteByIDs batch delete circuitsVirtualcircuittermination by ids
func (s *circuitsVirtualcircuittermination) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsVirtualcircuitterminationByIDsRequest) (*netbox_goV1.DeleteCircuitsVirtualcircuitterminationByIDsReply, error) {
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

	return &netbox_goV1.DeleteCircuitsVirtualcircuitterminationByIDsReply{}, nil
}

// GetByCondition get a circuitsVirtualcircuittermination by custom condition
func (s *circuitsVirtualcircuittermination) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsVirtualcircuitterminationByConditionRequest) (*netbox_goV1.GetCircuitsVirtualcircuitterminationByConditionReply, error) {
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

	data, err := convertCircuitsVirtualcircuittermination(record)
	if err != nil {
		logger.Warn("convertCircuitsVirtualcircuittermination error", logger.Err(err), logger.Any("circuitsVirtualcircuittermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCircuitsVirtualcircuittermination.Err()
	}

	return &netbox_goV1.GetCircuitsVirtualcircuitterminationByConditionReply{
		CircuitsVirtualcircuittermination: data,
	}, nil
}

// ListByIDs batch get circuitsVirtualcircuittermination by ids
func (s *circuitsVirtualcircuittermination) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuitterminationByIDsRequest) (*netbox_goV1.ListCircuitsVirtualcircuitterminationByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	circuitsVirtualcircuitterminationMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	circuitsVirtualcircuitterminations := []*netbox_goV1.CircuitsVirtualcircuittermination{}
	for _, id := range req.Ids {
		if v, ok := circuitsVirtualcircuitterminationMap[id]; ok {
			record, err := convertCircuitsVirtualcircuittermination(v)
			if err != nil {
				logger.Warn("convertCircuitsVirtualcircuittermination error", logger.Err(err), logger.Any("circuitsVirtualcircuittermination", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			circuitsVirtualcircuitterminations = append(circuitsVirtualcircuitterminations, record)
		}
	}

	return &netbox_goV1.ListCircuitsVirtualcircuitterminationByIDsReply{CircuitsVirtualcircuitterminations: circuitsVirtualcircuitterminations}, nil
}

// ListByLastID get a paginated list of circuitsVirtualcircuitterminations by last id
func (s *circuitsVirtualcircuittermination) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuitterminationByLastIDRequest) (*netbox_goV1.ListCircuitsVirtualcircuitterminationByLastIDReply, error) {
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

	circuitsVirtualcircuitterminations := []*netbox_goV1.CircuitsVirtualcircuittermination{}
	for _, record := range records {
		data, err := convertCircuitsVirtualcircuittermination(record)
		if err != nil {
			logger.Warn("convertCircuitsVirtualcircuittermination error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsVirtualcircuitterminations = append(circuitsVirtualcircuitterminations, data)
	}

	return &netbox_goV1.ListCircuitsVirtualcircuitterminationByLastIDReply{
		CircuitsVirtualcircuitterminations: circuitsVirtualcircuitterminations,
	}, nil
}

func convertCircuitsVirtualcircuittermination(record *model.CircuitsVirtualcircuittermination) (*netbox_goV1.CircuitsVirtualcircuittermination, error) {
	value := &netbox_goV1.CircuitsVirtualcircuittermination{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
