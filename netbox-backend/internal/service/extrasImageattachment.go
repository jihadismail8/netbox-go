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
		netbox_goV1.RegisterExtrasImageattachmentServer(server, NewExtrasImageattachmentServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasImageattachmentServer = (*extrasImageattachment)(nil)
var _ time.Time

type extrasImageattachment struct {
	netbox_goV1.UnimplementedExtrasImageattachmentServer

	iDao dao.ExtrasImageattachmentDao
}

// NewExtrasImageattachmentServer create a new service
func NewExtrasImageattachmentServer() netbox_goV1.ExtrasImageattachmentServer {
	return &extrasImageattachment{
		iDao: dao.NewExtrasImageattachmentDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasImageattachmentCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasImageattachment
func (s *extrasImageattachment) Create(ctx context.Context, req *netbox_goV1.CreateExtrasImageattachmentRequest) (*netbox_goV1.CreateExtrasImageattachmentReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasImageattachment{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasImageattachment.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasImageattachment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasImageattachmentReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasImageattachment by id
func (s *extrasImageattachment) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasImageattachmentByIDRequest) (*netbox_goV1.DeleteExtrasImageattachmentByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasImageattachmentByIDReply{}, nil
}

// UpdateByID update a extrasImageattachment by id
func (s *extrasImageattachment) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasImageattachmentByIDRequest) (*netbox_goV1.UpdateExtrasImageattachmentByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasImageattachment{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasImageattachment.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasImageattachment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasImageattachmentByIDReply{}, nil
}

// GetByID get a extrasImageattachment by id
func (s *extrasImageattachment) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasImageattachmentByIDRequest) (*netbox_goV1.GetExtrasImageattachmentByIDReply, error) {
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

	data, err := convertExtrasImageattachment(record)
	if err != nil {
		logger.Warn("convertExtrasImageattachment error", logger.Err(err), logger.Any("extrasImageattachment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasImageattachment.Err()
	}

	return &netbox_goV1.GetExtrasImageattachmentByIDReply{ExtrasImageattachment: data}, nil
}

// List get a paginated list of extrasImageattachments by custom conditions
func (s *extrasImageattachment) List(ctx context.Context, req *netbox_goV1.ListExtrasImageattachmentRequest) (*netbox_goV1.ListExtrasImageattachmentReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasImageattachment.Err()
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

	extrasImageattachments := []*netbox_goV1.ExtrasImageattachment{}
	for _, record := range records {
		data, err := convertExtrasImageattachment(record)
		if err != nil {
			logger.Warn("convertExtrasImageattachment error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasImageattachments = append(extrasImageattachments, data)
	}

	return &netbox_goV1.ListExtrasImageattachmentReply{
		Total:                  total,
		ExtrasImageattachments: extrasImageattachments,
	}, nil
}

// DeleteByIDs batch delete extrasImageattachment by ids
func (s *extrasImageattachment) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasImageattachmentByIDsRequest) (*netbox_goV1.DeleteExtrasImageattachmentByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasImageattachmentByIDsReply{}, nil
}

// GetByCondition get a extrasImageattachment by custom condition
func (s *extrasImageattachment) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasImageattachmentByConditionRequest) (*netbox_goV1.GetExtrasImageattachmentByConditionReply, error) {
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

	data, err := convertExtrasImageattachment(record)
	if err != nil {
		logger.Warn("convertExtrasImageattachment error", logger.Err(err), logger.Any("extrasImageattachment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasImageattachment.Err()
	}

	return &netbox_goV1.GetExtrasImageattachmentByConditionReply{
		ExtrasImageattachment: data,
	}, nil
}

// ListByIDs batch get extrasImageattachment by ids
func (s *extrasImageattachment) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasImageattachmentByIDsRequest) (*netbox_goV1.ListExtrasImageattachmentByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasImageattachmentMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasImageattachments := []*netbox_goV1.ExtrasImageattachment{}
	for _, id := range req.Ids {
		if v, ok := extrasImageattachmentMap[id]; ok {
			record, err := convertExtrasImageattachment(v)
			if err != nil {
				logger.Warn("convertExtrasImageattachment error", logger.Err(err), logger.Any("extrasImageattachment", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasImageattachments = append(extrasImageattachments, record)
		}
	}

	return &netbox_goV1.ListExtrasImageattachmentByIDsReply{ExtrasImageattachments: extrasImageattachments}, nil
}

// ListByLastID get a paginated list of extrasImageattachments by last id
func (s *extrasImageattachment) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasImageattachmentByLastIDRequest) (*netbox_goV1.ListExtrasImageattachmentByLastIDReply, error) {
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

	extrasImageattachments := []*netbox_goV1.ExtrasImageattachment{}
	for _, record := range records {
		data, err := convertExtrasImageattachment(record)
		if err != nil {
			logger.Warn("convertExtrasImageattachment error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasImageattachments = append(extrasImageattachments, data)
	}

	return &netbox_goV1.ListExtrasImageattachmentByLastIDReply{
		ExtrasImageattachments: extrasImageattachments,
	}, nil
}

func convertExtrasImageattachment(record *model.ExtrasImageattachment) (*netbox_goV1.ExtrasImageattachment, error) {
	value := &netbox_goV1.ExtrasImageattachment{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
