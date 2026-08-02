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
		netbox_goV1.RegisterTenancyContactassignmentServer(server, NewTenancyContactassignmentServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.TenancyContactassignmentServer = (*tenancyContactassignment)(nil)
var _ time.Time

type tenancyContactassignment struct {
	netbox_goV1.UnimplementedTenancyContactassignmentServer

	iDao dao.TenancyContactassignmentDao
}

// NewTenancyContactassignmentServer create a new service
func NewTenancyContactassignmentServer() netbox_goV1.TenancyContactassignmentServer {
	return &tenancyContactassignment{
		iDao: dao.NewTenancyContactassignmentDao(
			database.GetDB(), // db driver is postgresql
			cache.NewTenancyContactassignmentCache(database.GetCacheType()),
		),
	}
}

// Create a new tenancyContactassignment
func (s *tenancyContactassignment) Create(ctx context.Context, req *netbox_goV1.CreateTenancyContactassignmentRequest) (*netbox_goV1.CreateTenancyContactassignmentReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.TenancyContactassignment{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateTenancyContactassignment.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("tenancyContactassignment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateTenancyContactassignmentReply{Id: record.ID}, nil
}

// DeleteByID delete a tenancyContactassignment by id
func (s *tenancyContactassignment) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteTenancyContactassignmentByIDRequest) (*netbox_goV1.DeleteTenancyContactassignmentByIDReply, error) {
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

	return &netbox_goV1.DeleteTenancyContactassignmentByIDReply{}, nil
}

// UpdateByID update a tenancyContactassignment by id
func (s *tenancyContactassignment) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateTenancyContactassignmentByIDRequest) (*netbox_goV1.UpdateTenancyContactassignmentByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.TenancyContactassignment{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDTenancyContactassignment.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("tenancyContactassignment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateTenancyContactassignmentByIDReply{}, nil
}

// GetByID get a tenancyContactassignment by id
func (s *tenancyContactassignment) GetByID(ctx context.Context, req *netbox_goV1.GetTenancyContactassignmentByIDRequest) (*netbox_goV1.GetTenancyContactassignmentByIDReply, error) {
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

	data, err := convertTenancyContactassignment(record)
	if err != nil {
		logger.Warn("convertTenancyContactassignment error", logger.Err(err), logger.Any("tenancyContactassignment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDTenancyContactassignment.Err()
	}

	return &netbox_goV1.GetTenancyContactassignmentByIDReply{TenancyContactassignment: data}, nil
}

// List get a paginated list of tenancyContactassignments by custom conditions
func (s *tenancyContactassignment) List(ctx context.Context, req *netbox_goV1.ListTenancyContactassignmentRequest) (*netbox_goV1.ListTenancyContactassignmentReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListTenancyContactassignment.Err()
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

	tenancyContactassignments := []*netbox_goV1.TenancyContactassignment{}
	for _, record := range records {
		data, err := convertTenancyContactassignment(record)
		if err != nil {
			logger.Warn("convertTenancyContactassignment error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		tenancyContactassignments = append(tenancyContactassignments, data)
	}

	return &netbox_goV1.ListTenancyContactassignmentReply{
		Total:                     total,
		TenancyContactassignments: tenancyContactassignments,
	}, nil
}

// DeleteByIDs batch delete tenancyContactassignment by ids
func (s *tenancyContactassignment) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteTenancyContactassignmentByIDsRequest) (*netbox_goV1.DeleteTenancyContactassignmentByIDsReply, error) {
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

	return &netbox_goV1.DeleteTenancyContactassignmentByIDsReply{}, nil
}

// GetByCondition get a tenancyContactassignment by custom condition
func (s *tenancyContactassignment) GetByCondition(ctx context.Context, req *netbox_goV1.GetTenancyContactassignmentByConditionRequest) (*netbox_goV1.GetTenancyContactassignmentByConditionReply, error) {
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

	data, err := convertTenancyContactassignment(record)
	if err != nil {
		logger.Warn("convertTenancyContactassignment error", logger.Err(err), logger.Any("tenancyContactassignment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionTenancyContactassignment.Err()
	}

	return &netbox_goV1.GetTenancyContactassignmentByConditionReply{
		TenancyContactassignment: data,
	}, nil
}

// ListByIDs batch get tenancyContactassignment by ids
func (s *tenancyContactassignment) ListByIDs(ctx context.Context, req *netbox_goV1.ListTenancyContactassignmentByIDsRequest) (*netbox_goV1.ListTenancyContactassignmentByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	tenancyContactassignmentMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	tenancyContactassignments := []*netbox_goV1.TenancyContactassignment{}
	for _, id := range req.Ids {
		if v, ok := tenancyContactassignmentMap[id]; ok {
			record, err := convertTenancyContactassignment(v)
			if err != nil {
				logger.Warn("convertTenancyContactassignment error", logger.Err(err), logger.Any("tenancyContactassignment", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			tenancyContactassignments = append(tenancyContactassignments, record)
		}
	}

	return &netbox_goV1.ListTenancyContactassignmentByIDsReply{TenancyContactassignments: tenancyContactassignments}, nil
}

// ListByLastID get a paginated list of tenancyContactassignments by last id
func (s *tenancyContactassignment) ListByLastID(ctx context.Context, req *netbox_goV1.ListTenancyContactassignmentByLastIDRequest) (*netbox_goV1.ListTenancyContactassignmentByLastIDReply, error) {
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

	tenancyContactassignments := []*netbox_goV1.TenancyContactassignment{}
	for _, record := range records {
		data, err := convertTenancyContactassignment(record)
		if err != nil {
			logger.Warn("convertTenancyContactassignment error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		tenancyContactassignments = append(tenancyContactassignments, data)
	}

	return &netbox_goV1.ListTenancyContactassignmentByLastIDReply{
		TenancyContactassignments: tenancyContactassignments,
	}, nil
}

func convertTenancyContactassignment(record *model.TenancyContactassignment) (*netbox_goV1.TenancyContactassignment, error) {
	value := &netbox_goV1.TenancyContactassignment{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
