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
		netbox_goV1.RegisterCircuitsProvidernetworkServer(server, NewCircuitsProvidernetworkServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CircuitsProvidernetworkServer = (*circuitsProvidernetwork)(nil)
var _ time.Time

type circuitsProvidernetwork struct {
	netbox_goV1.UnimplementedCircuitsProvidernetworkServer

	iDao dao.CircuitsProvidernetworkDao
}

// NewCircuitsProvidernetworkServer create a new service
func NewCircuitsProvidernetworkServer() netbox_goV1.CircuitsProvidernetworkServer {
	return &circuitsProvidernetwork{
		iDao: dao.NewCircuitsProvidernetworkDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCircuitsProvidernetworkCache(database.GetCacheType()),
		),
	}
}

// Create a new circuitsProvidernetwork
func (s *circuitsProvidernetwork) Create(ctx context.Context, req *netbox_goV1.CreateCircuitsProvidernetworkRequest) (*netbox_goV1.CreateCircuitsProvidernetworkReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsProvidernetwork{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCircuitsProvidernetwork.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("circuitsProvidernetwork", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCircuitsProvidernetworkReply{Id: record.ID}, nil
}

// DeleteByID delete a circuitsProvidernetwork by id
func (s *circuitsProvidernetwork) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCircuitsProvidernetworkByIDRequest) (*netbox_goV1.DeleteCircuitsProvidernetworkByIDReply, error) {
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

	return &netbox_goV1.DeleteCircuitsProvidernetworkByIDReply{}, nil
}

// UpdateByID update a circuitsProvidernetwork by id
func (s *circuitsProvidernetwork) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCircuitsProvidernetworkByIDRequest) (*netbox_goV1.UpdateCircuitsProvidernetworkByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CircuitsProvidernetwork{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDCircuitsProvidernetwork.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("circuitsProvidernetwork", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCircuitsProvidernetworkByIDReply{}, nil
}

// GetByID get a circuitsProvidernetwork by id
func (s *circuitsProvidernetwork) GetByID(ctx context.Context, req *netbox_goV1.GetCircuitsProvidernetworkByIDRequest) (*netbox_goV1.GetCircuitsProvidernetworkByIDReply, error) {
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

	data, err := convertCircuitsProvidernetwork(record)
	if err != nil {
		logger.Warn("convertCircuitsProvidernetwork error", logger.Err(err), logger.Any("circuitsProvidernetwork", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDCircuitsProvidernetwork.Err()
	}

	return &netbox_goV1.GetCircuitsProvidernetworkByIDReply{CircuitsProvidernetwork: data}, nil
}

// List get a paginated list of circuitsProvidernetworks by custom conditions
func (s *circuitsProvidernetwork) List(ctx context.Context, req *netbox_goV1.ListCircuitsProvidernetworkRequest) (*netbox_goV1.ListCircuitsProvidernetworkReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCircuitsProvidernetwork.Err()
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

	circuitsProvidernetworks := []*netbox_goV1.CircuitsProvidernetwork{}
	for _, record := range records {
		data, err := convertCircuitsProvidernetwork(record)
		if err != nil {
			logger.Warn("convertCircuitsProvidernetwork error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsProvidernetworks = append(circuitsProvidernetworks, data)
	}

	return &netbox_goV1.ListCircuitsProvidernetworkReply{
		Total:                    total,
		CircuitsProvidernetworks: circuitsProvidernetworks,
	}, nil
}

// DeleteByIDs batch delete circuitsProvidernetwork by ids
func (s *circuitsProvidernetwork) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCircuitsProvidernetworkByIDsRequest) (*netbox_goV1.DeleteCircuitsProvidernetworkByIDsReply, error) {
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

	return &netbox_goV1.DeleteCircuitsProvidernetworkByIDsReply{}, nil
}

// GetByCondition get a circuitsProvidernetwork by custom condition
func (s *circuitsProvidernetwork) GetByCondition(ctx context.Context, req *netbox_goV1.GetCircuitsProvidernetworkByConditionRequest) (*netbox_goV1.GetCircuitsProvidernetworkByConditionReply, error) {
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

	data, err := convertCircuitsProvidernetwork(record)
	if err != nil {
		logger.Warn("convertCircuitsProvidernetwork error", logger.Err(err), logger.Any("circuitsProvidernetwork", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCircuitsProvidernetwork.Err()
	}

	return &netbox_goV1.GetCircuitsProvidernetworkByConditionReply{
		CircuitsProvidernetwork: data,
	}, nil
}

// ListByIDs batch get circuitsProvidernetwork by ids
func (s *circuitsProvidernetwork) ListByIDs(ctx context.Context, req *netbox_goV1.ListCircuitsProvidernetworkByIDsRequest) (*netbox_goV1.ListCircuitsProvidernetworkByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	circuitsProvidernetworkMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	circuitsProvidernetworks := []*netbox_goV1.CircuitsProvidernetwork{}
	for _, id := range req.Ids {
		if v, ok := circuitsProvidernetworkMap[id]; ok {
			record, err := convertCircuitsProvidernetwork(v)
			if err != nil {
				logger.Warn("convertCircuitsProvidernetwork error", logger.Err(err), logger.Any("circuitsProvidernetwork", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			circuitsProvidernetworks = append(circuitsProvidernetworks, record)
		}
	}

	return &netbox_goV1.ListCircuitsProvidernetworkByIDsReply{CircuitsProvidernetworks: circuitsProvidernetworks}, nil
}

// ListByLastID get a paginated list of circuitsProvidernetworks by last id
func (s *circuitsProvidernetwork) ListByLastID(ctx context.Context, req *netbox_goV1.ListCircuitsProvidernetworkByLastIDRequest) (*netbox_goV1.ListCircuitsProvidernetworkByLastIDReply, error) {
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

	circuitsProvidernetworks := []*netbox_goV1.CircuitsProvidernetwork{}
	for _, record := range records {
		data, err := convertCircuitsProvidernetwork(record)
		if err != nil {
			logger.Warn("convertCircuitsProvidernetwork error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		circuitsProvidernetworks = append(circuitsProvidernetworks, data)
	}

	return &netbox_goV1.ListCircuitsProvidernetworkByLastIDReply{
		CircuitsProvidernetworks: circuitsProvidernetworks,
	}, nil
}

func convertCircuitsProvidernetwork(record *model.CircuitsProvidernetwork) (*netbox_goV1.CircuitsProvidernetwork, error) {
	value := &netbox_goV1.CircuitsProvidernetwork{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
