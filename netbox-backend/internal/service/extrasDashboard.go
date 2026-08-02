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
		netbox_goV1.RegisterExtrasDashboardServer(server, NewExtrasDashboardServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasDashboardServer = (*extrasDashboard)(nil)
var _ time.Time

type extrasDashboard struct {
	netbox_goV1.UnimplementedExtrasDashboardServer

	iDao dao.ExtrasDashboardDao
}

// NewExtrasDashboardServer create a new service
func NewExtrasDashboardServer() netbox_goV1.ExtrasDashboardServer {
	return &extrasDashboard{
		iDao: dao.NewExtrasDashboardDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasDashboardCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasDashboard
func (s *extrasDashboard) Create(ctx context.Context, req *netbox_goV1.CreateExtrasDashboardRequest) (*netbox_goV1.CreateExtrasDashboardReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasDashboard{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasDashboard.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasDashboard", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasDashboardReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasDashboard by id
func (s *extrasDashboard) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasDashboardByIDRequest) (*netbox_goV1.DeleteExtrasDashboardByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasDashboardByIDReply{}, nil
}

// UpdateByID update a extrasDashboard by id
func (s *extrasDashboard) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasDashboardByIDRequest) (*netbox_goV1.UpdateExtrasDashboardByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasDashboard{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasDashboard.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasDashboard", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasDashboardByIDReply{}, nil
}

// GetByID get a extrasDashboard by id
func (s *extrasDashboard) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasDashboardByIDRequest) (*netbox_goV1.GetExtrasDashboardByIDReply, error) {
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

	data, err := convertExtrasDashboard(record)
	if err != nil {
		logger.Warn("convertExtrasDashboard error", logger.Err(err), logger.Any("extrasDashboard", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasDashboard.Err()
	}

	return &netbox_goV1.GetExtrasDashboardByIDReply{ExtrasDashboard: data}, nil
}

// List get a paginated list of extrasDashboards by custom conditions
func (s *extrasDashboard) List(ctx context.Context, req *netbox_goV1.ListExtrasDashboardRequest) (*netbox_goV1.ListExtrasDashboardReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasDashboard.Err()
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

	extrasDashboards := []*netbox_goV1.ExtrasDashboard{}
	for _, record := range records {
		data, err := convertExtrasDashboard(record)
		if err != nil {
			logger.Warn("convertExtrasDashboard error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasDashboards = append(extrasDashboards, data)
	}

	return &netbox_goV1.ListExtrasDashboardReply{
		Total:            total,
		ExtrasDashboards: extrasDashboards,
	}, nil
}

// DeleteByIDs batch delete extrasDashboard by ids
func (s *extrasDashboard) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasDashboardByIDsRequest) (*netbox_goV1.DeleteExtrasDashboardByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasDashboardByIDsReply{}, nil
}

// GetByCondition get a extrasDashboard by custom condition
func (s *extrasDashboard) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasDashboardByConditionRequest) (*netbox_goV1.GetExtrasDashboardByConditionReply, error) {
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

	data, err := convertExtrasDashboard(record)
	if err != nil {
		logger.Warn("convertExtrasDashboard error", logger.Err(err), logger.Any("extrasDashboard", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasDashboard.Err()
	}

	return &netbox_goV1.GetExtrasDashboardByConditionReply{
		ExtrasDashboard: data,
	}, nil
}

// ListByIDs batch get extrasDashboard by ids
func (s *extrasDashboard) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasDashboardByIDsRequest) (*netbox_goV1.ListExtrasDashboardByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasDashboardMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasDashboards := []*netbox_goV1.ExtrasDashboard{}
	for _, id := range req.Ids {
		if v, ok := extrasDashboardMap[id]; ok {
			record, err := convertExtrasDashboard(v)
			if err != nil {
				logger.Warn("convertExtrasDashboard error", logger.Err(err), logger.Any("extrasDashboard", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasDashboards = append(extrasDashboards, record)
		}
	}

	return &netbox_goV1.ListExtrasDashboardByIDsReply{ExtrasDashboards: extrasDashboards}, nil
}

// ListByLastID get a paginated list of extrasDashboards by last id
func (s *extrasDashboard) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasDashboardByLastIDRequest) (*netbox_goV1.ListExtrasDashboardByLastIDReply, error) {
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

	extrasDashboards := []*netbox_goV1.ExtrasDashboard{}
	for _, record := range records {
		data, err := convertExtrasDashboard(record)
		if err != nil {
			logger.Warn("convertExtrasDashboard error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasDashboards = append(extrasDashboards, data)
	}

	return &netbox_goV1.ListExtrasDashboardByLastIDReply{
		ExtrasDashboards: extrasDashboards,
	}, nil
}

func convertExtrasDashboard(record *model.ExtrasDashboard) (*netbox_goV1.ExtrasDashboard, error) {
	value := &netbox_goV1.ExtrasDashboard{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
