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
		netbox_goV1.RegisterVpnL2VpnServer(server, NewVpnL2VpnServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VpnL2VpnServer = (*vpnL2Vpn)(nil)
var _ time.Time

type vpnL2Vpn struct {
	netbox_goV1.UnimplementedVpnL2VpnServer

	iDao dao.VpnL2VpnDao
}

// NewVpnL2VpnServer create a new service
func NewVpnL2VpnServer() netbox_goV1.VpnL2VpnServer {
	return &vpnL2Vpn{
		iDao: dao.NewVpnL2VpnDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVpnL2VpnCache(database.GetCacheType()),
		),
	}
}

// Create a new vpnL2Vpn
func (s *vpnL2Vpn) Create(ctx context.Context, req *netbox_goV1.CreateVpnL2VpnRequest) (*netbox_goV1.CreateVpnL2VpnReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnL2Vpn{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVpnL2Vpn.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("vpnL2Vpn", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVpnL2VpnReply{Id: record.ID}, nil
}

// DeleteByID delete a vpnL2Vpn by id
func (s *vpnL2Vpn) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnL2VpnByIDRequest) (*netbox_goV1.DeleteVpnL2VpnByIDReply, error) {
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

	return &netbox_goV1.DeleteVpnL2VpnByIDReply{}, nil
}

// UpdateByID update a vpnL2Vpn by id
func (s *vpnL2Vpn) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnL2VpnByIDRequest) (*netbox_goV1.UpdateVpnL2VpnByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnL2Vpn{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVpnL2Vpn.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("vpnL2Vpn", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVpnL2VpnByIDReply{}, nil
}

// GetByID get a vpnL2Vpn by id
func (s *vpnL2Vpn) GetByID(ctx context.Context, req *netbox_goV1.GetVpnL2VpnByIDRequest) (*netbox_goV1.GetVpnL2VpnByIDReply, error) {
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

	data, err := convertVpnL2Vpn(record)
	if err != nil {
		logger.Warn("convertVpnL2Vpn error", logger.Err(err), logger.Any("vpnL2Vpn", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVpnL2Vpn.Err()
	}

	return &netbox_goV1.GetVpnL2VpnByIDReply{VpnL2Vpn: data}, nil
}

// List get a paginated list of vpnL2Vpns by custom conditions
func (s *vpnL2Vpn) List(ctx context.Context, req *netbox_goV1.ListVpnL2VpnRequest) (*netbox_goV1.ListVpnL2VpnReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVpnL2Vpn.Err()
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

	vpnL2Vpns := []*netbox_goV1.VpnL2Vpn{}
	for _, record := range records {
		data, err := convertVpnL2Vpn(record)
		if err != nil {
			logger.Warn("convertVpnL2Vpn error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnL2Vpns = append(vpnL2Vpns, data)
	}

	return &netbox_goV1.ListVpnL2VpnReply{
		Total:     total,
		VpnL2Vpns: vpnL2Vpns,
	}, nil
}

// DeleteByIDs batch delete vpnL2Vpn by ids
func (s *vpnL2Vpn) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnL2VpnByIDsRequest) (*netbox_goV1.DeleteVpnL2VpnByIDsReply, error) {
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

	return &netbox_goV1.DeleteVpnL2VpnByIDsReply{}, nil
}

// GetByCondition get a vpnL2Vpn by custom condition
func (s *vpnL2Vpn) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnL2VpnByConditionRequest) (*netbox_goV1.GetVpnL2VpnByConditionReply, error) {
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

	data, err := convertVpnL2Vpn(record)
	if err != nil {
		logger.Warn("convertVpnL2Vpn error", logger.Err(err), logger.Any("vpnL2Vpn", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVpnL2Vpn.Err()
	}

	return &netbox_goV1.GetVpnL2VpnByConditionReply{
		VpnL2Vpn: data,
	}, nil
}

// ListByIDs batch get vpnL2Vpn by ids
func (s *vpnL2Vpn) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnL2VpnByIDsRequest) (*netbox_goV1.ListVpnL2VpnByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	vpnL2VpnMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	vpnL2Vpns := []*netbox_goV1.VpnL2Vpn{}
	for _, id := range req.Ids {
		if v, ok := vpnL2VpnMap[id]; ok {
			record, err := convertVpnL2Vpn(v)
			if err != nil {
				logger.Warn("convertVpnL2Vpn error", logger.Err(err), logger.Any("vpnL2Vpn", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			vpnL2Vpns = append(vpnL2Vpns, record)
		}
	}

	return &netbox_goV1.ListVpnL2VpnByIDsReply{VpnL2Vpns: vpnL2Vpns}, nil
}

// ListByLastID get a paginated list of vpnL2Vpns by last id
func (s *vpnL2Vpn) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnL2VpnByLastIDRequest) (*netbox_goV1.ListVpnL2VpnByLastIDReply, error) {
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

	vpnL2Vpns := []*netbox_goV1.VpnL2Vpn{}
	for _, record := range records {
		data, err := convertVpnL2Vpn(record)
		if err != nil {
			logger.Warn("convertVpnL2Vpn error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnL2Vpns = append(vpnL2Vpns, data)
	}

	return &netbox_goV1.ListVpnL2VpnByLastIDReply{
		VpnL2Vpns: vpnL2Vpns,
	}, nil
}

func convertVpnL2Vpn(record *model.VpnL2Vpn) (*netbox_goV1.VpnL2Vpn, error) {
	value := &netbox_goV1.VpnL2Vpn{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
