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
		netbox_goV1.RegisterTenancyContactGroupsServer(server, NewTenancyContactGroupsServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.TenancyContactGroupsServer = (*tenancyContactGroups)(nil)
var _ time.Time

type tenancyContactGroups struct {
	netbox_goV1.UnimplementedTenancyContactGroupsServer

	iDao dao.TenancyContactGroupsDao
}

// NewTenancyContactGroupsServer create a new service
func NewTenancyContactGroupsServer() netbox_goV1.TenancyContactGroupsServer {
	return &tenancyContactGroups{
		iDao: dao.NewTenancyContactGroupsDao(
			database.GetDB(), // db driver is postgresql
			cache.NewTenancyContactGroupsCache(database.GetCacheType()),
		),
	}
}

// Create a new tenancyContactGroups
func (s *tenancyContactGroups) Create(ctx context.Context, req *netbox_goV1.CreateTenancyContactGroupsRequest) (*netbox_goV1.CreateTenancyContactGroupsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.TenancyContactGroups{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateTenancyContactGroups.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("tenancyContactGroups", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateTenancyContactGroupsReply{Id: record.ID}, nil
}

// DeleteByID delete a tenancyContactGroups by id
func (s *tenancyContactGroups) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteTenancyContactGroupsByIDRequest) (*netbox_goV1.DeleteTenancyContactGroupsByIDReply, error) {
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

	return &netbox_goV1.DeleteTenancyContactGroupsByIDReply{}, nil
}

// UpdateByID update a tenancyContactGroups by id
func (s *tenancyContactGroups) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateTenancyContactGroupsByIDRequest) (*netbox_goV1.UpdateTenancyContactGroupsByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.TenancyContactGroups{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDTenancyContactGroups.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("tenancyContactGroups", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateTenancyContactGroupsByIDReply{}, nil
}

// GetByID get a tenancyContactGroups by id
func (s *tenancyContactGroups) GetByID(ctx context.Context, req *netbox_goV1.GetTenancyContactGroupsByIDRequest) (*netbox_goV1.GetTenancyContactGroupsByIDReply, error) {
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

	data, err := convertTenancyContactGroups(record)
	if err != nil {
		logger.Warn("convertTenancyContactGroups error", logger.Err(err), logger.Any("tenancyContactGroups", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDTenancyContactGroups.Err()
	}

	return &netbox_goV1.GetTenancyContactGroupsByIDReply{TenancyContactGroups: data}, nil
}

// List get a paginated list of tenancyContactGroupss by custom conditions
func (s *tenancyContactGroups) List(ctx context.Context, req *netbox_goV1.ListTenancyContactGroupsRequest) (*netbox_goV1.ListTenancyContactGroupsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListTenancyContactGroups.Err()
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

	tenancyContactGroupss := []*netbox_goV1.TenancyContactGroups{}
	for _, record := range records {
		data, err := convertTenancyContactGroups(record)
		if err != nil {
			logger.Warn("convertTenancyContactGroups error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		tenancyContactGroupss = append(tenancyContactGroupss, data)
	}

	return &netbox_goV1.ListTenancyContactGroupsReply{
		Total:                 total,
		TenancyContactGroupss: tenancyContactGroupss,
	}, nil
}

// DeleteByIDs batch delete tenancyContactGroups by ids
func (s *tenancyContactGroups) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteTenancyContactGroupsByIDsRequest) (*netbox_goV1.DeleteTenancyContactGroupsByIDsReply, error) {
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

	return &netbox_goV1.DeleteTenancyContactGroupsByIDsReply{}, nil
}

// GetByCondition get a tenancyContactGroups by custom condition
func (s *tenancyContactGroups) GetByCondition(ctx context.Context, req *netbox_goV1.GetTenancyContactGroupsByConditionRequest) (*netbox_goV1.GetTenancyContactGroupsByConditionReply, error) {
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

	data, err := convertTenancyContactGroups(record)
	if err != nil {
		logger.Warn("convertTenancyContactGroups error", logger.Err(err), logger.Any("tenancyContactGroups", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionTenancyContactGroups.Err()
	}

	return &netbox_goV1.GetTenancyContactGroupsByConditionReply{
		TenancyContactGroups: data,
	}, nil
}

// ListByIDs batch get tenancyContactGroups by ids
func (s *tenancyContactGroups) ListByIDs(ctx context.Context, req *netbox_goV1.ListTenancyContactGroupsByIDsRequest) (*netbox_goV1.ListTenancyContactGroupsByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	tenancyContactGroupsMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	tenancyContactGroupss := []*netbox_goV1.TenancyContactGroups{}
	for _, id := range req.Ids {
		if v, ok := tenancyContactGroupsMap[id]; ok {
			record, err := convertTenancyContactGroups(v)
			if err != nil {
				logger.Warn("convertTenancyContactGroups error", logger.Err(err), logger.Any("tenancyContactGroups", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			tenancyContactGroupss = append(tenancyContactGroupss, record)
		}
	}

	return &netbox_goV1.ListTenancyContactGroupsByIDsReply{TenancyContactGroupss: tenancyContactGroupss}, nil
}

// ListByLastID get a paginated list of tenancyContactGroupss by last id
func (s *tenancyContactGroups) ListByLastID(ctx context.Context, req *netbox_goV1.ListTenancyContactGroupsByLastIDRequest) (*netbox_goV1.ListTenancyContactGroupsByLastIDReply, error) {
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

	tenancyContactGroupss := []*netbox_goV1.TenancyContactGroups{}
	for _, record := range records {
		data, err := convertTenancyContactGroups(record)
		if err != nil {
			logger.Warn("convertTenancyContactGroups error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		tenancyContactGroupss = append(tenancyContactGroupss, data)
	}

	return &netbox_goV1.ListTenancyContactGroupsByLastIDReply{
		TenancyContactGroupss: tenancyContactGroupss,
	}, nil
}

func convertTenancyContactGroups(record *model.TenancyContactGroups) (*netbox_goV1.TenancyContactGroups, error) {
	value := &netbox_goV1.TenancyContactGroups{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
