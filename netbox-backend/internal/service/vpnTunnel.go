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
		netbox_goV1.RegisterVpnTunnelServer(server, NewVpnTunnelServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VpnTunnelServer = (*vpnTunnel)(nil)
var _ time.Time

type vpnTunnel struct {
	netbox_goV1.UnimplementedVpnTunnelServer

	iDao dao.VpnTunnelDao
}

// NewVpnTunnelServer create a new service
func NewVpnTunnelServer() netbox_goV1.VpnTunnelServer {
	return &vpnTunnel{
		iDao: dao.NewVpnTunnelDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVpnTunnelCache(database.GetCacheType()),
		),
	}
}

// Create a new vpnTunnel
func (s *vpnTunnel) Create(ctx context.Context, req *netbox_goV1.CreateVpnTunnelRequest) (*netbox_goV1.CreateVpnTunnelReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnTunnel{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVpnTunnel.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("vpnTunnel", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVpnTunnelReply{Id: record.ID}, nil
}

// DeleteByID delete a vpnTunnel by id
func (s *vpnTunnel) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnTunnelByIDRequest) (*netbox_goV1.DeleteVpnTunnelByIDReply, error) {
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

	return &netbox_goV1.DeleteVpnTunnelByIDReply{}, nil
}

// UpdateByID update a vpnTunnel by id
func (s *vpnTunnel) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnTunnelByIDRequest) (*netbox_goV1.UpdateVpnTunnelByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnTunnel{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVpnTunnel.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("vpnTunnel", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVpnTunnelByIDReply{}, nil
}

// GetByID get a vpnTunnel by id
func (s *vpnTunnel) GetByID(ctx context.Context, req *netbox_goV1.GetVpnTunnelByIDRequest) (*netbox_goV1.GetVpnTunnelByIDReply, error) {
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

	data, err := convertVpnTunnel(record)
	if err != nil {
		logger.Warn("convertVpnTunnel error", logger.Err(err), logger.Any("vpnTunnel", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVpnTunnel.Err()
	}

	return &netbox_goV1.GetVpnTunnelByIDReply{VpnTunnel: data}, nil
}

// List get a paginated list of vpnTunnels by custom conditions
func (s *vpnTunnel) List(ctx context.Context, req *netbox_goV1.ListVpnTunnelRequest) (*netbox_goV1.ListVpnTunnelReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVpnTunnel.Err()
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

	vpnTunnels := []*netbox_goV1.VpnTunnel{}
	for _, record := range records {
		data, err := convertVpnTunnel(record)
		if err != nil {
			logger.Warn("convertVpnTunnel error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnTunnels = append(vpnTunnels, data)
	}

	return &netbox_goV1.ListVpnTunnelReply{
		Total:      total,
		VpnTunnels: vpnTunnels,
	}, nil
}

// DeleteByIDs batch delete vpnTunnel by ids
func (s *vpnTunnel) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnTunnelByIDsRequest) (*netbox_goV1.DeleteVpnTunnelByIDsReply, error) {
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

	return &netbox_goV1.DeleteVpnTunnelByIDsReply{}, nil
}

// GetByCondition get a vpnTunnel by custom condition
func (s *vpnTunnel) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnTunnelByConditionRequest) (*netbox_goV1.GetVpnTunnelByConditionReply, error) {
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

	data, err := convertVpnTunnel(record)
	if err != nil {
		logger.Warn("convertVpnTunnel error", logger.Err(err), logger.Any("vpnTunnel", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVpnTunnel.Err()
	}

	return &netbox_goV1.GetVpnTunnelByConditionReply{
		VpnTunnel: data,
	}, nil
}

// ListByIDs batch get vpnTunnel by ids
func (s *vpnTunnel) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnTunnelByIDsRequest) (*netbox_goV1.ListVpnTunnelByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	vpnTunnelMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	vpnTunnels := []*netbox_goV1.VpnTunnel{}
	for _, id := range req.Ids {
		if v, ok := vpnTunnelMap[id]; ok {
			record, err := convertVpnTunnel(v)
			if err != nil {
				logger.Warn("convertVpnTunnel error", logger.Err(err), logger.Any("vpnTunnel", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			vpnTunnels = append(vpnTunnels, record)
		}
	}

	return &netbox_goV1.ListVpnTunnelByIDsReply{VpnTunnels: vpnTunnels}, nil
}

// ListByLastID get a paginated list of vpnTunnels by last id
func (s *vpnTunnel) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnTunnelByLastIDRequest) (*netbox_goV1.ListVpnTunnelByLastIDReply, error) {
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

	vpnTunnels := []*netbox_goV1.VpnTunnel{}
	for _, record := range records {
		data, err := convertVpnTunnel(record)
		if err != nil {
			logger.Warn("convertVpnTunnel error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnTunnels = append(vpnTunnels, data)
	}

	return &netbox_goV1.ListVpnTunnelByLastIDReply{
		VpnTunnels: vpnTunnels,
	}, nil
}

func convertVpnTunnel(record *model.VpnTunnel) (*netbox_goV1.VpnTunnel, error) {
	value := &netbox_goV1.VpnTunnel{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
