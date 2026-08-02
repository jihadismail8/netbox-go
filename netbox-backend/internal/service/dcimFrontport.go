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
		netbox_goV1.RegisterDcimFrontportServer(server, NewDcimFrontportServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimFrontportServer = (*dcimFrontport)(nil)
var _ time.Time

type dcimFrontport struct {
	netbox_goV1.UnimplementedDcimFrontportServer

	iDao dao.DcimFrontportDao
}

// NewDcimFrontportServer create a new service
func NewDcimFrontportServer() netbox_goV1.DcimFrontportServer {
	return &dcimFrontport{
		iDao: dao.NewDcimFrontportDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimFrontportCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimFrontport
func (s *dcimFrontport) Create(ctx context.Context, req *netbox_goV1.CreateDcimFrontportRequest) (*netbox_goV1.CreateDcimFrontportReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimFrontport{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimFrontport.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimFrontport", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimFrontportReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimFrontport by id
func (s *dcimFrontport) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimFrontportByIDRequest) (*netbox_goV1.DeleteDcimFrontportByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimFrontportByIDReply{}, nil
}

// UpdateByID update a dcimFrontport by id
func (s *dcimFrontport) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimFrontportByIDRequest) (*netbox_goV1.UpdateDcimFrontportByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimFrontport{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimFrontport.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimFrontport", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimFrontportByIDReply{}, nil
}

// GetByID get a dcimFrontport by id
func (s *dcimFrontport) GetByID(ctx context.Context, req *netbox_goV1.GetDcimFrontportByIDRequest) (*netbox_goV1.GetDcimFrontportByIDReply, error) {
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

	data, err := convertDcimFrontport(record)
	if err != nil {
		logger.Warn("convertDcimFrontport error", logger.Err(err), logger.Any("dcimFrontport", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimFrontport.Err()
	}

	return &netbox_goV1.GetDcimFrontportByIDReply{DcimFrontport: data}, nil
}

// List get a paginated list of dcimFrontports by custom conditions
func (s *dcimFrontport) List(ctx context.Context, req *netbox_goV1.ListDcimFrontportRequest) (*netbox_goV1.ListDcimFrontportReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimFrontport.Err()
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

	dcimFrontports := []*netbox_goV1.DcimFrontport{}
	for _, record := range records {
		data, err := convertDcimFrontport(record)
		if err != nil {
			logger.Warn("convertDcimFrontport error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimFrontports = append(dcimFrontports, data)
	}

	return &netbox_goV1.ListDcimFrontportReply{
		Total:          total,
		DcimFrontports: dcimFrontports,
	}, nil
}

// DeleteByIDs batch delete dcimFrontport by ids
func (s *dcimFrontport) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimFrontportByIDsRequest) (*netbox_goV1.DeleteDcimFrontportByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimFrontportByIDsReply{}, nil
}

// GetByCondition get a dcimFrontport by custom condition
func (s *dcimFrontport) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimFrontportByConditionRequest) (*netbox_goV1.GetDcimFrontportByConditionReply, error) {
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

	data, err := convertDcimFrontport(record)
	if err != nil {
		logger.Warn("convertDcimFrontport error", logger.Err(err), logger.Any("dcimFrontport", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimFrontport.Err()
	}

	return &netbox_goV1.GetDcimFrontportByConditionReply{
		DcimFrontport: data,
	}, nil
}

// ListByIDs batch get dcimFrontport by ids
func (s *dcimFrontport) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimFrontportByIDsRequest) (*netbox_goV1.ListDcimFrontportByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimFrontportMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimFrontports := []*netbox_goV1.DcimFrontport{}
	for _, id := range req.Ids {
		if v, ok := dcimFrontportMap[id]; ok {
			record, err := convertDcimFrontport(v)
			if err != nil {
				logger.Warn("convertDcimFrontport error", logger.Err(err), logger.Any("dcimFrontport", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimFrontports = append(dcimFrontports, record)
		}
	}

	return &netbox_goV1.ListDcimFrontportByIDsReply{DcimFrontports: dcimFrontports}, nil
}

// ListByLastID get a paginated list of dcimFrontports by last id
func (s *dcimFrontport) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimFrontportByLastIDRequest) (*netbox_goV1.ListDcimFrontportByLastIDReply, error) {
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

	dcimFrontports := []*netbox_goV1.DcimFrontport{}
	for _, record := range records {
		data, err := convertDcimFrontport(record)
		if err != nil {
			logger.Warn("convertDcimFrontport error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimFrontports = append(dcimFrontports, data)
	}

	return &netbox_goV1.ListDcimFrontportByLastIDReply{
		DcimFrontports: dcimFrontports,
	}, nil
}

func convertDcimFrontport(record *model.DcimFrontport) (*netbox_goV1.DcimFrontport, error) {
	value := &netbox_goV1.DcimFrontport{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
