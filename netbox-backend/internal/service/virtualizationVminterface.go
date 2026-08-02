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
		netbox_goV1.RegisterVirtualizationVminterfaceServer(server, NewVirtualizationVminterfaceServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VirtualizationVminterfaceServer = (*virtualizationVminterface)(nil)
var _ time.Time

type virtualizationVminterface struct {
	netbox_goV1.UnimplementedVirtualizationVminterfaceServer

	iDao dao.VirtualizationVminterfaceDao
}

// NewVirtualizationVminterfaceServer create a new service
func NewVirtualizationVminterfaceServer() netbox_goV1.VirtualizationVminterfaceServer {
	return &virtualizationVminterface{
		iDao: dao.NewVirtualizationVminterfaceDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVirtualizationVminterfaceCache(database.GetCacheType()),
		),
	}
}

// Create a new virtualizationVminterface
func (s *virtualizationVminterface) Create(ctx context.Context, req *netbox_goV1.CreateVirtualizationVminterfaceRequest) (*netbox_goV1.CreateVirtualizationVminterfaceReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VirtualizationVminterface{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVirtualizationVminterface.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("virtualizationVminterface", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVirtualizationVminterfaceReply{Id: record.ID}, nil
}

// DeleteByID delete a virtualizationVminterface by id
func (s *virtualizationVminterface) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVirtualizationVminterfaceByIDRequest) (*netbox_goV1.DeleteVirtualizationVminterfaceByIDReply, error) {
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

	return &netbox_goV1.DeleteVirtualizationVminterfaceByIDReply{}, nil
}

// UpdateByID update a virtualizationVminterface by id
func (s *virtualizationVminterface) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVirtualizationVminterfaceByIDRequest) (*netbox_goV1.UpdateVirtualizationVminterfaceByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VirtualizationVminterface{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVirtualizationVminterface.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("virtualizationVminterface", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVirtualizationVminterfaceByIDReply{}, nil
}

// GetByID get a virtualizationVminterface by id
func (s *virtualizationVminterface) GetByID(ctx context.Context, req *netbox_goV1.GetVirtualizationVminterfaceByIDRequest) (*netbox_goV1.GetVirtualizationVminterfaceByIDReply, error) {
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

	data, err := convertVirtualizationVminterface(record)
	if err != nil {
		logger.Warn("convertVirtualizationVminterface error", logger.Err(err), logger.Any("virtualizationVminterface", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVirtualizationVminterface.Err()
	}

	return &netbox_goV1.GetVirtualizationVminterfaceByIDReply{VirtualizationVminterface: data}, nil
}

// List get a paginated list of virtualizationVminterfaces by custom conditions
func (s *virtualizationVminterface) List(ctx context.Context, req *netbox_goV1.ListVirtualizationVminterfaceRequest) (*netbox_goV1.ListVirtualizationVminterfaceReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVirtualizationVminterface.Err()
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

	virtualizationVminterfaces := []*netbox_goV1.VirtualizationVminterface{}
	for _, record := range records {
		data, err := convertVirtualizationVminterface(record)
		if err != nil {
			logger.Warn("convertVirtualizationVminterface error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		virtualizationVminterfaces = append(virtualizationVminterfaces, data)
	}

	return &netbox_goV1.ListVirtualizationVminterfaceReply{
		Total:                      total,
		VirtualizationVminterfaces: virtualizationVminterfaces,
	}, nil
}

// DeleteByIDs batch delete virtualizationVminterface by ids
func (s *virtualizationVminterface) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVirtualizationVminterfaceByIDsRequest) (*netbox_goV1.DeleteVirtualizationVminterfaceByIDsReply, error) {
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

	return &netbox_goV1.DeleteVirtualizationVminterfaceByIDsReply{}, nil
}

// GetByCondition get a virtualizationVminterface by custom condition
func (s *virtualizationVminterface) GetByCondition(ctx context.Context, req *netbox_goV1.GetVirtualizationVminterfaceByConditionRequest) (*netbox_goV1.GetVirtualizationVminterfaceByConditionReply, error) {
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

	data, err := convertVirtualizationVminterface(record)
	if err != nil {
		logger.Warn("convertVirtualizationVminterface error", logger.Err(err), logger.Any("virtualizationVminterface", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVirtualizationVminterface.Err()
	}

	return &netbox_goV1.GetVirtualizationVminterfaceByConditionReply{
		VirtualizationVminterface: data,
	}, nil
}

// ListByIDs batch get virtualizationVminterface by ids
func (s *virtualizationVminterface) ListByIDs(ctx context.Context, req *netbox_goV1.ListVirtualizationVminterfaceByIDsRequest) (*netbox_goV1.ListVirtualizationVminterfaceByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	virtualizationVminterfaceMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	virtualizationVminterfaces := []*netbox_goV1.VirtualizationVminterface{}
	for _, id := range req.Ids {
		if v, ok := virtualizationVminterfaceMap[id]; ok {
			record, err := convertVirtualizationVminterface(v)
			if err != nil {
				logger.Warn("convertVirtualizationVminterface error", logger.Err(err), logger.Any("virtualizationVminterface", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			virtualizationVminterfaces = append(virtualizationVminterfaces, record)
		}
	}

	return &netbox_goV1.ListVirtualizationVminterfaceByIDsReply{VirtualizationVminterfaces: virtualizationVminterfaces}, nil
}

// ListByLastID get a paginated list of virtualizationVminterfaces by last id
func (s *virtualizationVminterface) ListByLastID(ctx context.Context, req *netbox_goV1.ListVirtualizationVminterfaceByLastIDRequest) (*netbox_goV1.ListVirtualizationVminterfaceByLastIDReply, error) {
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

	virtualizationVminterfaces := []*netbox_goV1.VirtualizationVminterface{}
	for _, record := range records {
		data, err := convertVirtualizationVminterface(record)
		if err != nil {
			logger.Warn("convertVirtualizationVminterface error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		virtualizationVminterfaces = append(virtualizationVminterfaces, data)
	}

	return &netbox_goV1.ListVirtualizationVminterfaceByLastIDReply{
		VirtualizationVminterfaces: virtualizationVminterfaces,
	}, nil
}

func convertVirtualizationVminterface(record *model.VirtualizationVminterface) (*netbox_goV1.VirtualizationVminterface, error) {
	value := &netbox_goV1.VirtualizationVminterface{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
