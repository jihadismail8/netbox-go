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
		netbox_goV1.RegisterTenancyTenantServer(server, NewTenancyTenantServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.TenancyTenantServer = (*tenancyTenant)(nil)
var _ time.Time

type tenancyTenant struct {
	netbox_goV1.UnimplementedTenancyTenantServer

	iDao dao.TenancyTenantDao
}

// NewTenancyTenantServer create a new service
func NewTenancyTenantServer() netbox_goV1.TenancyTenantServer {
	return &tenancyTenant{
		iDao: dao.NewTenancyTenantDao(
			database.GetDB(), // db driver is postgresql
			cache.NewTenancyTenantCache(database.GetCacheType()),
		),
	}
}

// Create a new tenancyTenant
func (s *tenancyTenant) Create(ctx context.Context, req *netbox_goV1.CreateTenancyTenantRequest) (*netbox_goV1.CreateTenancyTenantReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.TenancyTenant{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateTenancyTenant.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("tenancyTenant", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateTenancyTenantReply{Id: record.ID}, nil
}

// DeleteByID delete a tenancyTenant by id
func (s *tenancyTenant) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteTenancyTenantByIDRequest) (*netbox_goV1.DeleteTenancyTenantByIDReply, error) {
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

	return &netbox_goV1.DeleteTenancyTenantByIDReply{}, nil
}

// UpdateByID update a tenancyTenant by id
func (s *tenancyTenant) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateTenancyTenantByIDRequest) (*netbox_goV1.UpdateTenancyTenantByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.TenancyTenant{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDTenancyTenant.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("tenancyTenant", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateTenancyTenantByIDReply{}, nil
}

// GetByID get a tenancyTenant by id
func (s *tenancyTenant) GetByID(ctx context.Context, req *netbox_goV1.GetTenancyTenantByIDRequest) (*netbox_goV1.GetTenancyTenantByIDReply, error) {
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

	data, err := convertTenancyTenant(record)
	if err != nil {
		logger.Warn("convertTenancyTenant error", logger.Err(err), logger.Any("tenancyTenant", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDTenancyTenant.Err()
	}

	return &netbox_goV1.GetTenancyTenantByIDReply{TenancyTenant: data}, nil
}

// List get a paginated list of tenancyTenants by custom conditions
func (s *tenancyTenant) List(ctx context.Context, req *netbox_goV1.ListTenancyTenantRequest) (*netbox_goV1.ListTenancyTenantReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListTenancyTenant.Err()
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

	tenancyTenants := []*netbox_goV1.TenancyTenant{}
	for _, record := range records {
		data, err := convertTenancyTenant(record)
		if err != nil {
			logger.Warn("convertTenancyTenant error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		tenancyTenants = append(tenancyTenants, data)
	}

	return &netbox_goV1.ListTenancyTenantReply{
		Total:          total,
		TenancyTenants: tenancyTenants,
	}, nil
}

// DeleteByIDs batch delete tenancyTenant by ids
func (s *tenancyTenant) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteTenancyTenantByIDsRequest) (*netbox_goV1.DeleteTenancyTenantByIDsReply, error) {
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

	return &netbox_goV1.DeleteTenancyTenantByIDsReply{}, nil
}

// GetByCondition get a tenancyTenant by custom condition
func (s *tenancyTenant) GetByCondition(ctx context.Context, req *netbox_goV1.GetTenancyTenantByConditionRequest) (*netbox_goV1.GetTenancyTenantByConditionReply, error) {
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

	data, err := convertTenancyTenant(record)
	if err != nil {
		logger.Warn("convertTenancyTenant error", logger.Err(err), logger.Any("tenancyTenant", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionTenancyTenant.Err()
	}

	return &netbox_goV1.GetTenancyTenantByConditionReply{
		TenancyTenant: data,
	}, nil
}

// ListByIDs batch get tenancyTenant by ids
func (s *tenancyTenant) ListByIDs(ctx context.Context, req *netbox_goV1.ListTenancyTenantByIDsRequest) (*netbox_goV1.ListTenancyTenantByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	tenancyTenantMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	tenancyTenants := []*netbox_goV1.TenancyTenant{}
	for _, id := range req.Ids {
		if v, ok := tenancyTenantMap[id]; ok {
			record, err := convertTenancyTenant(v)
			if err != nil {
				logger.Warn("convertTenancyTenant error", logger.Err(err), logger.Any("tenancyTenant", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			tenancyTenants = append(tenancyTenants, record)
		}
	}

	return &netbox_goV1.ListTenancyTenantByIDsReply{TenancyTenants: tenancyTenants}, nil
}

// ListByLastID get a paginated list of tenancyTenants by last id
func (s *tenancyTenant) ListByLastID(ctx context.Context, req *netbox_goV1.ListTenancyTenantByLastIDRequest) (*netbox_goV1.ListTenancyTenantByLastIDReply, error) {
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

	tenancyTenants := []*netbox_goV1.TenancyTenant{}
	for _, record := range records {
		data, err := convertTenancyTenant(record)
		if err != nil {
			logger.Warn("convertTenancyTenant error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		tenancyTenants = append(tenancyTenants, data)
	}

	return &netbox_goV1.ListTenancyTenantByLastIDReply{
		TenancyTenants: tenancyTenants,
	}, nil
}

func convertTenancyTenant(record *model.TenancyTenant) (*netbox_goV1.TenancyTenant, error) {
	value := &netbox_goV1.TenancyTenant{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
