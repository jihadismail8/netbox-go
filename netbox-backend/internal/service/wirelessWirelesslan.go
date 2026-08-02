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
		netbox_goV1.RegisterWirelessWirelesslanServer(server, NewWirelessWirelesslanServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.WirelessWirelesslanServer = (*wirelessWirelesslan)(nil)
var _ time.Time

type wirelessWirelesslan struct {
	netbox_goV1.UnimplementedWirelessWirelesslanServer

	iDao dao.WirelessWirelesslanDao
}

// NewWirelessWirelesslanServer create a new service
func NewWirelessWirelesslanServer() netbox_goV1.WirelessWirelesslanServer {
	return &wirelessWirelesslan{
		iDao: dao.NewWirelessWirelesslanDao(
			database.GetDB(), // db driver is postgresql
			cache.NewWirelessWirelesslanCache(database.GetCacheType()),
		),
	}
}

// Create a new wirelessWirelesslan
func (s *wirelessWirelesslan) Create(ctx context.Context, req *netbox_goV1.CreateWirelessWirelesslanRequest) (*netbox_goV1.CreateWirelessWirelesslanReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.WirelessWirelesslan{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateWirelessWirelesslan.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("wirelessWirelesslan", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateWirelessWirelesslanReply{Id: record.ID}, nil
}

// DeleteByID delete a wirelessWirelesslan by id
func (s *wirelessWirelesslan) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteWirelessWirelesslanByIDRequest) (*netbox_goV1.DeleteWirelessWirelesslanByIDReply, error) {
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

	return &netbox_goV1.DeleteWirelessWirelesslanByIDReply{}, nil
}

// UpdateByID update a wirelessWirelesslan by id
func (s *wirelessWirelesslan) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateWirelessWirelesslanByIDRequest) (*netbox_goV1.UpdateWirelessWirelesslanByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.WirelessWirelesslan{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDWirelessWirelesslan.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("wirelessWirelesslan", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateWirelessWirelesslanByIDReply{}, nil
}

// GetByID get a wirelessWirelesslan by id
func (s *wirelessWirelesslan) GetByID(ctx context.Context, req *netbox_goV1.GetWirelessWirelesslanByIDRequest) (*netbox_goV1.GetWirelessWirelesslanByIDReply, error) {
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

	data, err := convertWirelessWirelesslan(record)
	if err != nil {
		logger.Warn("convertWirelessWirelesslan error", logger.Err(err), logger.Any("wirelessWirelesslan", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDWirelessWirelesslan.Err()
	}

	return &netbox_goV1.GetWirelessWirelesslanByIDReply{WirelessWirelesslan: data}, nil
}

// List get a paginated list of wirelessWirelesslans by custom conditions
func (s *wirelessWirelesslan) List(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslanRequest) (*netbox_goV1.ListWirelessWirelesslanReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListWirelessWirelesslan.Err()
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

	wirelessWirelesslans := []*netbox_goV1.WirelessWirelesslan{}
	for _, record := range records {
		data, err := convertWirelessWirelesslan(record)
		if err != nil {
			logger.Warn("convertWirelessWirelesslan error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		wirelessWirelesslans = append(wirelessWirelesslans, data)
	}

	return &netbox_goV1.ListWirelessWirelesslanReply{
		Total:                total,
		WirelessWirelesslans: wirelessWirelesslans,
	}, nil
}

// DeleteByIDs batch delete wirelessWirelesslan by ids
func (s *wirelessWirelesslan) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteWirelessWirelesslanByIDsRequest) (*netbox_goV1.DeleteWirelessWirelesslanByIDsReply, error) {
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

	return &netbox_goV1.DeleteWirelessWirelesslanByIDsReply{}, nil
}

// GetByCondition get a wirelessWirelesslan by custom condition
func (s *wirelessWirelesslan) GetByCondition(ctx context.Context, req *netbox_goV1.GetWirelessWirelesslanByConditionRequest) (*netbox_goV1.GetWirelessWirelesslanByConditionReply, error) {
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

	data, err := convertWirelessWirelesslan(record)
	if err != nil {
		logger.Warn("convertWirelessWirelesslan error", logger.Err(err), logger.Any("wirelessWirelesslan", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionWirelessWirelesslan.Err()
	}

	return &netbox_goV1.GetWirelessWirelesslanByConditionReply{
		WirelessWirelesslan: data,
	}, nil
}

// ListByIDs batch get wirelessWirelesslan by ids
func (s *wirelessWirelesslan) ListByIDs(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslanByIDsRequest) (*netbox_goV1.ListWirelessWirelesslanByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	wirelessWirelesslanMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	wirelessWirelesslans := []*netbox_goV1.WirelessWirelesslan{}
	for _, id := range req.Ids {
		if v, ok := wirelessWirelesslanMap[id]; ok {
			record, err := convertWirelessWirelesslan(v)
			if err != nil {
				logger.Warn("convertWirelessWirelesslan error", logger.Err(err), logger.Any("wirelessWirelesslan", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			wirelessWirelesslans = append(wirelessWirelesslans, record)
		}
	}

	return &netbox_goV1.ListWirelessWirelesslanByIDsReply{WirelessWirelesslans: wirelessWirelesslans}, nil
}

// ListByLastID get a paginated list of wirelessWirelesslans by last id
func (s *wirelessWirelesslan) ListByLastID(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslanByLastIDRequest) (*netbox_goV1.ListWirelessWirelesslanByLastIDReply, error) {
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

	wirelessWirelesslans := []*netbox_goV1.WirelessWirelesslan{}
	for _, record := range records {
		data, err := convertWirelessWirelesslan(record)
		if err != nil {
			logger.Warn("convertWirelessWirelesslan error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		wirelessWirelesslans = append(wirelessWirelesslans, data)
	}

	return &netbox_goV1.ListWirelessWirelesslanByLastIDReply{
		WirelessWirelesslans: wirelessWirelesslans,
	}, nil
}

func convertWirelessWirelesslan(record *model.WirelessWirelesslan) (*netbox_goV1.WirelessWirelesslan, error) {
	value := &netbox_goV1.WirelessWirelesslan{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
