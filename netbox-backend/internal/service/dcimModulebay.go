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
		netbox_goV1.RegisterDcimModulebayServer(server, NewDcimModulebayServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimModulebayServer = (*dcimModulebay)(nil)
var _ time.Time

type dcimModulebay struct {
	netbox_goV1.UnimplementedDcimModulebayServer

	iDao dao.DcimModulebayDao
}

// NewDcimModulebayServer create a new service
func NewDcimModulebayServer() netbox_goV1.DcimModulebayServer {
	return &dcimModulebay{
		iDao: dao.NewDcimModulebayDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimModulebayCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimModulebay
func (s *dcimModulebay) Create(ctx context.Context, req *netbox_goV1.CreateDcimModulebayRequest) (*netbox_goV1.CreateDcimModulebayReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimModulebay{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimModulebay.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimModulebay", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimModulebayReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimModulebay by id
func (s *dcimModulebay) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimModulebayByIDRequest) (*netbox_goV1.DeleteDcimModulebayByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimModulebayByIDReply{}, nil
}

// UpdateByID update a dcimModulebay by id
func (s *dcimModulebay) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimModulebayByIDRequest) (*netbox_goV1.UpdateDcimModulebayByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimModulebay{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimModulebay.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimModulebay", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimModulebayByIDReply{}, nil
}

// GetByID get a dcimModulebay by id
func (s *dcimModulebay) GetByID(ctx context.Context, req *netbox_goV1.GetDcimModulebayByIDRequest) (*netbox_goV1.GetDcimModulebayByIDReply, error) {
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

	data, err := convertDcimModulebay(record)
	if err != nil {
		logger.Warn("convertDcimModulebay error", logger.Err(err), logger.Any("dcimModulebay", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimModulebay.Err()
	}

	return &netbox_goV1.GetDcimModulebayByIDReply{DcimModulebay: data}, nil
}

// List get a paginated list of dcimModulebays by custom conditions
func (s *dcimModulebay) List(ctx context.Context, req *netbox_goV1.ListDcimModulebayRequest) (*netbox_goV1.ListDcimModulebayReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimModulebay.Err()
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

	dcimModulebays := []*netbox_goV1.DcimModulebay{}
	for _, record := range records {
		data, err := convertDcimModulebay(record)
		if err != nil {
			logger.Warn("convertDcimModulebay error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimModulebays = append(dcimModulebays, data)
	}

	return &netbox_goV1.ListDcimModulebayReply{
		Total:          total,
		DcimModulebays: dcimModulebays,
	}, nil
}

// DeleteByIDs batch delete dcimModulebay by ids
func (s *dcimModulebay) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimModulebayByIDsRequest) (*netbox_goV1.DeleteDcimModulebayByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimModulebayByIDsReply{}, nil
}

// GetByCondition get a dcimModulebay by custom condition
func (s *dcimModulebay) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimModulebayByConditionRequest) (*netbox_goV1.GetDcimModulebayByConditionReply, error) {
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

	data, err := convertDcimModulebay(record)
	if err != nil {
		logger.Warn("convertDcimModulebay error", logger.Err(err), logger.Any("dcimModulebay", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimModulebay.Err()
	}

	return &netbox_goV1.GetDcimModulebayByConditionReply{
		DcimModulebay: data,
	}, nil
}

// ListByIDs batch get dcimModulebay by ids
func (s *dcimModulebay) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimModulebayByIDsRequest) (*netbox_goV1.ListDcimModulebayByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimModulebayMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimModulebays := []*netbox_goV1.DcimModulebay{}
	for _, id := range req.Ids {
		if v, ok := dcimModulebayMap[id]; ok {
			record, err := convertDcimModulebay(v)
			if err != nil {
				logger.Warn("convertDcimModulebay error", logger.Err(err), logger.Any("dcimModulebay", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimModulebays = append(dcimModulebays, record)
		}
	}

	return &netbox_goV1.ListDcimModulebayByIDsReply{DcimModulebays: dcimModulebays}, nil
}

// ListByLastID get a paginated list of dcimModulebays by last id
func (s *dcimModulebay) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimModulebayByLastIDRequest) (*netbox_goV1.ListDcimModulebayByLastIDReply, error) {
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

	dcimModulebays := []*netbox_goV1.DcimModulebay{}
	for _, record := range records {
		data, err := convertDcimModulebay(record)
		if err != nil {
			logger.Warn("convertDcimModulebay error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimModulebays = append(dcimModulebays, data)
	}

	return &netbox_goV1.ListDcimModulebayByLastIDReply{
		DcimModulebays: dcimModulebays,
	}, nil
}

func convertDcimModulebay(record *model.DcimModulebay) (*netbox_goV1.DcimModulebay, error) {
	value := &netbox_goV1.DcimModulebay{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
