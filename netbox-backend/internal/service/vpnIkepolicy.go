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
		netbox_goV1.RegisterVpnIkepolicyServer(server, NewVpnIkepolicyServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VpnIkepolicyServer = (*vpnIkepolicy)(nil)
var _ time.Time

type vpnIkepolicy struct {
	netbox_goV1.UnimplementedVpnIkepolicyServer

	iDao dao.VpnIkepolicyDao
}

// NewVpnIkepolicyServer create a new service
func NewVpnIkepolicyServer() netbox_goV1.VpnIkepolicyServer {
	return &vpnIkepolicy{
		iDao: dao.NewVpnIkepolicyDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVpnIkepolicyCache(database.GetCacheType()),
		),
	}
}

// Create a new vpnIkepolicy
func (s *vpnIkepolicy) Create(ctx context.Context, req *netbox_goV1.CreateVpnIkepolicyRequest) (*netbox_goV1.CreateVpnIkepolicyReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnIkepolicy{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVpnIkepolicy.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("vpnIkepolicy", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVpnIkepolicyReply{Id: record.ID}, nil
}

// DeleteByID delete a vpnIkepolicy by id
func (s *vpnIkepolicy) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnIkepolicyByIDRequest) (*netbox_goV1.DeleteVpnIkepolicyByIDReply, error) {
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

	return &netbox_goV1.DeleteVpnIkepolicyByIDReply{}, nil
}

// UpdateByID update a vpnIkepolicy by id
func (s *vpnIkepolicy) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnIkepolicyByIDRequest) (*netbox_goV1.UpdateVpnIkepolicyByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnIkepolicy{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVpnIkepolicy.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("vpnIkepolicy", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVpnIkepolicyByIDReply{}, nil
}

// GetByID get a vpnIkepolicy by id
func (s *vpnIkepolicy) GetByID(ctx context.Context, req *netbox_goV1.GetVpnIkepolicyByIDRequest) (*netbox_goV1.GetVpnIkepolicyByIDReply, error) {
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

	data, err := convertVpnIkepolicy(record)
	if err != nil {
		logger.Warn("convertVpnIkepolicy error", logger.Err(err), logger.Any("vpnIkepolicy", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVpnIkepolicy.Err()
	}

	return &netbox_goV1.GetVpnIkepolicyByIDReply{VpnIkepolicy: data}, nil
}

// List get a paginated list of vpnIkepolicys by custom conditions
func (s *vpnIkepolicy) List(ctx context.Context, req *netbox_goV1.ListVpnIkepolicyRequest) (*netbox_goV1.ListVpnIkepolicyReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVpnIkepolicy.Err()
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

	vpnIkepolicys := []*netbox_goV1.VpnIkepolicy{}
	for _, record := range records {
		data, err := convertVpnIkepolicy(record)
		if err != nil {
			logger.Warn("convertVpnIkepolicy error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnIkepolicys = append(vpnIkepolicys, data)
	}

	return &netbox_goV1.ListVpnIkepolicyReply{
		Total:         total,
		VpnIkepolicys: vpnIkepolicys,
	}, nil
}

// DeleteByIDs batch delete vpnIkepolicy by ids
func (s *vpnIkepolicy) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnIkepolicyByIDsRequest) (*netbox_goV1.DeleteVpnIkepolicyByIDsReply, error) {
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

	return &netbox_goV1.DeleteVpnIkepolicyByIDsReply{}, nil
}

// GetByCondition get a vpnIkepolicy by custom condition
func (s *vpnIkepolicy) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnIkepolicyByConditionRequest) (*netbox_goV1.GetVpnIkepolicyByConditionReply, error) {
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

	data, err := convertVpnIkepolicy(record)
	if err != nil {
		logger.Warn("convertVpnIkepolicy error", logger.Err(err), logger.Any("vpnIkepolicy", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVpnIkepolicy.Err()
	}

	return &netbox_goV1.GetVpnIkepolicyByConditionReply{
		VpnIkepolicy: data,
	}, nil
}

// ListByIDs batch get vpnIkepolicy by ids
func (s *vpnIkepolicy) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnIkepolicyByIDsRequest) (*netbox_goV1.ListVpnIkepolicyByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	vpnIkepolicyMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	vpnIkepolicys := []*netbox_goV1.VpnIkepolicy{}
	for _, id := range req.Ids {
		if v, ok := vpnIkepolicyMap[id]; ok {
			record, err := convertVpnIkepolicy(v)
			if err != nil {
				logger.Warn("convertVpnIkepolicy error", logger.Err(err), logger.Any("vpnIkepolicy", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			vpnIkepolicys = append(vpnIkepolicys, record)
		}
	}

	return &netbox_goV1.ListVpnIkepolicyByIDsReply{VpnIkepolicys: vpnIkepolicys}, nil
}

// ListByLastID get a paginated list of vpnIkepolicys by last id
func (s *vpnIkepolicy) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnIkepolicyByLastIDRequest) (*netbox_goV1.ListVpnIkepolicyByLastIDReply, error) {
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

	vpnIkepolicys := []*netbox_goV1.VpnIkepolicy{}
	for _, record := range records {
		data, err := convertVpnIkepolicy(record)
		if err != nil {
			logger.Warn("convertVpnIkepolicy error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnIkepolicys = append(vpnIkepolicys, data)
	}

	return &netbox_goV1.ListVpnIkepolicyByLastIDReply{
		VpnIkepolicys: vpnIkepolicys,
	}, nil
}

func convertVpnIkepolicy(record *model.VpnIkepolicy) (*netbox_goV1.VpnIkepolicy, error) {
	value := &netbox_goV1.VpnIkepolicy{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
