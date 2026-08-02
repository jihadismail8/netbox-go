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
		netbox_goV1.RegisterUsersObjectpermissionObjectTypesServer(server, NewUsersObjectpermissionObjectTypesServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.UsersObjectpermissionObjectTypesServer = (*usersObjectpermissionObjectTypes)(nil)
var _ time.Time

type usersObjectpermissionObjectTypes struct {
	netbox_goV1.UnimplementedUsersObjectpermissionObjectTypesServer

	iDao dao.UsersObjectpermissionObjectTypesDao
}

// NewUsersObjectpermissionObjectTypesServer create a new service
func NewUsersObjectpermissionObjectTypesServer() netbox_goV1.UsersObjectpermissionObjectTypesServer {
	return &usersObjectpermissionObjectTypes{
		iDao: dao.NewUsersObjectpermissionObjectTypesDao(
			database.GetDB(), // db driver is postgresql
			cache.NewUsersObjectpermissionObjectTypesCache(database.GetCacheType()),
		),
	}
}

// Create a new usersObjectpermissionObjectTypes
func (s *usersObjectpermissionObjectTypes) Create(ctx context.Context, req *netbox_goV1.CreateUsersObjectpermissionObjectTypesRequest) (*netbox_goV1.CreateUsersObjectpermissionObjectTypesReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.UsersObjectpermissionObjectTypes{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateUsersObjectpermissionObjectTypes.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("usersObjectpermissionObjectTypes", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateUsersObjectpermissionObjectTypesReply{Id: record.ID}, nil
}

// DeleteByID delete a usersObjectpermissionObjectTypes by id
func (s *usersObjectpermissionObjectTypes) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteUsersObjectpermissionObjectTypesByIDRequest) (*netbox_goV1.DeleteUsersObjectpermissionObjectTypesByIDReply, error) {
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

	return &netbox_goV1.DeleteUsersObjectpermissionObjectTypesByIDReply{}, nil
}

// UpdateByID update a usersObjectpermissionObjectTypes by id
func (s *usersObjectpermissionObjectTypes) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateUsersObjectpermissionObjectTypesByIDRequest) (*netbox_goV1.UpdateUsersObjectpermissionObjectTypesByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.UsersObjectpermissionObjectTypes{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDUsersObjectpermissionObjectTypes.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("usersObjectpermissionObjectTypes", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateUsersObjectpermissionObjectTypesByIDReply{}, nil
}

// GetByID get a usersObjectpermissionObjectTypes by id
func (s *usersObjectpermissionObjectTypes) GetByID(ctx context.Context, req *netbox_goV1.GetUsersObjectpermissionObjectTypesByIDRequest) (*netbox_goV1.GetUsersObjectpermissionObjectTypesByIDReply, error) {
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

	data, err := convertUsersObjectpermissionObjectTypes(record)
	if err != nil {
		logger.Warn("convertUsersObjectpermissionObjectTypes error", logger.Err(err), logger.Any("usersObjectpermissionObjectTypes", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDUsersObjectpermissionObjectTypes.Err()
	}

	return &netbox_goV1.GetUsersObjectpermissionObjectTypesByIDReply{UsersObjectpermissionObjectTypes: data}, nil
}

// List get a paginated list of usersObjectpermissionObjectTypess by custom conditions
func (s *usersObjectpermissionObjectTypes) List(ctx context.Context, req *netbox_goV1.ListUsersObjectpermissionObjectTypesRequest) (*netbox_goV1.ListUsersObjectpermissionObjectTypesReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListUsersObjectpermissionObjectTypes.Err()
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

	usersObjectpermissionObjectTypess := []*netbox_goV1.UsersObjectpermissionObjectTypes{}
	for _, record := range records {
		data, err := convertUsersObjectpermissionObjectTypes(record)
		if err != nil {
			logger.Warn("convertUsersObjectpermissionObjectTypes error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		usersObjectpermissionObjectTypess = append(usersObjectpermissionObjectTypess, data)
	}

	return &netbox_goV1.ListUsersObjectpermissionObjectTypesReply{
		Total:                             total,
		UsersObjectpermissionObjectTypess: usersObjectpermissionObjectTypess,
	}, nil
}

// DeleteByIDs batch delete usersObjectpermissionObjectTypes by ids
func (s *usersObjectpermissionObjectTypes) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteUsersObjectpermissionObjectTypesByIDsRequest) (*netbox_goV1.DeleteUsersObjectpermissionObjectTypesByIDsReply, error) {
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

	return &netbox_goV1.DeleteUsersObjectpermissionObjectTypesByIDsReply{}, nil
}

// GetByCondition get a usersObjectpermissionObjectTypes by custom condition
func (s *usersObjectpermissionObjectTypes) GetByCondition(ctx context.Context, req *netbox_goV1.GetUsersObjectpermissionObjectTypesByConditionRequest) (*netbox_goV1.GetUsersObjectpermissionObjectTypesByConditionReply, error) {
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

	data, err := convertUsersObjectpermissionObjectTypes(record)
	if err != nil {
		logger.Warn("convertUsersObjectpermissionObjectTypes error", logger.Err(err), logger.Any("usersObjectpermissionObjectTypes", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionUsersObjectpermissionObjectTypes.Err()
	}

	return &netbox_goV1.GetUsersObjectpermissionObjectTypesByConditionReply{
		UsersObjectpermissionObjectTypes: data,
	}, nil
}

// ListByIDs batch get usersObjectpermissionObjectTypes by ids
func (s *usersObjectpermissionObjectTypes) ListByIDs(ctx context.Context, req *netbox_goV1.ListUsersObjectpermissionObjectTypesByIDsRequest) (*netbox_goV1.ListUsersObjectpermissionObjectTypesByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	usersObjectpermissionObjectTypesMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	usersObjectpermissionObjectTypess := []*netbox_goV1.UsersObjectpermissionObjectTypes{}
	for _, id := range req.Ids {
		if v, ok := usersObjectpermissionObjectTypesMap[id]; ok {
			record, err := convertUsersObjectpermissionObjectTypes(v)
			if err != nil {
				logger.Warn("convertUsersObjectpermissionObjectTypes error", logger.Err(err), logger.Any("usersObjectpermissionObjectTypes", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			usersObjectpermissionObjectTypess = append(usersObjectpermissionObjectTypess, record)
		}
	}

	return &netbox_goV1.ListUsersObjectpermissionObjectTypesByIDsReply{UsersObjectpermissionObjectTypess: usersObjectpermissionObjectTypess}, nil
}

// ListByLastID get a paginated list of usersObjectpermissionObjectTypess by last id
func (s *usersObjectpermissionObjectTypes) ListByLastID(ctx context.Context, req *netbox_goV1.ListUsersObjectpermissionObjectTypesByLastIDRequest) (*netbox_goV1.ListUsersObjectpermissionObjectTypesByLastIDReply, error) {
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

	usersObjectpermissionObjectTypess := []*netbox_goV1.UsersObjectpermissionObjectTypes{}
	for _, record := range records {
		data, err := convertUsersObjectpermissionObjectTypes(record)
		if err != nil {
			logger.Warn("convertUsersObjectpermissionObjectTypes error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		usersObjectpermissionObjectTypess = append(usersObjectpermissionObjectTypess, data)
	}

	return &netbox_goV1.ListUsersObjectpermissionObjectTypesByLastIDReply{
		UsersObjectpermissionObjectTypess: usersObjectpermissionObjectTypess,
	}, nil
}

func convertUsersObjectpermissionObjectTypes(record *model.UsersObjectpermissionObjectTypes) (*netbox_goV1.UsersObjectpermissionObjectTypes, error) {
	value := &netbox_goV1.UsersObjectpermissionObjectTypes{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
