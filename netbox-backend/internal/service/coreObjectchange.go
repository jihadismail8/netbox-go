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
		netbox_goV1.RegisterCoreObjectchangeServer(server, NewCoreObjectchangeServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CoreObjectchangeServer = (*coreObjectchange)(nil)
var _ time.Time

type coreObjectchange struct {
	netbox_goV1.UnimplementedCoreObjectchangeServer

	iDao dao.CoreObjectchangeDao
}

// NewCoreObjectchangeServer create a new service
func NewCoreObjectchangeServer() netbox_goV1.CoreObjectchangeServer {
	return &coreObjectchange{
		iDao: dao.NewCoreObjectchangeDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCoreObjectchangeCache(database.GetCacheType()),
		),
	}
}

// Create a new coreObjectchange
func (s *coreObjectchange) Create(ctx context.Context, req *netbox_goV1.CreateCoreObjectchangeRequest) (*netbox_goV1.CreateCoreObjectchangeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CoreObjectchange{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCoreObjectchange.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("coreObjectchange", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCoreObjectchangeReply{Id: record.ID}, nil
}

// DeleteByID delete a coreObjectchange by id
func (s *coreObjectchange) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCoreObjectchangeByIDRequest) (*netbox_goV1.DeleteCoreObjectchangeByIDReply, error) {
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

	return &netbox_goV1.DeleteCoreObjectchangeByIDReply{}, nil
}

// UpdateByID update a coreObjectchange by id
func (s *coreObjectchange) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCoreObjectchangeByIDRequest) (*netbox_goV1.UpdateCoreObjectchangeByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CoreObjectchange{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDCoreObjectchange.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("coreObjectchange", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCoreObjectchangeByIDReply{}, nil
}

// GetByID get a coreObjectchange by id
func (s *coreObjectchange) GetByID(ctx context.Context, req *netbox_goV1.GetCoreObjectchangeByIDRequest) (*netbox_goV1.GetCoreObjectchangeByIDReply, error) {
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

	data, err := convertCoreObjectchange(record)
	if err != nil {
		logger.Warn("convertCoreObjectchange error", logger.Err(err), logger.Any("coreObjectchange", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDCoreObjectchange.Err()
	}

	return &netbox_goV1.GetCoreObjectchangeByIDReply{CoreObjectchange: data}, nil
}

// List get a paginated list of coreObjectchanges by custom conditions
func (s *coreObjectchange) List(ctx context.Context, req *netbox_goV1.ListCoreObjectchangeRequest) (*netbox_goV1.ListCoreObjectchangeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCoreObjectchange.Err()
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

	coreObjectchanges := []*netbox_goV1.CoreObjectchange{}
	for _, record := range records {
		data, err := convertCoreObjectchange(record)
		if err != nil {
			logger.Warn("convertCoreObjectchange error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		coreObjectchanges = append(coreObjectchanges, data)
	}

	return &netbox_goV1.ListCoreObjectchangeReply{
		Total:             total,
		CoreObjectchanges: coreObjectchanges,
	}, nil
}

// DeleteByIDs batch delete coreObjectchange by ids
func (s *coreObjectchange) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCoreObjectchangeByIDsRequest) (*netbox_goV1.DeleteCoreObjectchangeByIDsReply, error) {
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

	return &netbox_goV1.DeleteCoreObjectchangeByIDsReply{}, nil
}

// GetByCondition get a coreObjectchange by custom condition
func (s *coreObjectchange) GetByCondition(ctx context.Context, req *netbox_goV1.GetCoreObjectchangeByConditionRequest) (*netbox_goV1.GetCoreObjectchangeByConditionReply, error) {
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

	data, err := convertCoreObjectchange(record)
	if err != nil {
		logger.Warn("convertCoreObjectchange error", logger.Err(err), logger.Any("coreObjectchange", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCoreObjectchange.Err()
	}

	return &netbox_goV1.GetCoreObjectchangeByConditionReply{
		CoreObjectchange: data,
	}, nil
}

// ListByIDs batch get coreObjectchange by ids
func (s *coreObjectchange) ListByIDs(ctx context.Context, req *netbox_goV1.ListCoreObjectchangeByIDsRequest) (*netbox_goV1.ListCoreObjectchangeByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	coreObjectchangeMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	coreObjectchanges := []*netbox_goV1.CoreObjectchange{}
	for _, id := range req.Ids {
		if v, ok := coreObjectchangeMap[id]; ok {
			record, err := convertCoreObjectchange(v)
			if err != nil {
				logger.Warn("convertCoreObjectchange error", logger.Err(err), logger.Any("coreObjectchange", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			coreObjectchanges = append(coreObjectchanges, record)
		}
	}

	return &netbox_goV1.ListCoreObjectchangeByIDsReply{CoreObjectchanges: coreObjectchanges}, nil
}

// ListByLastID get a paginated list of coreObjectchanges by last id
func (s *coreObjectchange) ListByLastID(ctx context.Context, req *netbox_goV1.ListCoreObjectchangeByLastIDRequest) (*netbox_goV1.ListCoreObjectchangeByLastIDReply, error) {
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

	coreObjectchanges := []*netbox_goV1.CoreObjectchange{}
	for _, record := range records {
		data, err := convertCoreObjectchange(record)
		if err != nil {
			logger.Warn("convertCoreObjectchange error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		coreObjectchanges = append(coreObjectchanges, data)
	}

	return &netbox_goV1.ListCoreObjectchangeByLastIDReply{
		CoreObjectchanges: coreObjectchanges,
	}, nil
}

func convertCoreObjectchange(record *model.CoreObjectchange) (*netbox_goV1.CoreObjectchange, error) {
	value := &netbox_goV1.CoreObjectchange{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
