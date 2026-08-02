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
		netbox_goV1.RegisterCircuitsVirtualcircuittypeServer(server, NewCircuitsVirtualcircuittypeServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CircuitsVirtualcircuittypeServer = (*circuitsVirtualcircuittype)(nil)
var _ time.Time

type circuitsVirtualcircuittype struct {
	netbox_goV1.UnimplementedCircuitsVirtualcircuittypeServer

	iDao dao.CircuitsVirtualcircuittypeDao
}

// NewCircuitsVirtualcircuittypeServer create a new service
func NewCircuitsVirtualcircuittypeServer() netbox_goV1.CircuitsVirtualcircuittypeServer {
	return &circuitsVirtualcircuittype{
		iDao: dao.NewCircuitsVirtualcircuittypeDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCircuitsVirtualcircuittypeCache(database.GetCacheType()),
		),
	}
}

// Create a new circuitsVirtualcircuittype
func (s *circuitsVirtualcircuittype) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsVirtualcircuittypeRequest) (*netbox_goV1.CreateCircuitsVirtualcircuittypeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsVirtualcircuittype{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCircuitsVirtualcircuittype.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("circuitsVirtualcircuittype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCircuitsVirtualcircuittypeReply{Id: record.ID}, nil
}

// DeleteByID delete a circuitsVirtualcircuittype by id
func (s *circuitsVirtualcircuittype) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsVirtualcircuittypeByIDRequest) (*netbox_goV1.DeleteCircuitsVirtualcircuittypeByIDReply, error) {
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

	return &netbox_goV1.DeleteCircuitsVirtualcircuittypeByIDReply{}, nil
}

// UpdateByID update a circuitsVirtualcircuittype by id
func (s *circuitsVirtualcircuittype) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsVirtualcircuittypeByIDRequest) (*netbox_goV1.UpdateCircuitsVirtualcircuittypeByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsVirtualcircuittype{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDCircuitsVirtualcircuittype.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("circuitsVirtualcircuittype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCircuitsVirtualcircuittypeByIDReply{}, nil
}

// GetByID get a circuitsVirtualcircuittype by id
func (s *circuitsVirtualcircuittype) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsVirtualcircuittypeByIDRequest) (*netbox_goV1.GetCircuitsVirtualcircuittypeByIDReply, error) {
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

	data, err := convertCircuitsVirtualcircuittype(record)
	if err != nil {
		logger.Warn("convertCircuitsVirtualcircuittype error", logger.Err(err), logger.Any("circuitsVirtualcircuittype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDCircuitsVirtualcircuittype.Err()
	}

	return &netbox_goV1.GetCircuitsVirtualcircuittypeByIDReply{CircuitsVirtualcircuittype: data}, nil
}

// List get a paginated list of circuitsVirtualcircuittypes by custom conditions
func (s *circuitsVirtualcircuittype) List(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuittypeRequest) (*netbox_goV1.ListCircuitsVirtualcircuittypeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCircuitsVirtualcircuittype.Err()
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

	circuitsVirtualcircuittypes := []*netbox_goV1.CircuitsVirtualcircuittype{}
	for _, record := range records {
		data, err := convertCircuitsVirtualcircuittype(record)
		if err != nil {
			logger.Warn("convertCircuitsVirtualcircuittype error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsVirtualcircuittypes = append(circuitsVirtualcircuittypes, data)
	}

	return &netbox_goV1.ListCircuitsVirtualcircuittypeReply{
		Total:                       total,
		CircuitsVirtualcircuittypes: circuitsVirtualcircuittypes,
	}, nil
}

// DeleteByIDs batch delete circuitsVirtualcircuittype by ids
func (s *circuitsVirtualcircuittype) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsVirtualcircuittypeByIDsRequest) (*netbox_goV1.DeleteCircuitsVirtualcircuittypeByIDsReply, error) {
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

	return &netbox_goV1.DeleteCircuitsVirtualcircuittypeByIDsReply{}, nil
}

// GetByCondition get a circuitsVirtualcircuittype by custom condition
func (s *circuitsVirtualcircuittype) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsVirtualcircuittypeByConditionRequest) (*netbox_goV1.GetCircuitsVirtualcircuittypeByConditionReply, error) {
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

	data, err := convertCircuitsVirtualcircuittype(record)
	if err != nil {
		logger.Warn("convertCircuitsVirtualcircuittype error", logger.Err(err), logger.Any("circuitsVirtualcircuittype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCircuitsVirtualcircuittype.Err()
	}

	return &netbox_goV1.GetCircuitsVirtualcircuittypeByConditionReply{
		CircuitsVirtualcircuittype: data,
	}, nil
}

// ListByIDs batch get circuitsVirtualcircuittype by ids
func (s *circuitsVirtualcircuittype) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuittypeByIDsRequest) (*netbox_goV1.ListCircuitsVirtualcircuittypeByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	circuitsVirtualcircuittypeMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	circuitsVirtualcircuittypes := []*netbox_goV1.CircuitsVirtualcircuittype{}
	for _, id := range req.Ids {
		if v, ok := circuitsVirtualcircuittypeMap[id]; ok {
			record, err := convertCircuitsVirtualcircuittype(v)
			if err != nil {
				logger.Warn("convertCircuitsVirtualcircuittype error", logger.Err(err), logger.Any("circuitsVirtualcircuittype", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			circuitsVirtualcircuittypes = append(circuitsVirtualcircuittypes, record)
		}
	}

	return &netbox_goV1.ListCircuitsVirtualcircuittypeByIDsReply{CircuitsVirtualcircuittypes: circuitsVirtualcircuittypes}, nil
}

// ListByLastID get a paginated list of circuitsVirtualcircuittypes by last id
func (s *circuitsVirtualcircuittype) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsVirtualcircuittypeByLastIDRequest) (*netbox_goV1.ListCircuitsVirtualcircuittypeByLastIDReply, error) {
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

	circuitsVirtualcircuittypes := []*netbox_goV1.CircuitsVirtualcircuittype{}
	for _, record := range records {
		data, err := convertCircuitsVirtualcircuittype(record)
		if err != nil {
			logger.Warn("convertCircuitsVirtualcircuittype error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsVirtualcircuittypes = append(circuitsVirtualcircuittypes, data)
	}

	return &netbox_goV1.ListCircuitsVirtualcircuittypeByLastIDReply{
		CircuitsVirtualcircuittypes: circuitsVirtualcircuittypes,
	}, nil
}

func convertCircuitsVirtualcircuittype(record *model.CircuitsVirtualcircuittype) (*netbox_goV1.CircuitsVirtualcircuittype, error) {
	value := &netbox_goV1.CircuitsVirtualcircuittype{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
