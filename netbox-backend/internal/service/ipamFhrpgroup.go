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
		netbox_goV1.RegisterIpamFhrpgroupServer(server, NewIpamFhrpgroupServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.IpamFhrpgroupServer = (*ipamFhrpgroup)(nil)
var _ time.Time

type ipamFhrpgroup struct {
	netbox_goV1.UnimplementedIpamFhrpgroupServer

	iDao dao.IpamFhrpgroupDao
}

// NewIpamFhrpgroupServer create a new service
func NewIpamFhrpgroupServer() netbox_goV1.IpamFhrpgroupServer {
	return &ipamFhrpgroup{
		iDao: dao.NewIpamFhrpgroupDao(
			database.GetDB(), // db driver is postgresql
			cache.NewIpamFhrpgroupCache(database.GetCacheType()),
		),
	}
}

// Create a new ipamFhrpgroup
func (s *ipamFhrpgroup) Create(ctx context.Context, req *netbox_goV1.CreateIpamFhrpgroupRequest) (*netbox_goV1.CreateIpamFhrpgroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamFhrpgroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateIpamFhrpgroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("ipamFhrpgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateIpamFhrpgroupReply{Id: record.ID}, nil
}

// DeleteByID delete a ipamFhrpgroup by id
func (s *ipamFhrpgroup) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamFhrpgroupByIDRequest) (*netbox_goV1.DeleteIpamFhrpgroupByIDReply, error) {
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

	return &netbox_goV1.DeleteIpamFhrpgroupByIDReply{}, nil
}

// UpdateByID update a ipamFhrpgroup by id
func (s *ipamFhrpgroup) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamFhrpgroupByIDRequest) (*netbox_goV1.UpdateIpamFhrpgroupByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamFhrpgroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDIpamFhrpgroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("ipamFhrpgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateIpamFhrpgroupByIDReply{}, nil
}

// GetByID get a ipamFhrpgroup by id
func (s *ipamFhrpgroup) GetByID(ctx context.Context, req *netbox_goV1.GetIpamFhrpgroupByIDRequest) (*netbox_goV1.GetIpamFhrpgroupByIDReply, error) {
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

	data, err := convertIpamFhrpgroup(record)
	if err != nil {
		logger.Warn("convertIpamFhrpgroup error", logger.Err(err), logger.Any("ipamFhrpgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDIpamFhrpgroup.Err()
	}

	return &netbox_goV1.GetIpamFhrpgroupByIDReply{IpamFhrpgroup: data}, nil
}

// List get a paginated list of ipamFhrpgroups by custom conditions
func (s *ipamFhrpgroup) List(ctx context.Context, req *netbox_goV1.ListIpamFhrpgroupRequest) (*netbox_goV1.ListIpamFhrpgroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListIpamFhrpgroup.Err()
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

	ipamFhrpgroups := []*netbox_goV1.IpamFhrpgroup{}
	for _, record := range records {
		data, err := convertIpamFhrpgroup(record)
		if err != nil {
			logger.Warn("convertIpamFhrpgroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamFhrpgroups = append(ipamFhrpgroups, data)
	}

	return &netbox_goV1.ListIpamFhrpgroupReply{
		Total:          total,
		IpamFhrpgroups: ipamFhrpgroups,
	}, nil
}

// DeleteByIDs batch delete ipamFhrpgroup by ids
func (s *ipamFhrpgroup) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamFhrpgroupByIDsRequest) (*netbox_goV1.DeleteIpamFhrpgroupByIDsReply, error) {
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

	return &netbox_goV1.DeleteIpamFhrpgroupByIDsReply{}, nil
}

// GetByCondition get a ipamFhrpgroup by custom condition
func (s *ipamFhrpgroup) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamFhrpgroupByConditionRequest) (*netbox_goV1.GetIpamFhrpgroupByConditionReply, error) {
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

	data, err := convertIpamFhrpgroup(record)
	if err != nil {
		logger.Warn("convertIpamFhrpgroup error", logger.Err(err), logger.Any("ipamFhrpgroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionIpamFhrpgroup.Err()
	}

	return &netbox_goV1.GetIpamFhrpgroupByConditionReply{
		IpamFhrpgroup: data,
	}, nil
}

// ListByIDs batch get ipamFhrpgroup by ids
func (s *ipamFhrpgroup) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamFhrpgroupByIDsRequest) (*netbox_goV1.ListIpamFhrpgroupByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ipamFhrpgroupMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	ipamFhrpgroups := []*netbox_goV1.IpamFhrpgroup{}
	for _, id := range req.Ids {
		if v, ok := ipamFhrpgroupMap[id]; ok {
			record, err := convertIpamFhrpgroup(v)
			if err != nil {
				logger.Warn("convertIpamFhrpgroup error", logger.Err(err), logger.Any("ipamFhrpgroup", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			ipamFhrpgroups = append(ipamFhrpgroups, record)
		}
	}

	return &netbox_goV1.ListIpamFhrpgroupByIDsReply{IpamFhrpgroups: ipamFhrpgroups}, nil
}

// ListByLastID get a paginated list of ipamFhrpgroups by last id
func (s *ipamFhrpgroup) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamFhrpgroupByLastIDRequest) (*netbox_goV1.ListIpamFhrpgroupByLastIDReply, error) {
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

	ipamFhrpgroups := []*netbox_goV1.IpamFhrpgroup{}
	for _, record := range records {
		data, err := convertIpamFhrpgroup(record)
		if err != nil {
			logger.Warn("convertIpamFhrpgroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamFhrpgroups = append(ipamFhrpgroups, data)
	}

	return &netbox_goV1.ListIpamFhrpgroupByLastIDReply{
		IpamFhrpgroups: ipamFhrpgroups,
	}, nil
}

func convertIpamFhrpgroup(record *model.IpamFhrpgroup) (*netbox_goV1.IpamFhrpgroup, error) {
	value := &netbox_goV1.IpamFhrpgroup{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
