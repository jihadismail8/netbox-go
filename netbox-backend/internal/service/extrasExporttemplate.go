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
		netbox_goV1.RegisterExtrasExporttemplateServer(server, NewExtrasExporttemplateServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasExporttemplateServer = (*extrasExporttemplate)(nil)
var _ time.Time

type extrasExporttemplate struct {
	netbox_goV1.UnimplementedExtrasExporttemplateServer

	iDao dao.ExtrasExporttemplateDao
}

// NewExtrasExporttemplateServer create a new service
func NewExtrasExporttemplateServer() netbox_goV1.ExtrasExporttemplateServer {
	return &extrasExporttemplate{
		iDao: dao.NewExtrasExporttemplateDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasExporttemplateCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasExporttemplate
func (s *extrasExporttemplate) Create(ctx context.Context, req *netbox_goV1.CreateExtrasExporttemplateRequest) (*netbox_goV1.CreateExtrasExporttemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasExporttemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasExporttemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasExporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasExporttemplateReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasExporttemplate by id
func (s *extrasExporttemplate) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasExporttemplateByIDRequest) (*netbox_goV1.DeleteExtrasExporttemplateByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasExporttemplateByIDReply{}, nil
}

// UpdateByID update a extrasExporttemplate by id
func (s *extrasExporttemplate) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasExporttemplateByIDRequest) (*netbox_goV1.UpdateExtrasExporttemplateByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasExporttemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasExporttemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasExporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasExporttemplateByIDReply{}, nil
}

// GetByID get a extrasExporttemplate by id
func (s *extrasExporttemplate) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasExporttemplateByIDRequest) (*netbox_goV1.GetExtrasExporttemplateByIDReply, error) {
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

	data, err := convertExtrasExporttemplate(record)
	if err != nil {
		logger.Warn("convertExtrasExporttemplate error", logger.Err(err), logger.Any("extrasExporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasExporttemplate.Err()
	}

	return &netbox_goV1.GetExtrasExporttemplateByIDReply{ExtrasExporttemplate: data}, nil
}

// List get a paginated list of extrasExporttemplates by custom conditions
func (s *extrasExporttemplate) List(ctx context.Context, req *netbox_goV1.ListExtrasExporttemplateRequest) (*netbox_goV1.ListExtrasExporttemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasExporttemplate.Err()
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

	extrasExporttemplates := []*netbox_goV1.ExtrasExporttemplate{}
	for _, record := range records {
		data, err := convertExtrasExporttemplate(record)
		if err != nil {
			logger.Warn("convertExtrasExporttemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasExporttemplates = append(extrasExporttemplates, data)
	}

	return &netbox_goV1.ListExtrasExporttemplateReply{
		Total:                 total,
		ExtrasExporttemplates: extrasExporttemplates,
	}, nil
}

// DeleteByIDs batch delete extrasExporttemplate by ids
func (s *extrasExporttemplate) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasExporttemplateByIDsRequest) (*netbox_goV1.DeleteExtrasExporttemplateByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasExporttemplateByIDsReply{}, nil
}

// GetByCondition get a extrasExporttemplate by custom condition
func (s *extrasExporttemplate) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasExporttemplateByConditionRequest) (*netbox_goV1.GetExtrasExporttemplateByConditionReply, error) {
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

	data, err := convertExtrasExporttemplate(record)
	if err != nil {
		logger.Warn("convertExtrasExporttemplate error", logger.Err(err), logger.Any("extrasExporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasExporttemplate.Err()
	}

	return &netbox_goV1.GetExtrasExporttemplateByConditionReply{
		ExtrasExporttemplate: data,
	}, nil
}

// ListByIDs batch get extrasExporttemplate by ids
func (s *extrasExporttemplate) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasExporttemplateByIDsRequest) (*netbox_goV1.ListExtrasExporttemplateByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasExporttemplateMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasExporttemplates := []*netbox_goV1.ExtrasExporttemplate{}
	for _, id := range req.Ids {
		if v, ok := extrasExporttemplateMap[id]; ok {
			record, err := convertExtrasExporttemplate(v)
			if err != nil {
				logger.Warn("convertExtrasExporttemplate error", logger.Err(err), logger.Any("extrasExporttemplate", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasExporttemplates = append(extrasExporttemplates, record)
		}
	}

	return &netbox_goV1.ListExtrasExporttemplateByIDsReply{ExtrasExporttemplates: extrasExporttemplates}, nil
}

// ListByLastID get a paginated list of extrasExporttemplates by last id
func (s *extrasExporttemplate) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasExporttemplateByLastIDRequest) (*netbox_goV1.ListExtrasExporttemplateByLastIDReply, error) {
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

	extrasExporttemplates := []*netbox_goV1.ExtrasExporttemplate{}
	for _, record := range records {
		data, err := convertExtrasExporttemplate(record)
		if err != nil {
			logger.Warn("convertExtrasExporttemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasExporttemplates = append(extrasExporttemplates, data)
	}

	return &netbox_goV1.ListExtrasExporttemplateByLastIDReply{
		ExtrasExporttemplates: extrasExporttemplates,
	}, nil
}

func convertExtrasExporttemplate(record *model.ExtrasExporttemplate) (*netbox_goV1.ExtrasExporttemplate, error) {
	value := &netbox_goV1.ExtrasExporttemplate{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
