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
		netbox_goV1.RegisterIpamAsnServer(server, NewIpamAsnServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.IpamAsnServer = (*ipamAsn)(nil)
var _ time.Time

type ipamAsn struct {
	netbox_goV1.UnimplementedIpamAsnServer

	iDao dao.IpamAsnDao
}

// NewIpamAsnServer create a new service
func NewIpamAsnServer() netbox_goV1.IpamAsnServer {
	return &ipamAsn{
		iDao: dao.NewIpamAsnDao(
			database.GetDB(), // db driver is postgresql
			cache.NewIpamAsnCache(database.GetCacheType()),
		),
	}
}

// Create a new ipamAsn
func (s *ipamAsn) Create(ctx context.Context, req *netbox_goV1.CreateIpamAsnRequest) (*netbox_goV1.CreateIpamAsnReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamAsn{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateIpamAsn.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("ipamAsn", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateIpamAsnReply{Id: record.ID}, nil
}

// DeleteByID delete a ipamAsn by id
func (s *ipamAsn) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamAsnByIDRequest) (*netbox_goV1.DeleteIpamAsnByIDReply, error) {
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

	return &netbox_goV1.DeleteIpamAsnByIDReply{}, nil
}

// UpdateByID update a ipamAsn by id
func (s *ipamAsn) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamAsnByIDRequest) (*netbox_goV1.UpdateIpamAsnByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamAsn{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDIpamAsn.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("ipamAsn", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateIpamAsnByIDReply{}, nil
}

// GetByID get a ipamAsn by id
func (s *ipamAsn) GetByID(ctx context.Context, req *netbox_goV1.GetIpamAsnByIDRequest) (*netbox_goV1.GetIpamAsnByIDReply, error) {
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

	data, err := convertIpamAsn(record)
	if err != nil {
		logger.Warn("convertIpamAsn error", logger.Err(err), logger.Any("ipamAsn", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDIpamAsn.Err()
	}

	return &netbox_goV1.GetIpamAsnByIDReply{IpamAsn: data}, nil
}

// List get a paginated list of ipamAsns by custom conditions
func (s *ipamAsn) List(ctx context.Context, req *netbox_goV1.ListIpamAsnRequest) (*netbox_goV1.ListIpamAsnReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListIpamAsn.Err()
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

	ipamAsns := []*netbox_goV1.IpamAsn{}
	for _, record := range records {
		data, err := convertIpamAsn(record)
		if err != nil {
			logger.Warn("convertIpamAsn error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamAsns = append(ipamAsns, data)
	}

	return &netbox_goV1.ListIpamAsnReply{
		Total:    total,
		IpamAsns: ipamAsns,
	}, nil
}

// DeleteByIDs batch delete ipamAsn by ids
func (s *ipamAsn) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamAsnByIDsRequest) (*netbox_goV1.DeleteIpamAsnByIDsReply, error) {
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

	return &netbox_goV1.DeleteIpamAsnByIDsReply{}, nil
}

// GetByCondition get a ipamAsn by custom condition
func (s *ipamAsn) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamAsnByConditionRequest) (*netbox_goV1.GetIpamAsnByConditionReply, error) {
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

	data, err := convertIpamAsn(record)
	if err != nil {
		logger.Warn("convertIpamAsn error", logger.Err(err), logger.Any("ipamAsn", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionIpamAsn.Err()
	}

	return &netbox_goV1.GetIpamAsnByConditionReply{
		IpamAsn: data,
	}, nil
}

// ListByIDs batch get ipamAsn by ids
func (s *ipamAsn) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamAsnByIDsRequest) (*netbox_goV1.ListIpamAsnByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ipamAsnMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	ipamAsns := []*netbox_goV1.IpamAsn{}
	for _, id := range req.Ids {
		if v, ok := ipamAsnMap[id]; ok {
			record, err := convertIpamAsn(v)
			if err != nil {
				logger.Warn("convertIpamAsn error", logger.Err(err), logger.Any("ipamAsn", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			ipamAsns = append(ipamAsns, record)
		}
	}

	return &netbox_goV1.ListIpamAsnByIDsReply{IpamAsns: ipamAsns}, nil
}

// ListByLastID get a paginated list of ipamAsns by last id
func (s *ipamAsn) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamAsnByLastIDRequest) (*netbox_goV1.ListIpamAsnByLastIDReply, error) {
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

	ipamAsns := []*netbox_goV1.IpamAsn{}
	for _, record := range records {
		data, err := convertIpamAsn(record)
		if err != nil {
			logger.Warn("convertIpamAsn error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamAsns = append(ipamAsns, data)
	}

	return &netbox_goV1.ListIpamAsnByLastIDReply{
		IpamAsns: ipamAsns,
	}, nil
}

func convertIpamAsn(record *model.IpamAsn) (*netbox_goV1.IpamAsn, error) {
	value := &netbox_goV1.IpamAsn{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
