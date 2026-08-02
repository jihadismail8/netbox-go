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
		netbox_goV1.RegisterIpamAggregateServer(server, NewIpamAggregateServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.IpamAggregateServer = (*ipamAggregate)(nil)
var _ time.Time

type ipamAggregate struct {
	netbox_goV1.UnimplementedIpamAggregateServer

	iDao dao.IpamAggregateDao
}

// NewIpamAggregateServer create a new service
func NewIpamAggregateServer() netbox_goV1.IpamAggregateServer {
	return &ipamAggregate{
		iDao: dao.NewIpamAggregateDao(
			database.GetDB(), // db driver is postgresql
			cache.NewIpamAggregateCache(database.GetCacheType()),
		),
	}
}

// Create a new ipamAggregate
func (s *ipamAggregate) Create(ctx context.Context, req *netbox_goV1.CreateIpamAggregateRequest) (*netbox_goV1.CreateIpamAggregateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamAggregate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateIpamAggregate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("ipamAggregate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateIpamAggregateReply{Id: record.ID}, nil
}

// DeleteByID delete a ipamAggregate by id
func (s *ipamAggregate) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamAggregateByIDRequest) (*netbox_goV1.DeleteIpamAggregateByIDReply, error) {
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

	return &netbox_goV1.DeleteIpamAggregateByIDReply{}, nil
}

// UpdateByID update a ipamAggregate by id
func (s *ipamAggregate) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamAggregateByIDRequest) (*netbox_goV1.UpdateIpamAggregateByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamAggregate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDIpamAggregate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("ipamAggregate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateIpamAggregateByIDReply{}, nil
}

// GetByID get a ipamAggregate by id
func (s *ipamAggregate) GetByID(ctx context.Context, req *netbox_goV1.GetIpamAggregateByIDRequest) (*netbox_goV1.GetIpamAggregateByIDReply, error) {
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

	data, err := convertIpamAggregate(record)
	if err != nil {
		logger.Warn("convertIpamAggregate error", logger.Err(err), logger.Any("ipamAggregate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDIpamAggregate.Err()
	}

	return &netbox_goV1.GetIpamAggregateByIDReply{IpamAggregate: data}, nil
}

// List get a paginated list of ipamAggregates by custom conditions
func (s *ipamAggregate) List(ctx context.Context, req *netbox_goV1.ListIpamAggregateRequest) (*netbox_goV1.ListIpamAggregateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListIpamAggregate.Err()
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

	ipamAggregates := []*netbox_goV1.IpamAggregate{}
	for _, record := range records {
		data, err := convertIpamAggregate(record)
		if err != nil {
			logger.Warn("convertIpamAggregate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamAggregates = append(ipamAggregates, data)
	}

	return &netbox_goV1.ListIpamAggregateReply{
		Total:          total,
		IpamAggregates: ipamAggregates,
	}, nil
}

// DeleteByIDs batch delete ipamAggregate by ids
func (s *ipamAggregate) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamAggregateByIDsRequest) (*netbox_goV1.DeleteIpamAggregateByIDsReply, error) {
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

	return &netbox_goV1.DeleteIpamAggregateByIDsReply{}, nil
}

// GetByCondition get a ipamAggregate by custom condition
func (s *ipamAggregate) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamAggregateByConditionRequest) (*netbox_goV1.GetIpamAggregateByConditionReply, error) {
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

	data, err := convertIpamAggregate(record)
	if err != nil {
		logger.Warn("convertIpamAggregate error", logger.Err(err), logger.Any("ipamAggregate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionIpamAggregate.Err()
	}

	return &netbox_goV1.GetIpamAggregateByConditionReply{
		IpamAggregate: data,
	}, nil
}

// ListByIDs batch get ipamAggregate by ids
func (s *ipamAggregate) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamAggregateByIDsRequest) (*netbox_goV1.ListIpamAggregateByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ipamAggregateMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	ipamAggregates := []*netbox_goV1.IpamAggregate{}
	for _, id := range req.Ids {
		if v, ok := ipamAggregateMap[id]; ok {
			record, err := convertIpamAggregate(v)
			if err != nil {
				logger.Warn("convertIpamAggregate error", logger.Err(err), logger.Any("ipamAggregate", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			ipamAggregates = append(ipamAggregates, record)
		}
	}

	return &netbox_goV1.ListIpamAggregateByIDsReply{IpamAggregates: ipamAggregates}, nil
}

// ListByLastID get a paginated list of ipamAggregates by last id
func (s *ipamAggregate) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamAggregateByLastIDRequest) (*netbox_goV1.ListIpamAggregateByLastIDReply, error) {
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

	ipamAggregates := []*netbox_goV1.IpamAggregate{}
	for _, record := range records {
		data, err := convertIpamAggregate(record)
		if err != nil {
			logger.Warn("convertIpamAggregate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamAggregates = append(ipamAggregates, data)
	}

	return &netbox_goV1.ListIpamAggregateByLastIDReply{
		IpamAggregates: ipamAggregates,
	}, nil
}

func convertIpamAggregate(record *model.IpamAggregate) (*netbox_goV1.IpamAggregate, error) {
	value := &netbox_goV1.IpamAggregate{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
