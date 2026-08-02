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
		netbox_goV1.RegisterWirelessWirelesslangroupServer(server, NewWirelessWirelesslangroupServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.WirelessWirelesslangroupServer = (*wirelessWirelesslangroup)(nil)
var _ time.Time

type wirelessWirelesslangroup struct {
	netbox_goV1.UnimplementedWirelessWirelesslangroupServer

	iDao dao.WirelessWirelesslangroupDao
}

// NewWirelessWirelesslangroupServer create a new service
func NewWirelessWirelesslangroupServer() netbox_goV1.WirelessWirelesslangroupServer {
	return &wirelessWirelesslangroup{
		iDao: dao.NewWirelessWirelesslangroupDao(
			database.GetDB(), // db driver is postgresql
			cache.NewWirelessWirelesslangroupCache(database.GetCacheType()),
		),
	}
}

// Create a new wirelessWirelesslangroup
func (s *wirelessWirelesslangroup) Create(ctx context.Context, req *netbox_goV1.CreateWirelessWirelesslangroupRequest) (*netbox_goV1.CreateWirelessWirelesslangroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.WirelessWirelesslangroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateWirelessWirelesslangroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("wirelessWirelesslangroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateWirelessWirelesslangroupReply{Id: record.ID}, nil
}

// DeleteByID delete a wirelessWirelesslangroup by id
func (s *wirelessWirelesslangroup) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteWirelessWirelesslangroupByIDRequest) (*netbox_goV1.DeleteWirelessWirelesslangroupByIDReply, error) {
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

	return &netbox_goV1.DeleteWirelessWirelesslangroupByIDReply{}, nil
}

// UpdateByID update a wirelessWirelesslangroup by id
func (s *wirelessWirelesslangroup) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateWirelessWirelesslangroupByIDRequest) (*netbox_goV1.UpdateWirelessWirelesslangroupByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.WirelessWirelesslangroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDWirelessWirelesslangroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("wirelessWirelesslangroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateWirelessWirelesslangroupByIDReply{}, nil
}

// GetByID get a wirelessWirelesslangroup by id
func (s *wirelessWirelesslangroup) GetByID(ctx context.Context, req *netbox_goV1.GetWirelessWirelesslangroupByIDRequest) (*netbox_goV1.GetWirelessWirelesslangroupByIDReply, error) {
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

	data, err := convertWirelessWirelesslangroup(record)
	if err != nil {
		logger.Warn("convertWirelessWirelesslangroup error", logger.Err(err), logger.Any("wirelessWirelesslangroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDWirelessWirelesslangroup.Err()
	}

	return &netbox_goV1.GetWirelessWirelesslangroupByIDReply{WirelessWirelesslangroup: data}, nil
}

// List get a paginated list of wirelessWirelesslangroups by custom conditions
func (s *wirelessWirelesslangroup) List(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslangroupRequest) (*netbox_goV1.ListWirelessWirelesslangroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListWirelessWirelesslangroup.Err()
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

	wirelessWirelesslangroups := []*netbox_goV1.WirelessWirelesslangroup{}
	for _, record := range records {
		data, err := convertWirelessWirelesslangroup(record)
		if err != nil {
			logger.Warn("convertWirelessWirelesslangroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		wirelessWirelesslangroups = append(wirelessWirelesslangroups, data)
	}

	return &netbox_goV1.ListWirelessWirelesslangroupReply{
		Total:                     total,
		WirelessWirelesslangroups: wirelessWirelesslangroups,
	}, nil
}

// DeleteByIDs batch delete wirelessWirelesslangroup by ids
func (s *wirelessWirelesslangroup) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteWirelessWirelesslangroupByIDsRequest) (*netbox_goV1.DeleteWirelessWirelesslangroupByIDsReply, error) {
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

	return &netbox_goV1.DeleteWirelessWirelesslangroupByIDsReply{}, nil
}

// GetByCondition get a wirelessWirelesslangroup by custom condition
func (s *wirelessWirelesslangroup) GetByCondition(ctx context.Context, req *netbox_goV1.GetWirelessWirelesslangroupByConditionRequest) (*netbox_goV1.GetWirelessWirelesslangroupByConditionReply, error) {
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

	data, err := convertWirelessWirelesslangroup(record)
	if err != nil {
		logger.Warn("convertWirelessWirelesslangroup error", logger.Err(err), logger.Any("wirelessWirelesslangroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionWirelessWirelesslangroup.Err()
	}

	return &netbox_goV1.GetWirelessWirelesslangroupByConditionReply{
		WirelessWirelesslangroup: data,
	}, nil
}

// ListByIDs batch get wirelessWirelesslangroup by ids
func (s *wirelessWirelesslangroup) ListByIDs(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslangroupByIDsRequest) (*netbox_goV1.ListWirelessWirelesslangroupByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	wirelessWirelesslangroupMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	wirelessWirelesslangroups := []*netbox_goV1.WirelessWirelesslangroup{}
	for _, id := range req.Ids {
		if v, ok := wirelessWirelesslangroupMap[id]; ok {
			record, err := convertWirelessWirelesslangroup(v)
			if err != nil {
				logger.Warn("convertWirelessWirelesslangroup error", logger.Err(err), logger.Any("wirelessWirelesslangroup", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			wirelessWirelesslangroups = append(wirelessWirelesslangroups, record)
		}
	}

	return &netbox_goV1.ListWirelessWirelesslangroupByIDsReply{WirelessWirelesslangroups: wirelessWirelesslangroups}, nil
}

// ListByLastID get a paginated list of wirelessWirelesslangroups by last id
func (s *wirelessWirelesslangroup) ListByLastID(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslangroupByLastIDRequest) (*netbox_goV1.ListWirelessWirelesslangroupByLastIDReply, error) {
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

	wirelessWirelesslangroups := []*netbox_goV1.WirelessWirelesslangroup{}
	for _, record := range records {
		data, err := convertWirelessWirelesslangroup(record)
		if err != nil {
			logger.Warn("convertWirelessWirelesslangroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		wirelessWirelesslangroups = append(wirelessWirelesslangroups, data)
	}

	return &netbox_goV1.ListWirelessWirelesslangroupByLastIDReply{
		WirelessWirelesslangroups: wirelessWirelesslangroups,
	}, nil
}

func convertWirelessWirelesslangroup(record *model.WirelessWirelesslangroup) (*netbox_goV1.WirelessWirelesslangroup, error) {
	value := &netbox_goV1.WirelessWirelesslangroup{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
