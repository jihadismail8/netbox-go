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
		netbox_goV1.RegisterVpnIpsecproposalServer(server, NewVpnIpsecproposalServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VpnIpsecproposalServer = (*vpnIpsecproposal)(nil)
var _ time.Time

type vpnIpsecproposal struct {
	netbox_goV1.UnimplementedVpnIpsecproposalServer

	iDao dao.VpnIpsecproposalDao
}

// NewVpnIpsecproposalServer create a new service
func NewVpnIpsecproposalServer() netbox_goV1.VpnIpsecproposalServer {
	return &vpnIpsecproposal{
		iDao: dao.NewVpnIpsecproposalDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVpnIpsecproposalCache(database.GetCacheType()),
		),
	}
}

// Create a new vpnIpsecproposal
func (s *vpnIpsecproposal) Create(ctx context.Context, req *netbox_goV1.CreateVpnIpsecproposalRequest) (*netbox_goV1.CreateVpnIpsecproposalReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnIpsecproposal{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVpnIpsecproposal.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("vpnIpsecproposal", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVpnIpsecproposalReply{Id: record.ID}, nil
}

// DeleteByID delete a vpnIpsecproposal by id
func (s *vpnIpsecproposal) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnIpsecproposalByIDRequest) (*netbox_goV1.DeleteVpnIpsecproposalByIDReply, error) {
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

	return &netbox_goV1.DeleteVpnIpsecproposalByIDReply{}, nil
}

// UpdateByID update a vpnIpsecproposal by id
func (s *vpnIpsecproposal) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnIpsecproposalByIDRequest) (*netbox_goV1.UpdateVpnIpsecproposalByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnIpsecproposal{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVpnIpsecproposal.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("vpnIpsecproposal", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVpnIpsecproposalByIDReply{}, nil
}

// GetByID get a vpnIpsecproposal by id
func (s *vpnIpsecproposal) GetByID(ctx context.Context, req *netbox_goV1.GetVpnIpsecproposalByIDRequest) (*netbox_goV1.GetVpnIpsecproposalByIDReply, error) {
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

	data, err := convertVpnIpsecproposal(record)
	if err != nil {
		logger.Warn("convertVpnIpsecproposal error", logger.Err(err), logger.Any("vpnIpsecproposal", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVpnIpsecproposal.Err()
	}

	return &netbox_goV1.GetVpnIpsecproposalByIDReply{VpnIpsecproposal: data}, nil
}

// List get a paginated list of vpnIpsecproposals by custom conditions
func (s *vpnIpsecproposal) List(ctx context.Context, req *netbox_goV1.ListVpnIpsecproposalRequest) (*netbox_goV1.ListVpnIpsecproposalReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVpnIpsecproposal.Err()
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

	vpnIpsecproposals := []*netbox_goV1.VpnIpsecproposal{}
	for _, record := range records {
		data, err := convertVpnIpsecproposal(record)
		if err != nil {
			logger.Warn("convertVpnIpsecproposal error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnIpsecproposals = append(vpnIpsecproposals, data)
	}

	return &netbox_goV1.ListVpnIpsecproposalReply{
		Total:             total,
		VpnIpsecproposals: vpnIpsecproposals,
	}, nil
}

// DeleteByIDs batch delete vpnIpsecproposal by ids
func (s *vpnIpsecproposal) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnIpsecproposalByIDsRequest) (*netbox_goV1.DeleteVpnIpsecproposalByIDsReply, error) {
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

	return &netbox_goV1.DeleteVpnIpsecproposalByIDsReply{}, nil
}

// GetByCondition get a vpnIpsecproposal by custom condition
func (s *vpnIpsecproposal) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnIpsecproposalByConditionRequest) (*netbox_goV1.GetVpnIpsecproposalByConditionReply, error) {
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

	data, err := convertVpnIpsecproposal(record)
	if err != nil {
		logger.Warn("convertVpnIpsecproposal error", logger.Err(err), logger.Any("vpnIpsecproposal", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVpnIpsecproposal.Err()
	}

	return &netbox_goV1.GetVpnIpsecproposalByConditionReply{
		VpnIpsecproposal: data,
	}, nil
}

// ListByIDs batch get vpnIpsecproposal by ids
func (s *vpnIpsecproposal) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnIpsecproposalByIDsRequest) (*netbox_goV1.ListVpnIpsecproposalByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	vpnIpsecproposalMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	vpnIpsecproposals := []*netbox_goV1.VpnIpsecproposal{}
	for _, id := range req.Ids {
		if v, ok := vpnIpsecproposalMap[id]; ok {
			record, err := convertVpnIpsecproposal(v)
			if err != nil {
				logger.Warn("convertVpnIpsecproposal error", logger.Err(err), logger.Any("vpnIpsecproposal", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			vpnIpsecproposals = append(vpnIpsecproposals, record)
		}
	}

	return &netbox_goV1.ListVpnIpsecproposalByIDsReply{VpnIpsecproposals: vpnIpsecproposals}, nil
}

// ListByLastID get a paginated list of vpnIpsecproposals by last id
func (s *vpnIpsecproposal) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnIpsecproposalByLastIDRequest) (*netbox_goV1.ListVpnIpsecproposalByLastIDReply, error) {
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

	vpnIpsecproposals := []*netbox_goV1.VpnIpsecproposal{}
	for _, record := range records {
		data, err := convertVpnIpsecproposal(record)
		if err != nil {
			logger.Warn("convertVpnIpsecproposal error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnIpsecproposals = append(vpnIpsecproposals, data)
	}

	return &netbox_goV1.ListVpnIpsecproposalByLastIDReply{
		VpnIpsecproposals: vpnIpsecproposals,
	}, nil
}

func convertVpnIpsecproposal(record *model.VpnIpsecproposal) (*netbox_goV1.VpnIpsecproposal, error) {
	value := &netbox_goV1.VpnIpsecproposal{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
