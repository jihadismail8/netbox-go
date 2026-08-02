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
		netbox_goV1.RegisterUsersGroupServer(server, NewUsersGroupServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.UsersGroupServer = (*usersGroup)(nil)
var _ time.Time

type usersGroup struct {
	netbox_goV1.UnimplementedUsersGroupServer

	iDao dao.UsersGroupDao
}

// NewUsersGroupServer create a new service
func NewUsersGroupServer() netbox_goV1.UsersGroupServer {
	return &usersGroup{
		iDao: dao.NewUsersGroupDao(
			database.GetDB(), // db driver is postgresql
			cache.NewUsersGroupCache(database.GetCacheType()),
		),
	}
}

// Create a new usersGroup
func (s *usersGroup) Create(ctx context.Context, req *netbox_goV1.CreateUsersGroupRequest) (*netbox_goV1.CreateUsersGroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.UsersGroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateUsersGroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("usersGroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateUsersGroupReply{Id: record.ID}, nil
}

// DeleteByID delete a usersGroup by id
func (s *usersGroup) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteUsersGroupByIDRequest) (*netbox_goV1.DeleteUsersGroupByIDReply, error) {
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

	return &netbox_goV1.DeleteUsersGroupByIDReply{}, nil
}

// UpdateByID update a usersGroup by id
func (s *usersGroup) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateUsersGroupByIDRequest) (*netbox_goV1.UpdateUsersGroupByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.UsersGroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDUsersGroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("usersGroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateUsersGroupByIDReply{}, nil
}

// GetByID get a usersGroup by id
func (s *usersGroup) GetByID(ctx context.Context, req *netbox_goV1.GetUsersGroupByIDRequest) (*netbox_goV1.GetUsersGroupByIDReply, error) {
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

	data, err := convertUsersGroup(record)
	if err != nil {
		logger.Warn("convertUsersGroup error", logger.Err(err), logger.Any("usersGroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDUsersGroup.Err()
	}

	return &netbox_goV1.GetUsersGroupByIDReply{UsersGroup: data}, nil
}

// List get a paginated list of usersGroups by custom conditions
func (s *usersGroup) List(ctx context.Context, req *netbox_goV1.ListUsersGroupRequest) (*netbox_goV1.ListUsersGroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListUsersGroup.Err()
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

	usersGroups := []*netbox_goV1.UsersGroup{}
	for _, record := range records {
		data, err := convertUsersGroup(record)
		if err != nil {
			logger.Warn("convertUsersGroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		usersGroups = append(usersGroups, data)
	}

	return &netbox_goV1.ListUsersGroupReply{
		Total:       total,
		UsersGroups: usersGroups,
	}, nil
}

// DeleteByIDs batch delete usersGroup by ids
func (s *usersGroup) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteUsersGroupByIDsRequest) (*netbox_goV1.DeleteUsersGroupByIDsReply, error) {
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

	return &netbox_goV1.DeleteUsersGroupByIDsReply{}, nil
}

// GetByCondition get a usersGroup by custom condition
func (s *usersGroup) GetByCondition(ctx context.Context, req *netbox_goV1.GetUsersGroupByConditionRequest) (*netbox_goV1.GetUsersGroupByConditionReply, error) {
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

	data, err := convertUsersGroup(record)
	if err != nil {
		logger.Warn("convertUsersGroup error", logger.Err(err), logger.Any("usersGroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionUsersGroup.Err()
	}

	return &netbox_goV1.GetUsersGroupByConditionReply{
		UsersGroup: data,
	}, nil
}

// ListByIDs batch get usersGroup by ids
func (s *usersGroup) ListByIDs(ctx context.Context, req *netbox_goV1.ListUsersGroupByIDsRequest) (*netbox_goV1.ListUsersGroupByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	usersGroupMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	usersGroups := []*netbox_goV1.UsersGroup{}
	for _, id := range req.Ids {
		if v, ok := usersGroupMap[id]; ok {
			record, err := convertUsersGroup(v)
			if err != nil {
				logger.Warn("convertUsersGroup error", logger.Err(err), logger.Any("usersGroup", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			usersGroups = append(usersGroups, record)
		}
	}

	return &netbox_goV1.ListUsersGroupByIDsReply{UsersGroups: usersGroups}, nil
}

// ListByLastID get a paginated list of usersGroups by last id
func (s *usersGroup) ListByLastID(ctx context.Context, req *netbox_goV1.ListUsersGroupByLastIDRequest) (*netbox_goV1.ListUsersGroupByLastIDReply, error) {
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

	usersGroups := []*netbox_goV1.UsersGroup{}
	for _, record := range records {
		data, err := convertUsersGroup(record)
		if err != nil {
			logger.Warn("convertUsersGroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		usersGroups = append(usersGroups, data)
	}

	return &netbox_goV1.ListUsersGroupByLastIDReply{
		UsersGroups: usersGroups,
	}, nil
}

func convertUsersGroup(record *model.UsersGroup) (*netbox_goV1.UsersGroup, error) {
	value := &netbox_goV1.UsersGroup{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
