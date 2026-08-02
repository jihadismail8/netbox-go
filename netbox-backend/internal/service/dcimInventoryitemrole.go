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
		netbox_goV1.RegisterDcimInventoryitemroleServer(server, NewDcimInventoryitemroleServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimInventoryitemroleServer = (*dcimInventoryitemrole)(nil)
var _ time.Time

type dcimInventoryitemrole struct {
	netbox_goV1.UnimplementedDcimInventoryitemroleServer

	iDao dao.DcimInventoryitemroleDao
}

// NewDcimInventoryitemroleServer create a new service
func NewDcimInventoryitemroleServer() netbox_goV1.DcimInventoryitemroleServer {
	return &dcimInventoryitemrole{
		iDao: dao.NewDcimInventoryitemroleDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimInventoryitemroleCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimInventoryitemrole
func (s *dcimInventoryitemrole) Create(ctx context.Context, req *netbox_goV1.CreateDcimInventoryitemroleRequest) (*netbox_goV1.CreateDcimInventoryitemroleReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimInventoryitemrole{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimInventoryitemrole.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimInventoryitemrole", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimInventoryitemroleReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimInventoryitemrole by id
func (s *dcimInventoryitemrole) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimInventoryitemroleByIDRequest) (*netbox_goV1.DeleteDcimInventoryitemroleByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimInventoryitemroleByIDReply{}, nil
}

// UpdateByID update a dcimInventoryitemrole by id
func (s *dcimInventoryitemrole) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimInventoryitemroleByIDRequest) (*netbox_goV1.UpdateDcimInventoryitemroleByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimInventoryitemrole{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimInventoryitemrole.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimInventoryitemrole", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimInventoryitemroleByIDReply{}, nil
}

// GetByID get a dcimInventoryitemrole by id
func (s *dcimInventoryitemrole) GetByID(ctx context.Context, req *netbox_goV1.GetDcimInventoryitemroleByIDRequest) (*netbox_goV1.GetDcimInventoryitemroleByIDReply, error) {
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

	data, err := convertDcimInventoryitemrole(record)
	if err != nil {
		logger.Warn("convertDcimInventoryitemrole error", logger.Err(err), logger.Any("dcimInventoryitemrole", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimInventoryitemrole.Err()
	}

	return &netbox_goV1.GetDcimInventoryitemroleByIDReply{DcimInventoryitemrole: data}, nil
}

// List get a paginated list of dcimInventoryitemroles by custom conditions
func (s *dcimInventoryitemrole) List(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemroleRequest) (*netbox_goV1.ListDcimInventoryitemroleReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimInventoryitemrole.Err()
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

	dcimInventoryitemroles := []*netbox_goV1.DcimInventoryitemrole{}
	for _, record := range records {
		data, err := convertDcimInventoryitemrole(record)
		if err != nil {
			logger.Warn("convertDcimInventoryitemrole error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimInventoryitemroles = append(dcimInventoryitemroles, data)
	}

	return &netbox_goV1.ListDcimInventoryitemroleReply{
		Total:                  total,
		DcimInventoryitemroles: dcimInventoryitemroles,
	}, nil
}

// DeleteByIDs batch delete dcimInventoryitemrole by ids
func (s *dcimInventoryitemrole) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimInventoryitemroleByIDsRequest) (*netbox_goV1.DeleteDcimInventoryitemroleByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimInventoryitemroleByIDsReply{}, nil
}

// GetByCondition get a dcimInventoryitemrole by custom condition
func (s *dcimInventoryitemrole) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimInventoryitemroleByConditionRequest) (*netbox_goV1.GetDcimInventoryitemroleByConditionReply, error) {
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

	data, err := convertDcimInventoryitemrole(record)
	if err != nil {
		logger.Warn("convertDcimInventoryitemrole error", logger.Err(err), logger.Any("dcimInventoryitemrole", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimInventoryitemrole.Err()
	}

	return &netbox_goV1.GetDcimInventoryitemroleByConditionReply{
		DcimInventoryitemrole: data,
	}, nil
}

// ListByIDs batch get dcimInventoryitemrole by ids
func (s *dcimInventoryitemrole) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemroleByIDsRequest) (*netbox_goV1.ListDcimInventoryitemroleByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimInventoryitemroleMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimInventoryitemroles := []*netbox_goV1.DcimInventoryitemrole{}
	for _, id := range req.Ids {
		if v, ok := dcimInventoryitemroleMap[id]; ok {
			record, err := convertDcimInventoryitemrole(v)
			if err != nil {
				logger.Warn("convertDcimInventoryitemrole error", logger.Err(err), logger.Any("dcimInventoryitemrole", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimInventoryitemroles = append(dcimInventoryitemroles, record)
		}
	}

	return &netbox_goV1.ListDcimInventoryitemroleByIDsReply{DcimInventoryitemroles: dcimInventoryitemroles}, nil
}

// ListByLastID get a paginated list of dcimInventoryitemroles by last id
func (s *dcimInventoryitemrole) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemroleByLastIDRequest) (*netbox_goV1.ListDcimInventoryitemroleByLastIDReply, error) {
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

	dcimInventoryitemroles := []*netbox_goV1.DcimInventoryitemrole{}
	for _, record := range records {
		data, err := convertDcimInventoryitemrole(record)
		if err != nil {
			logger.Warn("convertDcimInventoryitemrole error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimInventoryitemroles = append(dcimInventoryitemroles, data)
	}

	return &netbox_goV1.ListDcimInventoryitemroleByLastIDReply{
		DcimInventoryitemroles: dcimInventoryitemroles,
	}, nil
}

func convertDcimInventoryitemrole(record *model.DcimInventoryitemrole) (*netbox_goV1.DcimInventoryitemrole, error) {
	value := &netbox_goV1.DcimInventoryitemrole{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
