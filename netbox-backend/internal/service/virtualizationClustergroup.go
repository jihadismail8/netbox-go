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
		netbox_goV1.RegisterVirtualizationClustergroupServer(server, NewVirtualizationClustergroupServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VirtualizationClustergroupServer = (*virtualizationClustergroup)(nil)
var _ time.Time

type virtualizationClustergroup struct {
	netbox_goV1.UnimplementedVirtualizationClustergroupServer

	iDao dao.VirtualizationClustergroupDao
}

// NewVirtualizationClustergroupServer create a new service
func NewVirtualizationClustergroupServer() netbox_goV1.VirtualizationClustergroupServer {
	return &virtualizationClustergroup{
		iDao: dao.NewVirtualizationClustergroupDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVirtualizationClustergroupCache(database.GetCacheType()),
		),
	}
}

// Create a new virtualizationClustergroup
func (s *virtualizationClustergroup) Create(ctx context.Context, req *netbox_goV1.CreateVirtualizationClustergroupRequest) (*netbox_goV1.CreateVirtualizationClustergroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VirtualizationClustergroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVirtualizationClustergroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("virtualizationClustergroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVirtualizationClustergroupReply{Id: record.ID}, nil
}

// DeleteByID delete a virtualizationClustergroup by id
func (s *virtualizationClustergroup) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVirtualizationClustergroupByIDRequest) (*netbox_goV1.DeleteVirtualizationClustergroupByIDReply, error) {
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

	return &netbox_goV1.DeleteVirtualizationClustergroupByIDReply{}, nil
}

// UpdateByID update a virtualizationClustergroup by id
func (s *virtualizationClustergroup) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVirtualizationClustergroupByIDRequest) (*netbox_goV1.UpdateVirtualizationClustergroupByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VirtualizationClustergroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVirtualizationClustergroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("virtualizationClustergroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVirtualizationClustergroupByIDReply{}, nil
}

// GetByID get a virtualizationClustergroup by id
func (s *virtualizationClustergroup) GetByID(ctx context.Context, req *netbox_goV1.GetVirtualizationClustergroupByIDRequest) (*netbox_goV1.GetVirtualizationClustergroupByIDReply, error) {
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

	data, err := convertVirtualizationClustergroup(record)
	if err != nil {
		logger.Warn("convertVirtualizationClustergroup error", logger.Err(err), logger.Any("virtualizationClustergroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVirtualizationClustergroup.Err()
	}

	return &netbox_goV1.GetVirtualizationClustergroupByIDReply{VirtualizationClustergroup: data}, nil
}

// List get a paginated list of virtualizationClustergroups by custom conditions
func (s *virtualizationClustergroup) List(ctx context.Context, req *netbox_goV1.ListVirtualizationClustergroupRequest) (*netbox_goV1.ListVirtualizationClustergroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVirtualizationClustergroup.Err()
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

	virtualizationClustergroups := []*netbox_goV1.VirtualizationClustergroup{}
	for _, record := range records {
		data, err := convertVirtualizationClustergroup(record)
		if err != nil {
			logger.Warn("convertVirtualizationClustergroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		virtualizationClustergroups = append(virtualizationClustergroups, data)
	}

	return &netbox_goV1.ListVirtualizationClustergroupReply{
		Total:                       total,
		VirtualizationClustergroups: virtualizationClustergroups,
	}, nil
}

// DeleteByIDs batch delete virtualizationClustergroup by ids
func (s *virtualizationClustergroup) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVirtualizationClustergroupByIDsRequest) (*netbox_goV1.DeleteVirtualizationClustergroupByIDsReply, error) {
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

	return &netbox_goV1.DeleteVirtualizationClustergroupByIDsReply{}, nil
}

// GetByCondition get a virtualizationClustergroup by custom condition
func (s *virtualizationClustergroup) GetByCondition(ctx context.Context, req *netbox_goV1.GetVirtualizationClustergroupByConditionRequest) (*netbox_goV1.GetVirtualizationClustergroupByConditionReply, error) {
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

	data, err := convertVirtualizationClustergroup(record)
	if err != nil {
		logger.Warn("convertVirtualizationClustergroup error", logger.Err(err), logger.Any("virtualizationClustergroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVirtualizationClustergroup.Err()
	}

	return &netbox_goV1.GetVirtualizationClustergroupByConditionReply{
		VirtualizationClustergroup: data,
	}, nil
}

// ListByIDs batch get virtualizationClustergroup by ids
func (s *virtualizationClustergroup) ListByIDs(ctx context.Context, req *netbox_goV1.ListVirtualizationClustergroupByIDsRequest) (*netbox_goV1.ListVirtualizationClustergroupByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	virtualizationClustergroupMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	virtualizationClustergroups := []*netbox_goV1.VirtualizationClustergroup{}
	for _, id := range req.Ids {
		if v, ok := virtualizationClustergroupMap[id]; ok {
			record, err := convertVirtualizationClustergroup(v)
			if err != nil {
				logger.Warn("convertVirtualizationClustergroup error", logger.Err(err), logger.Any("virtualizationClustergroup", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			virtualizationClustergroups = append(virtualizationClustergroups, record)
		}
	}

	return &netbox_goV1.ListVirtualizationClustergroupByIDsReply{VirtualizationClustergroups: virtualizationClustergroups}, nil
}

// ListByLastID get a paginated list of virtualizationClustergroups by last id
func (s *virtualizationClustergroup) ListByLastID(ctx context.Context, req *netbox_goV1.ListVirtualizationClustergroupByLastIDRequest) (*netbox_goV1.ListVirtualizationClustergroupByLastIDReply, error) {
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

	virtualizationClustergroups := []*netbox_goV1.VirtualizationClustergroup{}
	for _, record := range records {
		data, err := convertVirtualizationClustergroup(record)
		if err != nil {
			logger.Warn("convertVirtualizationClustergroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		virtualizationClustergroups = append(virtualizationClustergroups, data)
	}

	return &netbox_goV1.ListVirtualizationClustergroupByLastIDReply{
		VirtualizationClustergroups: virtualizationClustergroups,
	}, nil
}

func convertVirtualizationClustergroup(record *model.VirtualizationClustergroup) (*netbox_goV1.VirtualizationClustergroup, error) {
	value := &netbox_goV1.VirtualizationClustergroup{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
