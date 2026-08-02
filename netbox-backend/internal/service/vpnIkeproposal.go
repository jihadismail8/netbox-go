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
		netbox_goV1.RegisterVpnIkeproposalServer(server, NewVpnIkeproposalServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VpnIkeproposalServer = (*vpnIkeproposal)(nil)
var _ time.Time

type vpnIkeproposal struct {
	netbox_goV1.UnimplementedVpnIkeproposalServer

	iDao dao.VpnIkeproposalDao
}

// NewVpnIkeproposalServer create a new service
func NewVpnIkeproposalServer() netbox_goV1.VpnIkeproposalServer {
	return &vpnIkeproposal{
		iDao: dao.NewVpnIkeproposalDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVpnIkeproposalCache(database.GetCacheType()),
		),
	}
}

// Create a new vpnIkeproposal
func (s *vpnIkeproposal) Create(ctx context.Context, req *netbox_goV1.CreateVpnIkeproposalRequest) (*netbox_goV1.CreateVpnIkeproposalReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnIkeproposal{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVpnIkeproposal.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("vpnIkeproposal", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVpnIkeproposalReply{Id: record.ID}, nil
}

// DeleteByID delete a vpnIkeproposal by id
func (s *vpnIkeproposal) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnIkeproposalByIDRequest) (*netbox_goV1.DeleteVpnIkeproposalByIDReply, error) {
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

	return &netbox_goV1.DeleteVpnIkeproposalByIDReply{}, nil
}

// UpdateByID update a vpnIkeproposal by id
func (s *vpnIkeproposal) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnIkeproposalByIDRequest) (*netbox_goV1.UpdateVpnIkeproposalByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnIkeproposal{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVpnIkeproposal.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("vpnIkeproposal", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVpnIkeproposalByIDReply{}, nil
}

// GetByID get a vpnIkeproposal by id
func (s *vpnIkeproposal) GetByID(ctx context.Context, req *netbox_goV1.GetVpnIkeproposalByIDRequest) (*netbox_goV1.GetVpnIkeproposalByIDReply, error) {
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

	data, err := convertVpnIkeproposal(record)
	if err != nil {
		logger.Warn("convertVpnIkeproposal error", logger.Err(err), logger.Any("vpnIkeproposal", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVpnIkeproposal.Err()
	}

	return &netbox_goV1.GetVpnIkeproposalByIDReply{VpnIkeproposal: data}, nil
}

// List get a paginated list of vpnIkeproposals by custom conditions
func (s *vpnIkeproposal) List(ctx context.Context, req *netbox_goV1.ListVpnIkeproposalRequest) (*netbox_goV1.ListVpnIkeproposalReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVpnIkeproposal.Err()
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

	vpnIkeproposals := []*netbox_goV1.VpnIkeproposal{}
	for _, record := range records {
		data, err := convertVpnIkeproposal(record)
		if err != nil {
			logger.Warn("convertVpnIkeproposal error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnIkeproposals = append(vpnIkeproposals, data)
	}

	return &netbox_goV1.ListVpnIkeproposalReply{
		Total:           total,
		VpnIkeproposals: vpnIkeproposals,
	}, nil
}

// DeleteByIDs batch delete vpnIkeproposal by ids
func (s *vpnIkeproposal) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnIkeproposalByIDsRequest) (*netbox_goV1.DeleteVpnIkeproposalByIDsReply, error) {
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

	return &netbox_goV1.DeleteVpnIkeproposalByIDsReply{}, nil
}

// GetByCondition get a vpnIkeproposal by custom condition
func (s *vpnIkeproposal) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnIkeproposalByConditionRequest) (*netbox_goV1.GetVpnIkeproposalByConditionReply, error) {
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

	data, err := convertVpnIkeproposal(record)
	if err != nil {
		logger.Warn("convertVpnIkeproposal error", logger.Err(err), logger.Any("vpnIkeproposal", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVpnIkeproposal.Err()
	}

	return &netbox_goV1.GetVpnIkeproposalByConditionReply{
		VpnIkeproposal: data,
	}, nil
}

// ListByIDs batch get vpnIkeproposal by ids
func (s *vpnIkeproposal) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnIkeproposalByIDsRequest) (*netbox_goV1.ListVpnIkeproposalByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	vpnIkeproposalMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	vpnIkeproposals := []*netbox_goV1.VpnIkeproposal{}
	for _, id := range req.Ids {
		if v, ok := vpnIkeproposalMap[id]; ok {
			record, err := convertVpnIkeproposal(v)
			if err != nil {
				logger.Warn("convertVpnIkeproposal error", logger.Err(err), logger.Any("vpnIkeproposal", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			vpnIkeproposals = append(vpnIkeproposals, record)
		}
	}

	return &netbox_goV1.ListVpnIkeproposalByIDsReply{VpnIkeproposals: vpnIkeproposals}, nil
}

// ListByLastID get a paginated list of vpnIkeproposals by last id
func (s *vpnIkeproposal) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnIkeproposalByLastIDRequest) (*netbox_goV1.ListVpnIkeproposalByLastIDReply, error) {
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

	vpnIkeproposals := []*netbox_goV1.VpnIkeproposal{}
	for _, record := range records {
		data, err := convertVpnIkeproposal(record)
		if err != nil {
			logger.Warn("convertVpnIkeproposal error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnIkeproposals = append(vpnIkeproposals, data)
	}

	return &netbox_goV1.ListVpnIkeproposalByLastIDReply{
		VpnIkeproposals: vpnIkeproposals,
	}, nil
}

func convertVpnIkeproposal(record *model.VpnIkeproposal) (*netbox_goV1.VpnIkeproposal, error) {
	value := &netbox_goV1.VpnIkeproposal{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
