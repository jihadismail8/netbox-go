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
		netbox_goV1.RegisterExtrasNotificationServer(server, NewExtrasNotificationServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasNotificationServer = (*extrasNotification)(nil)
var _ time.Time

type extrasNotification struct {
	netbox_goV1.UnimplementedExtrasNotificationServer

	iDao dao.ExtrasNotificationDao
}

// NewExtrasNotificationServer create a new service
func NewExtrasNotificationServer() netbox_goV1.ExtrasNotificationServer {
	return &extrasNotification{
		iDao: dao.NewExtrasNotificationDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasNotificationCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasNotification
func (s *extrasNotification) Create(ctx context.Context, req *netbox_goV1.CreateExtrasNotificationRequest) (*netbox_goV1.CreateExtrasNotificationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasNotification{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasNotification.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasNotification", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasNotificationReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasNotification by id
func (s *extrasNotification) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasNotificationByIDRequest) (*netbox_goV1.DeleteExtrasNotificationByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasNotificationByIDReply{}, nil
}

// UpdateByID update a extrasNotification by id
func (s *extrasNotification) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasNotificationByIDRequest) (*netbox_goV1.UpdateExtrasNotificationByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasNotification{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasNotification.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasNotification", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasNotificationByIDReply{}, nil
}

// GetByID get a extrasNotification by id
func (s *extrasNotification) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasNotificationByIDRequest) (*netbox_goV1.GetExtrasNotificationByIDReply, error) {
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

	data, err := convertExtrasNotification(record)
	if err != nil {
		logger.Warn("convertExtrasNotification error", logger.Err(err), logger.Any("extrasNotification", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasNotification.Err()
	}

	return &netbox_goV1.GetExtrasNotificationByIDReply{ExtrasNotification: data}, nil
}

// List get a paginated list of extrasNotifications by custom conditions
func (s *extrasNotification) List(ctx context.Context, req *netbox_goV1.ListExtrasNotificationRequest) (*netbox_goV1.ListExtrasNotificationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasNotification.Err()
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

	extrasNotifications := []*netbox_goV1.ExtrasNotification{}
	for _, record := range records {
		data, err := convertExtrasNotification(record)
		if err != nil {
			logger.Warn("convertExtrasNotification error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasNotifications = append(extrasNotifications, data)
	}

	return &netbox_goV1.ListExtrasNotificationReply{
		Total:               total,
		ExtrasNotifications: extrasNotifications,
	}, nil
}

// DeleteByIDs batch delete extrasNotification by ids
func (s *extrasNotification) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasNotificationByIDsRequest) (*netbox_goV1.DeleteExtrasNotificationByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasNotificationByIDsReply{}, nil
}

// GetByCondition get a extrasNotification by custom condition
func (s *extrasNotification) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasNotificationByConditionRequest) (*netbox_goV1.GetExtrasNotificationByConditionReply, error) {
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

	data, err := convertExtrasNotification(record)
	if err != nil {
		logger.Warn("convertExtrasNotification error", logger.Err(err), logger.Any("extrasNotification", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasNotification.Err()
	}

	return &netbox_goV1.GetExtrasNotificationByConditionReply{
		ExtrasNotification: data,
	}, nil
}

// ListByIDs batch get extrasNotification by ids
func (s *extrasNotification) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasNotificationByIDsRequest) (*netbox_goV1.ListExtrasNotificationByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasNotificationMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasNotifications := []*netbox_goV1.ExtrasNotification{}
	for _, id := range req.Ids {
		if v, ok := extrasNotificationMap[id]; ok {
			record, err := convertExtrasNotification(v)
			if err != nil {
				logger.Warn("convertExtrasNotification error", logger.Err(err), logger.Any("extrasNotification", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasNotifications = append(extrasNotifications, record)
		}
	}

	return &netbox_goV1.ListExtrasNotificationByIDsReply{ExtrasNotifications: extrasNotifications}, nil
}

// ListByLastID get a paginated list of extrasNotifications by last id
func (s *extrasNotification) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasNotificationByLastIDRequest) (*netbox_goV1.ListExtrasNotificationByLastIDReply, error) {
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

	extrasNotifications := []*netbox_goV1.ExtrasNotification{}
	for _, record := range records {
		data, err := convertExtrasNotification(record)
		if err != nil {
			logger.Warn("convertExtrasNotification error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasNotifications = append(extrasNotifications, data)
	}

	return &netbox_goV1.ListExtrasNotificationByLastIDReply{
		ExtrasNotifications: extrasNotifications,
	}, nil
}

func convertExtrasNotification(record *model.ExtrasNotification) (*netbox_goV1.ExtrasNotification, error) {
	value := &netbox_goV1.ExtrasNotification{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
