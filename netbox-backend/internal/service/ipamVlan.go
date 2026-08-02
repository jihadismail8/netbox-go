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
		netbox_goV1.RegisterIpamVlanServer(server, NewIpamVlanServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.IpamVlanServer = (*ipamVlan)(nil)
var _ time.Time

type ipamVlan struct {
	netbox_goV1.UnimplementedIpamVlanServer

	iDao dao.IpamVlanDao
}

// NewIpamVlanServer create a new service
func NewIpamVlanServer() netbox_goV1.IpamVlanServer {
	return &ipamVlan{
		iDao: dao.NewIpamVlanDao(
			database.GetDB(), // db driver is postgresql
			cache.NewIpamVlanCache(database.GetCacheType()),
		),
	}
}

// Create a new ipamVlan
func (s *ipamVlan) Create(ctx context.Context, req *netbox_goV1.CreateIpamVlanRequest) (*netbox_goV1.CreateIpamVlanReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamVlan{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateIpamVlan.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("ipamVlan", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateIpamVlanReply{Id: record.ID}, nil
}

// DeleteByID delete a ipamVlan by id
func (s *ipamVlan) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamVlanByIDRequest) (*netbox_goV1.DeleteIpamVlanByIDReply, error) {
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

	return &netbox_goV1.DeleteIpamVlanByIDReply{}, nil
}

// UpdateByID update a ipamVlan by id
func (s *ipamVlan) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamVlanByIDRequest) (*netbox_goV1.UpdateIpamVlanByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamVlan{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDIpamVlan.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("ipamVlan", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateIpamVlanByIDReply{}, nil
}

// GetByID get a ipamVlan by id
func (s *ipamVlan) GetByID(ctx context.Context, req *netbox_goV1.GetIpamVlanByIDRequest) (*netbox_goV1.GetIpamVlanByIDReply, error) {
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

	data, err := convertIpamVlan(record)
	if err != nil {
		logger.Warn("convertIpamVlan error", logger.Err(err), logger.Any("ipamVlan", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDIpamVlan.Err()
	}

	return &netbox_goV1.GetIpamVlanByIDReply{IpamVlan: data}, nil
}

// List get a paginated list of ipamVlans by custom conditions
func (s *ipamVlan) List(ctx context.Context, req *netbox_goV1.ListIpamVlanRequest) (*netbox_goV1.ListIpamVlanReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListIpamVlan.Err()
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

	ipamVlans := []*netbox_goV1.IpamVlan{}
	for _, record := range records {
		data, err := convertIpamVlan(record)
		if err != nil {
			logger.Warn("convertIpamVlan error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamVlans = append(ipamVlans, data)
	}

	return &netbox_goV1.ListIpamVlanReply{
		Total:     total,
		IpamVlans: ipamVlans,
	}, nil
}

// DeleteByIDs batch delete ipamVlan by ids
func (s *ipamVlan) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamVlanByIDsRequest) (*netbox_goV1.DeleteIpamVlanByIDsReply, error) {
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

	return &netbox_goV1.DeleteIpamVlanByIDsReply{}, nil
}

// GetByCondition get a ipamVlan by custom condition
func (s *ipamVlan) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamVlanByConditionRequest) (*netbox_goV1.GetIpamVlanByConditionReply, error) {
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

	data, err := convertIpamVlan(record)
	if err != nil {
		logger.Warn("convertIpamVlan error", logger.Err(err), logger.Any("ipamVlan", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionIpamVlan.Err()
	}

	return &netbox_goV1.GetIpamVlanByConditionReply{
		IpamVlan: data,
	}, nil
}

// ListByIDs batch get ipamVlan by ids
func (s *ipamVlan) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamVlanByIDsRequest) (*netbox_goV1.ListIpamVlanByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ipamVlanMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	ipamVlans := []*netbox_goV1.IpamVlan{}
	for _, id := range req.Ids {
		if v, ok := ipamVlanMap[id]; ok {
			record, err := convertIpamVlan(v)
			if err != nil {
				logger.Warn("convertIpamVlan error", logger.Err(err), logger.Any("ipamVlan", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			ipamVlans = append(ipamVlans, record)
		}
	}

	return &netbox_goV1.ListIpamVlanByIDsReply{IpamVlans: ipamVlans}, nil
}

// ListByLastID get a paginated list of ipamVlans by last id
func (s *ipamVlan) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamVlanByLastIDRequest) (*netbox_goV1.ListIpamVlanByLastIDReply, error) {
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

	ipamVlans := []*netbox_goV1.IpamVlan{}
	for _, record := range records {
		data, err := convertIpamVlan(record)
		if err != nil {
			logger.Warn("convertIpamVlan error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamVlans = append(ipamVlans, data)
	}

	return &netbox_goV1.ListIpamVlanByLastIDReply{
		IpamVlans: ipamVlans,
	}, nil
}

func convertIpamVlan(record *model.IpamVlan) (*netbox_goV1.IpamVlan, error) {
	value := &netbox_goV1.IpamVlan{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
