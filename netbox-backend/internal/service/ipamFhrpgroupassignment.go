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
		netbox_goV1.RegisterIpamFhrpgroupassignmentServer(server, NewIpamFhrpgroupassignmentServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.IpamFhrpgroupassignmentServer = (*ipamFhrpgroupassignment)(nil)
var _ time.Time

type ipamFhrpgroupassignment struct {
	netbox_goV1.UnimplementedIpamFhrpgroupassignmentServer

	iDao dao.IpamFhrpgroupassignmentDao
}

// NewIpamFhrpgroupassignmentServer create a new service
func NewIpamFhrpgroupassignmentServer() netbox_goV1.IpamFhrpgroupassignmentServer {
	return &ipamFhrpgroupassignment{
		iDao: dao.NewIpamFhrpgroupassignmentDao(
			database.GetDB(), // db driver is postgresql
			cache.NewIpamFhrpgroupassignmentCache(database.GetCacheType()),
		),
	}
}

// Create a new ipamFhrpgroupassignment
func (s *ipamFhrpgroupassignment) Create(ctx context.Context, req *netbox_goV1.CreateIpamFhrpgroupassignmentRequest) (*netbox_goV1.CreateIpamFhrpgroupassignmentReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamFhrpgroupassignment{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateIpamFhrpgroupassignment.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("ipamFhrpgroupassignment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateIpamFhrpgroupassignmentReply{Id: record.ID}, nil
}

// DeleteByID delete a ipamFhrpgroupassignment by id
func (s *ipamFhrpgroupassignment) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamFhrpgroupassignmentByIDRequest) (*netbox_goV1.DeleteIpamFhrpgroupassignmentByIDReply, error) {
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

	return &netbox_goV1.DeleteIpamFhrpgroupassignmentByIDReply{}, nil
}

// UpdateByID update a ipamFhrpgroupassignment by id
func (s *ipamFhrpgroupassignment) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamFhrpgroupassignmentByIDRequest) (*netbox_goV1.UpdateIpamFhrpgroupassignmentByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamFhrpgroupassignment{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDIpamFhrpgroupassignment.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("ipamFhrpgroupassignment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateIpamFhrpgroupassignmentByIDReply{}, nil
}

// GetByID get a ipamFhrpgroupassignment by id
func (s *ipamFhrpgroupassignment) GetByID(ctx context.Context, req *netbox_goV1.GetIpamFhrpgroupassignmentByIDRequest) (*netbox_goV1.GetIpamFhrpgroupassignmentByIDReply, error) {
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

	data, err := convertIpamFhrpgroupassignment(record)
	if err != nil {
		logger.Warn("convertIpamFhrpgroupassignment error", logger.Err(err), logger.Any("ipamFhrpgroupassignment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDIpamFhrpgroupassignment.Err()
	}

	return &netbox_goV1.GetIpamFhrpgroupassignmentByIDReply{IpamFhrpgroupassignment: data}, nil
}

// List get a paginated list of ipamFhrpgroupassignments by custom conditions
func (s *ipamFhrpgroupassignment) List(ctx context.Context, req *netbox_goV1.ListIpamFhrpgroupassignmentRequest) (*netbox_goV1.ListIpamFhrpgroupassignmentReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListIpamFhrpgroupassignment.Err()
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

	ipamFhrpgroupassignments := []*netbox_goV1.IpamFhrpgroupassignment{}
	for _, record := range records {
		data, err := convertIpamFhrpgroupassignment(record)
		if err != nil {
			logger.Warn("convertIpamFhrpgroupassignment error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamFhrpgroupassignments = append(ipamFhrpgroupassignments, data)
	}

	return &netbox_goV1.ListIpamFhrpgroupassignmentReply{
		Total:                    total,
		IpamFhrpgroupassignments: ipamFhrpgroupassignments,
	}, nil
}

// DeleteByIDs batch delete ipamFhrpgroupassignment by ids
func (s *ipamFhrpgroupassignment) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamFhrpgroupassignmentByIDsRequest) (*netbox_goV1.DeleteIpamFhrpgroupassignmentByIDsReply, error) {
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

	return &netbox_goV1.DeleteIpamFhrpgroupassignmentByIDsReply{}, nil
}

// GetByCondition get a ipamFhrpgroupassignment by custom condition
func (s *ipamFhrpgroupassignment) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamFhrpgroupassignmentByConditionRequest) (*netbox_goV1.GetIpamFhrpgroupassignmentByConditionReply, error) {
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

	data, err := convertIpamFhrpgroupassignment(record)
	if err != nil {
		logger.Warn("convertIpamFhrpgroupassignment error", logger.Err(err), logger.Any("ipamFhrpgroupassignment", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionIpamFhrpgroupassignment.Err()
	}

	return &netbox_goV1.GetIpamFhrpgroupassignmentByConditionReply{
		IpamFhrpgroupassignment: data,
	}, nil
}

// ListByIDs batch get ipamFhrpgroupassignment by ids
func (s *ipamFhrpgroupassignment) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamFhrpgroupassignmentByIDsRequest) (*netbox_goV1.ListIpamFhrpgroupassignmentByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ipamFhrpgroupassignmentMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	ipamFhrpgroupassignments := []*netbox_goV1.IpamFhrpgroupassignment{}
	for _, id := range req.Ids {
		if v, ok := ipamFhrpgroupassignmentMap[id]; ok {
			record, err := convertIpamFhrpgroupassignment(v)
			if err != nil {
				logger.Warn("convertIpamFhrpgroupassignment error", logger.Err(err), logger.Any("ipamFhrpgroupassignment", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			ipamFhrpgroupassignments = append(ipamFhrpgroupassignments, record)
		}
	}

	return &netbox_goV1.ListIpamFhrpgroupassignmentByIDsReply{IpamFhrpgroupassignments: ipamFhrpgroupassignments}, nil
}

// ListByLastID get a paginated list of ipamFhrpgroupassignments by last id
func (s *ipamFhrpgroupassignment) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamFhrpgroupassignmentByLastIDRequest) (*netbox_goV1.ListIpamFhrpgroupassignmentByLastIDReply, error) {
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

	ipamFhrpgroupassignments := []*netbox_goV1.IpamFhrpgroupassignment{}
	for _, record := range records {
		data, err := convertIpamFhrpgroupassignment(record)
		if err != nil {
			logger.Warn("convertIpamFhrpgroupassignment error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamFhrpgroupassignments = append(ipamFhrpgroupassignments, data)
	}

	return &netbox_goV1.ListIpamFhrpgroupassignmentByLastIDReply{
		IpamFhrpgroupassignments: ipamFhrpgroupassignments,
	}, nil
}

func convertIpamFhrpgroupassignment(record *model.IpamFhrpgroupassignment) (*netbox_goV1.IpamFhrpgroupassignment, error) {
	value := &netbox_goV1.IpamFhrpgroupassignment{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
