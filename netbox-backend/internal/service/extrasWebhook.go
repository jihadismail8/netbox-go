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
		netbox_goV1.RegisterExtrasWebhookServer(server, NewExtrasWebhookServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasWebhookServer = (*extrasWebhook)(nil)
var _ time.Time

type extrasWebhook struct {
	netbox_goV1.UnimplementedExtrasWebhookServer

	iDao dao.ExtrasWebhookDao
}

// NewExtrasWebhookServer create a new service
func NewExtrasWebhookServer() netbox_goV1.ExtrasWebhookServer {
	return &extrasWebhook{
		iDao: dao.NewExtrasWebhookDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasWebhookCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasWebhook
func (s *extrasWebhook) Create(ctx context.Context, req *netbox_goV1.CreateExtrasWebhookRequest) (*netbox_goV1.CreateExtrasWebhookReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasWebhook{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasWebhook.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasWebhook", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasWebhookReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasWebhook by id
func (s *extrasWebhook) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasWebhookByIDRequest) (*netbox_goV1.DeleteExtrasWebhookByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasWebhookByIDReply{}, nil
}

// UpdateByID update a extrasWebhook by id
func (s *extrasWebhook) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasWebhookByIDRequest) (*netbox_goV1.UpdateExtrasWebhookByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasWebhook{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasWebhook.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasWebhook", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasWebhookByIDReply{}, nil
}

// GetByID get a extrasWebhook by id
func (s *extrasWebhook) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasWebhookByIDRequest) (*netbox_goV1.GetExtrasWebhookByIDReply, error) {
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

	data, err := convertExtrasWebhook(record)
	if err != nil {
		logger.Warn("convertExtrasWebhook error", logger.Err(err), logger.Any("extrasWebhook", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasWebhook.Err()
	}

	return &netbox_goV1.GetExtrasWebhookByIDReply{ExtrasWebhook: data}, nil
}

// List get a paginated list of extrasWebhooks by custom conditions
func (s *extrasWebhook) List(ctx context.Context, req *netbox_goV1.ListExtrasWebhookRequest) (*netbox_goV1.ListExtrasWebhookReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasWebhook.Err()
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

	extrasWebhooks := []*netbox_goV1.ExtrasWebhook{}
	for _, record := range records {
		data, err := convertExtrasWebhook(record)
		if err != nil {
			logger.Warn("convertExtrasWebhook error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasWebhooks = append(extrasWebhooks, data)
	}

	return &netbox_goV1.ListExtrasWebhookReply{
		Total:          total,
		ExtrasWebhooks: extrasWebhooks,
	}, nil
}

// DeleteByIDs batch delete extrasWebhook by ids
func (s *extrasWebhook) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasWebhookByIDsRequest) (*netbox_goV1.DeleteExtrasWebhookByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasWebhookByIDsReply{}, nil
}

// GetByCondition get a extrasWebhook by custom condition
func (s *extrasWebhook) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasWebhookByConditionRequest) (*netbox_goV1.GetExtrasWebhookByConditionReply, error) {
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

	data, err := convertExtrasWebhook(record)
	if err != nil {
		logger.Warn("convertExtrasWebhook error", logger.Err(err), logger.Any("extrasWebhook", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasWebhook.Err()
	}

	return &netbox_goV1.GetExtrasWebhookByConditionReply{
		ExtrasWebhook: data,
	}, nil
}

// ListByIDs batch get extrasWebhook by ids
func (s *extrasWebhook) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasWebhookByIDsRequest) (*netbox_goV1.ListExtrasWebhookByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasWebhookMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasWebhooks := []*netbox_goV1.ExtrasWebhook{}
	for _, id := range req.Ids {
		if v, ok := extrasWebhookMap[id]; ok {
			record, err := convertExtrasWebhook(v)
			if err != nil {
				logger.Warn("convertExtrasWebhook error", logger.Err(err), logger.Any("extrasWebhook", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasWebhooks = append(extrasWebhooks, record)
		}
	}

	return &netbox_goV1.ListExtrasWebhookByIDsReply{ExtrasWebhooks: extrasWebhooks}, nil
}

// ListByLastID get a paginated list of extrasWebhooks by last id
func (s *extrasWebhook) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasWebhookByLastIDRequest) (*netbox_goV1.ListExtrasWebhookByLastIDReply, error) {
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

	extrasWebhooks := []*netbox_goV1.ExtrasWebhook{}
	for _, record := range records {
		data, err := convertExtrasWebhook(record)
		if err != nil {
			logger.Warn("convertExtrasWebhook error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasWebhooks = append(extrasWebhooks, data)
	}

	return &netbox_goV1.ListExtrasWebhookByLastIDReply{
		ExtrasWebhooks: extrasWebhooks,
	}, nil
}

func convertExtrasWebhook(record *model.ExtrasWebhook) (*netbox_goV1.ExtrasWebhook, error) {
	value := &netbox_goV1.ExtrasWebhook{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
