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
		netbox_goV1.RegisterIpamVlangroupServer(server, NewIpamVlangroupServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.IpamVlangroupServer = (*ipamVlangroup)(nil)
var _ time.Time

type ipamVlangroup struct {
	netbox_goV1.UnimplementedIpamVlangroupServer

	iDao dao.IpamVlangroupDao
}

// NewIpamVlangroupServer create a new service
func NewIpamVlangroupServer() netbox_goV1.IpamVlangroupServer {
	return &ipamVlangroup{
		iDao: dao.NewIpamVlangroupDao(
			database.GetDB(), // db driver is postgresql
			cache.NewIpamVlangroupCache(database.GetCacheType()),
		),
	}
}

// Create a new ipamVlangroup
func (s *ipamVlangroup) Create(ctx context.Context, req *netbox_goV1.CreateIpamVlangroupRequest) (*netbox_goV1.CreateIpamVlangroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamVlangroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateIpamVlangroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("ipamVlangroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateIpamVlangroupReply{Id: record.ID}, nil
}

// DeleteByID delete a ipamVlangroup by id
func (s *ipamVlangroup) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamVlangroupByIDRequest) (*netbox_goV1.DeleteIpamVlangroupByIDReply, error) {
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

	return &netbox_goV1.DeleteIpamVlangroupByIDReply{}, nil
}

// UpdateByID update a ipamVlangroup by id
func (s *ipamVlangroup) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamVlangroupByIDRequest) (*netbox_goV1.UpdateIpamVlangroupByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamVlangroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDIpamVlangroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("ipamVlangroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateIpamVlangroupByIDReply{}, nil
}

// GetByID get a ipamVlangroup by id
func (s *ipamVlangroup) GetByID(ctx context.Context, req *netbox_goV1.GetIpamVlangroupByIDRequest) (*netbox_goV1.GetIpamVlangroupByIDReply, error) {
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

	data, err := convertIpamVlangroup(record)
	if err != nil {
		logger.Warn("convertIpamVlangroup error", logger.Err(err), logger.Any("ipamVlangroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDIpamVlangroup.Err()
	}

	return &netbox_goV1.GetIpamVlangroupByIDReply{IpamVlangroup: data}, nil
}

// List get a paginated list of ipamVlangroups by custom conditions
func (s *ipamVlangroup) List(ctx context.Context, req *netbox_goV1.ListIpamVlangroupRequest) (*netbox_goV1.ListIpamVlangroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListIpamVlangroup.Err()
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

	ipamVlangroups := []*netbox_goV1.IpamVlangroup{}
	for _, record := range records {
		data, err := convertIpamVlangroup(record)
		if err != nil {
			logger.Warn("convertIpamVlangroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamVlangroups = append(ipamVlangroups, data)
	}

	return &netbox_goV1.ListIpamVlangroupReply{
		Total:          total,
		IpamVlangroups: ipamVlangroups,
	}, nil
}

// DeleteByIDs batch delete ipamVlangroup by ids
func (s *ipamVlangroup) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamVlangroupByIDsRequest) (*netbox_goV1.DeleteIpamVlangroupByIDsReply, error) {
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

	return &netbox_goV1.DeleteIpamVlangroupByIDsReply{}, nil
}

// GetByCondition get a ipamVlangroup by custom condition
func (s *ipamVlangroup) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamVlangroupByConditionRequest) (*netbox_goV1.GetIpamVlangroupByConditionReply, error) {
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

	data, err := convertIpamVlangroup(record)
	if err != nil {
		logger.Warn("convertIpamVlangroup error", logger.Err(err), logger.Any("ipamVlangroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionIpamVlangroup.Err()
	}

	return &netbox_goV1.GetIpamVlangroupByConditionReply{
		IpamVlangroup: data,
	}, nil
}

// ListByIDs batch get ipamVlangroup by ids
func (s *ipamVlangroup) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamVlangroupByIDsRequest) (*netbox_goV1.ListIpamVlangroupByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ipamVlangroupMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	ipamVlangroups := []*netbox_goV1.IpamVlangroup{}
	for _, id := range req.Ids {
		if v, ok := ipamVlangroupMap[id]; ok {
			record, err := convertIpamVlangroup(v)
			if err != nil {
				logger.Warn("convertIpamVlangroup error", logger.Err(err), logger.Any("ipamVlangroup", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			ipamVlangroups = append(ipamVlangroups, record)
		}
	}

	return &netbox_goV1.ListIpamVlangroupByIDsReply{IpamVlangroups: ipamVlangroups}, nil
}

// ListByLastID get a paginated list of ipamVlangroups by last id
func (s *ipamVlangroup) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamVlangroupByLastIDRequest) (*netbox_goV1.ListIpamVlangroupByLastIDReply, error) {
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

	ipamVlangroups := []*netbox_goV1.IpamVlangroup{}
	for _, record := range records {
		data, err := convertIpamVlangroup(record)
		if err != nil {
			logger.Warn("convertIpamVlangroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamVlangroups = append(ipamVlangroups, data)
	}

	return &netbox_goV1.ListIpamVlangroupByLastIDReply{
		IpamVlangroups: ipamVlangroups,
	}, nil
}

func convertIpamVlangroup(record *model.IpamVlangroup) (*netbox_goV1.IpamVlangroup, error) {
	value := &netbox_goV1.IpamVlangroup{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
