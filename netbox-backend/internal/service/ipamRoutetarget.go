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
		netbox_goV1.RegisterIpamRoutetargetServer(server, NewIpamRoutetargetServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.IpamRoutetargetServer = (*ipamRoutetarget)(nil)
var _ time.Time

type ipamRoutetarget struct {
	netbox_goV1.UnimplementedIpamRoutetargetServer

	iDao dao.IpamRoutetargetDao
}

// NewIpamRoutetargetServer create a new service
func NewIpamRoutetargetServer() netbox_goV1.IpamRoutetargetServer {
	return &ipamRoutetarget{
		iDao: dao.NewIpamRoutetargetDao(
			database.GetDB(), // db driver is postgresql
			cache.NewIpamRoutetargetCache(database.GetCacheType()),
		),
	}
}

// Create a new ipamRoutetarget
func (s *ipamRoutetarget) Create(ctx context.Context, req *netbox_goV1.CreateIpamRoutetargetRequest) (*netbox_goV1.CreateIpamRoutetargetReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamRoutetarget{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateIpamRoutetarget.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("ipamRoutetarget", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateIpamRoutetargetReply{Id: record.ID}, nil
}

// DeleteByID delete a ipamRoutetarget by id
func (s *ipamRoutetarget) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamRoutetargetByIDRequest) (*netbox_goV1.DeleteIpamRoutetargetByIDReply, error) {
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

	return &netbox_goV1.DeleteIpamRoutetargetByIDReply{}, nil
}

// UpdateByID update a ipamRoutetarget by id
func (s *ipamRoutetarget) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamRoutetargetByIDRequest) (*netbox_goV1.UpdateIpamRoutetargetByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamRoutetarget{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDIpamRoutetarget.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("ipamRoutetarget", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateIpamRoutetargetByIDReply{}, nil
}

// GetByID get a ipamRoutetarget by id
func (s *ipamRoutetarget) GetByID(ctx context.Context, req *netbox_goV1.GetIpamRoutetargetByIDRequest) (*netbox_goV1.GetIpamRoutetargetByIDReply, error) {
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

	data, err := convertIpamRoutetarget(record)
	if err != nil {
		logger.Warn("convertIpamRoutetarget error", logger.Err(err), logger.Any("ipamRoutetarget", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDIpamRoutetarget.Err()
	}

	return &netbox_goV1.GetIpamRoutetargetByIDReply{IpamRoutetarget: data}, nil
}

// List get a paginated list of ipamRoutetargets by custom conditions
func (s *ipamRoutetarget) List(ctx context.Context, req *netbox_goV1.ListIpamRoutetargetRequest) (*netbox_goV1.ListIpamRoutetargetReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListIpamRoutetarget.Err()
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

	ipamRoutetargets := []*netbox_goV1.IpamRoutetarget{}
	for _, record := range records {
		data, err := convertIpamRoutetarget(record)
		if err != nil {
			logger.Warn("convertIpamRoutetarget error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamRoutetargets = append(ipamRoutetargets, data)
	}

	return &netbox_goV1.ListIpamRoutetargetReply{
		Total:            total,
		IpamRoutetargets: ipamRoutetargets,
	}, nil
}

// DeleteByIDs batch delete ipamRoutetarget by ids
func (s *ipamRoutetarget) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamRoutetargetByIDsRequest) (*netbox_goV1.DeleteIpamRoutetargetByIDsReply, error) {
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

	return &netbox_goV1.DeleteIpamRoutetargetByIDsReply{}, nil
}

// GetByCondition get a ipamRoutetarget by custom condition
func (s *ipamRoutetarget) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamRoutetargetByConditionRequest) (*netbox_goV1.GetIpamRoutetargetByConditionReply, error) {
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

	data, err := convertIpamRoutetarget(record)
	if err != nil {
		logger.Warn("convertIpamRoutetarget error", logger.Err(err), logger.Any("ipamRoutetarget", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionIpamRoutetarget.Err()
	}

	return &netbox_goV1.GetIpamRoutetargetByConditionReply{
		IpamRoutetarget: data,
	}, nil
}

// ListByIDs batch get ipamRoutetarget by ids
func (s *ipamRoutetarget) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamRoutetargetByIDsRequest) (*netbox_goV1.ListIpamRoutetargetByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ipamRoutetargetMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	ipamRoutetargets := []*netbox_goV1.IpamRoutetarget{}
	for _, id := range req.Ids {
		if v, ok := ipamRoutetargetMap[id]; ok {
			record, err := convertIpamRoutetarget(v)
			if err != nil {
				logger.Warn("convertIpamRoutetarget error", logger.Err(err), logger.Any("ipamRoutetarget", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			ipamRoutetargets = append(ipamRoutetargets, record)
		}
	}

	return &netbox_goV1.ListIpamRoutetargetByIDsReply{IpamRoutetargets: ipamRoutetargets}, nil
}

// ListByLastID get a paginated list of ipamRoutetargets by last id
func (s *ipamRoutetarget) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamRoutetargetByLastIDRequest) (*netbox_goV1.ListIpamRoutetargetByLastIDReply, error) {
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

	ipamRoutetargets := []*netbox_goV1.IpamRoutetarget{}
	for _, record := range records {
		data, err := convertIpamRoutetarget(record)
		if err != nil {
			logger.Warn("convertIpamRoutetarget error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamRoutetargets = append(ipamRoutetargets, data)
	}

	return &netbox_goV1.ListIpamRoutetargetByLastIDReply{
		IpamRoutetargets: ipamRoutetargets,
	}, nil
}

func convertIpamRoutetarget(record *model.IpamRoutetarget) (*netbox_goV1.IpamRoutetarget, error) {
	value := &netbox_goV1.IpamRoutetarget{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
