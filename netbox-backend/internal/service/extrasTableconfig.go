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
		netbox_goV1.RegisterExtrasTableconfigServer(server, NewExtrasTableconfigServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasTableconfigServer = (*extrasTableconfig)(nil)
var _ time.Time

type extrasTableconfig struct {
	netbox_goV1.UnimplementedExtrasTableconfigServer

	iDao dao.ExtrasTableconfigDao
}

// NewExtrasTableconfigServer create a new service
func NewExtrasTableconfigServer() netbox_goV1.ExtrasTableconfigServer {
	return &extrasTableconfig{
		iDao: dao.NewExtrasTableconfigDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasTableconfigCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasTableconfig
func (s *extrasTableconfig) Create(ctx context.Context, req *netbox_goV1.CreateExtrasTableconfigRequest) (*netbox_goV1.CreateExtrasTableconfigReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasTableconfig{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasTableconfig.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasTableconfig", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasTableconfigReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasTableconfig by id
func (s *extrasTableconfig) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasTableconfigByIDRequest) (*netbox_goV1.DeleteExtrasTableconfigByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasTableconfigByIDReply{}, nil
}

// UpdateByID update a extrasTableconfig by id
func (s *extrasTableconfig) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasTableconfigByIDRequest) (*netbox_goV1.UpdateExtrasTableconfigByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasTableconfig{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasTableconfig.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasTableconfig", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasTableconfigByIDReply{}, nil
}

// GetByID get a extrasTableconfig by id
func (s *extrasTableconfig) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasTableconfigByIDRequest) (*netbox_goV1.GetExtrasTableconfigByIDReply, error) {
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

	data, err := convertExtrasTableconfig(record)
	if err != nil {
		logger.Warn("convertExtrasTableconfig error", logger.Err(err), logger.Any("extrasTableconfig", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasTableconfig.Err()
	}

	return &netbox_goV1.GetExtrasTableconfigByIDReply{ExtrasTableconfig: data}, nil
}

// List get a paginated list of extrasTableconfigs by custom conditions
func (s *extrasTableconfig) List(ctx context.Context, req *netbox_goV1.ListExtrasTableconfigRequest) (*netbox_goV1.ListExtrasTableconfigReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasTableconfig.Err()
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

	extrasTableconfigs := []*netbox_goV1.ExtrasTableconfig{}
	for _, record := range records {
		data, err := convertExtrasTableconfig(record)
		if err != nil {
			logger.Warn("convertExtrasTableconfig error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasTableconfigs = append(extrasTableconfigs, data)
	}

	return &netbox_goV1.ListExtrasTableconfigReply{
		Total:              total,
		ExtrasTableconfigs: extrasTableconfigs,
	}, nil
}

// DeleteByIDs batch delete extrasTableconfig by ids
func (s *extrasTableconfig) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasTableconfigByIDsRequest) (*netbox_goV1.DeleteExtrasTableconfigByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasTableconfigByIDsReply{}, nil
}

// GetByCondition get a extrasTableconfig by custom condition
func (s *extrasTableconfig) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasTableconfigByConditionRequest) (*netbox_goV1.GetExtrasTableconfigByConditionReply, error) {
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

	data, err := convertExtrasTableconfig(record)
	if err != nil {
		logger.Warn("convertExtrasTableconfig error", logger.Err(err), logger.Any("extrasTableconfig", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasTableconfig.Err()
	}

	return &netbox_goV1.GetExtrasTableconfigByConditionReply{
		ExtrasTableconfig: data,
	}, nil
}

// ListByIDs batch get extrasTableconfig by ids
func (s *extrasTableconfig) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasTableconfigByIDsRequest) (*netbox_goV1.ListExtrasTableconfigByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasTableconfigMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasTableconfigs := []*netbox_goV1.ExtrasTableconfig{}
	for _, id := range req.Ids {
		if v, ok := extrasTableconfigMap[id]; ok {
			record, err := convertExtrasTableconfig(v)
			if err != nil {
				logger.Warn("convertExtrasTableconfig error", logger.Err(err), logger.Any("extrasTableconfig", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasTableconfigs = append(extrasTableconfigs, record)
		}
	}

	return &netbox_goV1.ListExtrasTableconfigByIDsReply{ExtrasTableconfigs: extrasTableconfigs}, nil
}

// ListByLastID get a paginated list of extrasTableconfigs by last id
func (s *extrasTableconfig) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasTableconfigByLastIDRequest) (*netbox_goV1.ListExtrasTableconfigByLastIDReply, error) {
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

	extrasTableconfigs := []*netbox_goV1.ExtrasTableconfig{}
	for _, record := range records {
		data, err := convertExtrasTableconfig(record)
		if err != nil {
			logger.Warn("convertExtrasTableconfig error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasTableconfigs = append(extrasTableconfigs, data)
	}

	return &netbox_goV1.ListExtrasTableconfigByLastIDReply{
		ExtrasTableconfigs: extrasTableconfigs,
	}, nil
}

func convertExtrasTableconfig(record *model.ExtrasTableconfig) (*netbox_goV1.ExtrasTableconfig, error) {
	value := &netbox_goV1.ExtrasTableconfig{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
