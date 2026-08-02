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
		netbox_goV1.RegisterIpamRoleServer(server, NewIpamRoleServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.IpamRoleServer = (*ipamRole)(nil)
var _ time.Time

type ipamRole struct {
	netbox_goV1.UnimplementedIpamRoleServer

	iDao dao.IpamRoleDao
}

// NewIpamRoleServer create a new service
func NewIpamRoleServer() netbox_goV1.IpamRoleServer {
	return &ipamRole{
		iDao: dao.NewIpamRoleDao(
			database.GetDB(), // db driver is postgresql
			cache.NewIpamRoleCache(database.GetCacheType()),
		),
	}
}

// Create a new ipamRole
func (s *ipamRole) Create(ctx context.Context, req *netbox_goV1.CreateIpamRoleRequest) (*netbox_goV1.CreateIpamRoleReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamRole{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateIpamRole.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("ipamRole", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateIpamRoleReply{Id: record.ID}, nil
}

// DeleteByID delete a ipamRole by id
func (s *ipamRole) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamRoleByIDRequest) (*netbox_goV1.DeleteIpamRoleByIDReply, error) {
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

	return &netbox_goV1.DeleteIpamRoleByIDReply{}, nil
}

// UpdateByID update a ipamRole by id
func (s *ipamRole) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamRoleByIDRequest) (*netbox_goV1.UpdateIpamRoleByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamRole{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDIpamRole.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("ipamRole", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateIpamRoleByIDReply{}, nil
}

// GetByID get a ipamRole by id
func (s *ipamRole) GetByID(ctx context.Context, req *netbox_goV1.GetIpamRoleByIDRequest) (*netbox_goV1.GetIpamRoleByIDReply, error) {
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

	data, err := convertIpamRole(record)
	if err != nil {
		logger.Warn("convertIpamRole error", logger.Err(err), logger.Any("ipamRole", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDIpamRole.Err()
	}

	return &netbox_goV1.GetIpamRoleByIDReply{IpamRole: data}, nil
}

// List get a paginated list of ipamRoles by custom conditions
func (s *ipamRole) List(ctx context.Context, req *netbox_goV1.ListIpamRoleRequest) (*netbox_goV1.ListIpamRoleReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListIpamRole.Err()
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

	ipamRoles := []*netbox_goV1.IpamRole{}
	for _, record := range records {
		data, err := convertIpamRole(record)
		if err != nil {
			logger.Warn("convertIpamRole error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamRoles = append(ipamRoles, data)
	}

	return &netbox_goV1.ListIpamRoleReply{
		Total:     total,
		IpamRoles: ipamRoles,
	}, nil
}

// DeleteByIDs batch delete ipamRole by ids
func (s *ipamRole) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamRoleByIDsRequest) (*netbox_goV1.DeleteIpamRoleByIDsReply, error) {
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

	return &netbox_goV1.DeleteIpamRoleByIDsReply{}, nil
}

// GetByCondition get a ipamRole by custom condition
func (s *ipamRole) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamRoleByConditionRequest) (*netbox_goV1.GetIpamRoleByConditionReply, error) {
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

	data, err := convertIpamRole(record)
	if err != nil {
		logger.Warn("convertIpamRole error", logger.Err(err), logger.Any("ipamRole", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionIpamRole.Err()
	}

	return &netbox_goV1.GetIpamRoleByConditionReply{
		IpamRole: data,
	}, nil
}

// ListByIDs batch get ipamRole by ids
func (s *ipamRole) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamRoleByIDsRequest) (*netbox_goV1.ListIpamRoleByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ipamRoleMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	ipamRoles := []*netbox_goV1.IpamRole{}
	for _, id := range req.Ids {
		if v, ok := ipamRoleMap[id]; ok {
			record, err := convertIpamRole(v)
			if err != nil {
				logger.Warn("convertIpamRole error", logger.Err(err), logger.Any("ipamRole", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			ipamRoles = append(ipamRoles, record)
		}
	}

	return &netbox_goV1.ListIpamRoleByIDsReply{IpamRoles: ipamRoles}, nil
}

// ListByLastID get a paginated list of ipamRoles by last id
func (s *ipamRole) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamRoleByLastIDRequest) (*netbox_goV1.ListIpamRoleByLastIDReply, error) {
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

	ipamRoles := []*netbox_goV1.IpamRole{}
	for _, record := range records {
		data, err := convertIpamRole(record)
		if err != nil {
			logger.Warn("convertIpamRole error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamRoles = append(ipamRoles, data)
	}

	return &netbox_goV1.ListIpamRoleByLastIDReply{
		IpamRoles: ipamRoles,
	}, nil
}

func convertIpamRole(record *model.IpamRole) (*netbox_goV1.IpamRole, error) {
	value := &netbox_goV1.IpamRole{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
