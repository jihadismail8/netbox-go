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
		netbox_goV1.RegisterDcimCableterminationServer(server, NewDcimCableterminationServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimCableterminationServer = (*dcimCabletermination)(nil)
var _ time.Time

type dcimCabletermination struct {
	netbox_goV1.UnimplementedDcimCableterminationServer

	iDao dao.DcimCableterminationDao
}

// NewDcimCableterminationServer create a new service
func NewDcimCableterminationServer() netbox_goV1.DcimCableterminationServer {
	return &dcimCabletermination{
		iDao: dao.NewDcimCableterminationDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimCableterminationCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimCabletermination
func (s *dcimCabletermination) Create(ctx context.Context, req *netbox_goV1.CreateDcimCableterminationRequest) (*netbox_goV1.CreateDcimCableterminationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimCabletermination{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimCabletermination.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimCabletermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimCableterminationReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimCabletermination by id
func (s *dcimCabletermination) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimCableterminationByIDRequest) (*netbox_goV1.DeleteDcimCableterminationByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimCableterminationByIDReply{}, nil
}

// UpdateByID update a dcimCabletermination by id
func (s *dcimCabletermination) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimCableterminationByIDRequest) (*netbox_goV1.UpdateDcimCableterminationByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimCabletermination{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimCabletermination.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimCabletermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimCableterminationByIDReply{}, nil
}

// GetByID get a dcimCabletermination by id
func (s *dcimCabletermination) GetByID(ctx context.Context, req *netbox_goV1.GetDcimCableterminationByIDRequest) (*netbox_goV1.GetDcimCableterminationByIDReply, error) {
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

	data, err := convertDcimCabletermination(record)
	if err != nil {
		logger.Warn("convertDcimCabletermination error", logger.Err(err), logger.Any("dcimCabletermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimCabletermination.Err()
	}

	return &netbox_goV1.GetDcimCableterminationByIDReply{DcimCabletermination: data}, nil
}

// List get a paginated list of dcimCableterminations by custom conditions
func (s *dcimCabletermination) List(ctx context.Context, req *netbox_goV1.ListDcimCableterminationRequest) (*netbox_goV1.ListDcimCableterminationReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimCabletermination.Err()
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

	dcimCableterminations := []*netbox_goV1.DcimCabletermination{}
	for _, record := range records {
		data, err := convertDcimCabletermination(record)
		if err != nil {
			logger.Warn("convertDcimCabletermination error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimCableterminations = append(dcimCableterminations, data)
	}

	return &netbox_goV1.ListDcimCableterminationReply{
		Total:                 total,
		DcimCableterminations: dcimCableterminations,
	}, nil
}

// DeleteByIDs batch delete dcimCabletermination by ids
func (s *dcimCabletermination) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimCableterminationByIDsRequest) (*netbox_goV1.DeleteDcimCableterminationByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimCableterminationByIDsReply{}, nil
}

// GetByCondition get a dcimCabletermination by custom condition
func (s *dcimCabletermination) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimCableterminationByConditionRequest) (*netbox_goV1.GetDcimCableterminationByConditionReply, error) {
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

	data, err := convertDcimCabletermination(record)
	if err != nil {
		logger.Warn("convertDcimCabletermination error", logger.Err(err), logger.Any("dcimCabletermination", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimCabletermination.Err()
	}

	return &netbox_goV1.GetDcimCableterminationByConditionReply{
		DcimCabletermination: data,
	}, nil
}

// ListByIDs batch get dcimCabletermination by ids
func (s *dcimCabletermination) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimCableterminationByIDsRequest) (*netbox_goV1.ListDcimCableterminationByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimCableterminationMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimCableterminations := []*netbox_goV1.DcimCabletermination{}
	for _, id := range req.Ids {
		if v, ok := dcimCableterminationMap[id]; ok {
			record, err := convertDcimCabletermination(v)
			if err != nil {
				logger.Warn("convertDcimCabletermination error", logger.Err(err), logger.Any("dcimCabletermination", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimCableterminations = append(dcimCableterminations, record)
		}
	}

	return &netbox_goV1.ListDcimCableterminationByIDsReply{DcimCableterminations: dcimCableterminations}, nil
}

// ListByLastID get a paginated list of dcimCableterminations by last id
func (s *dcimCabletermination) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimCableterminationByLastIDRequest) (*netbox_goV1.ListDcimCableterminationByLastIDReply, error) {
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

	dcimCableterminations := []*netbox_goV1.DcimCabletermination{}
	for _, record := range records {
		data, err := convertDcimCabletermination(record)
		if err != nil {
			logger.Warn("convertDcimCabletermination error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimCableterminations = append(dcimCableterminations, data)
	}

	return &netbox_goV1.ListDcimCableterminationByLastIDReply{
		DcimCableterminations: dcimCableterminations,
	}, nil
}

func convertDcimCabletermination(record *model.DcimCabletermination) (*netbox_goV1.DcimCabletermination, error) {
	value := &netbox_goV1.DcimCabletermination{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
