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
		netbox_goV1.RegisterCircuitsCircuitgroupassignmentServer(server, NewCircuitsCircuitgroupassignmentServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CircuitsCircuitgroupassignmentServer = (*circuitsCircuitgroupassignment)(nil)
var _ time.Time

type circuitsCircuitgroupassignment struct {
	netbox_goV1.UnimplementedCircuitsCircuitgroupassignmentServer

	iDao dao.CircuitsCircuitgroupassignmentDao
}

// NewCircuitsCircuitgroupassignmentServer create a new service
func NewCircuitsCircuitgroupassignmentServer() netbox_goV1.CircuitsCircuitgroupassignmentServer {
	return &circuitsCircuitgroupassignment{
		iDao: dao.NewCircuitsCircuitgroupassignmentDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCircuitsCircuitgroupassignmentCache(database.GetCacheType()),
		),
	}
}

// Create a new circuitsCircuitgroupassignment
func (s *circuitsCircuitgroupassignment) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsCircuitgroupassignmentRequest) (*netbox_goV1.CreateCircuitsCircuitgroupassignmentReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsCircuitgroupassignment{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCircuitsCircuitgroupassignment.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("circuitsCircuitgroupassignment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCircuitsCircuitgroupassignmentReply{Id: record.ID}, nil
}

// DeleteByID delete a circuitsCircuitgroupassignment by id
func (s *circuitsCircuitgroupassignment) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitgroupassignmentByIDRequest) (*netbox_goV1.DeleteCircuitsCircuitgroupassignmentByIDReply, error) {
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

	return &netbox_goV1.DeleteCircuitsCircuitgroupassignmentByIDReply{}, nil
}

// UpdateByID update a circuitsCircuitgroupassignment by id
func (s *circuitsCircuitgroupassignment) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsCircuitgroupassignmentByIDRequest) (*netbox_goV1.UpdateCircuitsCircuitgroupassignmentByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsCircuitgroupassignment{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDCircuitsCircuitgroupassignment.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("circuitsCircuitgroupassignment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCircuitsCircuitgroupassignmentByIDReply{}, nil
}

// GetByID get a circuitsCircuitgroupassignment by id
func (s *circuitsCircuitgroupassignment) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitgroupassignmentByIDRequest) (*netbox_goV1.GetCircuitsCircuitgroupassignmentByIDReply, error) {
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

	data, err := convertCircuitsCircuitgroupassignment(record)
	if err != nil {
		logger.Warn("convertCircuitsCircuitgroupassignment error", logger.Err(err), logger.Any("circuitsCircuitgroupassignment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDCircuitsCircuitgroupassignment.Err()
	}

	return &netbox_goV1.GetCircuitsCircuitgroupassignmentByIDReply{CircuitsCircuitgroupassignment: data}, nil
}

// List get a paginated list of circuitsCircuitgroupassignments by custom conditions
func (s *circuitsCircuitgroupassignment) List(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitgroupassignmentRequest) (*netbox_goV1.ListCircuitsCircuitgroupassignmentReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCircuitsCircuitgroupassignment.Err()
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

	circuitsCircuitgroupassignments := []*netbox_goV1.CircuitsCircuitgroupassignment{}
	for _, record := range records {
		data, err := convertCircuitsCircuitgroupassignment(record)
		if err != nil {
			logger.Warn("convertCircuitsCircuitgroupassignment error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsCircuitgroupassignments = append(circuitsCircuitgroupassignments, data)
	}

	return &netbox_goV1.ListCircuitsCircuitgroupassignmentReply{
		Total:                           total,
		CircuitsCircuitgroupassignments: circuitsCircuitgroupassignments,
	}, nil
}

// DeleteByIDs batch delete circuitsCircuitgroupassignment by ids
func (s *circuitsCircuitgroupassignment) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitgroupassignmentByIDsRequest) (*netbox_goV1.DeleteCircuitsCircuitgroupassignmentByIDsReply, error) {
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

	return &netbox_goV1.DeleteCircuitsCircuitgroupassignmentByIDsReply{}, nil
}

// GetByCondition get a circuitsCircuitgroupassignment by custom condition
func (s *circuitsCircuitgroupassignment) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitgroupassignmentByConditionRequest) (*netbox_goV1.GetCircuitsCircuitgroupassignmentByConditionReply, error) {
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

	data, err := convertCircuitsCircuitgroupassignment(record)
	if err != nil {
		logger.Warn("convertCircuitsCircuitgroupassignment error", logger.Err(err), logger.Any("circuitsCircuitgroupassignment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCircuitsCircuitgroupassignment.Err()
	}

	return &netbox_goV1.GetCircuitsCircuitgroupassignmentByConditionReply{
		CircuitsCircuitgroupassignment: data,
	}, nil
}

// ListByIDs batch get circuitsCircuitgroupassignment by ids
func (s *circuitsCircuitgroupassignment) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitgroupassignmentByIDsRequest) (*netbox_goV1.ListCircuitsCircuitgroupassignmentByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	circuitsCircuitgroupassignmentMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	circuitsCircuitgroupassignments := []*netbox_goV1.CircuitsCircuitgroupassignment{}
	for _, id := range req.Ids {
		if v, ok := circuitsCircuitgroupassignmentMap[id]; ok {
			record, err := convertCircuitsCircuitgroupassignment(v)
			if err != nil {
				logger.Warn("convertCircuitsCircuitgroupassignment error", logger.Err(err), logger.Any("circuitsCircuitgroupassignment", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			circuitsCircuitgroupassignments = append(circuitsCircuitgroupassignments, record)
		}
	}

	return &netbox_goV1.ListCircuitsCircuitgroupassignmentByIDsReply{CircuitsCircuitgroupassignments: circuitsCircuitgroupassignments}, nil
}

// ListByLastID get a paginated list of circuitsCircuitgroupassignments by last id
func (s *circuitsCircuitgroupassignment) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitgroupassignmentByLastIDRequest) (*netbox_goV1.ListCircuitsCircuitgroupassignmentByLastIDReply, error) {
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

	circuitsCircuitgroupassignments := []*netbox_goV1.CircuitsCircuitgroupassignment{}
	for _, record := range records {
		data, err := convertCircuitsCircuitgroupassignment(record)
		if err != nil {
			logger.Warn("convertCircuitsCircuitgroupassignment error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsCircuitgroupassignments = append(circuitsCircuitgroupassignments, data)
	}

	return &netbox_goV1.ListCircuitsCircuitgroupassignmentByLastIDReply{
		CircuitsCircuitgroupassignments: circuitsCircuitgroupassignments,
	}, nil
}

func convertCircuitsCircuitgroupassignment(record *model.CircuitsCircuitgroupassignment) (*netbox_goV1.CircuitsCircuitgroupassignment, error) {
	value := &netbox_goV1.CircuitsCircuitgroupassignment{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
