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
		netbox_goV1.RegisterVpnTunnelgroupServer(server, NewVpnTunnelgroupServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VpnTunnelgroupServer = (*vpnTunnelgroup)(nil)
var _ time.Time

type vpnTunnelgroup struct {
	netbox_goV1.UnimplementedVpnTunnelgroupServer

	iDao dao.VpnTunnelgroupDao
}

// NewVpnTunnelgroupServer create a new service
func NewVpnTunnelgroupServer() netbox_goV1.VpnTunnelgroupServer {
	return &vpnTunnelgroup{
		iDao: dao.NewVpnTunnelgroupDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVpnTunnelgroupCache(database.GetCacheType()),
		),
	}
}

// Create a new vpnTunnelgroup
func (s *vpnTunnelgroup) Create(ctx context.Context, req *netbox_goV1.CreateVpnTunnelgroupRequest) (*netbox_goV1.CreateVpnTunnelgroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnTunnelgroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVpnTunnelgroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("vpnTunnelgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVpnTunnelgroupReply{Id: record.ID}, nil
}

// DeleteByID delete a vpnTunnelgroup by id
func (s *vpnTunnelgroup) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnTunnelgroupByIDRequest) (*netbox_goV1.DeleteVpnTunnelgroupByIDReply, error) {
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

	return &netbox_goV1.DeleteVpnTunnelgroupByIDReply{}, nil
}

// UpdateByID update a vpnTunnelgroup by id
func (s *vpnTunnelgroup) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnTunnelgroupByIDRequest) (*netbox_goV1.UpdateVpnTunnelgroupByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnTunnelgroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVpnTunnelgroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("vpnTunnelgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVpnTunnelgroupByIDReply{}, nil
}

// GetByID get a vpnTunnelgroup by id
func (s *vpnTunnelgroup) GetByID(ctx context.Context, req *netbox_goV1.GetVpnTunnelgroupByIDRequest) (*netbox_goV1.GetVpnTunnelgroupByIDReply, error) {
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

	data, err := convertVpnTunnelgroup(record)
	if err != nil {
		logger.Warn("convertVpnTunnelgroup error", logger.Err(err), logger.Any("vpnTunnelgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVpnTunnelgroup.Err()
	}

	return &netbox_goV1.GetVpnTunnelgroupByIDReply{VpnTunnelgroup: data}, nil
}

// List get a paginated list of vpnTunnelgroups by custom conditions
func (s *vpnTunnelgroup) List(ctx context.Context, req *netbox_goV1.ListVpnTunnelgroupRequest) (*netbox_goV1.ListVpnTunnelgroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVpnTunnelgroup.Err()
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

	vpnTunnelgroups := []*netbox_goV1.VpnTunnelgroup{}
	for _, record := range records {
		data, err := convertVpnTunnelgroup(record)
		if err != nil {
			logger.Warn("convertVpnTunnelgroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnTunnelgroups = append(vpnTunnelgroups, data)
	}

	return &netbox_goV1.ListVpnTunnelgroupReply{
		Total:           total,
		VpnTunnelgroups: vpnTunnelgroups,
	}, nil
}

// DeleteByIDs batch delete vpnTunnelgroup by ids
func (s *vpnTunnelgroup) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnTunnelgroupByIDsRequest) (*netbox_goV1.DeleteVpnTunnelgroupByIDsReply, error) {
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

	return &netbox_goV1.DeleteVpnTunnelgroupByIDsReply{}, nil
}

// GetByCondition get a vpnTunnelgroup by custom condition
func (s *vpnTunnelgroup) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnTunnelgroupByConditionRequest) (*netbox_goV1.GetVpnTunnelgroupByConditionReply, error) {
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

	data, err := convertVpnTunnelgroup(record)
	if err != nil {
		logger.Warn("convertVpnTunnelgroup error", logger.Err(err), logger.Any("vpnTunnelgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVpnTunnelgroup.Err()
	}

	return &netbox_goV1.GetVpnTunnelgroupByConditionReply{
		VpnTunnelgroup: data,
	}, nil
}

// ListByIDs batch get vpnTunnelgroup by ids
func (s *vpnTunnelgroup) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnTunnelgroupByIDsRequest) (*netbox_goV1.ListVpnTunnelgroupByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	vpnTunnelgroupMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	vpnTunnelgroups := []*netbox_goV1.VpnTunnelgroup{}
	for _, id := range req.Ids {
		if v, ok := vpnTunnelgroupMap[id]; ok {
			record, err := convertVpnTunnelgroup(v)
			if err != nil {
				logger.Warn("convertVpnTunnelgroup error", logger.Err(err), logger.Any("vpnTunnelgroup", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			vpnTunnelgroups = append(vpnTunnelgroups, record)
		}
	}

	return &netbox_goV1.ListVpnTunnelgroupByIDsReply{VpnTunnelgroups: vpnTunnelgroups}, nil
}

// ListByLastID get a paginated list of vpnTunnelgroups by last id
func (s *vpnTunnelgroup) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnTunnelgroupByLastIDRequest) (*netbox_goV1.ListVpnTunnelgroupByLastIDReply, error) {
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

	vpnTunnelgroups := []*netbox_goV1.VpnTunnelgroup{}
	for _, record := range records {
		data, err := convertVpnTunnelgroup(record)
		if err != nil {
			logger.Warn("convertVpnTunnelgroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnTunnelgroups = append(vpnTunnelgroups, data)
	}

	return &netbox_goV1.ListVpnTunnelgroupByLastIDReply{
		VpnTunnelgroups: vpnTunnelgroups,
	}, nil
}

func convertVpnTunnelgroup(record *model.VpnTunnelgroup) (*netbox_goV1.VpnTunnelgroup, error) {
	value := &netbox_goV1.VpnTunnelgroup{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
