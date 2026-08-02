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
		netbox_goV1.RegisterIpamVlantranslationpolicyServer(server, NewIpamVlantranslationpolicyServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.IpamVlantranslationpolicyServer = (*ipamVlantranslationpolicy)(nil)
var _ time.Time

type ipamVlantranslationpolicy struct {
	netbox_goV1.UnimplementedIpamVlantranslationpolicyServer

	iDao dao.IpamVlantranslationpolicyDao
}

// NewIpamVlantranslationpolicyServer create a new service
func NewIpamVlantranslationpolicyServer() netbox_goV1.IpamVlantranslationpolicyServer {
	return &ipamVlantranslationpolicy{
		iDao: dao.NewIpamVlantranslationpolicyDao(
			database.GetDB(), // db driver is postgresql
			cache.NewIpamVlantranslationpolicyCache(database.GetCacheType()),
		),
	}
}

// Create a new ipamVlantranslationpolicy
func (s *ipamVlantranslationpolicy) Create(ctx context.Context, req *netbox_goV1.CreateIpamVlantranslationpolicyRequest) (*netbox_goV1.CreateIpamVlantranslationpolicyReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamVlantranslationpolicy{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateIpamVlantranslationpolicy.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("ipamVlantranslationpolicy", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateIpamVlantranslationpolicyReply{Id: record.ID}, nil
}

// DeleteByID delete a ipamVlantranslationpolicy by id
func (s *ipamVlantranslationpolicy) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamVlantranslationpolicyByIDRequest) (*netbox_goV1.DeleteIpamVlantranslationpolicyByIDReply, error) {
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

	return &netbox_goV1.DeleteIpamVlantranslationpolicyByIDReply{}, nil
}

// UpdateByID update a ipamVlantranslationpolicy by id
func (s *ipamVlantranslationpolicy) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamVlantranslationpolicyByIDRequest) (*netbox_goV1.UpdateIpamVlantranslationpolicyByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamVlantranslationpolicy{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDIpamVlantranslationpolicy.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("ipamVlantranslationpolicy", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateIpamVlantranslationpolicyByIDReply{}, nil
}

// GetByID get a ipamVlantranslationpolicy by id
func (s *ipamVlantranslationpolicy) GetByID(ctx context.Context, req *netbox_goV1.GetIpamVlantranslationpolicyByIDRequest) (*netbox_goV1.GetIpamVlantranslationpolicyByIDReply, error) {
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

	data, err := convertIpamVlantranslationpolicy(record)
	if err != nil {
		logger.Warn("convertIpamVlantranslationpolicy error", logger.Err(err), logger.Any("ipamVlantranslationpolicy", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDIpamVlantranslationpolicy.Err()
	}

	return &netbox_goV1.GetIpamVlantranslationpolicyByIDReply{IpamVlantranslationpolicy: data}, nil
}

// List get a paginated list of ipamVlantranslationpolicys by custom conditions
func (s *ipamVlantranslationpolicy) List(ctx context.Context, req *netbox_goV1.ListIpamVlantranslationpolicyRequest) (*netbox_goV1.ListIpamVlantranslationpolicyReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListIpamVlantranslationpolicy.Err()
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

	ipamVlantranslationpolicys := []*netbox_goV1.IpamVlantranslationpolicy{}
	for _, record := range records {
		data, err := convertIpamVlantranslationpolicy(record)
		if err != nil {
			logger.Warn("convertIpamVlantranslationpolicy error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamVlantranslationpolicys = append(ipamVlantranslationpolicys, data)
	}

	return &netbox_goV1.ListIpamVlantranslationpolicyReply{
		Total:                      total,
		IpamVlantranslationpolicys: ipamVlantranslationpolicys,
	}, nil
}

// DeleteByIDs batch delete ipamVlantranslationpolicy by ids
func (s *ipamVlantranslationpolicy) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamVlantranslationpolicyByIDsRequest) (*netbox_goV1.DeleteIpamVlantranslationpolicyByIDsReply, error) {
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

	return &netbox_goV1.DeleteIpamVlantranslationpolicyByIDsReply{}, nil
}

// GetByCondition get a ipamVlantranslationpolicy by custom condition
func (s *ipamVlantranslationpolicy) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamVlantranslationpolicyByConditionRequest) (*netbox_goV1.GetIpamVlantranslationpolicyByConditionReply, error) {
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

	data, err := convertIpamVlantranslationpolicy(record)
	if err != nil {
		logger.Warn("convertIpamVlantranslationpolicy error", logger.Err(err), logger.Any("ipamVlantranslationpolicy", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionIpamVlantranslationpolicy.Err()
	}

	return &netbox_goV1.GetIpamVlantranslationpolicyByConditionReply{
		IpamVlantranslationpolicy: data,
	}, nil
}

// ListByIDs batch get ipamVlantranslationpolicy by ids
func (s *ipamVlantranslationpolicy) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamVlantranslationpolicyByIDsRequest) (*netbox_goV1.ListIpamVlantranslationpolicyByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ipamVlantranslationpolicyMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	ipamVlantranslationpolicys := []*netbox_goV1.IpamVlantranslationpolicy{}
	for _, id := range req.Ids {
		if v, ok := ipamVlantranslationpolicyMap[id]; ok {
			record, err := convertIpamVlantranslationpolicy(v)
			if err != nil {
				logger.Warn("convertIpamVlantranslationpolicy error", logger.Err(err), logger.Any("ipamVlantranslationpolicy", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			ipamVlantranslationpolicys = append(ipamVlantranslationpolicys, record)
		}
	}

	return &netbox_goV1.ListIpamVlantranslationpolicyByIDsReply{IpamVlantranslationpolicys: ipamVlantranslationpolicys}, nil
}

// ListByLastID get a paginated list of ipamVlantranslationpolicys by last id
func (s *ipamVlantranslationpolicy) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamVlantranslationpolicyByLastIDRequest) (*netbox_goV1.ListIpamVlantranslationpolicyByLastIDReply, error) {
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

	ipamVlantranslationpolicys := []*netbox_goV1.IpamVlantranslationpolicy{}
	for _, record := range records {
		data, err := convertIpamVlantranslationpolicy(record)
		if err != nil {
			logger.Warn("convertIpamVlantranslationpolicy error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamVlantranslationpolicys = append(ipamVlantranslationpolicys, data)
	}

	return &netbox_goV1.ListIpamVlantranslationpolicyByLastIDReply{
		IpamVlantranslationpolicys: ipamVlantranslationpolicys,
	}, nil
}

func convertIpamVlantranslationpolicy(record *model.IpamVlantranslationpolicy) (*netbox_goV1.IpamVlantranslationpolicy, error) {
	value := &netbox_goV1.IpamVlantranslationpolicy{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
