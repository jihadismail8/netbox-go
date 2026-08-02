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
		netbox_goV1.RegisterTenancyTenantgroupServer(server, NewTenancyTenantgroupServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.TenancyTenantgroupServer = (*tenancyTenantgroup)(nil)
var _ time.Time

type tenancyTenantgroup struct {
	netbox_goV1.UnimplementedTenancyTenantgroupServer

	iDao dao.TenancyTenantgroupDao
}

// NewTenancyTenantgroupServer create a new service
func NewTenancyTenantgroupServer() netbox_goV1.TenancyTenantgroupServer {
	return &tenancyTenantgroup{
		iDao: dao.NewTenancyTenantgroupDao(
			database.GetDB(), // db driver is postgresql
			cache.NewTenancyTenantgroupCache(database.GetCacheType()),
		),
	}
}

// Create a new tenancyTenantgroup
func (s *tenancyTenantgroup) Create(ctx context.Context, req *netbox_goV1.CreateTenancyTenantgroupRequest) (*netbox_goV1.CreateTenancyTenantgroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.TenancyTenantgroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateTenancyTenantgroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("tenancyTenantgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateTenancyTenantgroupReply{Id: record.ID}, nil
}

// DeleteByID delete a tenancyTenantgroup by id
func (s *tenancyTenantgroup) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteTenancyTenantgroupByIDRequest) (*netbox_goV1.DeleteTenancyTenantgroupByIDReply, error) {
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

	return &netbox_goV1.DeleteTenancyTenantgroupByIDReply{}, nil
}

// UpdateByID update a tenancyTenantgroup by id
func (s *tenancyTenantgroup) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateTenancyTenantgroupByIDRequest) (*netbox_goV1.UpdateTenancyTenantgroupByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.TenancyTenantgroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDTenancyTenantgroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("tenancyTenantgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateTenancyTenantgroupByIDReply{}, nil
}

// GetByID get a tenancyTenantgroup by id
func (s *tenancyTenantgroup) GetByID(ctx context.Context, req *netbox_goV1.GetTenancyTenantgroupByIDRequest) (*netbox_goV1.GetTenancyTenantgroupByIDReply, error) {
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

	data, err := convertTenancyTenantgroup(record)
	if err != nil {
		logger.Warn("convertTenancyTenantgroup error", logger.Err(err), logger.Any("tenancyTenantgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDTenancyTenantgroup.Err()
	}

	return &netbox_goV1.GetTenancyTenantgroupByIDReply{TenancyTenantgroup: data}, nil
}

// List get a paginated list of tenancyTenantgroups by custom conditions
func (s *tenancyTenantgroup) List(ctx context.Context, req *netbox_goV1.ListTenancyTenantgroupRequest) (*netbox_goV1.ListTenancyTenantgroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListTenancyTenantgroup.Err()
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

	tenancyTenantgroups := []*netbox_goV1.TenancyTenantgroup{}
	for _, record := range records {
		data, err := convertTenancyTenantgroup(record)
		if err != nil {
			logger.Warn("convertTenancyTenantgroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		tenancyTenantgroups = append(tenancyTenantgroups, data)
	}

	return &netbox_goV1.ListTenancyTenantgroupReply{
		Total:               total,
		TenancyTenantgroups: tenancyTenantgroups,
	}, nil
}

// DeleteByIDs batch delete tenancyTenantgroup by ids
func (s *tenancyTenantgroup) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteTenancyTenantgroupByIDsRequest) (*netbox_goV1.DeleteTenancyTenantgroupByIDsReply, error) {
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

	return &netbox_goV1.DeleteTenancyTenantgroupByIDsReply{}, nil
}

// GetByCondition get a tenancyTenantgroup by custom condition
func (s *tenancyTenantgroup) GetByCondition(ctx context.Context, req *netbox_goV1.GetTenancyTenantgroupByConditionRequest) (*netbox_goV1.GetTenancyTenantgroupByConditionReply, error) {
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

	data, err := convertTenancyTenantgroup(record)
	if err != nil {
		logger.Warn("convertTenancyTenantgroup error", logger.Err(err), logger.Any("tenancyTenantgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionTenancyTenantgroup.Err()
	}

	return &netbox_goV1.GetTenancyTenantgroupByConditionReply{
		TenancyTenantgroup: data,
	}, nil
}

// ListByIDs batch get tenancyTenantgroup by ids
func (s *tenancyTenantgroup) ListByIDs(ctx context.Context, req *netbox_goV1.ListTenancyTenantgroupByIDsRequest) (*netbox_goV1.ListTenancyTenantgroupByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	tenancyTenantgroupMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	tenancyTenantgroups := []*netbox_goV1.TenancyTenantgroup{}
	for _, id := range req.Ids {
		if v, ok := tenancyTenantgroupMap[id]; ok {
			record, err := convertTenancyTenantgroup(v)
			if err != nil {
				logger.Warn("convertTenancyTenantgroup error", logger.Err(err), logger.Any("tenancyTenantgroup", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			tenancyTenantgroups = append(tenancyTenantgroups, record)
		}
	}

	return &netbox_goV1.ListTenancyTenantgroupByIDsReply{TenancyTenantgroups: tenancyTenantgroups}, nil
}

// ListByLastID get a paginated list of tenancyTenantgroups by last id
func (s *tenancyTenantgroup) ListByLastID(ctx context.Context, req *netbox_goV1.ListTenancyTenantgroupByLastIDRequest) (*netbox_goV1.ListTenancyTenantgroupByLastIDReply, error) {
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

	tenancyTenantgroups := []*netbox_goV1.TenancyTenantgroup{}
	for _, record := range records {
		data, err := convertTenancyTenantgroup(record)
		if err != nil {
			logger.Warn("convertTenancyTenantgroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		tenancyTenantgroups = append(tenancyTenantgroups, data)
	}

	return &netbox_goV1.ListTenancyTenantgroupByLastIDReply{
		TenancyTenantgroups: tenancyTenantgroups,
	}, nil
}

func convertTenancyTenantgroup(record *model.TenancyTenantgroup) (*netbox_goV1.TenancyTenantgroup, error) {
	value := &netbox_goV1.TenancyTenantgroup{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
