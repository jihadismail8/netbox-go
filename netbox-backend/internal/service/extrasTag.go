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
		netbox_goV1.RegisterExtrasTagServer(server, NewExtrasTagServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasTagServer = (*extrasTag)(nil)
var _ time.Time

type extrasTag struct {
	netbox_goV1.UnimplementedExtrasTagServer

	iDao dao.ExtrasTagDao
}

// NewExtrasTagServer create a new service
func NewExtrasTagServer() netbox_goV1.ExtrasTagServer {
	return &extrasTag{
		iDao: dao.NewExtrasTagDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasTagCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasTag
func (s *extrasTag) Create(ctx context.Context, req *netbox_goV1.CreateExtrasTagRequest) (*netbox_goV1.CreateExtrasTagReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasTag{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasTag.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasTag", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasTagReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasTag by id
func (s *extrasTag) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasTagByIDRequest) (*netbox_goV1.DeleteExtrasTagByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasTagByIDReply{}, nil
}

// UpdateByID update a extrasTag by id
func (s *extrasTag) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasTagByIDRequest) (*netbox_goV1.UpdateExtrasTagByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasTag{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasTag.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasTag", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasTagByIDReply{}, nil
}

// GetByID get a extrasTag by id
func (s *extrasTag) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasTagByIDRequest) (*netbox_goV1.GetExtrasTagByIDReply, error) {
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

	data, err := convertExtrasTag(record)
	if err != nil {
		logger.Warn("convertExtrasTag error", logger.Err(err), logger.Any("extrasTag", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasTag.Err()
	}

	return &netbox_goV1.GetExtrasTagByIDReply{ExtrasTag: data}, nil
}

// List get a paginated list of extrasTags by custom conditions
func (s *extrasTag) List(ctx context.Context, req *netbox_goV1.ListExtrasTagRequest) (*netbox_goV1.ListExtrasTagReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasTag.Err()
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

	extrasTags := []*netbox_goV1.ExtrasTag{}
	for _, record := range records {
		data, err := convertExtrasTag(record)
		if err != nil {
			logger.Warn("convertExtrasTag error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasTags = append(extrasTags, data)
	}

	return &netbox_goV1.ListExtrasTagReply{
		Total:      total,
		ExtrasTags: extrasTags,
	}, nil
}

// DeleteByIDs batch delete extrasTag by ids
func (s *extrasTag) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasTagByIDsRequest) (*netbox_goV1.DeleteExtrasTagByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasTagByIDsReply{}, nil
}

// GetByCondition get a extrasTag by custom condition
func (s *extrasTag) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasTagByConditionRequest) (*netbox_goV1.GetExtrasTagByConditionReply, error) {
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

	data, err := convertExtrasTag(record)
	if err != nil {
		logger.Warn("convertExtrasTag error", logger.Err(err), logger.Any("extrasTag", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasTag.Err()
	}

	return &netbox_goV1.GetExtrasTagByConditionReply{
		ExtrasTag: data,
	}, nil
}

// ListByIDs batch get extrasTag by ids
func (s *extrasTag) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasTagByIDsRequest) (*netbox_goV1.ListExtrasTagByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasTagMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasTags := []*netbox_goV1.ExtrasTag{}
	for _, id := range req.Ids {
		if v, ok := extrasTagMap[id]; ok {
			record, err := convertExtrasTag(v)
			if err != nil {
				logger.Warn("convertExtrasTag error", logger.Err(err), logger.Any("extrasTag", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasTags = append(extrasTags, record)
		}
	}

	return &netbox_goV1.ListExtrasTagByIDsReply{ExtrasTags: extrasTags}, nil
}

// ListByLastID get a paginated list of extrasTags by last id
func (s *extrasTag) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasTagByLastIDRequest) (*netbox_goV1.ListExtrasTagByLastIDReply, error) {
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

	extrasTags := []*netbox_goV1.ExtrasTag{}
	for _, record := range records {
		data, err := convertExtrasTag(record)
		if err != nil {
			logger.Warn("convertExtrasTag error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasTags = append(extrasTags, data)
	}

	return &netbox_goV1.ListExtrasTagByLastIDReply{
		ExtrasTags: extrasTags,
	}, nil
}

func convertExtrasTag(record *model.ExtrasTag) (*netbox_goV1.ExtrasTag, error) {
	value := &netbox_goV1.ExtrasTag{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
