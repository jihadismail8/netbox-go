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
		netbox_goV1.RegisterExtrasScriptServer(server, NewExtrasScriptServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasScriptServer = (*extrasScript)(nil)
var _ time.Time

type extrasScript struct {
	netbox_goV1.UnimplementedExtrasScriptServer

	iDao dao.ExtrasScriptDao
}

// NewExtrasScriptServer create a new service
func NewExtrasScriptServer() netbox_goV1.ExtrasScriptServer {
	return &extrasScript{
		iDao: dao.NewExtrasScriptDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasScriptCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasScript
func (s *extrasScript) Create(ctx context.Context, req *netbox_goV1.CreateExtrasScriptRequest) (*netbox_goV1.CreateExtrasScriptReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasScript{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasScript.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasScript", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasScriptReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasScript by id
func (s *extrasScript) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasScriptByIDRequest) (*netbox_goV1.DeleteExtrasScriptByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasScriptByIDReply{}, nil
}

// UpdateByID update a extrasScript by id
func (s *extrasScript) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasScriptByIDRequest) (*netbox_goV1.UpdateExtrasScriptByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasScript{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasScript.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasScript", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasScriptByIDReply{}, nil
}

// GetByID get a extrasScript by id
func (s *extrasScript) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasScriptByIDRequest) (*netbox_goV1.GetExtrasScriptByIDReply, error) {
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

	data, err := convertExtrasScript(record)
	if err != nil {
		logger.Warn("convertExtrasScript error", logger.Err(err), logger.Any("extrasScript", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasScript.Err()
	}

	return &netbox_goV1.GetExtrasScriptByIDReply{ExtrasScript: data}, nil
}

// List get a paginated list of extrasScripts by custom conditions
func (s *extrasScript) List(ctx context.Context, req *netbox_goV1.ListExtrasScriptRequest) (*netbox_goV1.ListExtrasScriptReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasScript.Err()
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

	extrasScripts := []*netbox_goV1.ExtrasScript{}
	for _, record := range records {
		data, err := convertExtrasScript(record)
		if err != nil {
			logger.Warn("convertExtrasScript error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasScripts = append(extrasScripts, data)
	}

	return &netbox_goV1.ListExtrasScriptReply{
		Total:         total,
		ExtrasScripts: extrasScripts,
	}, nil
}

// DeleteByIDs batch delete extrasScript by ids
func (s *extrasScript) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasScriptByIDsRequest) (*netbox_goV1.DeleteExtrasScriptByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasScriptByIDsReply{}, nil
}

// GetByCondition get a extrasScript by custom condition
func (s *extrasScript) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasScriptByConditionRequest) (*netbox_goV1.GetExtrasScriptByConditionReply, error) {
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

	data, err := convertExtrasScript(record)
	if err != nil {
		logger.Warn("convertExtrasScript error", logger.Err(err), logger.Any("extrasScript", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasScript.Err()
	}

	return &netbox_goV1.GetExtrasScriptByConditionReply{
		ExtrasScript: data,
	}, nil
}

// ListByIDs batch get extrasScript by ids
func (s *extrasScript) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasScriptByIDsRequest) (*netbox_goV1.ListExtrasScriptByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasScriptMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasScripts := []*netbox_goV1.ExtrasScript{}
	for _, id := range req.Ids {
		if v, ok := extrasScriptMap[id]; ok {
			record, err := convertExtrasScript(v)
			if err != nil {
				logger.Warn("convertExtrasScript error", logger.Err(err), logger.Any("extrasScript", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasScripts = append(extrasScripts, record)
		}
	}

	return &netbox_goV1.ListExtrasScriptByIDsReply{ExtrasScripts: extrasScripts}, nil
}

// ListByLastID get a paginated list of extrasScripts by last id
func (s *extrasScript) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasScriptByLastIDRequest) (*netbox_goV1.ListExtrasScriptByLastIDReply, error) {
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

	extrasScripts := []*netbox_goV1.ExtrasScript{}
	for _, record := range records {
		data, err := convertExtrasScript(record)
		if err != nil {
			logger.Warn("convertExtrasScript error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasScripts = append(extrasScripts, data)
	}

	return &netbox_goV1.ListExtrasScriptByLastIDReply{
		ExtrasScripts: extrasScripts,
	}, nil
}

func convertExtrasScript(record *model.ExtrasScript) (*netbox_goV1.ExtrasScript, error) {
	value := &netbox_goV1.ExtrasScript{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
