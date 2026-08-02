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
		netbox_goV1.RegisterCircuitsCircuitServer(server, NewCircuitsCircuitServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CircuitsCircuitServer = (*circuitsCircuit)(nil)
var _ time.Time

type circuitsCircuit struct {
	netbox_goV1.UnimplementedCircuitsCircuitServer

	iDao dao.CircuitsCircuitDao
}

// NewCircuitsCircuitServer create a new service
func NewCircuitsCircuitServer() netbox_goV1.CircuitsCircuitServer {
	return &circuitsCircuit{
		iDao: dao.NewCircuitsCircuitDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCircuitsCircuitCache(database.GetCacheType()),
		),
	}
}

// Create a new circuitsCircuit
func (s *circuitsCircuit) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsCircuitRequest) (*netbox_goV1.CreateCircuitsCircuitReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsCircuit{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCircuitsCircuit.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("circuitsCircuit", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCircuitsCircuitReply{Id: record.ID}, nil
}

// DeleteByID delete a circuitsCircuit by id
func (s *circuitsCircuit) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitByIDRequest) (*netbox_goV1.DeleteCircuitsCircuitByIDReply, error) {
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

	return &netbox_goV1.DeleteCircuitsCircuitByIDReply{}, nil
}

// UpdateByID update a circuitsCircuit by id
func (s *circuitsCircuit) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsCircuitByIDRequest) (*netbox_goV1.UpdateCircuitsCircuitByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsCircuit{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDCircuitsCircuit.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("circuitsCircuit", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCircuitsCircuitByIDReply{}, nil
}

// GetByID get a circuitsCircuit by id
func (s *circuitsCircuit) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitByIDRequest) (*netbox_goV1.GetCircuitsCircuitByIDReply, error) {
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

	data, err := convertCircuitsCircuit(record)
	if err != nil {
		logger.Warn("convertCircuitsCircuit error", logger.Err(err), logger.Any("circuitsCircuit", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDCircuitsCircuit.Err()
	}

	return &netbox_goV1.GetCircuitsCircuitByIDReply{CircuitsCircuit: data}, nil
}

// List get a paginated list of circuitsCircuits by custom conditions
func (s *circuitsCircuit) List(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitRequest) (*netbox_goV1.ListCircuitsCircuitReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCircuitsCircuit.Err()
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

	circuitsCircuits := []*netbox_goV1.CircuitsCircuit{}
	for _, record := range records {
		data, err := convertCircuitsCircuit(record)
		if err != nil {
			logger.Warn("convertCircuitsCircuit error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsCircuits = append(circuitsCircuits, data)
	}

	return &netbox_goV1.ListCircuitsCircuitReply{
		Total:            total,
		CircuitsCircuits: circuitsCircuits,
	}, nil
}

// DeleteByIDs batch delete circuitsCircuit by ids
func (s *circuitsCircuit) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitByIDsRequest) (*netbox_goV1.DeleteCircuitsCircuitByIDsReply, error) {
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

	return &netbox_goV1.DeleteCircuitsCircuitByIDsReply{}, nil
}

// GetByCondition get a circuitsCircuit by custom condition
func (s *circuitsCircuit) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitByConditionRequest) (*netbox_goV1.GetCircuitsCircuitByConditionReply, error) {
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

	data, err := convertCircuitsCircuit(record)
	if err != nil {
		logger.Warn("convertCircuitsCircuit error", logger.Err(err), logger.Any("circuitsCircuit", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCircuitsCircuit.Err()
	}

	return &netbox_goV1.GetCircuitsCircuitByConditionReply{
		CircuitsCircuit: data,
	}, nil
}

// ListByIDs batch get circuitsCircuit by ids
func (s *circuitsCircuit) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitByIDsRequest) (*netbox_goV1.ListCircuitsCircuitByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	circuitsCircuitMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	circuitsCircuits := []*netbox_goV1.CircuitsCircuit{}
	for _, id := range req.Ids {
		if v, ok := circuitsCircuitMap[id]; ok {
			record, err := convertCircuitsCircuit(v)
			if err != nil {
				logger.Warn("convertCircuitsCircuit error", logger.Err(err), logger.Any("circuitsCircuit", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			circuitsCircuits = append(circuitsCircuits, record)
		}
	}

	return &netbox_goV1.ListCircuitsCircuitByIDsReply{CircuitsCircuits: circuitsCircuits}, nil
}

// ListByLastID get a paginated list of circuitsCircuits by last id
func (s *circuitsCircuit) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitByLastIDRequest) (*netbox_goV1.ListCircuitsCircuitByLastIDReply, error) {
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

	circuitsCircuits := []*netbox_goV1.CircuitsCircuit{}
	for _, record := range records {
		data, err := convertCircuitsCircuit(record)
		if err != nil {
			logger.Warn("convertCircuitsCircuit error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsCircuits = append(circuitsCircuits, data)
	}

	return &netbox_goV1.ListCircuitsCircuitByLastIDReply{
		CircuitsCircuits: circuitsCircuits,
	}, nil
}

func convertCircuitsCircuit(record *model.CircuitsCircuit) (*netbox_goV1.CircuitsCircuit, error) {
	value := &netbox_goV1.CircuitsCircuit{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
