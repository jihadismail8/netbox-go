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
		netbox_goV1.RegisterUsersUserconfigServer(server, NewUsersUserconfigServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.UsersUserconfigServer = (*usersUserconfig)(nil)
var _ time.Time

type usersUserconfig struct {
	netbox_goV1.UnimplementedUsersUserconfigServer

	iDao dao.UsersUserconfigDao
}

// NewUsersUserconfigServer create a new service
func NewUsersUserconfigServer() netbox_goV1.UsersUserconfigServer {
	return &usersUserconfig{
		iDao: dao.NewUsersUserconfigDao(
			database.GetDB(), // db driver is postgresql
			cache.NewUsersUserconfigCache(database.GetCacheType()),
		),
	}
}

// Create a new usersUserconfig
func (s *usersUserconfig) Create(ctx context.Context, req *netbox_goV1.CreateUsersUserconfigRequest) (*netbox_goV1.CreateUsersUserconfigReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.UsersUserconfig{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateUsersUserconfig.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("usersUserconfig", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateUsersUserconfigReply{Id: record.ID}, nil
}

// DeleteByID delete a usersUserconfig by id
func (s *usersUserconfig) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteUsersUserconfigByIDRequest) (*netbox_goV1.DeleteUsersUserconfigByIDReply, error) {
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

	return &netbox_goV1.DeleteUsersUserconfigByIDReply{}, nil
}

// UpdateByID update a usersUserconfig by id
func (s *usersUserconfig) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateUsersUserconfigByIDRequest) (*netbox_goV1.UpdateUsersUserconfigByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.UsersUserconfig{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDUsersUserconfig.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("usersUserconfig", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateUsersUserconfigByIDReply{}, nil
}

// GetByID get a usersUserconfig by id
func (s *usersUserconfig) GetByID(ctx context.Context, req *netbox_goV1.GetUsersUserconfigByIDRequest) (*netbox_goV1.GetUsersUserconfigByIDReply, error) {
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

	data, err := convertUsersUserconfig(record)
	if err != nil {
		logger.Warn("convertUsersUserconfig error", logger.Err(err), logger.Any("usersUserconfig", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDUsersUserconfig.Err()
	}

	return &netbox_goV1.GetUsersUserconfigByIDReply{UsersUserconfig: data}, nil
}

// List get a paginated list of usersUserconfigs by custom conditions
func (s *usersUserconfig) List(ctx context.Context, req *netbox_goV1.ListUsersUserconfigRequest) (*netbox_goV1.ListUsersUserconfigReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListUsersUserconfig.Err()
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

	usersUserconfigs := []*netbox_goV1.UsersUserconfig{}
	for _, record := range records {
		data, err := convertUsersUserconfig(record)
		if err != nil {
			logger.Warn("convertUsersUserconfig error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		usersUserconfigs = append(usersUserconfigs, data)
	}

	return &netbox_goV1.ListUsersUserconfigReply{
		Total:            total,
		UsersUserconfigs: usersUserconfigs,
	}, nil
}

// DeleteByIDs batch delete usersUserconfig by ids
func (s *usersUserconfig) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteUsersUserconfigByIDsRequest) (*netbox_goV1.DeleteUsersUserconfigByIDsReply, error) {
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

	return &netbox_goV1.DeleteUsersUserconfigByIDsReply{}, nil
}

// GetByCondition get a usersUserconfig by custom condition
func (s *usersUserconfig) GetByCondition(ctx context.Context, req *netbox_goV1.GetUsersUserconfigByConditionRequest) (*netbox_goV1.GetUsersUserconfigByConditionReply, error) {
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

	data, err := convertUsersUserconfig(record)
	if err != nil {
		logger.Warn("convertUsersUserconfig error", logger.Err(err), logger.Any("usersUserconfig", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionUsersUserconfig.Err()
	}

	return &netbox_goV1.GetUsersUserconfigByConditionReply{
		UsersUserconfig: data,
	}, nil
}

// ListByIDs batch get usersUserconfig by ids
func (s *usersUserconfig) ListByIDs(ctx context.Context, req *netbox_goV1.ListUsersUserconfigByIDsRequest) (*netbox_goV1.ListUsersUserconfigByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	usersUserconfigMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	usersUserconfigs := []*netbox_goV1.UsersUserconfig{}
	for _, id := range req.Ids {
		if v, ok := usersUserconfigMap[id]; ok {
			record, err := convertUsersUserconfig(v)
			if err != nil {
				logger.Warn("convertUsersUserconfig error", logger.Err(err), logger.Any("usersUserconfig", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			usersUserconfigs = append(usersUserconfigs, record)
		}
	}

	return &netbox_goV1.ListUsersUserconfigByIDsReply{UsersUserconfigs: usersUserconfigs}, nil
}

// ListByLastID get a paginated list of usersUserconfigs by last id
func (s *usersUserconfig) ListByLastID(ctx context.Context, req *netbox_goV1.ListUsersUserconfigByLastIDRequest) (*netbox_goV1.ListUsersUserconfigByLastIDReply, error) {
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

	usersUserconfigs := []*netbox_goV1.UsersUserconfig{}
	for _, record := range records {
		data, err := convertUsersUserconfig(record)
		if err != nil {
			logger.Warn("convertUsersUserconfig error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		usersUserconfigs = append(usersUserconfigs, data)
	}

	return &netbox_goV1.ListUsersUserconfigByLastIDReply{
		UsersUserconfigs: usersUserconfigs,
	}, nil
}

func convertUsersUserconfig(record *model.UsersUserconfig) (*netbox_goV1.UsersUserconfig, error) {
	value := &netbox_goV1.UsersUserconfig{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
