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
		netbox_goV1.RegisterCircuitsCircuittypeServer(server, NewCircuitsCircuittypeServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CircuitsCircuittypeServer = (*circuitsCircuittype)(nil)
var _ time.Time

type circuitsCircuittype struct {
	netbox_goV1.UnimplementedCircuitsCircuittypeServer

	iDao dao.CircuitsCircuittypeDao
}

// NewCircuitsCircuittypeServer create a new service
func NewCircuitsCircuittypeServer() netbox_goV1.CircuitsCircuittypeServer {
	return &circuitsCircuittype{
		iDao: dao.NewCircuitsCircuittypeDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCircuitsCircuittypeCache(database.GetCacheType()),
		),
	}
}

// Create a new circuitsCircuittype
func (s *circuitsCircuittype) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsCircuittypeRequest) (*netbox_goV1.CreateCircuitsCircuittypeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsCircuittype{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCircuitsCircuittype.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("circuitsCircuittype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCircuitsCircuittypeReply{Id: record.ID}, nil
}

// DeleteByID delete a circuitsCircuittype by id
func (s *circuitsCircuittype) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuittypeByIDRequest) (*netbox_goV1.DeleteCircuitsCircuittypeByIDReply, error) {
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

	return &netbox_goV1.DeleteCircuitsCircuittypeByIDReply{}, nil
}

// UpdateByID update a circuitsCircuittype by id
func (s *circuitsCircuittype) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsCircuittypeByIDRequest) (*netbox_goV1.UpdateCircuitsCircuittypeByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsCircuittype{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDCircuitsCircuittype.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("circuitsCircuittype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCircuitsCircuittypeByIDReply{}, nil
}

// GetByID get a circuitsCircuittype by id
func (s *circuitsCircuittype) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsCircuittypeByIDRequest) (*netbox_goV1.GetCircuitsCircuittypeByIDReply, error) {
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

	data, err := convertCircuitsCircuittype(record)
	if err != nil {
		logger.Warn("convertCircuitsCircuittype error", logger.Err(err), logger.Any("circuitsCircuittype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDCircuitsCircuittype.Err()
	}

	return &netbox_goV1.GetCircuitsCircuittypeByIDReply{CircuitsCircuittype: data}, nil
}

// List get a paginated list of circuitsCircuittypes by custom conditions
func (s *circuitsCircuittype) List(ctx context.Context, req *netbox_goV1.ListCircuitsCircuittypeRequest) (*netbox_goV1.ListCircuitsCircuittypeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCircuitsCircuittype.Err()
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

	circuitsCircuittypes := []*netbox_goV1.CircuitsCircuittype{}
	for _, record := range records {
		data, err := convertCircuitsCircuittype(record)
		if err != nil {
			logger.Warn("convertCircuitsCircuittype error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsCircuittypes = append(circuitsCircuittypes, data)
	}

	return &netbox_goV1.ListCircuitsCircuittypeReply{
		Total:                total,
		CircuitsCircuittypes: circuitsCircuittypes,
	}, nil
}

// DeleteByIDs batch delete circuitsCircuittype by ids
func (s *circuitsCircuittype) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsCircuittypeByIDsRequest) (*netbox_goV1.DeleteCircuitsCircuittypeByIDsReply, error) {
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

	return &netbox_goV1.DeleteCircuitsCircuittypeByIDsReply{}, nil
}

// GetByCondition get a circuitsCircuittype by custom condition
func (s *circuitsCircuittype) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsCircuittypeByConditionRequest) (*netbox_goV1.GetCircuitsCircuittypeByConditionReply, error) {
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

	data, err := convertCircuitsCircuittype(record)
	if err != nil {
		logger.Warn("convertCircuitsCircuittype error", logger.Err(err), logger.Any("circuitsCircuittype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCircuitsCircuittype.Err()
	}

	return &netbox_goV1.GetCircuitsCircuittypeByConditionReply{
		CircuitsCircuittype: data,
	}, nil
}

// ListByIDs batch get circuitsCircuittype by ids
func (s *circuitsCircuittype) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsCircuittypeByIDsRequest) (*netbox_goV1.ListCircuitsCircuittypeByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	circuitsCircuittypeMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	circuitsCircuittypes := []*netbox_goV1.CircuitsCircuittype{}
	for _, id := range req.Ids {
		if v, ok := circuitsCircuittypeMap[id]; ok {
			record, err := convertCircuitsCircuittype(v)
			if err != nil {
				logger.Warn("convertCircuitsCircuittype error", logger.Err(err), logger.Any("circuitsCircuittype", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			circuitsCircuittypes = append(circuitsCircuittypes, record)
		}
	}

	return &netbox_goV1.ListCircuitsCircuittypeByIDsReply{CircuitsCircuittypes: circuitsCircuittypes}, nil
}

// ListByLastID get a paginated list of circuitsCircuittypes by last id
func (s *circuitsCircuittype) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsCircuittypeByLastIDRequest) (*netbox_goV1.ListCircuitsCircuittypeByLastIDReply, error) {
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

	circuitsCircuittypes := []*netbox_goV1.CircuitsCircuittype{}
	for _, record := range records {
		data, err := convertCircuitsCircuittype(record)
		if err != nil {
			logger.Warn("convertCircuitsCircuittype error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsCircuittypes = append(circuitsCircuittypes, data)
	}

	return &netbox_goV1.ListCircuitsCircuittypeByLastIDReply{
		CircuitsCircuittypes: circuitsCircuittypes,
	}, nil
}

func convertCircuitsCircuittype(record *model.CircuitsCircuittype) (*netbox_goV1.CircuitsCircuittype, error) {
	value := &netbox_goV1.CircuitsCircuittype{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
