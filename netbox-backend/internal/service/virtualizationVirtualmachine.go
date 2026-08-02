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
		netbox_goV1.RegisterVirtualizationVirtualmachineServer(server, NewVirtualizationVirtualmachineServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VirtualizationVirtualmachineServer = (*virtualizationVirtualmachine)(nil)
var _ time.Time

type virtualizationVirtualmachine struct {
	netbox_goV1.UnimplementedVirtualizationVirtualmachineServer

	iDao dao.VirtualizationVirtualmachineDao
}

// NewVirtualizationVirtualmachineServer create a new service
func NewVirtualizationVirtualmachineServer() netbox_goV1.VirtualizationVirtualmachineServer {
	return &virtualizationVirtualmachine{
		iDao: dao.NewVirtualizationVirtualmachineDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVirtualizationVirtualmachineCache(database.GetCacheType()),
		),
	}
}

// Create a new virtualizationVirtualmachine
func (s *virtualizationVirtualmachine) Create(ctx context.Context, req *netbox_goV1.CreateVirtualizationVirtualmachineRequest) (*netbox_goV1.CreateVirtualizationVirtualmachineReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VirtualizationVirtualmachine{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVirtualizationVirtualmachine.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("virtualizationVirtualmachine", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVirtualizationVirtualmachineReply{Id: record.ID}, nil
}

// DeleteByID delete a virtualizationVirtualmachine by id
func (s *virtualizationVirtualmachine) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVirtualizationVirtualmachineByIDRequest) (*netbox_goV1.DeleteVirtualizationVirtualmachineByIDReply, error) {
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

	return &netbox_goV1.DeleteVirtualizationVirtualmachineByIDReply{}, nil
}

// UpdateByID update a virtualizationVirtualmachine by id
func (s *virtualizationVirtualmachine) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVirtualizationVirtualmachineByIDRequest) (*netbox_goV1.UpdateVirtualizationVirtualmachineByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VirtualizationVirtualmachine{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVirtualizationVirtualmachine.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("virtualizationVirtualmachine", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVirtualizationVirtualmachineByIDReply{}, nil
}

// GetByID get a virtualizationVirtualmachine by id
func (s *virtualizationVirtualmachine) GetByID(ctx context.Context, req *netbox_goV1.GetVirtualizationVirtualmachineByIDRequest) (*netbox_goV1.GetVirtualizationVirtualmachineByIDReply, error) {
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

	data, err := convertVirtualizationVirtualmachine(record)
	if err != nil {
		logger.Warn("convertVirtualizationVirtualmachine error", logger.Err(err), logger.Any("virtualizationVirtualmachine", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVirtualizationVirtualmachine.Err()
	}

	return &netbox_goV1.GetVirtualizationVirtualmachineByIDReply{VirtualizationVirtualmachine: data}, nil
}

// List get a paginated list of virtualizationVirtualmachines by custom conditions
func (s *virtualizationVirtualmachine) List(ctx context.Context, req *netbox_goV1.ListVirtualizationVirtualmachineRequest) (*netbox_goV1.ListVirtualizationVirtualmachineReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVirtualizationVirtualmachine.Err()
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

	virtualizationVirtualmachines := []*netbox_goV1.VirtualizationVirtualmachine{}
	for _, record := range records {
		data, err := convertVirtualizationVirtualmachine(record)
		if err != nil {
			logger.Warn("convertVirtualizationVirtualmachine error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		virtualizationVirtualmachines = append(virtualizationVirtualmachines, data)
	}

	return &netbox_goV1.ListVirtualizationVirtualmachineReply{
		Total:                         total,
		VirtualizationVirtualmachines: virtualizationVirtualmachines,
	}, nil
}

// DeleteByIDs batch delete virtualizationVirtualmachine by ids
func (s *virtualizationVirtualmachine) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVirtualizationVirtualmachineByIDsRequest) (*netbox_goV1.DeleteVirtualizationVirtualmachineByIDsReply, error) {
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

	return &netbox_goV1.DeleteVirtualizationVirtualmachineByIDsReply{}, nil
}

// GetByCondition get a virtualizationVirtualmachine by custom condition
func (s *virtualizationVirtualmachine) GetByCondition(ctx context.Context, req *netbox_goV1.GetVirtualizationVirtualmachineByConditionRequest) (*netbox_goV1.GetVirtualizationVirtualmachineByConditionReply, error) {
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

	data, err := convertVirtualizationVirtualmachine(record)
	if err != nil {
		logger.Warn("convertVirtualizationVirtualmachine error", logger.Err(err), logger.Any("virtualizationVirtualmachine", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVirtualizationVirtualmachine.Err()
	}

	return &netbox_goV1.GetVirtualizationVirtualmachineByConditionReply{
		VirtualizationVirtualmachine: data,
	}, nil
}

// ListByIDs batch get virtualizationVirtualmachine by ids
func (s *virtualizationVirtualmachine) ListByIDs(ctx context.Context, req *netbox_goV1.ListVirtualizationVirtualmachineByIDsRequest) (*netbox_goV1.ListVirtualizationVirtualmachineByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	virtualizationVirtualmachineMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	virtualizationVirtualmachines := []*netbox_goV1.VirtualizationVirtualmachine{}
	for _, id := range req.Ids {
		if v, ok := virtualizationVirtualmachineMap[id]; ok {
			record, err := convertVirtualizationVirtualmachine(v)
			if err != nil {
				logger.Warn("convertVirtualizationVirtualmachine error", logger.Err(err), logger.Any("virtualizationVirtualmachine", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			virtualizationVirtualmachines = append(virtualizationVirtualmachines, record)
		}
	}

	return &netbox_goV1.ListVirtualizationVirtualmachineByIDsReply{VirtualizationVirtualmachines: virtualizationVirtualmachines}, nil
}

// ListByLastID get a paginated list of virtualizationVirtualmachines by last id
func (s *virtualizationVirtualmachine) ListByLastID(ctx context.Context, req *netbox_goV1.ListVirtualizationVirtualmachineByLastIDRequest) (*netbox_goV1.ListVirtualizationVirtualmachineByLastIDReply, error) {
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

	virtualizationVirtualmachines := []*netbox_goV1.VirtualizationVirtualmachine{}
	for _, record := range records {
		data, err := convertVirtualizationVirtualmachine(record)
		if err != nil {
			logger.Warn("convertVirtualizationVirtualmachine error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		virtualizationVirtualmachines = append(virtualizationVirtualmachines, data)
	}

	return &netbox_goV1.ListVirtualizationVirtualmachineByLastIDReply{
		VirtualizationVirtualmachines: virtualizationVirtualmachines,
	}, nil
}

func convertVirtualizationVirtualmachine(record *model.VirtualizationVirtualmachine) (*netbox_goV1.VirtualizationVirtualmachine, error) {
	value := &netbox_goV1.VirtualizationVirtualmachine{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
