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
		netbox_goV1.RegisterVirtualizationVirtualdiskServer(server, NewVirtualizationVirtualdiskServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VirtualizationVirtualdiskServer = (*virtualizationVirtualdisk)(nil)
var _ time.Time

type virtualizationVirtualdisk struct {
	netbox_goV1.UnimplementedVirtualizationVirtualdiskServer

	iDao dao.VirtualizationVirtualdiskDao
}

// NewVirtualizationVirtualdiskServer create a new service
func NewVirtualizationVirtualdiskServer() netbox_goV1.VirtualizationVirtualdiskServer {
	return &virtualizationVirtualdisk{
		iDao: dao.NewVirtualizationVirtualdiskDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVirtualizationVirtualdiskCache(database.GetCacheType()),
		),
	}
}

// Create a new virtualizationVirtualdisk
func (s *virtualizationVirtualdisk) Create(ctx context.Context, req *netbox_goV1.CreateVirtualizationVirtualdiskRequest) (*netbox_goV1.CreateVirtualizationVirtualdiskReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VirtualizationVirtualdisk{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVirtualizationVirtualdisk.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("virtualizationVirtualdisk", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVirtualizationVirtualdiskReply{Id: record.ID}, nil
}

// DeleteByID delete a virtualizationVirtualdisk by id
func (s *virtualizationVirtualdisk) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVirtualizationVirtualdiskByIDRequest) (*netbox_goV1.DeleteVirtualizationVirtualdiskByIDReply, error) {
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

	return &netbox_goV1.DeleteVirtualizationVirtualdiskByIDReply{}, nil
}

// UpdateByID update a virtualizationVirtualdisk by id
func (s *virtualizationVirtualdisk) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVirtualizationVirtualdiskByIDRequest) (*netbox_goV1.UpdateVirtualizationVirtualdiskByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VirtualizationVirtualdisk{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVirtualizationVirtualdisk.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("virtualizationVirtualdisk", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVirtualizationVirtualdiskByIDReply{}, nil
}

// GetByID get a virtualizationVirtualdisk by id
func (s *virtualizationVirtualdisk) GetByID(ctx context.Context, req *netbox_goV1.GetVirtualizationVirtualdiskByIDRequest) (*netbox_goV1.GetVirtualizationVirtualdiskByIDReply, error) {
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

	data, err := convertVirtualizationVirtualdisk(record)
	if err != nil {
		logger.Warn("convertVirtualizationVirtualdisk error", logger.Err(err), logger.Any("virtualizationVirtualdisk", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVirtualizationVirtualdisk.Err()
	}

	return &netbox_goV1.GetVirtualizationVirtualdiskByIDReply{VirtualizationVirtualdisk: data}, nil
}

// List get a paginated list of virtualizationVirtualdisks by custom conditions
func (s *virtualizationVirtualdisk) List(ctx context.Context, req *netbox_goV1.ListVirtualizationVirtualdiskRequest) (*netbox_goV1.ListVirtualizationVirtualdiskReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVirtualizationVirtualdisk.Err()
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

	virtualizationVirtualdisks := []*netbox_goV1.VirtualizationVirtualdisk{}
	for _, record := range records {
		data, err := convertVirtualizationVirtualdisk(record)
		if err != nil {
			logger.Warn("convertVirtualizationVirtualdisk error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		virtualizationVirtualdisks = append(virtualizationVirtualdisks, data)
	}

	return &netbox_goV1.ListVirtualizationVirtualdiskReply{
		Total:                      total,
		VirtualizationVirtualdisks: virtualizationVirtualdisks,
	}, nil
}

// DeleteByIDs batch delete virtualizationVirtualdisk by ids
func (s *virtualizationVirtualdisk) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVirtualizationVirtualdiskByIDsRequest) (*netbox_goV1.DeleteVirtualizationVirtualdiskByIDsReply, error) {
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

	return &netbox_goV1.DeleteVirtualizationVirtualdiskByIDsReply{}, nil
}

// GetByCondition get a virtualizationVirtualdisk by custom condition
func (s *virtualizationVirtualdisk) GetByCondition(ctx context.Context, req *netbox_goV1.GetVirtualizationVirtualdiskByConditionRequest) (*netbox_goV1.GetVirtualizationVirtualdiskByConditionReply, error) {
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

	data, err := convertVirtualizationVirtualdisk(record)
	if err != nil {
		logger.Warn("convertVirtualizationVirtualdisk error", logger.Err(err), logger.Any("virtualizationVirtualdisk", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVirtualizationVirtualdisk.Err()
	}

	return &netbox_goV1.GetVirtualizationVirtualdiskByConditionReply{
		VirtualizationVirtualdisk: data,
	}, nil
}

// ListByIDs batch get virtualizationVirtualdisk by ids
func (s *virtualizationVirtualdisk) ListByIDs(ctx context.Context, req *netbox_goV1.ListVirtualizationVirtualdiskByIDsRequest) (*netbox_goV1.ListVirtualizationVirtualdiskByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	virtualizationVirtualdiskMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	virtualizationVirtualdisks := []*netbox_goV1.VirtualizationVirtualdisk{}
	for _, id := range req.Ids {
		if v, ok := virtualizationVirtualdiskMap[id]; ok {
			record, err := convertVirtualizationVirtualdisk(v)
			if err != nil {
				logger.Warn("convertVirtualizationVirtualdisk error", logger.Err(err), logger.Any("virtualizationVirtualdisk", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			virtualizationVirtualdisks = append(virtualizationVirtualdisks, record)
		}
	}

	return &netbox_goV1.ListVirtualizationVirtualdiskByIDsReply{VirtualizationVirtualdisks: virtualizationVirtualdisks}, nil
}

// ListByLastID get a paginated list of virtualizationVirtualdisks by last id
func (s *virtualizationVirtualdisk) ListByLastID(ctx context.Context, req *netbox_goV1.ListVirtualizationVirtualdiskByLastIDRequest) (*netbox_goV1.ListVirtualizationVirtualdiskByLastIDReply, error) {
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

	virtualizationVirtualdisks := []*netbox_goV1.VirtualizationVirtualdisk{}
	for _, record := range records {
		data, err := convertVirtualizationVirtualdisk(record)
		if err != nil {
			logger.Warn("convertVirtualizationVirtualdisk error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		virtualizationVirtualdisks = append(virtualizationVirtualdisks, data)
	}

	return &netbox_goV1.ListVirtualizationVirtualdiskByLastIDReply{
		VirtualizationVirtualdisks: virtualizationVirtualdisks,
	}, nil
}

func convertVirtualizationVirtualdisk(record *model.VirtualizationVirtualdisk) (*netbox_goV1.VirtualizationVirtualdisk, error) {
	value := &netbox_goV1.VirtualizationVirtualdisk{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
