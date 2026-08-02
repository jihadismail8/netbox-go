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
		netbox_goV1.RegisterVirtualizationClusterServer(server, NewVirtualizationClusterServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VirtualizationClusterServer = (*virtualizationCluster)(nil)
var _ time.Time

type virtualizationCluster struct {
	netbox_goV1.UnimplementedVirtualizationClusterServer

	iDao dao.VirtualizationClusterDao
}

// NewVirtualizationClusterServer create a new service
func NewVirtualizationClusterServer() netbox_goV1.VirtualizationClusterServer {
	return &virtualizationCluster{
		iDao: dao.NewVirtualizationClusterDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVirtualizationClusterCache(database.GetCacheType()),
		),
	}
}

// Create a new virtualizationCluster
func (s *virtualizationCluster) Create(ctx context.Context, req *netbox_goV1.CreateVirtualizationClusterRequest) (*netbox_goV1.CreateVirtualizationClusterReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VirtualizationCluster{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVirtualizationCluster.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("virtualizationCluster", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVirtualizationClusterReply{Id: record.ID}, nil
}

// DeleteByID delete a virtualizationCluster by id
func (s *virtualizationCluster) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVirtualizationClusterByIDRequest) (*netbox_goV1.DeleteVirtualizationClusterByIDReply, error) {
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

	return &netbox_goV1.DeleteVirtualizationClusterByIDReply{}, nil
}

// UpdateByID update a virtualizationCluster by id
func (s *virtualizationCluster) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVirtualizationClusterByIDRequest) (*netbox_goV1.UpdateVirtualizationClusterByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VirtualizationCluster{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVirtualizationCluster.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("virtualizationCluster", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVirtualizationClusterByIDReply{}, nil
}

// GetByID get a virtualizationCluster by id
func (s *virtualizationCluster) GetByID(ctx context.Context, req *netbox_goV1.GetVirtualizationClusterByIDRequest) (*netbox_goV1.GetVirtualizationClusterByIDReply, error) {
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

	data, err := convertVirtualizationCluster(record)
	if err != nil {
		logger.Warn("convertVirtualizationCluster error", logger.Err(err), logger.Any("virtualizationCluster", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVirtualizationCluster.Err()
	}

	return &netbox_goV1.GetVirtualizationClusterByIDReply{VirtualizationCluster: data}, nil
}

// List get a paginated list of virtualizationClusters by custom conditions
func (s *virtualizationCluster) List(ctx context.Context, req *netbox_goV1.ListVirtualizationClusterRequest) (*netbox_goV1.ListVirtualizationClusterReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVirtualizationCluster.Err()
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

	virtualizationClusters := []*netbox_goV1.VirtualizationCluster{}
	for _, record := range records {
		data, err := convertVirtualizationCluster(record)
		if err != nil {
			logger.Warn("convertVirtualizationCluster error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		virtualizationClusters = append(virtualizationClusters, data)
	}

	return &netbox_goV1.ListVirtualizationClusterReply{
		Total:                  total,
		VirtualizationClusters: virtualizationClusters,
	}, nil
}

// DeleteByIDs batch delete virtualizationCluster by ids
func (s *virtualizationCluster) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVirtualizationClusterByIDsRequest) (*netbox_goV1.DeleteVirtualizationClusterByIDsReply, error) {
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

	return &netbox_goV1.DeleteVirtualizationClusterByIDsReply{}, nil
}

// GetByCondition get a virtualizationCluster by custom condition
func (s *virtualizationCluster) GetByCondition(ctx context.Context, req *netbox_goV1.GetVirtualizationClusterByConditionRequest) (*netbox_goV1.GetVirtualizationClusterByConditionReply, error) {
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

	data, err := convertVirtualizationCluster(record)
	if err != nil {
		logger.Warn("convertVirtualizationCluster error", logger.Err(err), logger.Any("virtualizationCluster", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVirtualizationCluster.Err()
	}

	return &netbox_goV1.GetVirtualizationClusterByConditionReply{
		VirtualizationCluster: data,
	}, nil
}

// ListByIDs batch get virtualizationCluster by ids
func (s *virtualizationCluster) ListByIDs(ctx context.Context, req *netbox_goV1.ListVirtualizationClusterByIDsRequest) (*netbox_goV1.ListVirtualizationClusterByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	virtualizationClusterMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	virtualizationClusters := []*netbox_goV1.VirtualizationCluster{}
	for _, id := range req.Ids {
		if v, ok := virtualizationClusterMap[id]; ok {
			record, err := convertVirtualizationCluster(v)
			if err != nil {
				logger.Warn("convertVirtualizationCluster error", logger.Err(err), logger.Any("virtualizationCluster", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			virtualizationClusters = append(virtualizationClusters, record)
		}
	}

	return &netbox_goV1.ListVirtualizationClusterByIDsReply{VirtualizationClusters: virtualizationClusters}, nil
}

// ListByLastID get a paginated list of virtualizationClusters by last id
func (s *virtualizationCluster) ListByLastID(ctx context.Context, req *netbox_goV1.ListVirtualizationClusterByLastIDRequest) (*netbox_goV1.ListVirtualizationClusterByLastIDReply, error) {
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

	virtualizationClusters := []*netbox_goV1.VirtualizationCluster{}
	for _, record := range records {
		data, err := convertVirtualizationCluster(record)
		if err != nil {
			logger.Warn("convertVirtualizationCluster error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		virtualizationClusters = append(virtualizationClusters, data)
	}

	return &netbox_goV1.ListVirtualizationClusterByLastIDReply{
		VirtualizationClusters: virtualizationClusters,
	}, nil
}

func convertVirtualizationCluster(record *model.VirtualizationCluster) (*netbox_goV1.VirtualizationCluster, error) {
	value := &netbox_goV1.VirtualizationCluster{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
