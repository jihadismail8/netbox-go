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
		netbox_goV1.RegisterExtrasCustomfieldchoicesetServer(server, NewExtrasCustomfieldchoicesetServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasCustomfieldchoicesetServer = (*extrasCustomfieldchoiceset)(nil)
var _ time.Time

type extrasCustomfieldchoiceset struct {
	netbox_goV1.UnimplementedExtrasCustomfieldchoicesetServer

	iDao dao.ExtrasCustomfieldchoicesetDao
}

// NewExtrasCustomfieldchoicesetServer create a new service
func NewExtrasCustomfieldchoicesetServer() netbox_goV1.ExtrasCustomfieldchoicesetServer {
	return &extrasCustomfieldchoiceset{
		iDao: dao.NewExtrasCustomfieldchoicesetDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasCustomfieldchoicesetCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasCustomfieldchoiceset
func (s *extrasCustomfieldchoiceset) Create(ctx context.Context, req *netbox_goV1.CreateExtrasCustomfieldchoicesetRequest) (*netbox_goV1.CreateExtrasCustomfieldchoicesetReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasCustomfieldchoiceset{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasCustomfieldchoiceset.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasCustomfieldchoiceset", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasCustomfieldchoicesetReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasCustomfieldchoiceset by id
func (s *extrasCustomfieldchoiceset) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasCustomfieldchoicesetByIDRequest) (*netbox_goV1.DeleteExtrasCustomfieldchoicesetByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasCustomfieldchoicesetByIDReply{}, nil
}

// UpdateByID update a extrasCustomfieldchoiceset by id
func (s *extrasCustomfieldchoiceset) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasCustomfieldchoicesetByIDRequest) (*netbox_goV1.UpdateExtrasCustomfieldchoicesetByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasCustomfieldchoiceset{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasCustomfieldchoiceset.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasCustomfieldchoiceset", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasCustomfieldchoicesetByIDReply{}, nil
}

// GetByID get a extrasCustomfieldchoiceset by id
func (s *extrasCustomfieldchoiceset) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasCustomfieldchoicesetByIDRequest) (*netbox_goV1.GetExtrasCustomfieldchoicesetByIDReply, error) {
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

	data, err := convertExtrasCustomfieldchoiceset(record)
	if err != nil {
		logger.Warn("convertExtrasCustomfieldchoiceset error", logger.Err(err), logger.Any("extrasCustomfieldchoiceset", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasCustomfieldchoiceset.Err()
	}

	return &netbox_goV1.GetExtrasCustomfieldchoicesetByIDReply{ExtrasCustomfieldchoiceset: data}, nil
}

// List get a paginated list of extrasCustomfieldchoicesets by custom conditions
func (s *extrasCustomfieldchoiceset) List(ctx context.Context, req *netbox_goV1.ListExtrasCustomfieldchoicesetRequest) (*netbox_goV1.ListExtrasCustomfieldchoicesetReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasCustomfieldchoiceset.Err()
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

	extrasCustomfieldchoicesets := []*netbox_goV1.ExtrasCustomfieldchoiceset{}
	for _, record := range records {
		data, err := convertExtrasCustomfieldchoiceset(record)
		if err != nil {
			logger.Warn("convertExtrasCustomfieldchoiceset error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasCustomfieldchoicesets = append(extrasCustomfieldchoicesets, data)
	}

	return &netbox_goV1.ListExtrasCustomfieldchoicesetReply{
		Total:                       total,
		ExtrasCustomfieldchoicesets: extrasCustomfieldchoicesets,
	}, nil
}

// DeleteByIDs batch delete extrasCustomfieldchoiceset by ids
func (s *extrasCustomfieldchoiceset) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasCustomfieldchoicesetByIDsRequest) (*netbox_goV1.DeleteExtrasCustomfieldchoicesetByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasCustomfieldchoicesetByIDsReply{}, nil
}

// GetByCondition get a extrasCustomfieldchoiceset by custom condition
func (s *extrasCustomfieldchoiceset) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasCustomfieldchoicesetByConditionRequest) (*netbox_goV1.GetExtrasCustomfieldchoicesetByConditionReply, error) {
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

	data, err := convertExtrasCustomfieldchoiceset(record)
	if err != nil {
		logger.Warn("convertExtrasCustomfieldchoiceset error", logger.Err(err), logger.Any("extrasCustomfieldchoiceset", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasCustomfieldchoiceset.Err()
	}

	return &netbox_goV1.GetExtrasCustomfieldchoicesetByConditionReply{
		ExtrasCustomfieldchoiceset: data,
	}, nil
}

// ListByIDs batch get extrasCustomfieldchoiceset by ids
func (s *extrasCustomfieldchoiceset) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasCustomfieldchoicesetByIDsRequest) (*netbox_goV1.ListExtrasCustomfieldchoicesetByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasCustomfieldchoicesetMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasCustomfieldchoicesets := []*netbox_goV1.ExtrasCustomfieldchoiceset{}
	for _, id := range req.Ids {
		if v, ok := extrasCustomfieldchoicesetMap[id]; ok {
			record, err := convertExtrasCustomfieldchoiceset(v)
			if err != nil {
				logger.Warn("convertExtrasCustomfieldchoiceset error", logger.Err(err), logger.Any("extrasCustomfieldchoiceset", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasCustomfieldchoicesets = append(extrasCustomfieldchoicesets, record)
		}
	}

	return &netbox_goV1.ListExtrasCustomfieldchoicesetByIDsReply{ExtrasCustomfieldchoicesets: extrasCustomfieldchoicesets}, nil
}

// ListByLastID get a paginated list of extrasCustomfieldchoicesets by last id
func (s *extrasCustomfieldchoiceset) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasCustomfieldchoicesetByLastIDRequest) (*netbox_goV1.ListExtrasCustomfieldchoicesetByLastIDReply, error) {
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

	extrasCustomfieldchoicesets := []*netbox_goV1.ExtrasCustomfieldchoiceset{}
	for _, record := range records {
		data, err := convertExtrasCustomfieldchoiceset(record)
		if err != nil {
			logger.Warn("convertExtrasCustomfieldchoiceset error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasCustomfieldchoicesets = append(extrasCustomfieldchoicesets, data)
	}

	return &netbox_goV1.ListExtrasCustomfieldchoicesetByLastIDReply{
		ExtrasCustomfieldchoicesets: extrasCustomfieldchoicesets,
	}, nil
}

func convertExtrasCustomfieldchoiceset(record *model.ExtrasCustomfieldchoiceset) (*netbox_goV1.ExtrasCustomfieldchoiceset, error) {
	value := &netbox_goV1.ExtrasCustomfieldchoiceset{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
