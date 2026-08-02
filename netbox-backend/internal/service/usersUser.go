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
		netbox_goV1.RegisterUsersUserServer(server, NewUsersUserServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.UsersUserServer = (*usersUser)(nil)
var _ time.Time

type usersUser struct {
	netbox_goV1.UnimplementedUsersUserServer

	iDao dao.UsersUserDao
}

// NewUsersUserServer create a new service
func NewUsersUserServer() netbox_goV1.UsersUserServer {
	return &usersUser{
		iDao: dao.NewUsersUserDao(
			database.GetDB(), // db driver is postgresql
			cache.NewUsersUserCache(database.GetCacheType()),
		),
	}
}

// Create a new usersUser
func (s *usersUser) Create(ctx context.Context, req *netbox_goV1.CreateUsersUserRequest) (*netbox_goV1.CreateUsersUserReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.UsersUser{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateUsersUser.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("usersUser", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateUsersUserReply{Id: record.ID}, nil
}

// DeleteByID delete a usersUser by id
func (s *usersUser) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteUsersUserByIDRequest) (*netbox_goV1.DeleteUsersUserByIDReply, error) {
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

	return &netbox_goV1.DeleteUsersUserByIDReply{}, nil
}

// UpdateByID update a usersUser by id
func (s *usersUser) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateUsersUserByIDRequest) (*netbox_goV1.UpdateUsersUserByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.UsersUser{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDUsersUser.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("usersUser", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateUsersUserByIDReply{}, nil
}

// GetByID get a usersUser by id
func (s *usersUser) GetByID(ctx context.Context, req *netbox_goV1.GetUsersUserByIDRequest) (*netbox_goV1.GetUsersUserByIDReply, error) {
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

	data, err := convertUsersUser(record)
	if err != nil {
		logger.Warn("convertUsersUser error", logger.Err(err), logger.Any("usersUser", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDUsersUser.Err()
	}

	return &netbox_goV1.GetUsersUserByIDReply{UsersUser: data}, nil
}

// List get a paginated list of usersUsers by custom conditions
func (s *usersUser) List(ctx context.Context, req *netbox_goV1.ListUsersUserRequest) (*netbox_goV1.ListUsersUserReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListUsersUser.Err()
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

	usersUsers := []*netbox_goV1.UsersUser{}
	for _, record := range records {
		data, err := convertUsersUser(record)
		if err != nil {
			logger.Warn("convertUsersUser error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		usersUsers = append(usersUsers, data)
	}

	return &netbox_goV1.ListUsersUserReply{
		Total:      total,
		UsersUsers: usersUsers,
	}, nil
}

// DeleteByIDs batch delete usersUser by ids
func (s *usersUser) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteUsersUserByIDsRequest) (*netbox_goV1.DeleteUsersUserByIDsReply, error) {
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

	return &netbox_goV1.DeleteUsersUserByIDsReply{}, nil
}

// GetByCondition get a usersUser by custom condition
func (s *usersUser) GetByCondition(ctx context.Context, req *netbox_goV1.GetUsersUserByConditionRequest) (*netbox_goV1.GetUsersUserByConditionReply, error) {
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

	data, err := convertUsersUser(record)
	if err != nil {
		logger.Warn("convertUsersUser error", logger.Err(err), logger.Any("usersUser", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionUsersUser.Err()
	}

	return &netbox_goV1.GetUsersUserByConditionReply{
		UsersUser: data,
	}, nil
}

// ListByIDs batch get usersUser by ids
func (s *usersUser) ListByIDs(ctx context.Context, req *netbox_goV1.ListUsersUserByIDsRequest) (*netbox_goV1.ListUsersUserByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	usersUserMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	usersUsers := []*netbox_goV1.UsersUser{}
	for _, id := range req.Ids {
		if v, ok := usersUserMap[id]; ok {
			record, err := convertUsersUser(v)
			if err != nil {
				logger.Warn("convertUsersUser error", logger.Err(err), logger.Any("usersUser", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			usersUsers = append(usersUsers, record)
		}
	}

	return &netbox_goV1.ListUsersUserByIDsReply{UsersUsers: usersUsers}, nil
}

// ListByLastID get a paginated list of usersUsers by last id
func (s *usersUser) ListByLastID(ctx context.Context, req *netbox_goV1.ListUsersUserByLastIDRequest) (*netbox_goV1.ListUsersUserByLastIDReply, error) {
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

	usersUsers := []*netbox_goV1.UsersUser{}
	for _, record := range records {
		data, err := convertUsersUser(record)
		if err != nil {
			logger.Warn("convertUsersUser error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		usersUsers = append(usersUsers, data)
	}

	return &netbox_goV1.ListUsersUserByLastIDReply{
		UsersUsers: usersUsers,
	}, nil
}

func convertUsersUser(record *model.UsersUser) (*netbox_goV1.UsersUser, error) {
	value := &netbox_goV1.UsersUser{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
