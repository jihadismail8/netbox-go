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
		netbox_goV1.RegisterDcimPlatformServer(server, NewDcimPlatformServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimPlatformServer = (*dcimPlatform)(nil)
var _ time.Time

type dcimPlatform struct {
	netbox_goV1.UnimplementedDcimPlatformServer

	iDao dao.DcimPlatformDao
}

// NewDcimPlatformServer create a new service
func NewDcimPlatformServer() netbox_goV1.DcimPlatformServer {
	return &dcimPlatform{
		iDao: dao.NewDcimPlatformDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimPlatformCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimPlatform
func (s *dcimPlatform) Create(ctx context.Context, req *netbox_goV1.CreateDcimPlatformRequest) (*netbox_goV1.CreateDcimPlatformReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimPlatform{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimPlatform.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimPlatform", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimPlatformReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimPlatform by id
func (s *dcimPlatform) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimPlatformByIDRequest) (*netbox_goV1.DeleteDcimPlatformByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimPlatformByIDReply{}, nil
}

// UpdateByID update a dcimPlatform by id
func (s *dcimPlatform) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimPlatformByIDRequest) (*netbox_goV1.UpdateDcimPlatformByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimPlatform{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimPlatform.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimPlatform", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimPlatformByIDReply{}, nil
}

// GetByID get a dcimPlatform by id
func (s *dcimPlatform) GetByID(ctx context.Context, req *netbox_goV1.GetDcimPlatformByIDRequest) (*netbox_goV1.GetDcimPlatformByIDReply, error) {
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

	data, err := convertDcimPlatform(record)
	if err != nil {
		logger.Warn("convertDcimPlatform error", logger.Err(err), logger.Any("dcimPlatform", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimPlatform.Err()
	}

	return &netbox_goV1.GetDcimPlatformByIDReply{DcimPlatform: data}, nil
}

// List get a paginated list of dcimPlatforms by custom conditions
func (s *dcimPlatform) List(ctx context.Context, req *netbox_goV1.ListDcimPlatformRequest) (*netbox_goV1.ListDcimPlatformReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimPlatform.Err()
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

	dcimPlatforms := []*netbox_goV1.DcimPlatform{}
	for _, record := range records {
		data, err := convertDcimPlatform(record)
		if err != nil {
			logger.Warn("convertDcimPlatform error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimPlatforms = append(dcimPlatforms, data)
	}

	return &netbox_goV1.ListDcimPlatformReply{
		Total:         total,
		DcimPlatforms: dcimPlatforms,
	}, nil
}

// DeleteByIDs batch delete dcimPlatform by ids
func (s *dcimPlatform) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimPlatformByIDsRequest) (*netbox_goV1.DeleteDcimPlatformByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimPlatformByIDsReply{}, nil
}

// GetByCondition get a dcimPlatform by custom condition
func (s *dcimPlatform) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimPlatformByConditionRequest) (*netbox_goV1.GetDcimPlatformByConditionReply, error) {
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

	data, err := convertDcimPlatform(record)
	if err != nil {
		logger.Warn("convertDcimPlatform error", logger.Err(err), logger.Any("dcimPlatform", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimPlatform.Err()
	}

	return &netbox_goV1.GetDcimPlatformByConditionReply{
		DcimPlatform: data,
	}, nil
}

// ListByIDs batch get dcimPlatform by ids
func (s *dcimPlatform) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimPlatformByIDsRequest) (*netbox_goV1.ListDcimPlatformByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimPlatformMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimPlatforms := []*netbox_goV1.DcimPlatform{}
	for _, id := range req.Ids {
		if v, ok := dcimPlatformMap[id]; ok {
			record, err := convertDcimPlatform(v)
			if err != nil {
				logger.Warn("convertDcimPlatform error", logger.Err(err), logger.Any("dcimPlatform", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimPlatforms = append(dcimPlatforms, record)
		}
	}

	return &netbox_goV1.ListDcimPlatformByIDsReply{DcimPlatforms: dcimPlatforms}, nil
}

// ListByLastID get a paginated list of dcimPlatforms by last id
func (s *dcimPlatform) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimPlatformByLastIDRequest) (*netbox_goV1.ListDcimPlatformByLastIDReply, error) {
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

	dcimPlatforms := []*netbox_goV1.DcimPlatform{}
	for _, record := range records {
		data, err := convertDcimPlatform(record)
		if err != nil {
			logger.Warn("convertDcimPlatform error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimPlatforms = append(dcimPlatforms, data)
	}

	return &netbox_goV1.ListDcimPlatformByLastIDReply{
		DcimPlatforms: dcimPlatforms,
	}, nil
}

func convertDcimPlatform(record *model.DcimPlatform) (*netbox_goV1.DcimPlatform, error) {
	value := &netbox_goV1.DcimPlatform{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
