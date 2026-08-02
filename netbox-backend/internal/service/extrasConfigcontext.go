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
		netbox_goV1.RegisterExtrasConfigcontextServer(server, NewExtrasConfigcontextServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasConfigcontextServer = (*extrasConfigcontext)(nil)
var _ time.Time

type extrasConfigcontext struct {
	netbox_goV1.UnimplementedExtrasConfigcontextServer

	iDao dao.ExtrasConfigcontextDao
}

// NewExtrasConfigcontextServer create a new service
func NewExtrasConfigcontextServer() netbox_goV1.ExtrasConfigcontextServer {
	return &extrasConfigcontext{
		iDao: dao.NewExtrasConfigcontextDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasConfigcontextCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasConfigcontext
func (s *extrasConfigcontext) Create(ctx context.Context, req *netbox_goV1.CreateExtrasConfigcontextRequest) (*netbox_goV1.CreateExtrasConfigcontextReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasConfigcontext{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasConfigcontext.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasConfigcontext", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasConfigcontextReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasConfigcontext by id
func (s *extrasConfigcontext) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasConfigcontextByIDRequest) (*netbox_goV1.DeleteExtrasConfigcontextByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasConfigcontextByIDReply{}, nil
}

// UpdateByID update a extrasConfigcontext by id
func (s *extrasConfigcontext) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasConfigcontextByIDRequest) (*netbox_goV1.UpdateExtrasConfigcontextByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasConfigcontext{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasConfigcontext.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasConfigcontext", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasConfigcontextByIDReply{}, nil
}

// GetByID get a extrasConfigcontext by id
func (s *extrasConfigcontext) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasConfigcontextByIDRequest) (*netbox_goV1.GetExtrasConfigcontextByIDReply, error) {
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

	data, err := convertExtrasConfigcontext(record)
	if err != nil {
		logger.Warn("convertExtrasConfigcontext error", logger.Err(err), logger.Any("extrasConfigcontext", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasConfigcontext.Err()
	}

	return &netbox_goV1.GetExtrasConfigcontextByIDReply{ExtrasConfigcontext: data}, nil
}

// List get a paginated list of extrasConfigcontexts by custom conditions
func (s *extrasConfigcontext) List(ctx context.Context, req *netbox_goV1.ListExtrasConfigcontextRequest) (*netbox_goV1.ListExtrasConfigcontextReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasConfigcontext.Err()
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

	extrasConfigcontexts := []*netbox_goV1.ExtrasConfigcontext{}
	for _, record := range records {
		data, err := convertExtrasConfigcontext(record)
		if err != nil {
			logger.Warn("convertExtrasConfigcontext error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasConfigcontexts = append(extrasConfigcontexts, data)
	}

	return &netbox_goV1.ListExtrasConfigcontextReply{
		Total:                total,
		ExtrasConfigcontexts: extrasConfigcontexts,
	}, nil
}

// DeleteByIDs batch delete extrasConfigcontext by ids
func (s *extrasConfigcontext) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasConfigcontextByIDsRequest) (*netbox_goV1.DeleteExtrasConfigcontextByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasConfigcontextByIDsReply{}, nil
}

// GetByCondition get a extrasConfigcontext by custom condition
func (s *extrasConfigcontext) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasConfigcontextByConditionRequest) (*netbox_goV1.GetExtrasConfigcontextByConditionReply, error) {
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

	data, err := convertExtrasConfigcontext(record)
	if err != nil {
		logger.Warn("convertExtrasConfigcontext error", logger.Err(err), logger.Any("extrasConfigcontext", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasConfigcontext.Err()
	}

	return &netbox_goV1.GetExtrasConfigcontextByConditionReply{
		ExtrasConfigcontext: data,
	}, nil
}

// ListByIDs batch get extrasConfigcontext by ids
func (s *extrasConfigcontext) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasConfigcontextByIDsRequest) (*netbox_goV1.ListExtrasConfigcontextByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasConfigcontextMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasConfigcontexts := []*netbox_goV1.ExtrasConfigcontext{}
	for _, id := range req.Ids {
		if v, ok := extrasConfigcontextMap[id]; ok {
			record, err := convertExtrasConfigcontext(v)
			if err != nil {
				logger.Warn("convertExtrasConfigcontext error", logger.Err(err), logger.Any("extrasConfigcontext", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasConfigcontexts = append(extrasConfigcontexts, record)
		}
	}

	return &netbox_goV1.ListExtrasConfigcontextByIDsReply{ExtrasConfigcontexts: extrasConfigcontexts}, nil
}

// ListByLastID get a paginated list of extrasConfigcontexts by last id
func (s *extrasConfigcontext) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasConfigcontextByLastIDRequest) (*netbox_goV1.ListExtrasConfigcontextByLastIDReply, error) {
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

	extrasConfigcontexts := []*netbox_goV1.ExtrasConfigcontext{}
	for _, record := range records {
		data, err := convertExtrasConfigcontext(record)
		if err != nil {
			logger.Warn("convertExtrasConfigcontext error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasConfigcontexts = append(extrasConfigcontexts, data)
	}

	return &netbox_goV1.ListExtrasConfigcontextByLastIDReply{
		ExtrasConfigcontexts: extrasConfigcontexts,
	}, nil
}

func convertExtrasConfigcontext(record *model.ExtrasConfigcontext) (*netbox_goV1.ExtrasConfigcontext, error) {
	value := &netbox_goV1.ExtrasConfigcontext{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
