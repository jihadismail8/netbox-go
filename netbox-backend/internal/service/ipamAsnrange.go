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
		netbox_goV1.RegisterIpamAsnrangeServer(server, NewIpamAsnrangeServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.IpamAsnrangeServer = (*ipamAsnrange)(nil)
var _ time.Time

type ipamAsnrange struct {
	netbox_goV1.UnimplementedIpamAsnrangeServer

	iDao dao.IpamAsnrangeDao
}

// NewIpamAsnrangeServer create a new service
func NewIpamAsnrangeServer() netbox_goV1.IpamAsnrangeServer {
	return &ipamAsnrange{
		iDao: dao.NewIpamAsnrangeDao(
			database.GetDB(), // db driver is postgresql
			cache.NewIpamAsnrangeCache(database.GetCacheType()),
		),
	}
}

// Create a new ipamAsnrange
func (s *ipamAsnrange) Create(ctx context.Context, req *netbox_goV1.CreateIpamAsnrangeRequest) (*netbox_goV1.CreateIpamAsnrangeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamAsnrange{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateIpamAsnrange.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("ipamAsnrange", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateIpamAsnrangeReply{Id: record.ID}, nil
}

// DeleteByID delete a ipamAsnrange by id
func (s *ipamAsnrange) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamAsnrangeByIDRequest) (*netbox_goV1.DeleteIpamAsnrangeByIDReply, error) {
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

	return &netbox_goV1.DeleteIpamAsnrangeByIDReply{}, nil
}

// UpdateByID update a ipamAsnrange by id
func (s *ipamAsnrange) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamAsnrangeByIDRequest) (*netbox_goV1.UpdateIpamAsnrangeByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamAsnrange{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDIpamAsnrange.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("ipamAsnrange", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateIpamAsnrangeByIDReply{}, nil
}

// GetByID get a ipamAsnrange by id
func (s *ipamAsnrange) GetByID(ctx context.Context, req *netbox_goV1.GetIpamAsnrangeByIDRequest) (*netbox_goV1.GetIpamAsnrangeByIDReply, error) {
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

	data, err := convertIpamAsnrange(record)
	if err != nil {
		logger.Warn("convertIpamAsnrange error", logger.Err(err), logger.Any("ipamAsnrange", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDIpamAsnrange.Err()
	}

	return &netbox_goV1.GetIpamAsnrangeByIDReply{IpamAsnrange: data}, nil
}

// List get a paginated list of ipamAsnranges by custom conditions
func (s *ipamAsnrange) List(ctx context.Context, req *netbox_goV1.ListIpamAsnrangeRequest) (*netbox_goV1.ListIpamAsnrangeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListIpamAsnrange.Err()
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

	ipamAsnranges := []*netbox_goV1.IpamAsnrange{}
	for _, record := range records {
		data, err := convertIpamAsnrange(record)
		if err != nil {
			logger.Warn("convertIpamAsnrange error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamAsnranges = append(ipamAsnranges, data)
	}

	return &netbox_goV1.ListIpamAsnrangeReply{
		Total:         total,
		IpamAsnranges: ipamAsnranges,
	}, nil
}

// DeleteByIDs batch delete ipamAsnrange by ids
func (s *ipamAsnrange) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamAsnrangeByIDsRequest) (*netbox_goV1.DeleteIpamAsnrangeByIDsReply, error) {
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

	return &netbox_goV1.DeleteIpamAsnrangeByIDsReply{}, nil
}

// GetByCondition get a ipamAsnrange by custom condition
func (s *ipamAsnrange) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamAsnrangeByConditionRequest) (*netbox_goV1.GetIpamAsnrangeByConditionReply, error) {
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

	data, err := convertIpamAsnrange(record)
	if err != nil {
		logger.Warn("convertIpamAsnrange error", logger.Err(err), logger.Any("ipamAsnrange", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionIpamAsnrange.Err()
	}

	return &netbox_goV1.GetIpamAsnrangeByConditionReply{
		IpamAsnrange: data,
	}, nil
}

// ListByIDs batch get ipamAsnrange by ids
func (s *ipamAsnrange) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamAsnrangeByIDsRequest) (*netbox_goV1.ListIpamAsnrangeByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ipamAsnrangeMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	ipamAsnranges := []*netbox_goV1.IpamAsnrange{}
	for _, id := range req.Ids {
		if v, ok := ipamAsnrangeMap[id]; ok {
			record, err := convertIpamAsnrange(v)
			if err != nil {
				logger.Warn("convertIpamAsnrange error", logger.Err(err), logger.Any("ipamAsnrange", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			ipamAsnranges = append(ipamAsnranges, record)
		}
	}

	return &netbox_goV1.ListIpamAsnrangeByIDsReply{IpamAsnranges: ipamAsnranges}, nil
}

// ListByLastID get a paginated list of ipamAsnranges by last id
func (s *ipamAsnrange) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamAsnrangeByLastIDRequest) (*netbox_goV1.ListIpamAsnrangeByLastIDReply, error) {
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

	ipamAsnranges := []*netbox_goV1.IpamAsnrange{}
	for _, record := range records {
		data, err := convertIpamAsnrange(record)
		if err != nil {
			logger.Warn("convertIpamAsnrange error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamAsnranges = append(ipamAsnranges, data)
	}

	return &netbox_goV1.ListIpamAsnrangeByLastIDReply{
		IpamAsnranges: ipamAsnranges,
	}, nil
}

func convertIpamAsnrange(record *model.IpamAsnrange) (*netbox_goV1.IpamAsnrange, error) {
	value := &netbox_goV1.IpamAsnrange{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
