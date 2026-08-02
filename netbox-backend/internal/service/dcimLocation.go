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
		netbox_goV1.RegisterDcimLocationServer(server, NewDcimLocationServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimLocationServer = (*dcimLocation)(nil)
var _ time.Time

type dcimLocation struct {
	netbox_goV1.UnimplementedDcimLocationServer

	iDao dao.DcimLocationDao
}

// NewDcimLocationServer create a new service
func NewDcimLocationServer() netbox_goV1.DcimLocationServer {
	return &dcimLocation{
		iDao: dao.NewDcimLocationDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimLocationCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimLocation
func (s *dcimLocation) Create(ctx context.Context, req *netbox_goV1.CreateDcimLocationRequest) (*netbox_goV1.CreateDcimLocationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimLocation{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimLocation.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimLocation", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimLocationReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimLocation by id
func (s *dcimLocation) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimLocationByIDRequest) (*netbox_goV1.DeleteDcimLocationByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimLocationByIDReply{}, nil
}

// UpdateByID update a dcimLocation by id
func (s *dcimLocation) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimLocationByIDRequest) (*netbox_goV1.UpdateDcimLocationByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimLocation{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimLocation.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimLocation", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimLocationByIDReply{}, nil
}

// GetByID get a dcimLocation by id
func (s *dcimLocation) GetByID(ctx context.Context, req *netbox_goV1.GetDcimLocationByIDRequest) (*netbox_goV1.GetDcimLocationByIDReply, error) {
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

	data, err := convertDcimLocation(record)
	if err != nil {
		logger.Warn("convertDcimLocation error", logger.Err(err), logger.Any("dcimLocation", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimLocation.Err()
	}

	return &netbox_goV1.GetDcimLocationByIDReply{DcimLocation: data}, nil
}

// List get a paginated list of dcimLocations by custom conditions
func (s *dcimLocation) List(ctx context.Context, req *netbox_goV1.ListDcimLocationRequest) (*netbox_goV1.ListDcimLocationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimLocation.Err()
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

	dcimLocations := []*netbox_goV1.DcimLocation{}
	for _, record := range records {
		data, err := convertDcimLocation(record)
		if err != nil {
			logger.Warn("convertDcimLocation error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimLocations = append(dcimLocations, data)
	}

	return &netbox_goV1.ListDcimLocationReply{
		Total:         total,
		DcimLocations: dcimLocations,
	}, nil
}

// DeleteByIDs batch delete dcimLocation by ids
func (s *dcimLocation) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimLocationByIDsRequest) (*netbox_goV1.DeleteDcimLocationByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimLocationByIDsReply{}, nil
}

// GetByCondition get a dcimLocation by custom condition
func (s *dcimLocation) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimLocationByConditionRequest) (*netbox_goV1.GetDcimLocationByConditionReply, error) {
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

	data, err := convertDcimLocation(record)
	if err != nil {
		logger.Warn("convertDcimLocation error", logger.Err(err), logger.Any("dcimLocation", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimLocation.Err()
	}

	return &netbox_goV1.GetDcimLocationByConditionReply{
		DcimLocation: data,
	}, nil
}

// ListByIDs batch get dcimLocation by ids
func (s *dcimLocation) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimLocationByIDsRequest) (*netbox_goV1.ListDcimLocationByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimLocationMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimLocations := []*netbox_goV1.DcimLocation{}
	for _, id := range req.Ids {
		if v, ok := dcimLocationMap[id]; ok {
			record, err := convertDcimLocation(v)
			if err != nil {
				logger.Warn("convertDcimLocation error", logger.Err(err), logger.Any("dcimLocation", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimLocations = append(dcimLocations, record)
		}
	}

	return &netbox_goV1.ListDcimLocationByIDsReply{DcimLocations: dcimLocations}, nil
}

// ListByLastID get a paginated list of dcimLocations by last id
func (s *dcimLocation) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimLocationByLastIDRequest) (*netbox_goV1.ListDcimLocationByLastIDReply, error) {
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

	dcimLocations := []*netbox_goV1.DcimLocation{}
	for _, record := range records {
		data, err := convertDcimLocation(record)
		if err != nil {
			logger.Warn("convertDcimLocation error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimLocations = append(dcimLocations, data)
	}

	return &netbox_goV1.ListDcimLocationByLastIDReply{
		DcimLocations: dcimLocations,
	}, nil
}

func convertDcimLocation(record *model.DcimLocation) (*netbox_goV1.DcimLocation, error) {
	value := &netbox_goV1.DcimLocation{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
