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
		netbox_goV1.RegisterExtrasCustomfieldServer(server, NewExtrasCustomfieldServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasCustomfieldServer = (*extrasCustomfield)(nil)
var _ time.Time

type extrasCustomfield struct {
	netbox_goV1.UnimplementedExtrasCustomfieldServer

	iDao dao.ExtrasCustomfieldDao
}

// NewExtrasCustomfieldServer create a new service
func NewExtrasCustomfieldServer() netbox_goV1.ExtrasCustomfieldServer {
	return &extrasCustomfield{
		iDao: dao.NewExtrasCustomfieldDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasCustomfieldCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasCustomfield
func (s *extrasCustomfield) Create(ctx context.Context, req *netbox_goV1.CreateExtrasCustomfieldRequest) (*netbox_goV1.CreateExtrasCustomfieldReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasCustomfield{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasCustomfield.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasCustomfield", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasCustomfieldReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasCustomfield by id
func (s *extrasCustomfield) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasCustomfieldByIDRequest) (*netbox_goV1.DeleteExtrasCustomfieldByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasCustomfieldByIDReply{}, nil
}

// UpdateByID update a extrasCustomfield by id
func (s *extrasCustomfield) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasCustomfieldByIDRequest) (*netbox_goV1.UpdateExtrasCustomfieldByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasCustomfield{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasCustomfield.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasCustomfield", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasCustomfieldByIDReply{}, nil
}

// GetByID get a extrasCustomfield by id
func (s *extrasCustomfield) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasCustomfieldByIDRequest) (*netbox_goV1.GetExtrasCustomfieldByIDReply, error) {
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

	data, err := convertExtrasCustomfield(record)
	if err != nil {
		logger.Warn("convertExtrasCustomfield error", logger.Err(err), logger.Any("extrasCustomfield", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasCustomfield.Err()
	}

	return &netbox_goV1.GetExtrasCustomfieldByIDReply{ExtrasCustomfield: data}, nil
}

// List get a paginated list of extrasCustomfields by custom conditions
func (s *extrasCustomfield) List(ctx context.Context, req *netbox_goV1.ListExtrasCustomfieldRequest) (*netbox_goV1.ListExtrasCustomfieldReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasCustomfield.Err()
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

	extrasCustomfields := []*netbox_goV1.ExtrasCustomfield{}
	for _, record := range records {
		data, err := convertExtrasCustomfield(record)
		if err != nil {
			logger.Warn("convertExtrasCustomfield error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasCustomfields = append(extrasCustomfields, data)
	}

	return &netbox_goV1.ListExtrasCustomfieldReply{
		Total:              total,
		ExtrasCustomfields: extrasCustomfields,
	}, nil
}

// DeleteByIDs batch delete extrasCustomfield by ids
func (s *extrasCustomfield) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasCustomfieldByIDsRequest) (*netbox_goV1.DeleteExtrasCustomfieldByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasCustomfieldByIDsReply{}, nil
}

// GetByCondition get a extrasCustomfield by custom condition
func (s *extrasCustomfield) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasCustomfieldByConditionRequest) (*netbox_goV1.GetExtrasCustomfieldByConditionReply, error) {
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

	data, err := convertExtrasCustomfield(record)
	if err != nil {
		logger.Warn("convertExtrasCustomfield error", logger.Err(err), logger.Any("extrasCustomfield", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasCustomfield.Err()
	}

	return &netbox_goV1.GetExtrasCustomfieldByConditionReply{
		ExtrasCustomfield: data,
	}, nil
}

// ListByIDs batch get extrasCustomfield by ids
func (s *extrasCustomfield) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasCustomfieldByIDsRequest) (*netbox_goV1.ListExtrasCustomfieldByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasCustomfieldMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasCustomfields := []*netbox_goV1.ExtrasCustomfield{}
	for _, id := range req.Ids {
		if v, ok := extrasCustomfieldMap[id]; ok {
			record, err := convertExtrasCustomfield(v)
			if err != nil {
				logger.Warn("convertExtrasCustomfield error", logger.Err(err), logger.Any("extrasCustomfield", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasCustomfields = append(extrasCustomfields, record)
		}
	}

	return &netbox_goV1.ListExtrasCustomfieldByIDsReply{ExtrasCustomfields: extrasCustomfields}, nil
}

// ListByLastID get a paginated list of extrasCustomfields by last id
func (s *extrasCustomfield) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasCustomfieldByLastIDRequest) (*netbox_goV1.ListExtrasCustomfieldByLastIDReply, error) {
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

	extrasCustomfields := []*netbox_goV1.ExtrasCustomfield{}
	for _, record := range records {
		data, err := convertExtrasCustomfield(record)
		if err != nil {
			logger.Warn("convertExtrasCustomfield error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasCustomfields = append(extrasCustomfields, data)
	}

	return &netbox_goV1.ListExtrasCustomfieldByLastIDReply{
		ExtrasCustomfields: extrasCustomfields,
	}, nil
}

func convertExtrasCustomfield(record *model.ExtrasCustomfield) (*netbox_goV1.ExtrasCustomfield, error) {
	value := &netbox_goV1.ExtrasCustomfield{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
