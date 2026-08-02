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
		netbox_goV1.RegisterTenancyContactgroupServer(server, NewTenancyContactgroupServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.TenancyContactgroupServer = (*tenancyContactgroup)(nil)
var _ time.Time

type tenancyContactgroup struct {
	netbox_goV1.UnimplementedTenancyContactgroupServer

	iDao dao.TenancyContactgroupDao
}

// NewTenancyContactgroupServer create a new service
func NewTenancyContactgroupServer() netbox_goV1.TenancyContactgroupServer {
	return &tenancyContactgroup{
		iDao: dao.NewTenancyContactgroupDao(
			database.GetDB(), // db driver is postgresql
			cache.NewTenancyContactgroupCache(database.GetCacheType()),
		),
	}
}

// Create a new tenancyContactgroup
func (s *tenancyContactgroup) Create(ctx context.Context, req *netbox_goV1.CreateTenancyContactgroupRequest) (*netbox_goV1.CreateTenancyContactgroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.TenancyContactgroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateTenancyContactgroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("tenancyContactgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateTenancyContactgroupReply{Id: record.ID}, nil
}

// DeleteByID delete a tenancyContactgroup by id
func (s *tenancyContactgroup) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteTenancyContactgroupByIDRequest) (*netbox_goV1.DeleteTenancyContactgroupByIDReply, error) {
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

	return &netbox_goV1.DeleteTenancyContactgroupByIDReply{}, nil
}

// UpdateByID update a tenancyContactgroup by id
func (s *tenancyContactgroup) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateTenancyContactgroupByIDRequest) (*netbox_goV1.UpdateTenancyContactgroupByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.TenancyContactgroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDTenancyContactgroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("tenancyContactgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateTenancyContactgroupByIDReply{}, nil
}

// GetByID get a tenancyContactgroup by id
func (s *tenancyContactgroup) GetByID(ctx context.Context, req *netbox_goV1.GetTenancyContactgroupByIDRequest) (*netbox_goV1.GetTenancyContactgroupByIDReply, error) {
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

	data, err := convertTenancyContactgroup(record)
	if err != nil {
		logger.Warn("convertTenancyContactgroup error", logger.Err(err), logger.Any("tenancyContactgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDTenancyContactgroup.Err()
	}

	return &netbox_goV1.GetTenancyContactgroupByIDReply{TenancyContactgroup: data}, nil
}

// List get a paginated list of tenancyContactgroups by custom conditions
func (s *tenancyContactgroup) List(ctx context.Context, req *netbox_goV1.ListTenancyContactgroupRequest) (*netbox_goV1.ListTenancyContactgroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListTenancyContactgroup.Err()
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

	tenancyContactgroups := []*netbox_goV1.TenancyContactgroup{}
	for _, record := range records {
		data, err := convertTenancyContactgroup(record)
		if err != nil {
			logger.Warn("convertTenancyContactgroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		tenancyContactgroups = append(tenancyContactgroups, data)
	}

	return &netbox_goV1.ListTenancyContactgroupReply{
		Total:                total,
		TenancyContactgroups: tenancyContactgroups,
	}, nil
}

// DeleteByIDs batch delete tenancyContactgroup by ids
func (s *tenancyContactgroup) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteTenancyContactgroupByIDsRequest) (*netbox_goV1.DeleteTenancyContactgroupByIDsReply, error) {
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

	return &netbox_goV1.DeleteTenancyContactgroupByIDsReply{}, nil
}

// GetByCondition get a tenancyContactgroup by custom condition
func (s *tenancyContactgroup) GetByCondition(ctx context.Context, req *netbox_goV1.GetTenancyContactgroupByConditionRequest) (*netbox_goV1.GetTenancyContactgroupByConditionReply, error) {
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

	data, err := convertTenancyContactgroup(record)
	if err != nil {
		logger.Warn("convertTenancyContactgroup error", logger.Err(err), logger.Any("tenancyContactgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionTenancyContactgroup.Err()
	}

	return &netbox_goV1.GetTenancyContactgroupByConditionReply{
		TenancyContactgroup: data,
	}, nil
}

// ListByIDs batch get tenancyContactgroup by ids
func (s *tenancyContactgroup) ListByIDs(ctx context.Context, req *netbox_goV1.ListTenancyContactgroupByIDsRequest) (*netbox_goV1.ListTenancyContactgroupByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	tenancyContactgroupMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	tenancyContactgroups := []*netbox_goV1.TenancyContactgroup{}
	for _, id := range req.Ids {
		if v, ok := tenancyContactgroupMap[id]; ok {
			record, err := convertTenancyContactgroup(v)
			if err != nil {
				logger.Warn("convertTenancyContactgroup error", logger.Err(err), logger.Any("tenancyContactgroup", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			tenancyContactgroups = append(tenancyContactgroups, record)
		}
	}

	return &netbox_goV1.ListTenancyContactgroupByIDsReply{TenancyContactgroups: tenancyContactgroups}, nil
}

// ListByLastID get a paginated list of tenancyContactgroups by last id
func (s *tenancyContactgroup) ListByLastID(ctx context.Context, req *netbox_goV1.ListTenancyContactgroupByLastIDRequest) (*netbox_goV1.ListTenancyContactgroupByLastIDReply, error) {
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

	tenancyContactgroups := []*netbox_goV1.TenancyContactgroup{}
	for _, record := range records {
		data, err := convertTenancyContactgroup(record)
		if err != nil {
			logger.Warn("convertTenancyContactgroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		tenancyContactgroups = append(tenancyContactgroups, data)
	}

	return &netbox_goV1.ListTenancyContactgroupByLastIDReply{
		TenancyContactgroups: tenancyContactgroups,
	}, nil
}

func convertTenancyContactgroup(record *model.TenancyContactgroup) (*netbox_goV1.TenancyContactgroup, error) {
	value := &netbox_goV1.TenancyContactgroup{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
