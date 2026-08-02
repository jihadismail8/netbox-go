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
		netbox_goV1.RegisterDcimPowerportServer(server, NewDcimPowerportServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimPowerportServer = (*dcimPowerport)(nil)
var _ time.Time

type dcimPowerport struct {
	netbox_goV1.UnimplementedDcimPowerportServer

	iDao dao.DcimPowerportDao
}

// NewDcimPowerportServer create a new service
func NewDcimPowerportServer() netbox_goV1.DcimPowerportServer {
	return &dcimPowerport{
		iDao: dao.NewDcimPowerportDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimPowerportCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimPowerport
func (s *dcimPowerport) Create(ctx context.Context, req *netbox_goV1.CreateDcimPowerportRequest) (*netbox_goV1.CreateDcimPowerportReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimPowerport{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimPowerport.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimPowerport", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimPowerportReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimPowerport by id
func (s *dcimPowerport) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimPowerportByIDRequest) (*netbox_goV1.DeleteDcimPowerportByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimPowerportByIDReply{}, nil
}

// UpdateByID update a dcimPowerport by id
func (s *dcimPowerport) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimPowerportByIDRequest) (*netbox_goV1.UpdateDcimPowerportByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimPowerport{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimPowerport.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimPowerport", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimPowerportByIDReply{}, nil
}

// GetByID get a dcimPowerport by id
func (s *dcimPowerport) GetByID(ctx context.Context, req *netbox_goV1.GetDcimPowerportByIDRequest) (*netbox_goV1.GetDcimPowerportByIDReply, error) {
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

	data, err := convertDcimPowerport(record)
	if err != nil {
		logger.Warn("convertDcimPowerport error", logger.Err(err), logger.Any("dcimPowerport", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimPowerport.Err()
	}

	return &netbox_goV1.GetDcimPowerportByIDReply{DcimPowerport: data}, nil
}

// List get a paginated list of dcimPowerports by custom conditions
func (s *dcimPowerport) List(ctx context.Context, req *netbox_goV1.ListDcimPowerportRequest) (*netbox_goV1.ListDcimPowerportReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimPowerport.Err()
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

	dcimPowerports := []*netbox_goV1.DcimPowerport{}
	for _, record := range records {
		data, err := convertDcimPowerport(record)
		if err != nil {
			logger.Warn("convertDcimPowerport error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimPowerports = append(dcimPowerports, data)
	}

	return &netbox_goV1.ListDcimPowerportReply{
		Total:          total,
		DcimPowerports: dcimPowerports,
	}, nil
}

// DeleteByIDs batch delete dcimPowerport by ids
func (s *dcimPowerport) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimPowerportByIDsRequest) (*netbox_goV1.DeleteDcimPowerportByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimPowerportByIDsReply{}, nil
}

// GetByCondition get a dcimPowerport by custom condition
func (s *dcimPowerport) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimPowerportByConditionRequest) (*netbox_goV1.GetDcimPowerportByConditionReply, error) {
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

	data, err := convertDcimPowerport(record)
	if err != nil {
		logger.Warn("convertDcimPowerport error", logger.Err(err), logger.Any("dcimPowerport", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimPowerport.Err()
	}

	return &netbox_goV1.GetDcimPowerportByConditionReply{
		DcimPowerport: data,
	}, nil
}

// ListByIDs batch get dcimPowerport by ids
func (s *dcimPowerport) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimPowerportByIDsRequest) (*netbox_goV1.ListDcimPowerportByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimPowerportMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimPowerports := []*netbox_goV1.DcimPowerport{}
	for _, id := range req.Ids {
		if v, ok := dcimPowerportMap[id]; ok {
			record, err := convertDcimPowerport(v)
			if err != nil {
				logger.Warn("convertDcimPowerport error", logger.Err(err), logger.Any("dcimPowerport", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimPowerports = append(dcimPowerports, record)
		}
	}

	return &netbox_goV1.ListDcimPowerportByIDsReply{DcimPowerports: dcimPowerports}, nil
}

// ListByLastID get a paginated list of dcimPowerports by last id
func (s *dcimPowerport) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimPowerportByLastIDRequest) (*netbox_goV1.ListDcimPowerportByLastIDReply, error) {
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

	dcimPowerports := []*netbox_goV1.DcimPowerport{}
	for _, record := range records {
		data, err := convertDcimPowerport(record)
		if err != nil {
			logger.Warn("convertDcimPowerport error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimPowerports = append(dcimPowerports, data)
	}

	return &netbox_goV1.ListDcimPowerportByLastIDReply{
		DcimPowerports: dcimPowerports,
	}, nil
}

func convertDcimPowerport(record *model.DcimPowerport) (*netbox_goV1.DcimPowerport, error) {
	value := &netbox_goV1.DcimPowerport{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
