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
		netbox_goV1.RegisterVpnTunnelterminationServer(server, NewVpnTunnelterminationServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VpnTunnelterminationServer = (*vpnTunneltermination)(nil)
var _ time.Time

type vpnTunneltermination struct {
	netbox_goV1.UnimplementedVpnTunnelterminationServer

	iDao dao.VpnTunnelterminationDao
}

// NewVpnTunnelterminationServer create a new service
func NewVpnTunnelterminationServer() netbox_goV1.VpnTunnelterminationServer {
	return &vpnTunneltermination{
		iDao: dao.NewVpnTunnelterminationDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVpnTunnelterminationCache(database.GetCacheType()),
		),
	}
}

// Create a new vpnTunneltermination
func (s *vpnTunneltermination) Create(ctx context.Context, req *netbox_goV1.CreateVpnTunnelterminationRequest) (*netbox_goV1.CreateVpnTunnelterminationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnTunneltermination{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVpnTunneltermination.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("vpnTunneltermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVpnTunnelterminationReply{Id: record.ID}, nil
}

// DeleteByID delete a vpnTunneltermination by id
func (s *vpnTunneltermination) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnTunnelterminationByIDRequest) (*netbox_goV1.DeleteVpnTunnelterminationByIDReply, error) {
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

	return &netbox_goV1.DeleteVpnTunnelterminationByIDReply{}, nil
}

// UpdateByID update a vpnTunneltermination by id
func (s *vpnTunneltermination) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnTunnelterminationByIDRequest) (*netbox_goV1.UpdateVpnTunnelterminationByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnTunneltermination{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVpnTunneltermination.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("vpnTunneltermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVpnTunnelterminationByIDReply{}, nil
}

// GetByID get a vpnTunneltermination by id
func (s *vpnTunneltermination) GetByID(ctx context.Context, req *netbox_goV1.GetVpnTunnelterminationByIDRequest) (*netbox_goV1.GetVpnTunnelterminationByIDReply, error) {
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

	data, err := convertVpnTunneltermination(record)
	if err != nil {
		logger.Warn("convertVpnTunneltermination error", logger.Err(err), logger.Any("vpnTunneltermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVpnTunneltermination.Err()
	}

	return &netbox_goV1.GetVpnTunnelterminationByIDReply{VpnTunneltermination: data}, nil
}

// List get a paginated list of vpnTunnelterminations by custom conditions
func (s *vpnTunneltermination) List(ctx context.Context, req *netbox_goV1.ListVpnTunnelterminationRequest) (*netbox_goV1.ListVpnTunnelterminationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVpnTunneltermination.Err()
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

	vpnTunnelterminations := []*netbox_goV1.VpnTunneltermination{}
	for _, record := range records {
		data, err := convertVpnTunneltermination(record)
		if err != nil {
			logger.Warn("convertVpnTunneltermination error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnTunnelterminations = append(vpnTunnelterminations, data)
	}

	return &netbox_goV1.ListVpnTunnelterminationReply{
		Total:                 total,
		VpnTunnelterminations: vpnTunnelterminations,
	}, nil
}

// DeleteByIDs batch delete vpnTunneltermination by ids
func (s *vpnTunneltermination) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnTunnelterminationByIDsRequest) (*netbox_goV1.DeleteVpnTunnelterminationByIDsReply, error) {
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

	return &netbox_goV1.DeleteVpnTunnelterminationByIDsReply{}, nil
}

// GetByCondition get a vpnTunneltermination by custom condition
func (s *vpnTunneltermination) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnTunnelterminationByConditionRequest) (*netbox_goV1.GetVpnTunnelterminationByConditionReply, error) {
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

	data, err := convertVpnTunneltermination(record)
	if err != nil {
		logger.Warn("convertVpnTunneltermination error", logger.Err(err), logger.Any("vpnTunneltermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVpnTunneltermination.Err()
	}

	return &netbox_goV1.GetVpnTunnelterminationByConditionReply{
		VpnTunneltermination: data,
	}, nil
}

// ListByIDs batch get vpnTunneltermination by ids
func (s *vpnTunneltermination) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnTunnelterminationByIDsRequest) (*netbox_goV1.ListVpnTunnelterminationByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	vpnTunnelterminationMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	vpnTunnelterminations := []*netbox_goV1.VpnTunneltermination{}
	for _, id := range req.Ids {
		if v, ok := vpnTunnelterminationMap[id]; ok {
			record, err := convertVpnTunneltermination(v)
			if err != nil {
				logger.Warn("convertVpnTunneltermination error", logger.Err(err), logger.Any("vpnTunneltermination", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			vpnTunnelterminations = append(vpnTunnelterminations, record)
		}
	}

	return &netbox_goV1.ListVpnTunnelterminationByIDsReply{VpnTunnelterminations: vpnTunnelterminations}, nil
}

// ListByLastID get a paginated list of vpnTunnelterminations by last id
func (s *vpnTunneltermination) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnTunnelterminationByLastIDRequest) (*netbox_goV1.ListVpnTunnelterminationByLastIDReply, error) {
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

	vpnTunnelterminations := []*netbox_goV1.VpnTunneltermination{}
	for _, record := range records {
		data, err := convertVpnTunneltermination(record)
		if err != nil {
			logger.Warn("convertVpnTunneltermination error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnTunnelterminations = append(vpnTunnelterminations, data)
	}

	return &netbox_goV1.ListVpnTunnelterminationByLastIDReply{
		VpnTunnelterminations: vpnTunnelterminations,
	}, nil
}

func convertVpnTunneltermination(record *model.VpnTunneltermination) (*netbox_goV1.VpnTunneltermination, error) {
	value := &netbox_goV1.VpnTunneltermination{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
