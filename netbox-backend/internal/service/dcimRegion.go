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
		netbox_goV1.RegisterDcimRegionServer(server, NewDcimRegionServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimRegionServer = (*dcimRegion)(nil)
var _ time.Time

type dcimRegion struct {
	netbox_goV1.UnimplementedDcimRegionServer

	iDao dao.DcimRegionDao
}

// NewDcimRegionServer create a new service
func NewDcimRegionServer() netbox_goV1.DcimRegionServer {
	return &dcimRegion{
		iDao: dao.NewDcimRegionDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimRegionCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimRegion
func (s *dcimRegion) Create(ctx context.Context, req *netbox_goV1.CreateDcimRegionRequest) (*netbox_goV1.CreateDcimRegionReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimRegion{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimRegion.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimRegion", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimRegionReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimRegion by id
func (s *dcimRegion) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimRegionByIDRequest) (*netbox_goV1.DeleteDcimRegionByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimRegionByIDReply{}, nil
}

// UpdateByID update a dcimRegion by id
func (s *dcimRegion) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimRegionByIDRequest) (*netbox_goV1.UpdateDcimRegionByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimRegion{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimRegion.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimRegion", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimRegionByIDReply{}, nil
}

// GetByID get a dcimRegion by id
func (s *dcimRegion) GetByID(ctx context.Context, req *netbox_goV1.GetDcimRegionByIDRequest) (*netbox_goV1.GetDcimRegionByIDReply, error) {
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

	data, err := convertDcimRegion(record)
	if err != nil {
		logger.Warn("convertDcimRegion error", logger.Err(err), logger.Any("dcimRegion", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimRegion.Err()
	}

	return &netbox_goV1.GetDcimRegionByIDReply{DcimRegion: data}, nil
}

// List get a paginated list of dcimRegions by custom conditions
func (s *dcimRegion) List(ctx context.Context, req *netbox_goV1.ListDcimRegionRequest) (*netbox_goV1.ListDcimRegionReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimRegion.Err()
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

	dcimRegions := []*netbox_goV1.DcimRegion{}
	for _, record := range records {
		data, err := convertDcimRegion(record)
		if err != nil {
			logger.Warn("convertDcimRegion error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimRegions = append(dcimRegions, data)
	}

	return &netbox_goV1.ListDcimRegionReply{
		Total:       total,
		DcimRegions: dcimRegions,
	}, nil
}

// DeleteByIDs batch delete dcimRegion by ids
func (s *dcimRegion) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimRegionByIDsRequest) (*netbox_goV1.DeleteDcimRegionByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimRegionByIDsReply{}, nil
}

// GetByCondition get a dcimRegion by custom condition
func (s *dcimRegion) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimRegionByConditionRequest) (*netbox_goV1.GetDcimRegionByConditionReply, error) {
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

	data, err := convertDcimRegion(record)
	if err != nil {
		logger.Warn("convertDcimRegion error", logger.Err(err), logger.Any("dcimRegion", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimRegion.Err()
	}

	return &netbox_goV1.GetDcimRegionByConditionReply{
		DcimRegion: data,
	}, nil
}

// ListByIDs batch get dcimRegion by ids
func (s *dcimRegion) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimRegionByIDsRequest) (*netbox_goV1.ListDcimRegionByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimRegionMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimRegions := []*netbox_goV1.DcimRegion{}
	for _, id := range req.Ids {
		if v, ok := dcimRegionMap[id]; ok {
			record, err := convertDcimRegion(v)
			if err != nil {
				logger.Warn("convertDcimRegion error", logger.Err(err), logger.Any("dcimRegion", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimRegions = append(dcimRegions, record)
		}
	}

	return &netbox_goV1.ListDcimRegionByIDsReply{DcimRegions: dcimRegions}, nil
}

// ListByLastID get a paginated list of dcimRegions by last id
func (s *dcimRegion) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimRegionByLastIDRequest) (*netbox_goV1.ListDcimRegionByLastIDReply, error) {
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

	dcimRegions := []*netbox_goV1.DcimRegion{}
	for _, record := range records {
		data, err := convertDcimRegion(record)
		if err != nil {
			logger.Warn("convertDcimRegion error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimRegions = append(dcimRegions, data)
	}

	return &netbox_goV1.ListDcimRegionByLastIDReply{
		DcimRegions: dcimRegions,
	}, nil
}

func convertDcimRegion(record *model.DcimRegion) (*netbox_goV1.DcimRegion, error) {
	value := &netbox_goV1.DcimRegion{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
