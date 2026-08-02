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
		netbox_goV1.RegisterWirelessWirelesslinkServer(server, NewWirelessWirelesslinkServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.WirelessWirelesslinkServer = (*wirelessWirelesslink)(nil)
var _ time.Time

type wirelessWirelesslink struct {
	netbox_goV1.UnimplementedWirelessWirelesslinkServer

	iDao dao.WirelessWirelesslinkDao
}

// NewWirelessWirelesslinkServer create a new service
func NewWirelessWirelesslinkServer() netbox_goV1.WirelessWirelesslinkServer {
	return &wirelessWirelesslink{
		iDao: dao.NewWirelessWirelesslinkDao(
			database.GetDB(), // db driver is postgresql
			cache.NewWirelessWirelesslinkCache(database.GetCacheType()),
		),
	}
}

// Create a new wirelessWirelesslink
func (s *wirelessWirelesslink) Create(ctx context.Context, req *netbox_goV1.CreateWirelessWirelesslinkRequest) (*netbox_goV1.CreateWirelessWirelesslinkReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.WirelessWirelesslink{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateWirelessWirelesslink.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("wirelessWirelesslink", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateWirelessWirelesslinkReply{Id: record.ID}, nil
}

// DeleteByID delete a wirelessWirelesslink by id
func (s *wirelessWirelesslink) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteWirelessWirelesslinkByIDRequest) (*netbox_goV1.DeleteWirelessWirelesslinkByIDReply, error) {
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

	return &netbox_goV1.DeleteWirelessWirelesslinkByIDReply{}, nil
}

// UpdateByID update a wirelessWirelesslink by id
func (s *wirelessWirelesslink) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateWirelessWirelesslinkByIDRequest) (*netbox_goV1.UpdateWirelessWirelesslinkByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.WirelessWirelesslink{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDWirelessWirelesslink.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("wirelessWirelesslink", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateWirelessWirelesslinkByIDReply{}, nil
}

// GetByID get a wirelessWirelesslink by id
func (s *wirelessWirelesslink) GetByID(ctx context.Context, req *netbox_goV1.GetWirelessWirelesslinkByIDRequest) (*netbox_goV1.GetWirelessWirelesslinkByIDReply, error) {
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

	data, err := convertWirelessWirelesslink(record)
	if err != nil {
		logger.Warn("convertWirelessWirelesslink error", logger.Err(err), logger.Any("wirelessWirelesslink", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDWirelessWirelesslink.Err()
	}

	return &netbox_goV1.GetWirelessWirelesslinkByIDReply{WirelessWirelesslink: data}, nil
}

// List get a paginated list of wirelessWirelesslinks by custom conditions
func (s *wirelessWirelesslink) List(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslinkRequest) (*netbox_goV1.ListWirelessWirelesslinkReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListWirelessWirelesslink.Err()
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

	wirelessWirelesslinks := []*netbox_goV1.WirelessWirelesslink{}
	for _, record := range records {
		data, err := convertWirelessWirelesslink(record)
		if err != nil {
			logger.Warn("convertWirelessWirelesslink error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		wirelessWirelesslinks = append(wirelessWirelesslinks, data)
	}

	return &netbox_goV1.ListWirelessWirelesslinkReply{
		Total:                 total,
		WirelessWirelesslinks: wirelessWirelesslinks,
	}, nil
}

// DeleteByIDs batch delete wirelessWirelesslink by ids
func (s *wirelessWirelesslink) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteWirelessWirelesslinkByIDsRequest) (*netbox_goV1.DeleteWirelessWirelesslinkByIDsReply, error) {
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

	return &netbox_goV1.DeleteWirelessWirelesslinkByIDsReply{}, nil
}

// GetByCondition get a wirelessWirelesslink by custom condition
func (s *wirelessWirelesslink) GetByCondition(ctx context.Context, req *netbox_goV1.GetWirelessWirelesslinkByConditionRequest) (*netbox_goV1.GetWirelessWirelesslinkByConditionReply, error) {
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

	data, err := convertWirelessWirelesslink(record)
	if err != nil {
		logger.Warn("convertWirelessWirelesslink error", logger.Err(err), logger.Any("wirelessWirelesslink", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionWirelessWirelesslink.Err()
	}

	return &netbox_goV1.GetWirelessWirelesslinkByConditionReply{
		WirelessWirelesslink: data,
	}, nil
}

// ListByIDs batch get wirelessWirelesslink by ids
func (s *wirelessWirelesslink) ListByIDs(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslinkByIDsRequest) (*netbox_goV1.ListWirelessWirelesslinkByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	wirelessWirelesslinkMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	wirelessWirelesslinks := []*netbox_goV1.WirelessWirelesslink{}
	for _, id := range req.Ids {
		if v, ok := wirelessWirelesslinkMap[id]; ok {
			record, err := convertWirelessWirelesslink(v)
			if err != nil {
				logger.Warn("convertWirelessWirelesslink error", logger.Err(err), logger.Any("wirelessWirelesslink", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			wirelessWirelesslinks = append(wirelessWirelesslinks, record)
		}
	}

	return &netbox_goV1.ListWirelessWirelesslinkByIDsReply{WirelessWirelesslinks: wirelessWirelesslinks}, nil
}

// ListByLastID get a paginated list of wirelessWirelesslinks by last id
func (s *wirelessWirelesslink) ListByLastID(ctx context.Context, req *netbox_goV1.ListWirelessWirelesslinkByLastIDRequest) (*netbox_goV1.ListWirelessWirelesslinkByLastIDReply, error) {
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

	wirelessWirelesslinks := []*netbox_goV1.WirelessWirelesslink{}
	for _, record := range records {
		data, err := convertWirelessWirelesslink(record)
		if err != nil {
			logger.Warn("convertWirelessWirelesslink error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		wirelessWirelesslinks = append(wirelessWirelesslinks, data)
	}

	return &netbox_goV1.ListWirelessWirelesslinkByLastIDReply{
		WirelessWirelesslinks: wirelessWirelesslinks,
	}, nil
}

func convertWirelessWirelesslink(record *model.WirelessWirelesslink) (*netbox_goV1.WirelessWirelesslink, error) {
	value := &netbox_goV1.WirelessWirelesslink{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
