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
		netbox_goV1.RegisterDcimRackreservationServer(server, NewDcimRackreservationServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimRackreservationServer = (*dcimRackreservation)(nil)
var _ time.Time

type dcimRackreservation struct {
	netbox_goV1.UnimplementedDcimRackreservationServer

	iDao dao.DcimRackreservationDao
}

// NewDcimRackreservationServer create a new service
func NewDcimRackreservationServer() netbox_goV1.DcimRackreservationServer {
	return &dcimRackreservation{
		iDao: dao.NewDcimRackreservationDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimRackreservationCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimRackreservation
func (s *dcimRackreservation) Create(ctx context.Context, req *netbox_goV1.CreateDcimRackreservationRequest) (*netbox_goV1.CreateDcimRackreservationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimRackreservation{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimRackreservation.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimRackreservation", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimRackreservationReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimRackreservation by id
func (s *dcimRackreservation) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimRackreservationByIDRequest) (*netbox_goV1.DeleteDcimRackreservationByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimRackreservationByIDReply{}, nil
}

// UpdateByID update a dcimRackreservation by id
func (s *dcimRackreservation) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimRackreservationByIDRequest) (*netbox_goV1.UpdateDcimRackreservationByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimRackreservation{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimRackreservation.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimRackreservation", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimRackreservationByIDReply{}, nil
}

// GetByID get a dcimRackreservation by id
func (s *dcimRackreservation) GetByID(ctx context.Context, req *netbox_goV1.GetDcimRackreservationByIDRequest) (*netbox_goV1.GetDcimRackreservationByIDReply, error) {
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

	data, err := convertDcimRackreservation(record)
	if err != nil {
		logger.Warn("convertDcimRackreservation error", logger.Err(err), logger.Any("dcimRackreservation", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimRackreservation.Err()
	}

	return &netbox_goV1.GetDcimRackreservationByIDReply{DcimRackreservation: data}, nil
}

// List get a paginated list of dcimRackreservations by custom conditions
func (s *dcimRackreservation) List(ctx context.Context, req *netbox_goV1.ListDcimRackreservationRequest) (*netbox_goV1.ListDcimRackreservationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimRackreservation.Err()
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

	dcimRackreservations := []*netbox_goV1.DcimRackreservation{}
	for _, record := range records {
		data, err := convertDcimRackreservation(record)
		if err != nil {
			logger.Warn("convertDcimRackreservation error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimRackreservations = append(dcimRackreservations, data)
	}

	return &netbox_goV1.ListDcimRackreservationReply{
		Total:                total,
		DcimRackreservations: dcimRackreservations,
	}, nil
}

// DeleteByIDs batch delete dcimRackreservation by ids
func (s *dcimRackreservation) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimRackreservationByIDsRequest) (*netbox_goV1.DeleteDcimRackreservationByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimRackreservationByIDsReply{}, nil
}

// GetByCondition get a dcimRackreservation by custom condition
func (s *dcimRackreservation) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimRackreservationByConditionRequest) (*netbox_goV1.GetDcimRackreservationByConditionReply, error) {
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

	data, err := convertDcimRackreservation(record)
	if err != nil {
		logger.Warn("convertDcimRackreservation error", logger.Err(err), logger.Any("dcimRackreservation", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimRackreservation.Err()
	}

	return &netbox_goV1.GetDcimRackreservationByConditionReply{
		DcimRackreservation: data,
	}, nil
}

// ListByIDs batch get dcimRackreservation by ids
func (s *dcimRackreservation) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimRackreservationByIDsRequest) (*netbox_goV1.ListDcimRackreservationByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimRackreservationMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimRackreservations := []*netbox_goV1.DcimRackreservation{}
	for _, id := range req.Ids {
		if v, ok := dcimRackreservationMap[id]; ok {
			record, err := convertDcimRackreservation(v)
			if err != nil {
				logger.Warn("convertDcimRackreservation error", logger.Err(err), logger.Any("dcimRackreservation", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimRackreservations = append(dcimRackreservations, record)
		}
	}

	return &netbox_goV1.ListDcimRackreservationByIDsReply{DcimRackreservations: dcimRackreservations}, nil
}

// ListByLastID get a paginated list of dcimRackreservations by last id
func (s *dcimRackreservation) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimRackreservationByLastIDRequest) (*netbox_goV1.ListDcimRackreservationByLastIDReply, error) {
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

	dcimRackreservations := []*netbox_goV1.DcimRackreservation{}
	for _, record := range records {
		data, err := convertDcimRackreservation(record)
		if err != nil {
			logger.Warn("convertDcimRackreservation error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimRackreservations = append(dcimRackreservations, data)
	}

	return &netbox_goV1.ListDcimRackreservationByLastIDReply{
		DcimRackreservations: dcimRackreservations,
	}, nil
}

func convertDcimRackreservation(record *model.DcimRackreservation) (*netbox_goV1.DcimRackreservation, error) {
	value := &netbox_goV1.DcimRackreservation{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
