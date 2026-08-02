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
		netbox_goV1.RegisterCircuitsVirtualcircuitServer(server, NewCircuitsVirtualcircuitServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CircuitsVirtualcircuitServer = (*circuitsVirtualcircuit)(nil)
var _ time.Time

type circuitsVirtualcircuit struct {
	netbox_goV1.UnimplementedCircuitsVirtualcircuitServer

	iDao dao.CircuitsVirtualcircuitDao
}

// NewCircuitsVirtualcircuitServer create a new service
func NewCircuitsVirtualcircuitServer() netbox_goV1.CircuitsVirtualcircuitServer {
	return &circuitsVirtualcircuit{
		iDao: dao.NewCircuitsVirtualcircuitDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCircuitsVirtualcircuitCache(database.GetCacheType()),
		),
	}
}

// Create a new circuitsVirtualcircuit
func (s *circuitsVirtualcircuit) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsVirtualcircuitRequest) (*netbox_goV1.CreateCircuitsVirtualcircuitReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsVirtualcircuit{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCircuitsVirtualcircuit.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("circuitsVirtualcircuit", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCircuitsVirtualcircuitReply{Id: record.ID}, nil
}

// DeleteByID delete a circuitsVirtualcircuit by id
func (s *circuitsVirtualcircuit) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsVirtualcircuitByIDRequest) (*netbox_goV1.DeleteCircuitsVirtualcircuitByIDReply, error) {
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

	return &netbox_goV1.DeleteCircuitsVirtualcircuitByIDReply{}, nil
}

// UpdateByID update a circuitsVirtualcircuit by id
func (s *circuitsVirtualcircuit) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsVirtualcircuitByIDRequest) (*netbox_goV1.UpdateCircuitsVirtualcircuitByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsVirtualcircuit{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDCircuitsVirtualcircuit.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("circuitsVirtualcircuit", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCircuitsVirtualcircuitByIDReply{}, nil
}

// GetByID get a circuitsVirtualcircuit by id
func (s *circuitsVirtualcircuit) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsVirtualcircuitByIDRequest) (*netbox_goV1.GetCircuitsVirtualcircuitByIDReply, error) {
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

	data, err := convertCircuitsVirtualcircuit(record)
	if err != nil {
		logger.Warn("convertCircuitsVirtualcircuit error", logger.Err(err), logger.Any("circuitsVirtualcircuit", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDCircuitsVirtualcircuit.Err()
	}

	return &netbox_goV1.GetCircuitsVirtualcircuitByIDReply{CircuitsVirtualcircuit: data}, nil
}

// List get a paginated list of circuitsVirtualcircuits by custom conditions
func (s *circuitsVirtualcircuit) List(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuitRequest) (*netbox_goV1.ListCircuitsVirtualcircuitReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCircuitsVirtualcircuit.Err()
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

	circuitsVirtualcircuits := []*netbox_goV1.CircuitsVirtualcircuit{}
	for _, record := range records {
		data, err := convertCircuitsVirtualcircuit(record)
		if err != nil {
			logger.Warn("convertCircuitsVirtualcircuit error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsVirtualcircuits = append(circuitsVirtualcircuits, data)
	}

	return &netbox_goV1.ListCircuitsVirtualcircuitReply{
		Total:                   total,
		CircuitsVirtualcircuits: circuitsVirtualcircuits,
	}, nil
}

// DeleteByIDs batch delete circuitsVirtualcircuit by ids
func (s *circuitsVirtualcircuit) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsVirtualcircuitByIDsRequest) (*netbox_goV1.DeleteCircuitsVirtualcircuitByIDsReply, error) {
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

	return &netbox_goV1.DeleteCircuitsVirtualcircuitByIDsReply{}, nil
}

// GetByCondition get a circuitsVirtualcircuit by custom condition
func (s *circuitsVirtualcircuit) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsVirtualcircuitByConditionRequest) (*netbox_goV1.GetCircuitsVirtualcircuitByConditionReply, error) {
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

	data, err := convertCircuitsVirtualcircuit(record)
	if err != nil {
		logger.Warn("convertCircuitsVirtualcircuit error", logger.Err(err), logger.Any("circuitsVirtualcircuit", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCircuitsVirtualcircuit.Err()
	}

	return &netbox_goV1.GetCircuitsVirtualcircuitByConditionReply{
		CircuitsVirtualcircuit: data,
	}, nil
}

// ListByIDs batch get circuitsVirtualcircuit by ids
func (s *circuitsVirtualcircuit) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuitByIDsRequest) (*netbox_goV1.ListCircuitsVirtualcircuitByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	circuitsVirtualcircuitMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	circuitsVirtualcircuits := []*netbox_goV1.CircuitsVirtualcircuit{}
	for _, id := range req.Ids {
		if v, ok := circuitsVirtualcircuitMap[id]; ok {
			record, err := convertCircuitsVirtualcircuit(v)
			if err != nil {
				logger.Warn("convertCircuitsVirtualcircuit error", logger.Err(err), logger.Any("circuitsVirtualcircuit", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			circuitsVirtualcircuits = append(circuitsVirtualcircuits, record)
		}
	}

	return &netbox_goV1.ListCircuitsVirtualcircuitByIDsReply{CircuitsVirtualcircuits: circuitsVirtualcircuits}, nil
}

// ListByLastID get a paginated list of circuitsVirtualcircuits by last id
func (s *circuitsVirtualcircuit) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuitByLastIDRequest) (*netbox_goV1.ListCircuitsVirtualcircuitByLastIDReply, error) {
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

	circuitsVirtualcircuits := []*netbox_goV1.CircuitsVirtualcircuit{}
	for _, record := range records {
		data, err := convertCircuitsVirtualcircuit(record)
		if err != nil {
			logger.Warn("convertCircuitsVirtualcircuit error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsVirtualcircuits = append(circuitsVirtualcircuits, data)
	}

	return &netbox_goV1.ListCircuitsVirtualcircuitByLastIDReply{
		CircuitsVirtualcircuits: circuitsVirtualcircuits,
	}, nil
}

func convertCircuitsVirtualcircuit(record *model.CircuitsVirtualcircuit) (*netbox_goV1.CircuitsVirtualcircuit, error) {
	value := &netbox_goV1.CircuitsVirtualcircuit{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
