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
		netbox_goV1.RegisterDcimMacaddressServer(server, NewDcimMacaddressServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimMacaddressServer = (*dcimMacaddress)(nil)
var _ time.Time

type dcimMacaddress struct {
	netbox_goV1.UnimplementedDcimMacaddressServer

	iDao dao.DcimMacaddressDao
}

// NewDcimMacaddressServer create a new service
func NewDcimMacaddressServer() netbox_goV1.DcimMacaddressServer {
	return &dcimMacaddress{
		iDao: dao.NewDcimMacaddressDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimMacaddressCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimMacaddress
func (s *dcimMacaddress) Create(ctx context.Context, req *netbox_goV1.CreateDcimMacaddressRequest) (*netbox_goV1.CreateDcimMacaddressReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimMacaddress{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimMacaddress.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimMacaddress", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimMacaddressReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimMacaddress by id
func (s *dcimMacaddress) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimMacaddressByIDRequest) (*netbox_goV1.DeleteDcimMacaddressByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimMacaddressByIDReply{}, nil
}

// UpdateByID update a dcimMacaddress by id
func (s *dcimMacaddress) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimMacaddressByIDRequest) (*netbox_goV1.UpdateDcimMacaddressByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimMacaddress{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimMacaddress.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimMacaddress", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimMacaddressByIDReply{}, nil
}

// GetByID get a dcimMacaddress by id
func (s *dcimMacaddress) GetByID(ctx context.Context, req *netbox_goV1.GetDcimMacaddressByIDRequest) (*netbox_goV1.GetDcimMacaddressByIDReply, error) {
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

	data, err := convertDcimMacaddress(record)
	if err != nil {
		logger.Warn("convertDcimMacaddress error", logger.Err(err), logger.Any("dcimMacaddress", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimMacaddress.Err()
	}

	return &netbox_goV1.GetDcimMacaddressByIDReply{DcimMacaddress: data}, nil
}

// List get a paginated list of dcimMacaddresss by custom conditions
func (s *dcimMacaddress) List(ctx context.Context, req *netbox_goV1.ListDcimMacaddressRequest) (*netbox_goV1.ListDcimMacaddressReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimMacaddress.Err()
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

	dcimMacaddresss := []*netbox_goV1.DcimMacaddress{}
	for _, record := range records {
		data, err := convertDcimMacaddress(record)
		if err != nil {
			logger.Warn("convertDcimMacaddress error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimMacaddresss = append(dcimMacaddresss, data)
	}

	return &netbox_goV1.ListDcimMacaddressReply{
		Total:           total,
		DcimMacaddresss: dcimMacaddresss,
	}, nil
}

// DeleteByIDs batch delete dcimMacaddress by ids
func (s *dcimMacaddress) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimMacaddressByIDsRequest) (*netbox_goV1.DeleteDcimMacaddressByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimMacaddressByIDsReply{}, nil
}

// GetByCondition get a dcimMacaddress by custom condition
func (s *dcimMacaddress) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimMacaddressByConditionRequest) (*netbox_goV1.GetDcimMacaddressByConditionReply, error) {
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

	data, err := convertDcimMacaddress(record)
	if err != nil {
		logger.Warn("convertDcimMacaddress error", logger.Err(err), logger.Any("dcimMacaddress", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimMacaddress.Err()
	}

	return &netbox_goV1.GetDcimMacaddressByConditionReply{
		DcimMacaddress: data,
	}, nil
}

// ListByIDs batch get dcimMacaddress by ids
func (s *dcimMacaddress) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimMacaddressByIDsRequest) (*netbox_goV1.ListDcimMacaddressByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimMacaddressMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimMacaddresss := []*netbox_goV1.DcimMacaddress{}
	for _, id := range req.Ids {
		if v, ok := dcimMacaddressMap[id]; ok {
			record, err := convertDcimMacaddress(v)
			if err != nil {
				logger.Warn("convertDcimMacaddress error", logger.Err(err), logger.Any("dcimMacaddress", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimMacaddresss = append(dcimMacaddresss, record)
		}
	}

	return &netbox_goV1.ListDcimMacaddressByIDsReply{DcimMacaddresss: dcimMacaddresss}, nil
}

// ListByLastID get a paginated list of dcimMacaddresss by last id
func (s *dcimMacaddress) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimMacaddressByLastIDRequest) (*netbox_goV1.ListDcimMacaddressByLastIDReply, error) {
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

	dcimMacaddresss := []*netbox_goV1.DcimMacaddress{}
	for _, record := range records {
		data, err := convertDcimMacaddress(record)
		if err != nil {
			logger.Warn("convertDcimMacaddress error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimMacaddresss = append(dcimMacaddresss, data)
	}

	return &netbox_goV1.ListDcimMacaddressByLastIDReply{
		DcimMacaddresss: dcimMacaddresss,
	}, nil
}

func convertDcimMacaddress(record *model.DcimMacaddress) (*netbox_goV1.DcimMacaddress, error) {
	value := &netbox_goV1.DcimMacaddress{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
