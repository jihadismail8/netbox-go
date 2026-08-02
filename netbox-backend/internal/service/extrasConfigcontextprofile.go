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
		netbox_goV1.RegisterExtrasConfigcontextprofileServer(server, NewExtrasConfigcontextprofileServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasConfigcontextprofileServer = (*extrasConfigcontextprofile)(nil)
var _ time.Time

type extrasConfigcontextprofile struct {
	netbox_goV1.UnimplementedExtrasConfigcontextprofileServer

	iDao dao.ExtrasConfigcontextprofileDao
}

// NewExtrasConfigcontextprofileServer create a new service
func NewExtrasConfigcontextprofileServer() netbox_goV1.ExtrasConfigcontextprofileServer {
	return &extrasConfigcontextprofile{
		iDao: dao.NewExtrasConfigcontextprofileDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasConfigcontextprofileCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasConfigcontextprofile
func (s *extrasConfigcontextprofile) Create(ctx context.Context, req *netbox_goV1.CreateExtrasConfigcontextprofileRequest) (*netbox_goV1.CreateExtrasConfigcontextprofileReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasConfigcontextprofile{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasConfigcontextprofile.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasConfigcontextprofile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasConfigcontextprofileReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasConfigcontextprofile by id
func (s *extrasConfigcontextprofile) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasConfigcontextprofileByIDRequest) (*netbox_goV1.DeleteExtrasConfigcontextprofileByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasConfigcontextprofileByIDReply{}, nil
}

// UpdateByID update a extrasConfigcontextprofile by id
func (s *extrasConfigcontextprofile) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasConfigcontextprofileByIDRequest) (*netbox_goV1.UpdateExtrasConfigcontextprofileByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasConfigcontextprofile{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasConfigcontextprofile.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasConfigcontextprofile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasConfigcontextprofileByIDReply{}, nil
}

// GetByID get a extrasConfigcontextprofile by id
func (s *extrasConfigcontextprofile) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasConfigcontextprofileByIDRequest) (*netbox_goV1.GetExtrasConfigcontextprofileByIDReply, error) {
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

	data, err := convertExtrasConfigcontextprofile(record)
	if err != nil {
		logger.Warn("convertExtrasConfigcontextprofile error", logger.Err(err), logger.Any("extrasConfigcontextprofile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasConfigcontextprofile.Err()
	}

	return &netbox_goV1.GetExtrasConfigcontextprofileByIDReply{ExtrasConfigcontextprofile: data}, nil
}

// List get a paginated list of extrasConfigcontextprofiles by custom conditions
func (s *extrasConfigcontextprofile) List(ctx context.Context, req *netbox_goV1.ListExtrasConfigcontextprofileRequest) (*netbox_goV1.ListExtrasConfigcontextprofileReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasConfigcontextprofile.Err()
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

	extrasConfigcontextprofiles := []*netbox_goV1.ExtrasConfigcontextprofile{}
	for _, record := range records {
		data, err := convertExtrasConfigcontextprofile(record)
		if err != nil {
			logger.Warn("convertExtrasConfigcontextprofile error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasConfigcontextprofiles = append(extrasConfigcontextprofiles, data)
	}

	return &netbox_goV1.ListExtrasConfigcontextprofileReply{
		Total:                       total,
		ExtrasConfigcontextprofiles: extrasConfigcontextprofiles,
	}, nil
}

// DeleteByIDs batch delete extrasConfigcontextprofile by ids
func (s *extrasConfigcontextprofile) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasConfigcontextprofileByIDsRequest) (*netbox_goV1.DeleteExtrasConfigcontextprofileByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasConfigcontextprofileByIDsReply{}, nil
}

// GetByCondition get a extrasConfigcontextprofile by custom condition
func (s *extrasConfigcontextprofile) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasConfigcontextprofileByConditionRequest) (*netbox_goV1.GetExtrasConfigcontextprofileByConditionReply, error) {
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

	data, err := convertExtrasConfigcontextprofile(record)
	if err != nil {
		logger.Warn("convertExtrasConfigcontextprofile error", logger.Err(err), logger.Any("extrasConfigcontextprofile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasConfigcontextprofile.Err()
	}

	return &netbox_goV1.GetExtrasConfigcontextprofileByConditionReply{
		ExtrasConfigcontextprofile: data,
	}, nil
}

// ListByIDs batch get extrasConfigcontextprofile by ids
func (s *extrasConfigcontextprofile) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasConfigcontextprofileByIDsRequest) (*netbox_goV1.ListExtrasConfigcontextprofileByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasConfigcontextprofileMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasConfigcontextprofiles := []*netbox_goV1.ExtrasConfigcontextprofile{}
	for _, id := range req.Ids {
		if v, ok := extrasConfigcontextprofileMap[id]; ok {
			record, err := convertExtrasConfigcontextprofile(v)
			if err != nil {
				logger.Warn("convertExtrasConfigcontextprofile error", logger.Err(err), logger.Any("extrasConfigcontextprofile", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasConfigcontextprofiles = append(extrasConfigcontextprofiles, record)
		}
	}

	return &netbox_goV1.ListExtrasConfigcontextprofileByIDsReply{ExtrasConfigcontextprofiles: extrasConfigcontextprofiles}, nil
}

// ListByLastID get a paginated list of extrasConfigcontextprofiles by last id
func (s *extrasConfigcontextprofile) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasConfigcontextprofileByLastIDRequest) (*netbox_goV1.ListExtrasConfigcontextprofileByLastIDReply, error) {
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

	extrasConfigcontextprofiles := []*netbox_goV1.ExtrasConfigcontextprofile{}
	for _, record := range records {
		data, err := convertExtrasConfigcontextprofile(record)
		if err != nil {
			logger.Warn("convertExtrasConfigcontextprofile error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasConfigcontextprofiles = append(extrasConfigcontextprofiles, data)
	}

	return &netbox_goV1.ListExtrasConfigcontextprofileByLastIDReply{
		ExtrasConfigcontextprofiles: extrasConfigcontextprofiles,
	}, nil
}

func convertExtrasConfigcontextprofile(record *model.ExtrasConfigcontextprofile) (*netbox_goV1.ExtrasConfigcontextprofile, error) {
	value := &netbox_goV1.ExtrasConfigcontextprofile{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
