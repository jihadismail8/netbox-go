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
		netbox_goV1.RegisterUsersTokenServer(server, NewUsersTokenServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.UsersTokenServer = (*usersToken)(nil)
var _ time.Time

type usersToken struct {
	netbox_goV1.UnimplementedUsersTokenServer

	iDao dao.UsersTokenDao
}

// NewUsersTokenServer create a new service
func NewUsersTokenServer() netbox_goV1.UsersTokenServer {
	return &usersToken{
		iDao: dao.NewUsersTokenDao(
			database.GetDB(), // db driver is postgresql
			cache.NewUsersTokenCache(database.GetCacheType()),
		),
	}
}

// Create a new usersToken
func (s *usersToken) Create(ctx context.Context, req *netbox_goV1.CreateUsersTokenRequest) (*netbox_goV1.CreateUsersTokenReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.UsersToken{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateUsersToken.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("usersToken", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateUsersTokenReply{Id: record.ID}, nil
}

// DeleteByID delete a usersToken by id
func (s *usersToken) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteUsersTokenByIDRequest) (*netbox_goV1.DeleteUsersTokenByIDReply, error) {
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

	return &netbox_goV1.DeleteUsersTokenByIDReply{}, nil
}

// UpdateByID update a usersToken by id
func (s *usersToken) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateUsersTokenByIDRequest) (*netbox_goV1.UpdateUsersTokenByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.UsersToken{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDUsersToken.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("usersToken", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateUsersTokenByIDReply{}, nil
}

// GetByID get a usersToken by id
func (s *usersToken) GetByID(ctx context.Context, req *netbox_goV1.GetUsersTokenByIDRequest) (*netbox_goV1.GetUsersTokenByIDReply, error) {
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

	data, err := convertUsersToken(record)
	if err != nil {
		logger.Warn("convertUsersToken error", logger.Err(err), logger.Any("usersToken", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDUsersToken.Err()
	}

	return &netbox_goV1.GetUsersTokenByIDReply{UsersToken: data}, nil
}

// List get a paginated list of usersTokens by custom conditions
func (s *usersToken) List(ctx context.Context, req *netbox_goV1.ListUsersTokenRequest) (*netbox_goV1.ListUsersTokenReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListUsersToken.Err()
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

	usersTokens := []*netbox_goV1.UsersToken{}
	for _, record := range records {
		data, err := convertUsersToken(record)
		if err != nil {
			logger.Warn("convertUsersToken error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		usersTokens = append(usersTokens, data)
	}

	return &netbox_goV1.ListUsersTokenReply{
		Total:       total,
		UsersTokens: usersTokens,
	}, nil
}

// DeleteByIDs batch delete usersToken by ids
func (s *usersToken) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteUsersTokenByIDsRequest) (*netbox_goV1.DeleteUsersTokenByIDsReply, error) {
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

	return &netbox_goV1.DeleteUsersTokenByIDsReply{}, nil
}

// GetByCondition get a usersToken by custom condition
func (s *usersToken) GetByCondition(ctx context.Context, req *netbox_goV1.GetUsersTokenByConditionRequest) (*netbox_goV1.GetUsersTokenByConditionReply, error) {
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

	data, err := convertUsersToken(record)
	if err != nil {
		logger.Warn("convertUsersToken error", logger.Err(err), logger.Any("usersToken", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionUsersToken.Err()
	}

	return &netbox_goV1.GetUsersTokenByConditionReply{
		UsersToken: data,
	}, nil
}

// ListByIDs batch get usersToken by ids
func (s *usersToken) ListByIDs(ctx context.Context, req *netbox_goV1.ListUsersTokenByIDsRequest) (*netbox_goV1.ListUsersTokenByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	usersTokenMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	usersTokens := []*netbox_goV1.UsersToken{}
	for _, id := range req.Ids {
		if v, ok := usersTokenMap[id]; ok {
			record, err := convertUsersToken(v)
			if err != nil {
				logger.Warn("convertUsersToken error", logger.Err(err), logger.Any("usersToken", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			usersTokens = append(usersTokens, record)
		}
	}

	return &netbox_goV1.ListUsersTokenByIDsReply{UsersTokens: usersTokens}, nil
}

// ListByLastID get a paginated list of usersTokens by last id
func (s *usersToken) ListByLastID(ctx context.Context, req *netbox_goV1.ListUsersTokenByLastIDRequest) (*netbox_goV1.ListUsersTokenByLastIDReply, error) {
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

	usersTokens := []*netbox_goV1.UsersToken{}
	for _, record := range records {
		data, err := convertUsersToken(record)
		if err != nil {
			logger.Warn("convertUsersToken error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		usersTokens = append(usersTokens, data)
	}

	return &netbox_goV1.ListUsersTokenByLastIDReply{
		UsersTokens: usersTokens,
	}, nil
}

func convertUsersToken(record *model.UsersToken) (*netbox_goV1.UsersToken, error) {
	value := &netbox_goV1.UsersToken{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
