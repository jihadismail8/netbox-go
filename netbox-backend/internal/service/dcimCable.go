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
		netbox_goV1.RegisterDcimCableServer(server, NewDcimCableServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimCableServer = (*dcimCable)(nil)
var _ time.Time

type dcimCable struct {
	netbox_goV1.UnimplementedDcimCableServer

	iDao dao.DcimCableDao
}

// NewDcimCableServer create a new service
func NewDcimCableServer() netbox_goV1.DcimCableServer {
	return &dcimCable{
		iDao: dao.NewDcimCableDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimCableCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimCable
func (s *dcimCable) Create(ctx context.Context, req *netbox_goV1.CreateDcimCableRequest) (*netbox_goV1.CreateDcimCableReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimCable{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimCable.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimCable", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimCableReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimCable by id
func (s *dcimCable) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimCableByIDRequest) (*netbox_goV1.DeleteDcimCableByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimCableByIDReply{}, nil
}

// UpdateByID update a dcimCable by id
func (s *dcimCable) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimCableByIDRequest) (*netbox_goV1.UpdateDcimCableByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimCable{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimCable.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimCable", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimCableByIDReply{}, nil
}

// GetByID get a dcimCable by id
func (s *dcimCable) GetByID(ctx context.Context, req *netbox_goV1.GetDcimCableByIDRequest) (*netbox_goV1.GetDcimCableByIDReply, error) {
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

	data, err := convertDcimCable(record)
	if err != nil {
		logger.Warn("convertDcimCable error", logger.Err(err), logger.Any("dcimCable", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimCable.Err()
	}

	return &netbox_goV1.GetDcimCableByIDReply{DcimCable: data}, nil
}

// List get a paginated list of dcimCables by custom conditions
func (s *dcimCable) List(ctx context.Context, req *netbox_goV1.ListDcimCableRequest) (*netbox_goV1.ListDcimCableReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimCable.Err()
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

	dcimCables := []*netbox_goV1.DcimCable{}
	for _, record := range records {
		data, err := convertDcimCable(record)
		if err != nil {
			logger.Warn("convertDcimCable error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimCables = append(dcimCables, data)
	}

	return &netbox_goV1.ListDcimCableReply{
		Total:      total,
		DcimCables: dcimCables,
	}, nil
}

// DeleteByIDs batch delete dcimCable by ids
func (s *dcimCable) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimCableByIDsRequest) (*netbox_goV1.DeleteDcimCableByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimCableByIDsReply{}, nil
}

// GetByCondition get a dcimCable by custom condition
func (s *dcimCable) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimCableByConditionRequest) (*netbox_goV1.GetDcimCableByConditionReply, error) {
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

	data, err := convertDcimCable(record)
	if err != nil {
		logger.Warn("convertDcimCable error", logger.Err(err), logger.Any("dcimCable", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimCable.Err()
	}

	return &netbox_goV1.GetDcimCableByConditionReply{
		DcimCable: data,
	}, nil
}

// ListByIDs batch get dcimCable by ids
func (s *dcimCable) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimCableByIDsRequest) (*netbox_goV1.ListDcimCableByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimCableMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimCables := []*netbox_goV1.DcimCable{}
	for _, id := range req.Ids {
		if v, ok := dcimCableMap[id]; ok {
			record, err := convertDcimCable(v)
			if err != nil {
				logger.Warn("convertDcimCable error", logger.Err(err), logger.Any("dcimCable", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimCables = append(dcimCables, record)
		}
	}

	return &netbox_goV1.ListDcimCableByIDsReply{DcimCables: dcimCables}, nil
}

// ListByLastID get a paginated list of dcimCables by last id
func (s *dcimCable) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimCableByLastIDRequest) (*netbox_goV1.ListDcimCableByLastIDReply, error) {
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

	dcimCables := []*netbox_goV1.DcimCable{}
	for _, record := range records {
		data, err := convertDcimCable(record)
		if err != nil {
			logger.Warn("convertDcimCable error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimCables = append(dcimCables, data)
	}

	return &netbox_goV1.ListDcimCableByLastIDReply{
		DcimCables: dcimCables,
	}, nil
}

func convertDcimCable(record *model.DcimCable) (*netbox_goV1.DcimCable, error) {
	value := &netbox_goV1.DcimCable{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
