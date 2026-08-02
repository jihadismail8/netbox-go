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
		netbox_goV1.RegisterDcimPoweroutletServer(server, NewDcimPoweroutletServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimPoweroutletServer = (*dcimPoweroutlet)(nil)
var _ time.Time

type dcimPoweroutlet struct {
	netbox_goV1.UnimplementedDcimPoweroutletServer

	iDao dao.DcimPoweroutletDao
}

// NewDcimPoweroutletServer create a new service
func NewDcimPoweroutletServer() netbox_goV1.DcimPoweroutletServer {
	return &dcimPoweroutlet{
		iDao: dao.NewDcimPoweroutletDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimPoweroutletCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimPoweroutlet
func (s *dcimPoweroutlet) Create(ctx context.Context, req *netbox_goV1.CreateDcimPoweroutletRequest) (*netbox_goV1.CreateDcimPoweroutletReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimPoweroutlet{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimPoweroutlet.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimPoweroutlet", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimPoweroutletReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimPoweroutlet by id
func (s *dcimPoweroutlet) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimPoweroutletByIDRequest) (*netbox_goV1.DeleteDcimPoweroutletByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimPoweroutletByIDReply{}, nil
}

// UpdateByID update a dcimPoweroutlet by id
func (s *dcimPoweroutlet) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimPoweroutletByIDRequest) (*netbox_goV1.UpdateDcimPoweroutletByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimPoweroutlet{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimPoweroutlet.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimPoweroutlet", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimPoweroutletByIDReply{}, nil
}

// GetByID get a dcimPoweroutlet by id
func (s *dcimPoweroutlet) GetByID(ctx context.Context, req *netbox_goV1.GetDcimPoweroutletByIDRequest) (*netbox_goV1.GetDcimPoweroutletByIDReply, error) {
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

	data, err := convertDcimPoweroutlet(record)
	if err != nil {
		logger.Warn("convertDcimPoweroutlet error", logger.Err(err), logger.Any("dcimPoweroutlet", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimPoweroutlet.Err()
	}

	return &netbox_goV1.GetDcimPoweroutletByIDReply{DcimPoweroutlet: data}, nil
}

// List get a paginated list of dcimPoweroutlets by custom conditions
func (s *dcimPoweroutlet) List(ctx context.Context, req *netbox_goV1.ListDcimPoweroutletRequest) (*netbox_goV1.ListDcimPoweroutletReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimPoweroutlet.Err()
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

	dcimPoweroutlets := []*netbox_goV1.DcimPoweroutlet{}
	for _, record := range records {
		data, err := convertDcimPoweroutlet(record)
		if err != nil {
			logger.Warn("convertDcimPoweroutlet error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimPoweroutlets = append(dcimPoweroutlets, data)
	}

	return &netbox_goV1.ListDcimPoweroutletReply{
		Total:            total,
		DcimPoweroutlets: dcimPoweroutlets,
	}, nil
}

// DeleteByIDs batch delete dcimPoweroutlet by ids
func (s *dcimPoweroutlet) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimPoweroutletByIDsRequest) (*netbox_goV1.DeleteDcimPoweroutletByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimPoweroutletByIDsReply{}, nil
}

// GetByCondition get a dcimPoweroutlet by custom condition
func (s *dcimPoweroutlet) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimPoweroutletByConditionRequest) (*netbox_goV1.GetDcimPoweroutletByConditionReply, error) {
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

	data, err := convertDcimPoweroutlet(record)
	if err != nil {
		logger.Warn("convertDcimPoweroutlet error", logger.Err(err), logger.Any("dcimPoweroutlet", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimPoweroutlet.Err()
	}

	return &netbox_goV1.GetDcimPoweroutletByConditionReply{
		DcimPoweroutlet: data,
	}, nil
}

// ListByIDs batch get dcimPoweroutlet by ids
func (s *dcimPoweroutlet) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimPoweroutletByIDsRequest) (*netbox_goV1.ListDcimPoweroutletByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimPoweroutletMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimPoweroutlets := []*netbox_goV1.DcimPoweroutlet{}
	for _, id := range req.Ids {
		if v, ok := dcimPoweroutletMap[id]; ok {
			record, err := convertDcimPoweroutlet(v)
			if err != nil {
				logger.Warn("convertDcimPoweroutlet error", logger.Err(err), logger.Any("dcimPoweroutlet", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimPoweroutlets = append(dcimPoweroutlets, record)
		}
	}

	return &netbox_goV1.ListDcimPoweroutletByIDsReply{DcimPoweroutlets: dcimPoweroutlets}, nil
}

// ListByLastID get a paginated list of dcimPoweroutlets by last id
func (s *dcimPoweroutlet) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimPoweroutletByLastIDRequest) (*netbox_goV1.ListDcimPoweroutletByLastIDReply, error) {
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

	dcimPoweroutlets := []*netbox_goV1.DcimPoweroutlet{}
	for _, record := range records {
		data, err := convertDcimPoweroutlet(record)
		if err != nil {
			logger.Warn("convertDcimPoweroutlet error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimPoweroutlets = append(dcimPoweroutlets, data)
	}

	return &netbox_goV1.ListDcimPoweroutletByLastIDReply{
		DcimPoweroutlets: dcimPoweroutlets,
	}, nil
}

func convertDcimPoweroutlet(record *model.DcimPoweroutlet) (*netbox_goV1.DcimPoweroutlet, error) {
	value := &netbox_goV1.DcimPoweroutlet{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
