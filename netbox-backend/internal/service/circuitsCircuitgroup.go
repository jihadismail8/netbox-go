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
		netbox_goV1.RegisterCircuitsCircuitgroupServer(server, NewCircuitsCircuitgroupServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CircuitsCircuitgroupServer = (*circuitsCircuitgroup)(nil)
var _ time.Time

type circuitsCircuitgroup struct {
	netbox_goV1.UnimplementedCircuitsCircuitgroupServer

	iDao dao.CircuitsCircuitgroupDao
}

// NewCircuitsCircuitgroupServer create a new service
func NewCircuitsCircuitgroupServer() netbox_goV1.CircuitsCircuitgroupServer {
	return &circuitsCircuitgroup{
		iDao: dao.NewCircuitsCircuitgroupDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCircuitsCircuitgroupCache(database.GetCacheType()),
		),
	}
}

// Create a new circuitsCircuitgroup
func (s *circuitsCircuitgroup) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsCircuitgroupRequest) (*netbox_goV1.CreateCircuitsCircuitgroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsCircuitgroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCircuitsCircuitgroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("circuitsCircuitgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCircuitsCircuitgroupReply{Id: record.ID}, nil
}

// DeleteByID delete a circuitsCircuitgroup by id
func (s *circuitsCircuitgroup) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitgroupByIDRequest) (*netbox_goV1.DeleteCircuitsCircuitgroupByIDReply, error) {
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

	return &netbox_goV1.DeleteCircuitsCircuitgroupByIDReply{}, nil
}

// UpdateByID update a circuitsCircuitgroup by id
func (s *circuitsCircuitgroup) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsCircuitgroupByIDRequest) (*netbox_goV1.UpdateCircuitsCircuitgroupByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsCircuitgroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDCircuitsCircuitgroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("circuitsCircuitgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCircuitsCircuitgroupByIDReply{}, nil
}

// GetByID get a circuitsCircuitgroup by id
func (s *circuitsCircuitgroup) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitgroupByIDRequest) (*netbox_goV1.GetCircuitsCircuitgroupByIDReply, error) {
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

	data, err := convertCircuitsCircuitgroup(record)
	if err != nil {
		logger.Warn("convertCircuitsCircuitgroup error", logger.Err(err), logger.Any("circuitsCircuitgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDCircuitsCircuitgroup.Err()
	}

	return &netbox_goV1.GetCircuitsCircuitgroupByIDReply{CircuitsCircuitgroup: data}, nil
}

// List get a paginated list of circuitsCircuitgroups by custom conditions
func (s *circuitsCircuitgroup) List(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitgroupRequest) (*netbox_goV1.ListCircuitsCircuitgroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCircuitsCircuitgroup.Err()
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

	circuitsCircuitgroups := []*netbox_goV1.CircuitsCircuitgroup{}
	for _, record := range records {
		data, err := convertCircuitsCircuitgroup(record)
		if err != nil {
			logger.Warn("convertCircuitsCircuitgroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsCircuitgroups = append(circuitsCircuitgroups, data)
	}

	return &netbox_goV1.ListCircuitsCircuitgroupReply{
		Total:                 total,
		CircuitsCircuitgroups: circuitsCircuitgroups,
	}, nil
}

// DeleteByIDs batch delete circuitsCircuitgroup by ids
func (s *circuitsCircuitgroup) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuitgroupByIDsRequest) (*netbox_goV1.DeleteCircuitsCircuitgroupByIDsReply, error) {
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

	return &netbox_goV1.DeleteCircuitsCircuitgroupByIDsReply{}, nil
}

// GetByCondition get a circuitsCircuitgroup by custom condition
func (s *circuitsCircuitgroup) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsCircuitgroupByConditionRequest) (*netbox_goV1.GetCircuitsCircuitgroupByConditionReply, error) {
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

	data, err := convertCircuitsCircuitgroup(record)
	if err != nil {
		logger.Warn("convertCircuitsCircuitgroup error", logger.Err(err), logger.Any("circuitsCircuitgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCircuitsCircuitgroup.Err()
	}

	return &netbox_goV1.GetCircuitsCircuitgroupByConditionReply{
		CircuitsCircuitgroup: data,
	}, nil
}

// ListByIDs batch get circuitsCircuitgroup by ids
func (s *circuitsCircuitgroup) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitgroupByIDsRequest) (*netbox_goV1.ListCircuitsCircuitgroupByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	circuitsCircuitgroupMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	circuitsCircuitgroups := []*netbox_goV1.CircuitsCircuitgroup{}
	for _, id := range req.Ids {
		if v, ok := circuitsCircuitgroupMap[id]; ok {
			record, err := convertCircuitsCircuitgroup(v)
			if err != nil {
				logger.Warn("convertCircuitsCircuitgroup error", logger.Err(err), logger.Any("circuitsCircuitgroup", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			circuitsCircuitgroups = append(circuitsCircuitgroups, record)
		}
	}

	return &netbox_goV1.ListCircuitsCircuitgroupByIDsReply{CircuitsCircuitgroups: circuitsCircuitgroups}, nil
}

// ListByLastID get a paginated list of circuitsCircuitgroups by last id
func (s *circuitsCircuitgroup) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsCircuitgroupByLastIDRequest) (*netbox_goV1.ListCircuitsCircuitgroupByLastIDReply, error) {
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

	circuitsCircuitgroups := []*netbox_goV1.CircuitsCircuitgroup{}
	for _, record := range records {
		data, err := convertCircuitsCircuitgroup(record)
		if err != nil {
			logger.Warn("convertCircuitsCircuitgroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsCircuitgroups = append(circuitsCircuitgroups, data)
	}

	return &netbox_goV1.ListCircuitsCircuitgroupByLastIDReply{
		CircuitsCircuitgroups: circuitsCircuitgroups,
	}, nil
}

func convertCircuitsCircuitgroup(record *model.CircuitsCircuitgroup) (*netbox_goV1.CircuitsCircuitgroup, error) {
	value := &netbox_goV1.CircuitsCircuitgroup{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
