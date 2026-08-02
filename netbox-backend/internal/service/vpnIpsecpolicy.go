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
		netbox_goV1.RegisterVpnIpsecpolicyServer(server, NewVpnIpsecpolicyServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VpnIpsecpolicyServer = (*vpnIpsecpolicy)(nil)
var _ time.Time

type vpnIpsecpolicy struct {
	netbox_goV1.UnimplementedVpnIpsecpolicyServer

	iDao dao.VpnIpsecpolicyDao
}

// NewVpnIpsecpolicyServer create a new service
func NewVpnIpsecpolicyServer() netbox_goV1.VpnIpsecpolicyServer {
	return &vpnIpsecpolicy{
		iDao: dao.NewVpnIpsecpolicyDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVpnIpsecpolicyCache(database.GetCacheType()),
		),
	}
}

// Create a new vpnIpsecpolicy
func (s *vpnIpsecpolicy) Create(ctx context.Context, req *netbox_goV1.CreateVpnIpsecpolicyRequest) (*netbox_goV1.CreateVpnIpsecpolicyReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnIpsecpolicy{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVpnIpsecpolicy.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("vpnIpsecpolicy", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVpnIpsecpolicyReply{Id: record.ID}, nil
}

// DeleteByID delete a vpnIpsecpolicy by id
func (s *vpnIpsecpolicy) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnIpsecpolicyByIDRequest) (*netbox_goV1.DeleteVpnIpsecpolicyByIDReply, error) {
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

	return &netbox_goV1.DeleteVpnIpsecpolicyByIDReply{}, nil
}

// UpdateByID update a vpnIpsecpolicy by id
func (s *vpnIpsecpolicy) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnIpsecpolicyByIDRequest) (*netbox_goV1.UpdateVpnIpsecpolicyByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnIpsecpolicy{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVpnIpsecpolicy.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("vpnIpsecpolicy", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVpnIpsecpolicyByIDReply{}, nil
}

// GetByID get a vpnIpsecpolicy by id
func (s *vpnIpsecpolicy) GetByID(ctx context.Context, req *netbox_goV1.GetVpnIpsecpolicyByIDRequest) (*netbox_goV1.GetVpnIpsecpolicyByIDReply, error) {
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

	data, err := convertVpnIpsecpolicy(record)
	if err != nil {
		logger.Warn("convertVpnIpsecpolicy error", logger.Err(err), logger.Any("vpnIpsecpolicy", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVpnIpsecpolicy.Err()
	}

	return &netbox_goV1.GetVpnIpsecpolicyByIDReply{VpnIpsecpolicy: data}, nil
}

// List get a paginated list of vpnIpsecpolicys by custom conditions
func (s *vpnIpsecpolicy) List(ctx context.Context, req *netbox_goV1.ListVpnIpsecpolicyRequest) (*netbox_goV1.ListVpnIpsecpolicyReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVpnIpsecpolicy.Err()
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

	vpnIpsecpolicys := []*netbox_goV1.VpnIpsecpolicy{}
	for _, record := range records {
		data, err := convertVpnIpsecpolicy(record)
		if err != nil {
			logger.Warn("convertVpnIpsecpolicy error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnIpsecpolicys = append(vpnIpsecpolicys, data)
	}

	return &netbox_goV1.ListVpnIpsecpolicyReply{
		Total:           total,
		VpnIpsecpolicys: vpnIpsecpolicys,
	}, nil
}

// DeleteByIDs batch delete vpnIpsecpolicy by ids
func (s *vpnIpsecpolicy) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnIpsecpolicyByIDsRequest) (*netbox_goV1.DeleteVpnIpsecpolicyByIDsReply, error) {
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

	return &netbox_goV1.DeleteVpnIpsecpolicyByIDsReply{}, nil
}

// GetByCondition get a vpnIpsecpolicy by custom condition
func (s *vpnIpsecpolicy) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnIpsecpolicyByConditionRequest) (*netbox_goV1.GetVpnIpsecpolicyByConditionReply, error) {
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

	data, err := convertVpnIpsecpolicy(record)
	if err != nil {
		logger.Warn("convertVpnIpsecpolicy error", logger.Err(err), logger.Any("vpnIpsecpolicy", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVpnIpsecpolicy.Err()
	}

	return &netbox_goV1.GetVpnIpsecpolicyByConditionReply{
		VpnIpsecpolicy: data,
	}, nil
}

// ListByIDs batch get vpnIpsecpolicy by ids
func (s *vpnIpsecpolicy) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnIpsecpolicyByIDsRequest) (*netbox_goV1.ListVpnIpsecpolicyByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	vpnIpsecpolicyMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	vpnIpsecpolicys := []*netbox_goV1.VpnIpsecpolicy{}
	for _, id := range req.Ids {
		if v, ok := vpnIpsecpolicyMap[id]; ok {
			record, err := convertVpnIpsecpolicy(v)
			if err != nil {
				logger.Warn("convertVpnIpsecpolicy error", logger.Err(err), logger.Any("vpnIpsecpolicy", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			vpnIpsecpolicys = append(vpnIpsecpolicys, record)
		}
	}

	return &netbox_goV1.ListVpnIpsecpolicyByIDsReply{VpnIpsecpolicys: vpnIpsecpolicys}, nil
}

// ListByLastID get a paginated list of vpnIpsecpolicys by last id
func (s *vpnIpsecpolicy) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnIpsecpolicyByLastIDRequest) (*netbox_goV1.ListVpnIpsecpolicyByLastIDReply, error) {
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

	vpnIpsecpolicys := []*netbox_goV1.VpnIpsecpolicy{}
	for _, record := range records {
		data, err := convertVpnIpsecpolicy(record)
		if err != nil {
			logger.Warn("convertVpnIpsecpolicy error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnIpsecpolicys = append(vpnIpsecpolicys, data)
	}

	return &netbox_goV1.ListVpnIpsecpolicyByLastIDReply{
		VpnIpsecpolicys: vpnIpsecpolicys,
	}, nil
}

func convertVpnIpsecpolicy(record *model.VpnIpsecpolicy) (*netbox_goV1.VpnIpsecpolicy, error) {
	value := &netbox_goV1.VpnIpsecpolicy{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
