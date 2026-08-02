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
		netbox_goV1.RegisterTenancyContactroleServer(server, NewTenancyContactroleServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.TenancyContactroleServer = (*tenancyContactrole)(nil)
var _ time.Time

type tenancyContactrole struct {
	netbox_goV1.UnimplementedTenancyContactroleServer

	iDao dao.TenancyContactroleDao
}

// NewTenancyContactroleServer create a new service
func NewTenancyContactroleServer() netbox_goV1.TenancyContactroleServer {
	return &tenancyContactrole{
		iDao: dao.NewTenancyContactroleDao(
			database.GetDB(), // db driver is postgresql
			cache.NewTenancyContactroleCache(database.GetCacheType()),
		),
	}
}

// Create a new tenancyContactrole
func (s *tenancyContactrole) Create(ctx context.Context, req *netbox_goV1.CreateTenancyContactroleRequest) (*netbox_goV1.CreateTenancyContactroleReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.TenancyContactrole{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateTenancyContactrole.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("tenancyContactrole", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateTenancyContactroleReply{Id: record.ID}, nil
}

// DeleteByID delete a tenancyContactrole by id
func (s *tenancyContactrole) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteTenancyContactroleByIDRequest) (*netbox_goV1.DeleteTenancyContactroleByIDReply, error) {
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

	return &netbox_goV1.DeleteTenancyContactroleByIDReply{}, nil
}

// UpdateByID update a tenancyContactrole by id
func (s *tenancyContactrole) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateTenancyContactroleByIDRequest) (*netbox_goV1.UpdateTenancyContactroleByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.TenancyContactrole{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDTenancyContactrole.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("tenancyContactrole", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateTenancyContactroleByIDReply{}, nil
}

// GetByID get a tenancyContactrole by id
func (s *tenancyContactrole) GetByID(ctx context.Context, req *netbox_goV1.GetTenancyContactroleByIDRequest) (*netbox_goV1.GetTenancyContactroleByIDReply, error) {
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

	data, err := convertTenancyContactrole(record)
	if err != nil {
		logger.Warn("convertTenancyContactrole error", logger.Err(err), logger.Any("tenancyContactrole", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDTenancyContactrole.Err()
	}

	return &netbox_goV1.GetTenancyContactroleByIDReply{TenancyContactrole: data}, nil
}

// List get a paginated list of tenancyContactroles by custom conditions
func (s *tenancyContactrole) List(ctx context.Context, req *netbox_goV1.ListTenancyContactroleRequest) (*netbox_goV1.ListTenancyContactroleReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListTenancyContactrole.Err()
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

	tenancyContactroles := []*netbox_goV1.TenancyContactrole{}
	for _, record := range records {
		data, err := convertTenancyContactrole(record)
		if err != nil {
			logger.Warn("convertTenancyContactrole error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		tenancyContactroles = append(tenancyContactroles, data)
	}

	return &netbox_goV1.ListTenancyContactroleReply{
		Total:               total,
		TenancyContactroles: tenancyContactroles,
	}, nil
}

// DeleteByIDs batch delete tenancyContactrole by ids
func (s *tenancyContactrole) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteTenancyContactroleByIDsRequest) (*netbox_goV1.DeleteTenancyContactroleByIDsReply, error) {
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

	return &netbox_goV1.DeleteTenancyContactroleByIDsReply{}, nil
}

// GetByCondition get a tenancyContactrole by custom condition
func (s *tenancyContactrole) GetByCondition(ctx context.Context, req *netbox_goV1.GetTenancyContactroleByConditionRequest) (*netbox_goV1.GetTenancyContactroleByConditionReply, error) {
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

	data, err := convertTenancyContactrole(record)
	if err != nil {
		logger.Warn("convertTenancyContactrole error", logger.Err(err), logger.Any("tenancyContactrole", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionTenancyContactrole.Err()
	}

	return &netbox_goV1.GetTenancyContactroleByConditionReply{
		TenancyContactrole: data,
	}, nil
}

// ListByIDs batch get tenancyContactrole by ids
func (s *tenancyContactrole) ListByIDs(ctx context.Context, req *netbox_goV1.ListTenancyContactroleByIDsRequest) (*netbox_goV1.ListTenancyContactroleByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	tenancyContactroleMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	tenancyContactroles := []*netbox_goV1.TenancyContactrole{}
	for _, id := range req.Ids {
		if v, ok := tenancyContactroleMap[id]; ok {
			record, err := convertTenancyContactrole(v)
			if err != nil {
				logger.Warn("convertTenancyContactrole error", logger.Err(err), logger.Any("tenancyContactrole", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			tenancyContactroles = append(tenancyContactroles, record)
		}
	}

	return &netbox_goV1.ListTenancyContactroleByIDsReply{TenancyContactroles: tenancyContactroles}, nil
}

// ListByLastID get a paginated list of tenancyContactroles by last id
func (s *tenancyContactrole) ListByLastID(ctx context.Context, req *netbox_goV1.ListTenancyContactroleByLastIDRequest) (*netbox_goV1.ListTenancyContactroleByLastIDReply, error) {
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

	tenancyContactroles := []*netbox_goV1.TenancyContactrole{}
	for _, record := range records {
		data, err := convertTenancyContactrole(record)
		if err != nil {
			logger.Warn("convertTenancyContactrole error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		tenancyContactroles = append(tenancyContactroles, data)
	}

	return &netbox_goV1.ListTenancyContactroleByLastIDReply{
		TenancyContactroles: tenancyContactroles,
	}, nil
}

func convertTenancyContactrole(record *model.TenancyContactrole) (*netbox_goV1.TenancyContactrole, error) {
	value := &netbox_goV1.TenancyContactrole{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
