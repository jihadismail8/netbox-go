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
		netbox_goV1.RegisterDcimModulebaytemplateServer(server, NewDcimModulebaytemplateServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimModulebaytemplateServer = (*dcimModulebaytemplate)(nil)
var _ time.Time

type dcimModulebaytemplate struct {
	netbox_goV1.UnimplementedDcimModulebaytemplateServer

	iDao dao.DcimModulebaytemplateDao
}

// NewDcimModulebaytemplateServer create a new service
func NewDcimModulebaytemplateServer() netbox_goV1.DcimModulebaytemplateServer {
	return &dcimModulebaytemplate{
		iDao: dao.NewDcimModulebaytemplateDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimModulebaytemplateCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimModulebaytemplate
func (s *dcimModulebaytemplate) Create(ctx context.Context, req *netbox_goV1.CreateDcimModulebaytemplateRequest) (*netbox_goV1.CreateDcimModulebaytemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimModulebaytemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimModulebaytemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimModulebaytemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimModulebaytemplateReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimModulebaytemplate by id
func (s *dcimModulebaytemplate) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimModulebaytemplateByIDRequest) (*netbox_goV1.DeleteDcimModulebaytemplateByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimModulebaytemplateByIDReply{}, nil
}

// UpdateByID update a dcimModulebaytemplate by id
func (s *dcimModulebaytemplate) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimModulebaytemplateByIDRequest) (*netbox_goV1.UpdateDcimModulebaytemplateByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimModulebaytemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimModulebaytemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimModulebaytemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimModulebaytemplateByIDReply{}, nil
}

// GetByID get a dcimModulebaytemplate by id
func (s *dcimModulebaytemplate) GetByID(ctx context.Context, req *netbox_goV1.GetDcimModulebaytemplateByIDRequest) (*netbox_goV1.GetDcimModulebaytemplateByIDReply, error) {
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

	data, err := convertDcimModulebaytemplate(record)
	if err != nil {
		logger.Warn("convertDcimModulebaytemplate error", logger.Err(err), logger.Any("dcimModulebaytemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimModulebaytemplate.Err()
	}

	return &netbox_goV1.GetDcimModulebaytemplateByIDReply{DcimModulebaytemplate: data}, nil
}

// List get a paginated list of dcimModulebaytemplates by custom conditions
func (s *dcimModulebaytemplate) List(ctx context.Context, req *netbox_goV1.ListDcimModulebaytemplateRequest) (*netbox_goV1.ListDcimModulebaytemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimModulebaytemplate.Err()
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

	dcimModulebaytemplates := []*netbox_goV1.DcimModulebaytemplate{}
	for _, record := range records {
		data, err := convertDcimModulebaytemplate(record)
		if err != nil {
			logger.Warn("convertDcimModulebaytemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimModulebaytemplates = append(dcimModulebaytemplates, data)
	}

	return &netbox_goV1.ListDcimModulebaytemplateReply{
		Total:                  total,
		DcimModulebaytemplates: dcimModulebaytemplates,
	}, nil
}

// DeleteByIDs batch delete dcimModulebaytemplate by ids
func (s *dcimModulebaytemplate) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimModulebaytemplateByIDsRequest) (*netbox_goV1.DeleteDcimModulebaytemplateByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimModulebaytemplateByIDsReply{}, nil
}

// GetByCondition get a dcimModulebaytemplate by custom condition
func (s *dcimModulebaytemplate) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimModulebaytemplateByConditionRequest) (*netbox_goV1.GetDcimModulebaytemplateByConditionReply, error) {
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

	data, err := convertDcimModulebaytemplate(record)
	if err != nil {
		logger.Warn("convertDcimModulebaytemplate error", logger.Err(err), logger.Any("dcimModulebaytemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimModulebaytemplate.Err()
	}

	return &netbox_goV1.GetDcimModulebaytemplateByConditionReply{
		DcimModulebaytemplate: data,
	}, nil
}

// ListByIDs batch get dcimModulebaytemplate by ids
func (s *dcimModulebaytemplate) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimModulebaytemplateByIDsRequest) (*netbox_goV1.ListDcimModulebaytemplateByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimModulebaytemplateMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimModulebaytemplates := []*netbox_goV1.DcimModulebaytemplate{}
	for _, id := range req.Ids {
		if v, ok := dcimModulebaytemplateMap[id]; ok {
			record, err := convertDcimModulebaytemplate(v)
			if err != nil {
				logger.Warn("convertDcimModulebaytemplate error", logger.Err(err), logger.Any("dcimModulebaytemplate", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimModulebaytemplates = append(dcimModulebaytemplates, record)
		}
	}

	return &netbox_goV1.ListDcimModulebaytemplateByIDsReply{DcimModulebaytemplates: dcimModulebaytemplates}, nil
}

// ListByLastID get a paginated list of dcimModulebaytemplates by last id
func (s *dcimModulebaytemplate) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimModulebaytemplateByLastIDRequest) (*netbox_goV1.ListDcimModulebaytemplateByLastIDReply, error) {
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

	dcimModulebaytemplates := []*netbox_goV1.DcimModulebaytemplate{}
	for _, record := range records {
		data, err := convertDcimModulebaytemplate(record)
		if err != nil {
			logger.Warn("convertDcimModulebaytemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimModulebaytemplates = append(dcimModulebaytemplates, data)
	}

	return &netbox_goV1.ListDcimModulebaytemplateByLastIDReply{
		DcimModulebaytemplates: dcimModulebaytemplates,
	}, nil
}

func convertDcimModulebaytemplate(record *model.DcimModulebaytemplate) (*netbox_goV1.DcimModulebaytemplate, error) {
	value := &netbox_goV1.DcimModulebaytemplate{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
