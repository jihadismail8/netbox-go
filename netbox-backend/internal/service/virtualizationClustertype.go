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
		netbox_goV1.RegisterVirtualizationClustertypeServer(server, NewVirtualizationClustertypeServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VirtualizationClustertypeServer = (*virtualizationClustertype)(nil)
var _ time.Time

type virtualizationClustertype struct {
	netbox_goV1.UnimplementedVirtualizationClustertypeServer

	iDao dao.VirtualizationClustertypeDao
}

// NewVirtualizationClustertypeServer create a new service
func NewVirtualizationClustertypeServer() netbox_goV1.VirtualizationClustertypeServer {
	return &virtualizationClustertype{
		iDao: dao.NewVirtualizationClustertypeDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVirtualizationClustertypeCache(database.GetCacheType()),
		),
	}
}

// Create a new virtualizationClustertype
func (s *virtualizationClustertype) Create(ctx context.Context, req *netbox_goV1.CreateVirtualizationClustertypeRequest) (*netbox_goV1.CreateVirtualizationClustertypeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VirtualizationClustertype{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVirtualizationClustertype.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("virtualizationClustertype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVirtualizationClustertypeReply{Id: record.ID}, nil
}

// DeleteByID delete a virtualizationClustertype by id
func (s *virtualizationClustertype) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVirtualizationClustertypeByIDRequest) (*netbox_goV1.DeleteVirtualizationClustertypeByIDReply, error) {
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

	return &netbox_goV1.DeleteVirtualizationClustertypeByIDReply{}, nil
}

// UpdateByID update a virtualizationClustertype by id
func (s *virtualizationClustertype) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVirtualizationClustertypeByIDRequest) (*netbox_goV1.UpdateVirtualizationClustertypeByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VirtualizationClustertype{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVirtualizationClustertype.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("virtualizationClustertype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVirtualizationClustertypeByIDReply{}, nil
}

// GetByID get a virtualizationClustertype by id
func (s *virtualizationClustertype) GetByID(ctx context.Context, req *netbox_goV1.GetVirtualizationClustertypeByIDRequest) (*netbox_goV1.GetVirtualizationClustertypeByIDReply, error) {
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

	data, err := convertVirtualizationClustertype(record)
	if err != nil {
		logger.Warn("convertVirtualizationClustertype error", logger.Err(err), logger.Any("virtualizationClustertype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVirtualizationClustertype.Err()
	}

	return &netbox_goV1.GetVirtualizationClustertypeByIDReply{VirtualizationClustertype: data}, nil
}

// List get a paginated list of virtualizationClustertypes by custom conditions
func (s *virtualizationClustertype) List(ctx context.Context, req *netbox_goV1.ListVirtualizationClustertypeRequest) (*netbox_goV1.ListVirtualizationClustertypeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVirtualizationClustertype.Err()
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

	virtualizationClustertypes := []*netbox_goV1.VirtualizationClustertype{}
	for _, record := range records {
		data, err := convertVirtualizationClustertype(record)
		if err != nil {
			logger.Warn("convertVirtualizationClustertype error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		virtualizationClustertypes = append(virtualizationClustertypes, data)
	}

	return &netbox_goV1.ListVirtualizationClustertypeReply{
		Total:                      total,
		VirtualizationClustertypes: virtualizationClustertypes,
	}, nil
}

// DeleteByIDs batch delete virtualizationClustertype by ids
func (s *virtualizationClustertype) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVirtualizationClustertypeByIDsRequest) (*netbox_goV1.DeleteVirtualizationClustertypeByIDsReply, error) {
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

	return &netbox_goV1.DeleteVirtualizationClustertypeByIDsReply{}, nil
}

// GetByCondition get a virtualizationClustertype by custom condition
func (s *virtualizationClustertype) GetByCondition(ctx context.Context, req *netbox_goV1.GetVirtualizationClustertypeByConditionRequest) (*netbox_goV1.GetVirtualizationClustertypeByConditionReply, error) {
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

	data, err := convertVirtualizationClustertype(record)
	if err != nil {
		logger.Warn("convertVirtualizationClustertype error", logger.Err(err), logger.Any("virtualizationClustertype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVirtualizationClustertype.Err()
	}

	return &netbox_goV1.GetVirtualizationClustertypeByConditionReply{
		VirtualizationClustertype: data,
	}, nil
}

// ListByIDs batch get virtualizationClustertype by ids
func (s *virtualizationClustertype) ListByIDs(ctx context.Context, req *netbox_goV1.ListVirtualizationClustertypeByIDsRequest) (*netbox_goV1.ListVirtualizationClustertypeByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	virtualizationClustertypeMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	virtualizationClustertypes := []*netbox_goV1.VirtualizationClustertype{}
	for _, id := range req.Ids {
		if v, ok := virtualizationClustertypeMap[id]; ok {
			record, err := convertVirtualizationClustertype(v)
			if err != nil {
				logger.Warn("convertVirtualizationClustertype error", logger.Err(err), logger.Any("virtualizationClustertype", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			virtualizationClustertypes = append(virtualizationClustertypes, record)
		}
	}

	return &netbox_goV1.ListVirtualizationClustertypeByIDsReply{VirtualizationClustertypes: virtualizationClustertypes}, nil
}

// ListByLastID get a paginated list of virtualizationClustertypes by last id
func (s *virtualizationClustertype) ListByLastID(ctx context.Context, req *netbox_goV1.ListVirtualizationClustertypeByLastIDRequest) (*netbox_goV1.ListVirtualizationClustertypeByLastIDReply, error) {
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

	virtualizationClustertypes := []*netbox_goV1.VirtualizationClustertype{}
	for _, record := range records {
		data, err := convertVirtualizationClustertype(record)
		if err != nil {
			logger.Warn("convertVirtualizationClustertype error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		virtualizationClustertypes = append(virtualizationClustertypes, data)
	}

	return &netbox_goV1.ListVirtualizationClustertypeByLastIDReply{
		VirtualizationClustertypes: virtualizationClustertypes,
	}, nil
}

func convertVirtualizationClustertype(record *model.VirtualizationClustertype) (*netbox_goV1.VirtualizationClustertype, error) {
	value := &netbox_goV1.VirtualizationClustertype{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
