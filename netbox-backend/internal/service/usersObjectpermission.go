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
		netbox_goV1.RegisterUsersObjectpermissionServer(server, NewUsersObjectpermissionServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.UsersObjectpermissionServer = (*usersObjectpermission)(nil)
var _ time.Time

type usersObjectpermission struct {
	netbox_goV1.UnimplementedUsersObjectpermissionServer

	iDao dao.UsersObjectpermissionDao
}

// NewUsersObjectpermissionServer create a new service
func NewUsersObjectpermissionServer() netbox_goV1.UsersObjectpermissionServer {
	return &usersObjectpermission{
		iDao: dao.NewUsersObjectpermissionDao(
			database.GetDB(), // db driver is postgresql
			cache.NewUsersObjectpermissionCache(database.GetCacheType()),
		),
	}
}

// Create a new usersObjectpermission
func (s *usersObjectpermission) Create(ctx context.Context, req *netbox_goV1.CreateUsersObjectpermissionRequest) (*netbox_goV1.CreateUsersObjectpermissionReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.UsersObjectpermission{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateUsersObjectpermission.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("usersObjectpermission", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateUsersObjectpermissionReply{Id: record.ID}, nil
}

// DeleteByID delete a usersObjectpermission by id
func (s *usersObjectpermission) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteUsersObjectpermissionByIDRequest) (*netbox_goV1.DeleteUsersObjectpermissionByIDReply, error) {
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

	return &netbox_goV1.DeleteUsersObjectpermissionByIDReply{}, nil
}

// UpdateByID update a usersObjectpermission by id
func (s *usersObjectpermission) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateUsersObjectpermissionByIDRequest) (*netbox_goV1.UpdateUsersObjectpermissionByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.UsersObjectpermission{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDUsersObjectpermission.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("usersObjectpermission", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateUsersObjectpermissionByIDReply{}, nil
}

// GetByID get a usersObjectpermission by id
func (s *usersObjectpermission) GetByID(ctx context.Context, req *netbox_goV1.GetUsersObjectpermissionByIDRequest) (*netbox_goV1.GetUsersObjectpermissionByIDReply, error) {
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

	data, err := convertUsersObjectpermission(record)
	if err != nil {
		logger.Warn("convertUsersObjectpermission error", logger.Err(err), logger.Any("usersObjectpermission", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDUsersObjectpermission.Err()
	}

	return &netbox_goV1.GetUsersObjectpermissionByIDReply{UsersObjectpermission: data}, nil
}

// List get a paginated list of usersObjectpermissions by custom conditions
func (s *usersObjectpermission) List(ctx context.Context, req *netbox_goV1.ListUsersObjectpermissionRequest) (*netbox_goV1.ListUsersObjectpermissionReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListUsersObjectpermission.Err()
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

	usersObjectpermissions := []*netbox_goV1.UsersObjectpermission{}
	for _, record := range records {
		data, err := convertUsersObjectpermission(record)
		if err != nil {
			logger.Warn("convertUsersObjectpermission error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		usersObjectpermissions = append(usersObjectpermissions, data)
	}

	return &netbox_goV1.ListUsersObjectpermissionReply{
		Total:                  total,
		UsersObjectpermissions: usersObjectpermissions,
	}, nil
}

// DeleteByIDs batch delete usersObjectpermission by ids
func (s *usersObjectpermission) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteUsersObjectpermissionByIDsRequest) (*netbox_goV1.DeleteUsersObjectpermissionByIDsReply, error) {
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

	return &netbox_goV1.DeleteUsersObjectpermissionByIDsReply{}, nil
}

// GetByCondition get a usersObjectpermission by custom condition
func (s *usersObjectpermission) GetByCondition(ctx context.Context, req *netbox_goV1.GetUsersObjectpermissionByConditionRequest) (*netbox_goV1.GetUsersObjectpermissionByConditionReply, error) {
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

	data, err := convertUsersObjectpermission(record)
	if err != nil {
		logger.Warn("convertUsersObjectpermission error", logger.Err(err), logger.Any("usersObjectpermission", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionUsersObjectpermission.Err()
	}

	return &netbox_goV1.GetUsersObjectpermissionByConditionReply{
		UsersObjectpermission: data,
	}, nil
}

// ListByIDs batch get usersObjectpermission by ids
func (s *usersObjectpermission) ListByIDs(ctx context.Context, req *netbox_goV1.ListUsersObjectpermissionByIDsRequest) (*netbox_goV1.ListUsersObjectpermissionByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	usersObjectpermissionMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	usersObjectpermissions := []*netbox_goV1.UsersObjectpermission{}
	for _, id := range req.Ids {
		if v, ok := usersObjectpermissionMap[id]; ok {
			record, err := convertUsersObjectpermission(v)
			if err != nil {
				logger.Warn("convertUsersObjectpermission error", logger.Err(err), logger.Any("usersObjectpermission", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			usersObjectpermissions = append(usersObjectpermissions, record)
		}
	}

	return &netbox_goV1.ListUsersObjectpermissionByIDsReply{UsersObjectpermissions: usersObjectpermissions}, nil
}

// ListByLastID get a paginated list of usersObjectpermissions by last id
func (s *usersObjectpermission) ListByLastID(ctx context.Context, req *netbox_goV1.ListUsersObjectpermissionByLastIDRequest) (*netbox_goV1.ListUsersObjectpermissionByLastIDReply, error) {
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

	usersObjectpermissions := []*netbox_goV1.UsersObjectpermission{}
	for _, record := range records {
		data, err := convertUsersObjectpermission(record)
		if err != nil {
			logger.Warn("convertUsersObjectpermission error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		usersObjectpermissions = append(usersObjectpermissions, data)
	}

	return &netbox_goV1.ListUsersObjectpermissionByLastIDReply{
		UsersObjectpermissions: usersObjectpermissions,
	}, nil
}

func convertUsersObjectpermission(record *model.UsersObjectpermission) (*netbox_goV1.UsersObjectpermission, error) {
	value := &netbox_goV1.UsersObjectpermission{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
