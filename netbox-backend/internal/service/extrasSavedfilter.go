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
		netbox_goV1.RegisterExtrasSavedfilterServer(server, NewExtrasSavedfilterServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasSavedfilterServer = (*extrasSavedfilter)(nil)
var _ time.Time

type extrasSavedfilter struct {
	netbox_goV1.UnimplementedExtrasSavedfilterServer

	iDao dao.ExtrasSavedfilterDao
}

// NewExtrasSavedfilterServer create a new service
func NewExtrasSavedfilterServer() netbox_goV1.ExtrasSavedfilterServer {
	return &extrasSavedfilter{
		iDao: dao.NewExtrasSavedfilterDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasSavedfilterCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasSavedfilter
func (s *extrasSavedfilter) Create(ctx context.Context, req *netbox_goV1.CreateExtrasSavedfilterRequest) (*netbox_goV1.CreateExtrasSavedfilterReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasSavedfilter{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasSavedfilter.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasSavedfilter", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasSavedfilterReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasSavedfilter by id
func (s *extrasSavedfilter) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasSavedfilterByIDRequest) (*netbox_goV1.DeleteExtrasSavedfilterByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasSavedfilterByIDReply{}, nil
}

// UpdateByID update a extrasSavedfilter by id
func (s *extrasSavedfilter) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasSavedfilterByIDRequest) (*netbox_goV1.UpdateExtrasSavedfilterByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasSavedfilter{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasSavedfilter.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasSavedfilter", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasSavedfilterByIDReply{}, nil
}

// GetByID get a extrasSavedfilter by id
func (s *extrasSavedfilter) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasSavedfilterByIDRequest) (*netbox_goV1.GetExtrasSavedfilterByIDReply, error) {
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

	data, err := convertExtrasSavedfilter(record)
	if err != nil {
		logger.Warn("convertExtrasSavedfilter error", logger.Err(err), logger.Any("extrasSavedfilter", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasSavedfilter.Err()
	}

	return &netbox_goV1.GetExtrasSavedfilterByIDReply{ExtrasSavedfilter: data}, nil
}

// List get a paginated list of extrasSavedfilters by custom conditions
func (s *extrasSavedfilter) List(ctx context.Context, req *netbox_goV1.ListExtrasSavedfilterRequest) (*netbox_goV1.ListExtrasSavedfilterReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasSavedfilter.Err()
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

	extrasSavedfilters := []*netbox_goV1.ExtrasSavedfilter{}
	for _, record := range records {
		data, err := convertExtrasSavedfilter(record)
		if err != nil {
			logger.Warn("convertExtrasSavedfilter error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasSavedfilters = append(extrasSavedfilters, data)
	}

	return &netbox_goV1.ListExtrasSavedfilterReply{
		Total:              total,
		ExtrasSavedfilters: extrasSavedfilters,
	}, nil
}

// DeleteByIDs batch delete extrasSavedfilter by ids
func (s *extrasSavedfilter) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasSavedfilterByIDsRequest) (*netbox_goV1.DeleteExtrasSavedfilterByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasSavedfilterByIDsReply{}, nil
}

// GetByCondition get a extrasSavedfilter by custom condition
func (s *extrasSavedfilter) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasSavedfilterByConditionRequest) (*netbox_goV1.GetExtrasSavedfilterByConditionReply, error) {
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

	data, err := convertExtrasSavedfilter(record)
	if err != nil {
		logger.Warn("convertExtrasSavedfilter error", logger.Err(err), logger.Any("extrasSavedfilter", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasSavedfilter.Err()
	}

	return &netbox_goV1.GetExtrasSavedfilterByConditionReply{
		ExtrasSavedfilter: data,
	}, nil
}

// ListByIDs batch get extrasSavedfilter by ids
func (s *extrasSavedfilter) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasSavedfilterByIDsRequest) (*netbox_goV1.ListExtrasSavedfilterByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasSavedfilterMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasSavedfilters := []*netbox_goV1.ExtrasSavedfilter{}
	for _, id := range req.Ids {
		if v, ok := extrasSavedfilterMap[id]; ok {
			record, err := convertExtrasSavedfilter(v)
			if err != nil {
				logger.Warn("convertExtrasSavedfilter error", logger.Err(err), logger.Any("extrasSavedfilter", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasSavedfilters = append(extrasSavedfilters, record)
		}
	}

	return &netbox_goV1.ListExtrasSavedfilterByIDsReply{ExtrasSavedfilters: extrasSavedfilters}, nil
}

// ListByLastID get a paginated list of extrasSavedfilters by last id
func (s *extrasSavedfilter) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasSavedfilterByLastIDRequest) (*netbox_goV1.ListExtrasSavedfilterByLastIDReply, error) {
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

	extrasSavedfilters := []*netbox_goV1.ExtrasSavedfilter{}
	for _, record := range records {
		data, err := convertExtrasSavedfilter(record)
		if err != nil {
			logger.Warn("convertExtrasSavedfilter error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasSavedfilters = append(extrasSavedfilters, data)
	}

	return &netbox_goV1.ListExtrasSavedfilterByLastIDReply{
		ExtrasSavedfilters: extrasSavedfilters,
	}, nil
}

func convertExtrasSavedfilter(record *model.ExtrasSavedfilter) (*netbox_goV1.ExtrasSavedfilter, error) {
	value := &netbox_goV1.ExtrasSavedfilter{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
