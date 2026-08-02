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
		netbox_goV1.RegisterDcimVirtualchassisServer(server, NewDcimVirtualchassisServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimVirtualchassisServer = (*dcimVirtualchassis)(nil)
var _ time.Time

type dcimVirtualchassis struct {
	netbox_goV1.UnimplementedDcimVirtualchassisServer

	iDao dao.DcimVirtualchassisDao
}

// NewDcimVirtualchassisServer create a new service
func NewDcimVirtualchassisServer() netbox_goV1.DcimVirtualchassisServer {
	return &dcimVirtualchassis{
		iDao: dao.NewDcimVirtualchassisDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimVirtualchassisCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimVirtualchassis
func (s *dcimVirtualchassis) Create(ctx context.Context, req *netbox_goV1.CreateDcimVirtualchassisRequest) (*netbox_goV1.CreateDcimVirtualchassisReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimVirtualchassis{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimVirtualchassis.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimVirtualchassis", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimVirtualchassisReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimVirtualchassis by id
func (s *dcimVirtualchassis) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimVirtualchassisByIDRequest) (*netbox_goV1.DeleteDcimVirtualchassisByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimVirtualchassisByIDReply{}, nil
}

// UpdateByID update a dcimVirtualchassis by id
func (s *dcimVirtualchassis) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimVirtualchassisByIDRequest) (*netbox_goV1.UpdateDcimVirtualchassisByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimVirtualchassis{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimVirtualchassis.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimVirtualchassis", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimVirtualchassisByIDReply{}, nil
}

// GetByID get a dcimVirtualchassis by id
func (s *dcimVirtualchassis) GetByID(ctx context.Context, req *netbox_goV1.GetDcimVirtualchassisByIDRequest) (*netbox_goV1.GetDcimVirtualchassisByIDReply, error) {
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

	data, err := convertDcimVirtualchassis(record)
	if err != nil {
		logger.Warn("convertDcimVirtualchassis error", logger.Err(err), logger.Any("dcimVirtualchassis", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimVirtualchassis.Err()
	}

	return &netbox_goV1.GetDcimVirtualchassisByIDReply{DcimVirtualchassis: data}, nil
}

// List get a paginated list of dcimVirtualchassiss by custom conditions
func (s *dcimVirtualchassis) List(ctx context.Context, req *netbox_goV1.ListDcimVirtualchassisRequest) (*netbox_goV1.ListDcimVirtualchassisReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimVirtualchassis.Err()
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

	dcimVirtualchassiss := []*netbox_goV1.DcimVirtualchassis{}
	for _, record := range records {
		data, err := convertDcimVirtualchassis(record)
		if err != nil {
			logger.Warn("convertDcimVirtualchassis error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimVirtualchassiss = append(dcimVirtualchassiss, data)
	}

	return &netbox_goV1.ListDcimVirtualchassisReply{
		Total:               total,
		DcimVirtualchassiss: dcimVirtualchassiss,
	}, nil
}

// DeleteByIDs batch delete dcimVirtualchassis by ids
func (s *dcimVirtualchassis) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimVirtualchassisByIDsRequest) (*netbox_goV1.DeleteDcimVirtualchassisByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimVirtualchassisByIDsReply{}, nil
}

// GetByCondition get a dcimVirtualchassis by custom condition
func (s *dcimVirtualchassis) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimVirtualchassisByConditionRequest) (*netbox_goV1.GetDcimVirtualchassisByConditionReply, error) {
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

	data, err := convertDcimVirtualchassis(record)
	if err != nil {
		logger.Warn("convertDcimVirtualchassis error", logger.Err(err), logger.Any("dcimVirtualchassis", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimVirtualchassis.Err()
	}

	return &netbox_goV1.GetDcimVirtualchassisByConditionReply{
		DcimVirtualchassis: data,
	}, nil
}

// ListByIDs batch get dcimVirtualchassis by ids
func (s *dcimVirtualchassis) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimVirtualchassisByIDsRequest) (*netbox_goV1.ListDcimVirtualchassisByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimVirtualchassisMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimVirtualchassiss := []*netbox_goV1.DcimVirtualchassis{}
	for _, id := range req.Ids {
		if v, ok := dcimVirtualchassisMap[id]; ok {
			record, err := convertDcimVirtualchassis(v)
			if err != nil {
				logger.Warn("convertDcimVirtualchassis error", logger.Err(err), logger.Any("dcimVirtualchassis", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimVirtualchassiss = append(dcimVirtualchassiss, record)
		}
	}

	return &netbox_goV1.ListDcimVirtualchassisByIDsReply{DcimVirtualchassiss: dcimVirtualchassiss}, nil
}

// ListByLastID get a paginated list of dcimVirtualchassiss by last id
func (s *dcimVirtualchassis) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimVirtualchassisByLastIDRequest) (*netbox_goV1.ListDcimVirtualchassisByLastIDReply, error) {
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

	dcimVirtualchassiss := []*netbox_goV1.DcimVirtualchassis{}
	for _, record := range records {
		data, err := convertDcimVirtualchassis(record)
		if err != nil {
			logger.Warn("convertDcimVirtualchassis error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimVirtualchassiss = append(dcimVirtualchassiss, data)
	}

	return &netbox_goV1.ListDcimVirtualchassisByLastIDReply{
		DcimVirtualchassiss: dcimVirtualchassiss,
	}, nil
}

func convertDcimVirtualchassis(record *model.DcimVirtualchassis) (*netbox_goV1.DcimVirtualchassis, error) {
	value := &netbox_goV1.DcimVirtualchassis{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
