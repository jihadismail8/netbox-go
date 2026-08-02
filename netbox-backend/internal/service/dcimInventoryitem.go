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
		netbox_goV1.RegisterDcimInventoryitemServer(server, NewDcimInventoryitemServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimInventoryitemServer = (*dcimInventoryitem)(nil)
var _ time.Time

type dcimInventoryitem struct {
	netbox_goV1.UnimplementedDcimInventoryitemServer

	iDao dao.DcimInventoryitemDao
}

// NewDcimInventoryitemServer create a new service
func NewDcimInventoryitemServer() netbox_goV1.DcimInventoryitemServer {
	return &dcimInventoryitem{
		iDao: dao.NewDcimInventoryitemDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimInventoryitemCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimInventoryitem
func (s *dcimInventoryitem) Create(ctx context.Context, req *netbox_goV1.CreateDcimInventoryitemRequest) (*netbox_goV1.CreateDcimInventoryitemReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimInventoryitem{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimInventoryitem.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimInventoryitem", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimInventoryitemReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimInventoryitem by id
func (s *dcimInventoryitem) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimInventoryitemByIDRequest) (*netbox_goV1.DeleteDcimInventoryitemByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimInventoryitemByIDReply{}, nil
}

// UpdateByID update a dcimInventoryitem by id
func (s *dcimInventoryitem) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimInventoryitemByIDRequest) (*netbox_goV1.UpdateDcimInventoryitemByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimInventoryitem{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimInventoryitem.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimInventoryitem", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimInventoryitemByIDReply{}, nil
}

// GetByID get a dcimInventoryitem by id
func (s *dcimInventoryitem) GetByID(ctx context.Context, req *netbox_goV1.GetDcimInventoryitemByIDRequest) (*netbox_goV1.GetDcimInventoryitemByIDReply, error) {
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

	data, err := convertDcimInventoryitem(record)
	if err != nil {
		logger.Warn("convertDcimInventoryitem error", logger.Err(err), logger.Any("dcimInventoryitem", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimInventoryitem.Err()
	}

	return &netbox_goV1.GetDcimInventoryitemByIDReply{DcimInventoryitem: data}, nil
}

// List get a paginated list of dcimInventoryitems by custom conditions
func (s *dcimInventoryitem) List(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemRequest) (*netbox_goV1.ListDcimInventoryitemReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimInventoryitem.Err()
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

	dcimInventoryitems := []*netbox_goV1.DcimInventoryitem{}
	for _, record := range records {
		data, err := convertDcimInventoryitem(record)
		if err != nil {
			logger.Warn("convertDcimInventoryitem error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimInventoryitems = append(dcimInventoryitems, data)
	}

	return &netbox_goV1.ListDcimInventoryitemReply{
		Total:              total,
		DcimInventoryitems: dcimInventoryitems,
	}, nil
}

// DeleteByIDs batch delete dcimInventoryitem by ids
func (s *dcimInventoryitem) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimInventoryitemByIDsRequest) (*netbox_goV1.DeleteDcimInventoryitemByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimInventoryitemByIDsReply{}, nil
}

// GetByCondition get a dcimInventoryitem by custom condition
func (s *dcimInventoryitem) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimInventoryitemByConditionRequest) (*netbox_goV1.GetDcimInventoryitemByConditionReply, error) {
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

	data, err := convertDcimInventoryitem(record)
	if err != nil {
		logger.Warn("convertDcimInventoryitem error", logger.Err(err), logger.Any("dcimInventoryitem", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimInventoryitem.Err()
	}

	return &netbox_goV1.GetDcimInventoryitemByConditionReply{
		DcimInventoryitem: data,
	}, nil
}

// ListByIDs batch get dcimInventoryitem by ids
func (s *dcimInventoryitem) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemByIDsRequest) (*netbox_goV1.ListDcimInventoryitemByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimInventoryitemMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimInventoryitems := []*netbox_goV1.DcimInventoryitem{}
	for _, id := range req.Ids {
		if v, ok := dcimInventoryitemMap[id]; ok {
			record, err := convertDcimInventoryitem(v)
			if err != nil {
				logger.Warn("convertDcimInventoryitem error", logger.Err(err), logger.Any("dcimInventoryitem", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimInventoryitems = append(dcimInventoryitems, record)
		}
	}

	return &netbox_goV1.ListDcimInventoryitemByIDsReply{DcimInventoryitems: dcimInventoryitems}, nil
}

// ListByLastID get a paginated list of dcimInventoryitems by last id
func (s *dcimInventoryitem) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemByLastIDRequest) (*netbox_goV1.ListDcimInventoryitemByLastIDReply, error) {
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

	dcimInventoryitems := []*netbox_goV1.DcimInventoryitem{}
	for _, record := range records {
		data, err := convertDcimInventoryitem(record)
		if err != nil {
			logger.Warn("convertDcimInventoryitem error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimInventoryitems = append(dcimInventoryitems, data)
	}

	return &netbox_goV1.ListDcimInventoryitemByLastIDReply{
		DcimInventoryitems: dcimInventoryitems,
	}, nil
}

func convertDcimInventoryitem(record *model.DcimInventoryitem) (*netbox_goV1.DcimInventoryitem, error) {
	value := &netbox_goV1.DcimInventoryitem{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
