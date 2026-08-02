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
		netbox_goV1.RegisterDcimModuleServer(server, NewDcimModuleServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimModuleServer = (*dcimModule)(nil)
var _ time.Time

type dcimModule struct {
	netbox_goV1.UnimplementedDcimModuleServer

	iDao dao.DcimModuleDao
}

// NewDcimModuleServer create a new service
func NewDcimModuleServer() netbox_goV1.DcimModuleServer {
	return &dcimModule{
		iDao: dao.NewDcimModuleDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimModuleCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimModule
func (s *dcimModule) Create(ctx context.Context, req *netbox_goV1.CreateDcimModuleRequest) (*netbox_goV1.CreateDcimModuleReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimModule{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimModule.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimModule", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimModuleReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimModule by id
func (s *dcimModule) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimModuleByIDRequest) (*netbox_goV1.DeleteDcimModuleByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimModuleByIDReply{}, nil
}

// UpdateByID update a dcimModule by id
func (s *dcimModule) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimModuleByIDRequest) (*netbox_goV1.UpdateDcimModuleByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimModule{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimModule.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimModule", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimModuleByIDReply{}, nil
}

// GetByID get a dcimModule by id
func (s *dcimModule) GetByID(ctx context.Context, req *netbox_goV1.GetDcimModuleByIDRequest) (*netbox_goV1.GetDcimModuleByIDReply, error) {
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

	data, err := convertDcimModule(record)
	if err != nil {
		logger.Warn("convertDcimModule error", logger.Err(err), logger.Any("dcimModule", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimModule.Err()
	}

	return &netbox_goV1.GetDcimModuleByIDReply{DcimModule: data}, nil
}

// List get a paginated list of dcimModules by custom conditions
func (s *dcimModule) List(ctx context.Context, req *netbox_goV1.ListDcimModuleRequest) (*netbox_goV1.ListDcimModuleReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimModule.Err()
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

	dcimModules := []*netbox_goV1.DcimModule{}
	for _, record := range records {
		data, err := convertDcimModule(record)
		if err != nil {
			logger.Warn("convertDcimModule error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimModules = append(dcimModules, data)
	}

	return &netbox_goV1.ListDcimModuleReply{
		Total:       total,
		DcimModules: dcimModules,
	}, nil
}

// DeleteByIDs batch delete dcimModule by ids
func (s *dcimModule) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimModuleByIDsRequest) (*netbox_goV1.DeleteDcimModuleByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimModuleByIDsReply{}, nil
}

// GetByCondition get a dcimModule by custom condition
func (s *dcimModule) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimModuleByConditionRequest) (*netbox_goV1.GetDcimModuleByConditionReply, error) {
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

	data, err := convertDcimModule(record)
	if err != nil {
		logger.Warn("convertDcimModule error", logger.Err(err), logger.Any("dcimModule", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimModule.Err()
	}

	return &netbox_goV1.GetDcimModuleByConditionReply{
		DcimModule: data,
	}, nil
}

// ListByIDs batch get dcimModule by ids
func (s *dcimModule) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimModuleByIDsRequest) (*netbox_goV1.ListDcimModuleByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimModuleMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimModules := []*netbox_goV1.DcimModule{}
	for _, id := range req.Ids {
		if v, ok := dcimModuleMap[id]; ok {
			record, err := convertDcimModule(v)
			if err != nil {
				logger.Warn("convertDcimModule error", logger.Err(err), logger.Any("dcimModule", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimModules = append(dcimModules, record)
		}
	}

	return &netbox_goV1.ListDcimModuleByIDsReply{DcimModules: dcimModules}, nil
}

// ListByLastID get a paginated list of dcimModules by last id
func (s *dcimModule) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimModuleByLastIDRequest) (*netbox_goV1.ListDcimModuleByLastIDReply, error) {
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

	dcimModules := []*netbox_goV1.DcimModule{}
	for _, record := range records {
		data, err := convertDcimModule(record)
		if err != nil {
			logger.Warn("convertDcimModule error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimModules = append(dcimModules, data)
	}

	return &netbox_goV1.ListDcimModuleByLastIDReply{
		DcimModules: dcimModules,
	}, nil
}

func convertDcimModule(record *model.DcimModule) (*netbox_goV1.DcimModule, error) {
	value := &netbox_goV1.DcimModule{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
