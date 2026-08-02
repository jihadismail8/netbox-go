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
		netbox_goV1.RegisterExtrasTaggeditemServer(server, NewExtrasTaggeditemServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasTaggeditemServer = (*extrasTaggeditem)(nil)
var _ time.Time

type extrasTaggeditem struct {
	netbox_goV1.UnimplementedExtrasTaggeditemServer

	iDao dao.ExtrasTaggeditemDao
}

// NewExtrasTaggeditemServer create a new service
func NewExtrasTaggeditemServer() netbox_goV1.ExtrasTaggeditemServer {
	return &extrasTaggeditem{
		iDao: dao.NewExtrasTaggeditemDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasTaggeditemCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasTaggeditem
func (s *extrasTaggeditem) Create(ctx context.Context, req *netbox_goV1.CreateExtrasTaggeditemRequest) (*netbox_goV1.CreateExtrasTaggeditemReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasTaggeditem{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasTaggeditem.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasTaggeditem", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasTaggeditemReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasTaggeditem by id
func (s *extrasTaggeditem) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasTaggeditemByIDRequest) (*netbox_goV1.DeleteExtrasTaggeditemByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasTaggeditemByIDReply{}, nil
}

// UpdateByID update a extrasTaggeditem by id
func (s *extrasTaggeditem) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasTaggeditemByIDRequest) (*netbox_goV1.UpdateExtrasTaggeditemByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasTaggeditem{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasTaggeditem.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasTaggeditem", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasTaggeditemByIDReply{}, nil
}

// GetByID get a extrasTaggeditem by id
func (s *extrasTaggeditem) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasTaggeditemByIDRequest) (*netbox_goV1.GetExtrasTaggeditemByIDReply, error) {
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

	data, err := convertExtrasTaggeditem(record)
	if err != nil {
		logger.Warn("convertExtrasTaggeditem error", logger.Err(err), logger.Any("extrasTaggeditem", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasTaggeditem.Err()
	}

	return &netbox_goV1.GetExtrasTaggeditemByIDReply{ExtrasTaggeditem: data}, nil
}

// List get a paginated list of extrasTaggeditems by custom conditions
func (s *extrasTaggeditem) List(ctx context.Context, req *netbox_goV1.ListExtrasTaggeditemRequest) (*netbox_goV1.ListExtrasTaggeditemReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasTaggeditem.Err()
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

	extrasTaggeditems := []*netbox_goV1.ExtrasTaggeditem{}
	for _, record := range records {
		data, err := convertExtrasTaggeditem(record)
		if err != nil {
			logger.Warn("convertExtrasTaggeditem error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasTaggeditems = append(extrasTaggeditems, data)
	}

	return &netbox_goV1.ListExtrasTaggeditemReply{
		Total:             total,
		ExtrasTaggeditems: extrasTaggeditems,
	}, nil
}

// DeleteByIDs batch delete extrasTaggeditem by ids
func (s *extrasTaggeditem) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasTaggeditemByIDsRequest) (*netbox_goV1.DeleteExtrasTaggeditemByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasTaggeditemByIDsReply{}, nil
}

// GetByCondition get a extrasTaggeditem by custom condition
func (s *extrasTaggeditem) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasTaggeditemByConditionRequest) (*netbox_goV1.GetExtrasTaggeditemByConditionReply, error) {
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

	data, err := convertExtrasTaggeditem(record)
	if err != nil {
		logger.Warn("convertExtrasTaggeditem error", logger.Err(err), logger.Any("extrasTaggeditem", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasTaggeditem.Err()
	}

	return &netbox_goV1.GetExtrasTaggeditemByConditionReply{
		ExtrasTaggeditem: data,
	}, nil
}

// ListByIDs batch get extrasTaggeditem by ids
func (s *extrasTaggeditem) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasTaggeditemByIDsRequest) (*netbox_goV1.ListExtrasTaggeditemByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasTaggeditemMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasTaggeditems := []*netbox_goV1.ExtrasTaggeditem{}
	for _, id := range req.Ids {
		if v, ok := extrasTaggeditemMap[id]; ok {
			record, err := convertExtrasTaggeditem(v)
			if err != nil {
				logger.Warn("convertExtrasTaggeditem error", logger.Err(err), logger.Any("extrasTaggeditem", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasTaggeditems = append(extrasTaggeditems, record)
		}
	}

	return &netbox_goV1.ListExtrasTaggeditemByIDsReply{ExtrasTaggeditems: extrasTaggeditems}, nil
}

// ListByLastID get a paginated list of extrasTaggeditems by last id
func (s *extrasTaggeditem) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasTaggeditemByLastIDRequest) (*netbox_goV1.ListExtrasTaggeditemByLastIDReply, error) {
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

	extrasTaggeditems := []*netbox_goV1.ExtrasTaggeditem{}
	for _, record := range records {
		data, err := convertExtrasTaggeditem(record)
		if err != nil {
			logger.Warn("convertExtrasTaggeditem error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasTaggeditems = append(extrasTaggeditems, data)
	}

	return &netbox_goV1.ListExtrasTaggeditemByLastIDReply{
		ExtrasTaggeditems: extrasTaggeditems,
	}, nil
}

func convertExtrasTaggeditem(record *model.ExtrasTaggeditem) (*netbox_goV1.ExtrasTaggeditem, error) {
	value := &netbox_goV1.ExtrasTaggeditem{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
