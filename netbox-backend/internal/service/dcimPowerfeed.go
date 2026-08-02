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
		netbox_goV1.RegisterDcimPowerfeedServer(server, NewDcimPowerfeedServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimPowerfeedServer = (*dcimPowerfeed)(nil)
var _ time.Time

type dcimPowerfeed struct {
	netbox_goV1.UnimplementedDcimPowerfeedServer

	iDao dao.DcimPowerfeedDao
}

// NewDcimPowerfeedServer create a new service
func NewDcimPowerfeedServer() netbox_goV1.DcimPowerfeedServer {
	return &dcimPowerfeed{
		iDao: dao.NewDcimPowerfeedDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimPowerfeedCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimPowerfeed
func (s *dcimPowerfeed) Create(ctx context.Context, req *netbox_goV1.CreateDcimPowerfeedRequest) (*netbox_goV1.CreateDcimPowerfeedReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimPowerfeed{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimPowerfeed.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimPowerfeed", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimPowerfeedReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimPowerfeed by id
func (s *dcimPowerfeed) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimPowerfeedByIDRequest) (*netbox_goV1.DeleteDcimPowerfeedByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimPowerfeedByIDReply{}, nil
}

// UpdateByID update a dcimPowerfeed by id
func (s *dcimPowerfeed) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimPowerfeedByIDRequest) (*netbox_goV1.UpdateDcimPowerfeedByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimPowerfeed{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimPowerfeed.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimPowerfeed", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimPowerfeedByIDReply{}, nil
}

// GetByID get a dcimPowerfeed by id
func (s *dcimPowerfeed) GetByID(ctx context.Context, req *netbox_goV1.GetDcimPowerfeedByIDRequest) (*netbox_goV1.GetDcimPowerfeedByIDReply, error) {
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

	data, err := convertDcimPowerfeed(record)
	if err != nil {
		logger.Warn("convertDcimPowerfeed error", logger.Err(err), logger.Any("dcimPowerfeed", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimPowerfeed.Err()
	}

	return &netbox_goV1.GetDcimPowerfeedByIDReply{DcimPowerfeed: data}, nil
}

// List get a paginated list of dcimPowerfeeds by custom conditions
func (s *dcimPowerfeed) List(ctx context.Context, req *netbox_goV1.ListDcimPowerfeedRequest) (*netbox_goV1.ListDcimPowerfeedReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimPowerfeed.Err()
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

	dcimPowerfeeds := []*netbox_goV1.DcimPowerfeed{}
	for _, record := range records {
		data, err := convertDcimPowerfeed(record)
		if err != nil {
			logger.Warn("convertDcimPowerfeed error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimPowerfeeds = append(dcimPowerfeeds, data)
	}

	return &netbox_goV1.ListDcimPowerfeedReply{
		Total:          total,
		DcimPowerfeeds: dcimPowerfeeds,
	}, nil
}

// DeleteByIDs batch delete dcimPowerfeed by ids
func (s *dcimPowerfeed) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimPowerfeedByIDsRequest) (*netbox_goV1.DeleteDcimPowerfeedByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimPowerfeedByIDsReply{}, nil
}

// GetByCondition get a dcimPowerfeed by custom condition
func (s *dcimPowerfeed) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimPowerfeedByConditionRequest) (*netbox_goV1.GetDcimPowerfeedByConditionReply, error) {
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

	data, err := convertDcimPowerfeed(record)
	if err != nil {
		logger.Warn("convertDcimPowerfeed error", logger.Err(err), logger.Any("dcimPowerfeed", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimPowerfeed.Err()
	}

	return &netbox_goV1.GetDcimPowerfeedByConditionReply{
		DcimPowerfeed: data,
	}, nil
}

// ListByIDs batch get dcimPowerfeed by ids
func (s *dcimPowerfeed) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimPowerfeedByIDsRequest) (*netbox_goV1.ListDcimPowerfeedByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimPowerfeedMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimPowerfeeds := []*netbox_goV1.DcimPowerfeed{}
	for _, id := range req.Ids {
		if v, ok := dcimPowerfeedMap[id]; ok {
			record, err := convertDcimPowerfeed(v)
			if err != nil {
				logger.Warn("convertDcimPowerfeed error", logger.Err(err), logger.Any("dcimPowerfeed", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimPowerfeeds = append(dcimPowerfeeds, record)
		}
	}

	return &netbox_goV1.ListDcimPowerfeedByIDsReply{DcimPowerfeeds: dcimPowerfeeds}, nil
}

// ListByLastID get a paginated list of dcimPowerfeeds by last id
func (s *dcimPowerfeed) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimPowerfeedByLastIDRequest) (*netbox_goV1.ListDcimPowerfeedByLastIDReply, error) {
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

	dcimPowerfeeds := []*netbox_goV1.DcimPowerfeed{}
	for _, record := range records {
		data, err := convertDcimPowerfeed(record)
		if err != nil {
			logger.Warn("convertDcimPowerfeed error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimPowerfeeds = append(dcimPowerfeeds, data)
	}

	return &netbox_goV1.ListDcimPowerfeedByLastIDReply{
		DcimPowerfeeds: dcimPowerfeeds,
	}, nil
}

func convertDcimPowerfeed(record *model.DcimPowerfeed) (*netbox_goV1.DcimPowerfeed, error) {
	value := &netbox_goV1.DcimPowerfeed{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
