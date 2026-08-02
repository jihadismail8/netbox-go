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
		netbox_goV1.RegisterDcimInventoryitemtemplateServer(server, NewDcimInventoryitemtemplateServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimInventoryitemtemplateServer = (*dcimInventoryitemtemplate)(nil)
var _ time.Time

type dcimInventoryitemtemplate struct {
	netbox_goV1.UnimplementedDcimInventoryitemtemplateServer

	iDao dao.DcimInventoryitemtemplateDao
}

// NewDcimInventoryitemtemplateServer create a new service
func NewDcimInventoryitemtemplateServer() netbox_goV1.DcimInventoryitemtemplateServer {
	return &dcimInventoryitemtemplate{
		iDao: dao.NewDcimInventoryitemtemplateDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimInventoryitemtemplateCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimInventoryitemtemplate
func (s *dcimInventoryitemtemplate) Create(ctx context.Context, req *netbox_goV1.CreateDcimInventoryitemtemplateRequest) (*netbox_goV1.CreateDcimInventoryitemtemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimInventoryitemtemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimInventoryitemtemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimInventoryitemtemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimInventoryitemtemplateReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimInventoryitemtemplate by id
func (s *dcimInventoryitemtemplate) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimInventoryitemtemplateByIDRequest) (*netbox_goV1.DeleteDcimInventoryitemtemplateByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimInventoryitemtemplateByIDReply{}, nil
}

// UpdateByID update a dcimInventoryitemtemplate by id
func (s *dcimInventoryitemtemplate) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimInventoryitemtemplateByIDRequest) (*netbox_goV1.UpdateDcimInventoryitemtemplateByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimInventoryitemtemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimInventoryitemtemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimInventoryitemtemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimInventoryitemtemplateByIDReply{}, nil
}

// GetByID get a dcimInventoryitemtemplate by id
func (s *dcimInventoryitemtemplate) GetByID(ctx context.Context, req *netbox_goV1.GetDcimInventoryitemtemplateByIDRequest) (*netbox_goV1.GetDcimInventoryitemtemplateByIDReply, error) {
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

	data, err := convertDcimInventoryitemtemplate(record)
	if err != nil {
		logger.Warn("convertDcimInventoryitemtemplate error", logger.Err(err), logger.Any("dcimInventoryitemtemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimInventoryitemtemplate.Err()
	}

	return &netbox_goV1.GetDcimInventoryitemtemplateByIDReply{DcimInventoryitemtemplate: data}, nil
}

// List get a paginated list of dcimInventoryitemtemplates by custom conditions
func (s *dcimInventoryitemtemplate) List(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemtemplateRequest) (*netbox_goV1.ListDcimInventoryitemtemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimInventoryitemtemplate.Err()
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

	dcimInventoryitemtemplates := []*netbox_goV1.DcimInventoryitemtemplate{}
	for _, record := range records {
		data, err := convertDcimInventoryitemtemplate(record)
		if err != nil {
			logger.Warn("convertDcimInventoryitemtemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimInventoryitemtemplates = append(dcimInventoryitemtemplates, data)
	}

	return &netbox_goV1.ListDcimInventoryitemtemplateReply{
		Total:                      total,
		DcimInventoryitemtemplates: dcimInventoryitemtemplates,
	}, nil
}

// DeleteByIDs batch delete dcimInventoryitemtemplate by ids
func (s *dcimInventoryitemtemplate) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimInventoryitemtemplateByIDsRequest) (*netbox_goV1.DeleteDcimInventoryitemtemplateByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimInventoryitemtemplateByIDsReply{}, nil
}

// GetByCondition get a dcimInventoryitemtemplate by custom condition
func (s *dcimInventoryitemtemplate) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimInventoryitemtemplateByConditionRequest) (*netbox_goV1.GetDcimInventoryitemtemplateByConditionReply, error) {
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

	data, err := convertDcimInventoryitemtemplate(record)
	if err != nil {
		logger.Warn("convertDcimInventoryitemtemplate error", logger.Err(err), logger.Any("dcimInventoryitemtemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimInventoryitemtemplate.Err()
	}

	return &netbox_goV1.GetDcimInventoryitemtemplateByConditionReply{
		DcimInventoryitemtemplate: data,
	}, nil
}

// ListByIDs batch get dcimInventoryitemtemplate by ids
func (s *dcimInventoryitemtemplate) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemtemplateByIDsRequest) (*netbox_goV1.ListDcimInventoryitemtemplateByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimInventoryitemtemplateMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimInventoryitemtemplates := []*netbox_goV1.DcimInventoryitemtemplate{}
	for _, id := range req.Ids {
		if v, ok := dcimInventoryitemtemplateMap[id]; ok {
			record, err := convertDcimInventoryitemtemplate(v)
			if err != nil {
				logger.Warn("convertDcimInventoryitemtemplate error", logger.Err(err), logger.Any("dcimInventoryitemtemplate", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimInventoryitemtemplates = append(dcimInventoryitemtemplates, record)
		}
	}

	return &netbox_goV1.ListDcimInventoryitemtemplateByIDsReply{DcimInventoryitemtemplates: dcimInventoryitemtemplates}, nil
}

// ListByLastID get a paginated list of dcimInventoryitemtemplates by last id
func (s *dcimInventoryitemtemplate) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimInventoryitemtemplateByLastIDRequest) (*netbox_goV1.ListDcimInventoryitemtemplateByLastIDReply, error) {
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

	dcimInventoryitemtemplates := []*netbox_goV1.DcimInventoryitemtemplate{}
	for _, record := range records {
		data, err := convertDcimInventoryitemtemplate(record)
		if err != nil {
			logger.Warn("convertDcimInventoryitemtemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimInventoryitemtemplates = append(dcimInventoryitemtemplates, data)
	}

	return &netbox_goV1.ListDcimInventoryitemtemplateByLastIDReply{
		DcimInventoryitemtemplates: dcimInventoryitemtemplates,
	}, nil
}

func convertDcimInventoryitemtemplate(record *model.DcimInventoryitemtemplate) (*netbox_goV1.DcimInventoryitemtemplate, error) {
	value := &netbox_goV1.DcimInventoryitemtemplate{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
