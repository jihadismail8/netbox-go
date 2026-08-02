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
		netbox_goV1.RegisterExtrasCustomlinkServer(server, NewExtrasCustomlinkServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasCustomlinkServer = (*extrasCustomlink)(nil)
var _ time.Time

type extrasCustomlink struct {
	netbox_goV1.UnimplementedExtrasCustomlinkServer

	iDao dao.ExtrasCustomlinkDao
}

// NewExtrasCustomlinkServer create a new service
func NewExtrasCustomlinkServer() netbox_goV1.ExtrasCustomlinkServer {
	return &extrasCustomlink{
		iDao: dao.NewExtrasCustomlinkDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasCustomlinkCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasCustomlink
func (s *extrasCustomlink) Create(ctx context.Context, req *netbox_goV1.CreateExtrasCustomlinkRequest) (*netbox_goV1.CreateExtrasCustomlinkReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasCustomlink{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasCustomlink.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasCustomlink", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasCustomlinkReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasCustomlink by id
func (s *extrasCustomlink) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasCustomlinkByIDRequest) (*netbox_goV1.DeleteExtrasCustomlinkByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasCustomlinkByIDReply{}, nil
}

// UpdateByID update a extrasCustomlink by id
func (s *extrasCustomlink) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasCustomlinkByIDRequest) (*netbox_goV1.UpdateExtrasCustomlinkByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasCustomlink{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasCustomlink.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasCustomlink", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasCustomlinkByIDReply{}, nil
}

// GetByID get a extrasCustomlink by id
func (s *extrasCustomlink) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasCustomlinkByIDRequest) (*netbox_goV1.GetExtrasCustomlinkByIDReply, error) {
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

	data, err := convertExtrasCustomlink(record)
	if err != nil {
		logger.Warn("convertExtrasCustomlink error", logger.Err(err), logger.Any("extrasCustomlink", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasCustomlink.Err()
	}

	return &netbox_goV1.GetExtrasCustomlinkByIDReply{ExtrasCustomlink: data}, nil
}

// List get a paginated list of extrasCustomlinks by custom conditions
func (s *extrasCustomlink) List(ctx context.Context, req *netbox_goV1.ListExtrasCustomlinkRequest) (*netbox_goV1.ListExtrasCustomlinkReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasCustomlink.Err()
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

	extrasCustomlinks := []*netbox_goV1.ExtrasCustomlink{}
	for _, record := range records {
		data, err := convertExtrasCustomlink(record)
		if err != nil {
			logger.Warn("convertExtrasCustomlink error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasCustomlinks = append(extrasCustomlinks, data)
	}

	return &netbox_goV1.ListExtrasCustomlinkReply{
		Total:             total,
		ExtrasCustomlinks: extrasCustomlinks,
	}, nil
}

// DeleteByIDs batch delete extrasCustomlink by ids
func (s *extrasCustomlink) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasCustomlinkByIDsRequest) (*netbox_goV1.DeleteExtrasCustomlinkByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasCustomlinkByIDsReply{}, nil
}

// GetByCondition get a extrasCustomlink by custom condition
func (s *extrasCustomlink) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasCustomlinkByConditionRequest) (*netbox_goV1.GetExtrasCustomlinkByConditionReply, error) {
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

	data, err := convertExtrasCustomlink(record)
	if err != nil {
		logger.Warn("convertExtrasCustomlink error", logger.Err(err), logger.Any("extrasCustomlink", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasCustomlink.Err()
	}

	return &netbox_goV1.GetExtrasCustomlinkByConditionReply{
		ExtrasCustomlink: data,
	}, nil
}

// ListByIDs batch get extrasCustomlink by ids
func (s *extrasCustomlink) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasCustomlinkByIDsRequest) (*netbox_goV1.ListExtrasCustomlinkByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasCustomlinkMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasCustomlinks := []*netbox_goV1.ExtrasCustomlink{}
	for _, id := range req.Ids {
		if v, ok := extrasCustomlinkMap[id]; ok {
			record, err := convertExtrasCustomlink(v)
			if err != nil {
				logger.Warn("convertExtrasCustomlink error", logger.Err(err), logger.Any("extrasCustomlink", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasCustomlinks = append(extrasCustomlinks, record)
		}
	}

	return &netbox_goV1.ListExtrasCustomlinkByIDsReply{ExtrasCustomlinks: extrasCustomlinks}, nil
}

// ListByLastID get a paginated list of extrasCustomlinks by last id
func (s *extrasCustomlink) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasCustomlinkByLastIDRequest) (*netbox_goV1.ListExtrasCustomlinkByLastIDReply, error) {
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

	extrasCustomlinks := []*netbox_goV1.ExtrasCustomlink{}
	for _, record := range records {
		data, err := convertExtrasCustomlink(record)
		if err != nil {
			logger.Warn("convertExtrasCustomlink error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasCustomlinks = append(extrasCustomlinks, data)
	}

	return &netbox_goV1.ListExtrasCustomlinkByLastIDReply{
		ExtrasCustomlinks: extrasCustomlinks,
	}, nil
}

func convertExtrasCustomlink(record *model.ExtrasCustomlink) (*netbox_goV1.ExtrasCustomlink, error) {
	value := &netbox_goV1.ExtrasCustomlink{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
