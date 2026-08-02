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
		netbox_goV1.RegisterDcimDevicebayServer(server, NewDcimDevicebayServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimDevicebayServer = (*dcimDevicebay)(nil)
var _ time.Time

type dcimDevicebay struct {
	netbox_goV1.UnimplementedDcimDevicebayServer

	iDao dao.DcimDevicebayDao
}

// NewDcimDevicebayServer create a new service
func NewDcimDevicebayServer() netbox_goV1.DcimDevicebayServer {
	return &dcimDevicebay{
		iDao: dao.NewDcimDevicebayDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimDevicebayCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimDevicebay
func (s *dcimDevicebay) Create(ctx context.Context, req *netbox_goV1.CreateDcimDevicebayRequest) (*netbox_goV1.CreateDcimDevicebayReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimDevicebay{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimDevicebay.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimDevicebay", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimDevicebayReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimDevicebay by id
func (s *dcimDevicebay) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimDevicebayByIDRequest) (*netbox_goV1.DeleteDcimDevicebayByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimDevicebayByIDReply{}, nil
}

// UpdateByID update a dcimDevicebay by id
func (s *dcimDevicebay) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimDevicebayByIDRequest) (*netbox_goV1.UpdateDcimDevicebayByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimDevicebay{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimDevicebay.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimDevicebay", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimDevicebayByIDReply{}, nil
}

// GetByID get a dcimDevicebay by id
func (s *dcimDevicebay) GetByID(ctx context.Context, req *netbox_goV1.GetDcimDevicebayByIDRequest) (*netbox_goV1.GetDcimDevicebayByIDReply, error) {
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

	data, err := convertDcimDevicebay(record)
	if err != nil {
		logger.Warn("convertDcimDevicebay error", logger.Err(err), logger.Any("dcimDevicebay", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimDevicebay.Err()
	}

	return &netbox_goV1.GetDcimDevicebayByIDReply{DcimDevicebay: data}, nil
}

// List get a paginated list of dcimDevicebays by custom conditions
func (s *dcimDevicebay) List(ctx context.Context, req *netbox_goV1.ListDcimDevicebayRequest) (*netbox_goV1.ListDcimDevicebayReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimDevicebay.Err()
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

	dcimDevicebays := []*netbox_goV1.DcimDevicebay{}
	for _, record := range records {
		data, err := convertDcimDevicebay(record)
		if err != nil {
			logger.Warn("convertDcimDevicebay error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimDevicebays = append(dcimDevicebays, data)
	}

	return &netbox_goV1.ListDcimDevicebayReply{
		Total:          total,
		DcimDevicebays: dcimDevicebays,
	}, nil
}

// DeleteByIDs batch delete dcimDevicebay by ids
func (s *dcimDevicebay) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimDevicebayByIDsRequest) (*netbox_goV1.DeleteDcimDevicebayByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimDevicebayByIDsReply{}, nil
}

// GetByCondition get a dcimDevicebay by custom condition
func (s *dcimDevicebay) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimDevicebayByConditionRequest) (*netbox_goV1.GetDcimDevicebayByConditionReply, error) {
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

	data, err := convertDcimDevicebay(record)
	if err != nil {
		logger.Warn("convertDcimDevicebay error", logger.Err(err), logger.Any("dcimDevicebay", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimDevicebay.Err()
	}

	return &netbox_goV1.GetDcimDevicebayByConditionReply{
		DcimDevicebay: data,
	}, nil
}

// ListByIDs batch get dcimDevicebay by ids
func (s *dcimDevicebay) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimDevicebayByIDsRequest) (*netbox_goV1.ListDcimDevicebayByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimDevicebayMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimDevicebays := []*netbox_goV1.DcimDevicebay{}
	for _, id := range req.Ids {
		if v, ok := dcimDevicebayMap[id]; ok {
			record, err := convertDcimDevicebay(v)
			if err != nil {
				logger.Warn("convertDcimDevicebay error", logger.Err(err), logger.Any("dcimDevicebay", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimDevicebays = append(dcimDevicebays, record)
		}
	}

	return &netbox_goV1.ListDcimDevicebayByIDsReply{DcimDevicebays: dcimDevicebays}, nil
}

// ListByLastID get a paginated list of dcimDevicebays by last id
func (s *dcimDevicebay) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimDevicebayByLastIDRequest) (*netbox_goV1.ListDcimDevicebayByLastIDReply, error) {
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

	dcimDevicebays := []*netbox_goV1.DcimDevicebay{}
	for _, record := range records {
		data, err := convertDcimDevicebay(record)
		if err != nil {
			logger.Warn("convertDcimDevicebay error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimDevicebays = append(dcimDevicebays, data)
	}

	return &netbox_goV1.ListDcimDevicebayByLastIDReply{
		DcimDevicebays: dcimDevicebays,
	}, nil
}

func convertDcimDevicebay(record *model.DcimDevicebay) (*netbox_goV1.DcimDevicebay, error) {
	value := &netbox_goV1.DcimDevicebay{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
