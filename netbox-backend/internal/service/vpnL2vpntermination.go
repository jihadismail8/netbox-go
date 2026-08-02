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
		netbox_goV1.RegisterVpnL2VpnterminationServer(server, NewVpnL2VpnterminationServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VpnL2VpnterminationServer = (*vpnL2Vpntermination)(nil)
var _ time.Time

type vpnL2Vpntermination struct {
	netbox_goV1.UnimplementedVpnL2VpnterminationServer

	iDao dao.VpnL2VpnterminationDao
}

// NewVpnL2VpnterminationServer create a new service
func NewVpnL2VpnterminationServer() netbox_goV1.VpnL2VpnterminationServer {
	return &vpnL2Vpntermination{
		iDao: dao.NewVpnL2VpnterminationDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVpnL2VpnterminationCache(database.GetCacheType()),
		),
	}
}

// Create a new vpnL2Vpntermination
func (s *vpnL2Vpntermination) Create(ctx context.Context, req *netbox_goV1.CreateVpnL2VpnterminationRequest) (*netbox_goV1.CreateVpnL2VpnterminationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnL2Vpntermination{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVpnL2Vpntermination.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("vpnL2Vpntermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVpnL2VpnterminationReply{Id: record.ID}, nil
}

// DeleteByID delete a vpnL2Vpntermination by id
func (s *vpnL2Vpntermination) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnL2VpnterminationByIDRequest) (*netbox_goV1.DeleteVpnL2VpnterminationByIDReply, error) {
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

	return &netbox_goV1.DeleteVpnL2VpnterminationByIDReply{}, nil
}

// UpdateByID update a vpnL2Vpntermination by id
func (s *vpnL2Vpntermination) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnL2VpnterminationByIDRequest) (*netbox_goV1.UpdateVpnL2VpnterminationByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnL2Vpntermination{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVpnL2Vpntermination.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("vpnL2Vpntermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVpnL2VpnterminationByIDReply{}, nil
}

// GetByID get a vpnL2Vpntermination by id
func (s *vpnL2Vpntermination) GetByID(ctx context.Context, req *netbox_goV1.GetVpnL2VpnterminationByIDRequest) (*netbox_goV1.GetVpnL2VpnterminationByIDReply, error) {
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

	data, err := convertVpnL2Vpntermination(record)
	if err != nil {
		logger.Warn("convertVpnL2Vpntermination error", logger.Err(err), logger.Any("vpnL2Vpntermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVpnL2Vpntermination.Err()
	}

	return &netbox_goV1.GetVpnL2VpnterminationByIDReply{VpnL2Vpntermination: data}, nil
}

// List get a paginated list of vpnL2Vpnterminations by custom conditions
func (s *vpnL2Vpntermination) List(ctx context.Context, req *netbox_goV1.ListVpnL2VpnterminationRequest) (*netbox_goV1.ListVpnL2VpnterminationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVpnL2Vpntermination.Err()
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

	vpnL2Vpnterminations := []*netbox_goV1.VpnL2Vpntermination{}
	for _, record := range records {
		data, err := convertVpnL2Vpntermination(record)
		if err != nil {
			logger.Warn("convertVpnL2Vpntermination error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnL2Vpnterminations = append(vpnL2Vpnterminations, data)
	}

	return &netbox_goV1.ListVpnL2VpnterminationReply{
		Total:                total,
		VpnL2Vpnterminations: vpnL2Vpnterminations,
	}, nil
}

// DeleteByIDs batch delete vpnL2Vpntermination by ids
func (s *vpnL2Vpntermination) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnL2VpnterminationByIDsRequest) (*netbox_goV1.DeleteVpnL2VpnterminationByIDsReply, error) {
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

	return &netbox_goV1.DeleteVpnL2VpnterminationByIDsReply{}, nil
}

// GetByCondition get a vpnL2Vpntermination by custom condition
func (s *vpnL2Vpntermination) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnL2VpnterminationByConditionRequest) (*netbox_goV1.GetVpnL2VpnterminationByConditionReply, error) {
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

	data, err := convertVpnL2Vpntermination(record)
	if err != nil {
		logger.Warn("convertVpnL2Vpntermination error", logger.Err(err), logger.Any("vpnL2Vpntermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVpnL2Vpntermination.Err()
	}

	return &netbox_goV1.GetVpnL2VpnterminationByConditionReply{
		VpnL2Vpntermination: data,
	}, nil
}

// ListByIDs batch get vpnL2Vpntermination by ids
func (s *vpnL2Vpntermination) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnL2VpnterminationByIDsRequest) (*netbox_goV1.ListVpnL2VpnterminationByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	vpnL2VpnterminationMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	vpnL2Vpnterminations := []*netbox_goV1.VpnL2Vpntermination{}
	for _, id := range req.Ids {
		if v, ok := vpnL2VpnterminationMap[id]; ok {
			record, err := convertVpnL2Vpntermination(v)
			if err != nil {
				logger.Warn("convertVpnL2Vpntermination error", logger.Err(err), logger.Any("vpnL2Vpntermination", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			vpnL2Vpnterminations = append(vpnL2Vpnterminations, record)
		}
	}

	return &netbox_goV1.ListVpnL2VpnterminationByIDsReply{VpnL2Vpnterminations: vpnL2Vpnterminations}, nil
}

// ListByLastID get a paginated list of vpnL2Vpnterminations by last id
func (s *vpnL2Vpntermination) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnL2VpnterminationByLastIDRequest) (*netbox_goV1.ListVpnL2VpnterminationByLastIDReply, error) {
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

	vpnL2Vpnterminations := []*netbox_goV1.VpnL2Vpntermination{}
	for _, record := range records {
		data, err := convertVpnL2Vpntermination(record)
		if err != nil {
			logger.Warn("convertVpnL2Vpntermination error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnL2Vpnterminations = append(vpnL2Vpnterminations, data)
	}

	return &netbox_goV1.ListVpnL2VpnterminationByLastIDReply{
		VpnL2Vpnterminations: vpnL2Vpnterminations,
	}, nil
}

func convertVpnL2Vpntermination(record *model.VpnL2Vpntermination) (*netbox_goV1.VpnL2Vpntermination, error) {
	value := &netbox_goV1.VpnL2Vpntermination{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
