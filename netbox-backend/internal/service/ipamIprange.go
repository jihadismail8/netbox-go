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
		netbox_goV1.RegisterIpamIprangeServer(server, NewIpamIprangeServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.IpamIprangeServer = (*ipamIprange)(nil)
var _ time.Time

type ipamIprange struct {
	netbox_goV1.UnimplementedIpamIprangeServer

	iDao dao.IpamIprangeDao
}

// NewIpamIprangeServer create a new service
func NewIpamIprangeServer() netbox_goV1.IpamIprangeServer {
	return &ipamIprange{
		iDao: dao.NewIpamIprangeDao(
			database.GetDB(), // db driver is postgresql
			cache.NewIpamIprangeCache(database.GetCacheType()),
		),
	}
}

// Create a new ipamIprange
func (s *ipamIprange) Create(ctx context.Context, req *netbox_goV1.CreateIpamIprangeRequest) (*netbox_goV1.CreateIpamIprangeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamIprange{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateIpamIprange.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("ipamIprange", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateIpamIprangeReply{Id: record.ID}, nil
}

// DeleteByID delete a ipamIprange by id
func (s *ipamIprange) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamIprangeByIDRequest) (*netbox_goV1.DeleteIpamIprangeByIDReply, error) {
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

	return &netbox_goV1.DeleteIpamIprangeByIDReply{}, nil
}

// UpdateByID update a ipamIprange by id
func (s *ipamIprange) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamIprangeByIDRequest) (*netbox_goV1.UpdateIpamIprangeByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamIprange{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDIpamIprange.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("ipamIprange", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateIpamIprangeByIDReply{}, nil
}

// GetByID get a ipamIprange by id
func (s *ipamIprange) GetByID(ctx context.Context, req *netbox_goV1.GetIpamIprangeByIDRequest) (*netbox_goV1.GetIpamIprangeByIDReply, error) {
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

	data, err := convertIpamIprange(record)
	if err != nil {
		logger.Warn("convertIpamIprange error", logger.Err(err), logger.Any("ipamIprange", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDIpamIprange.Err()
	}

	return &netbox_goV1.GetIpamIprangeByIDReply{IpamIprange: data}, nil
}

// List get a paginated list of ipamIpranges by custom conditions
func (s *ipamIprange) List(ctx context.Context, req *netbox_goV1.ListIpamIprangeRequest) (*netbox_goV1.ListIpamIprangeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListIpamIprange.Err()
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

	ipamIpranges := []*netbox_goV1.IpamIprange{}
	for _, record := range records {
		data, err := convertIpamIprange(record)
		if err != nil {
			logger.Warn("convertIpamIprange error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamIpranges = append(ipamIpranges, data)
	}

	return &netbox_goV1.ListIpamIprangeReply{
		Total:        total,
		IpamIpranges: ipamIpranges,
	}, nil
}

// DeleteByIDs batch delete ipamIprange by ids
func (s *ipamIprange) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamIprangeByIDsRequest) (*netbox_goV1.DeleteIpamIprangeByIDsReply, error) {
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

	return &netbox_goV1.DeleteIpamIprangeByIDsReply{}, nil
}

// GetByCondition get a ipamIprange by custom condition
func (s *ipamIprange) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamIprangeByConditionRequest) (*netbox_goV1.GetIpamIprangeByConditionReply, error) {
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

	data, err := convertIpamIprange(record)
	if err != nil {
		logger.Warn("convertIpamIprange error", logger.Err(err), logger.Any("ipamIprange", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionIpamIprange.Err()
	}

	return &netbox_goV1.GetIpamIprangeByConditionReply{
		IpamIprange: data,
	}, nil
}

// ListByIDs batch get ipamIprange by ids
func (s *ipamIprange) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamIprangeByIDsRequest) (*netbox_goV1.ListIpamIprangeByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ipamIprangeMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	ipamIpranges := []*netbox_goV1.IpamIprange{}
	for _, id := range req.Ids {
		if v, ok := ipamIprangeMap[id]; ok {
			record, err := convertIpamIprange(v)
			if err != nil {
				logger.Warn("convertIpamIprange error", logger.Err(err), logger.Any("ipamIprange", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			ipamIpranges = append(ipamIpranges, record)
		}
	}

	return &netbox_goV1.ListIpamIprangeByIDsReply{IpamIpranges: ipamIpranges}, nil
}

// ListByLastID get a paginated list of ipamIpranges by last id
func (s *ipamIprange) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamIprangeByLastIDRequest) (*netbox_goV1.ListIpamIprangeByLastIDReply, error) {
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

	ipamIpranges := []*netbox_goV1.IpamIprange{}
	for _, record := range records {
		data, err := convertIpamIprange(record)
		if err != nil {
			logger.Warn("convertIpamIprange error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamIpranges = append(ipamIpranges, data)
	}

	return &netbox_goV1.ListIpamIprangeByLastIDReply{
		IpamIpranges: ipamIpranges,
	}, nil
}

func convertIpamIprange(record *model.IpamIprange) (*netbox_goV1.IpamIprange, error) {
	value := &netbox_goV1.IpamIprange{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
