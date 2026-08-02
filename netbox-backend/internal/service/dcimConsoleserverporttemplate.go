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
		netbox_goV1.RegisterDcimConsoleserverporttemplateServer(server, NewDcimConsoleserverporttemplateServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimConsoleserverporttemplateServer = (*dcimConsoleserverporttemplate)(nil)
var _ time.Time

type dcimConsoleserverporttemplate struct {
	netbox_goV1.UnimplementedDcimConsoleserverporttemplateServer

	iDao dao.DcimConsoleserverporttemplateDao
}

// NewDcimConsoleserverporttemplateServer create a new service
func NewDcimConsoleserverporttemplateServer() netbox_goV1.DcimConsoleserverporttemplateServer {
	return &dcimConsoleserverporttemplate{
		iDao: dao.NewDcimConsoleserverporttemplateDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimConsoleserverporttemplateCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimConsoleserverporttemplate
func (s *dcimConsoleserverporttemplate) Create(ctx context.Context, req *netbox_goV1.CreateDcimConsoleserverporttemplateRequest) (*netbox_goV1.CreateDcimConsoleserverporttemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimConsoleserverporttemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimConsoleserverporttemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimConsoleserverporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimConsoleserverporttemplateReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimConsoleserverporttemplate by id
func (s *dcimConsoleserverporttemplate) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimConsoleserverporttemplateByIDRequest) (*netbox_goV1.DeleteDcimConsoleserverporttemplateByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimConsoleserverporttemplateByIDReply{}, nil
}

// UpdateByID update a dcimConsoleserverporttemplate by id
func (s *dcimConsoleserverporttemplate) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimConsoleserverporttemplateByIDRequest) (*netbox_goV1.UpdateDcimConsoleserverporttemplateByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimConsoleserverporttemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimConsoleserverporttemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimConsoleserverporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimConsoleserverporttemplateByIDReply{}, nil
}

// GetByID get a dcimConsoleserverporttemplate by id
func (s *dcimConsoleserverporttemplate) GetByID(ctx context.Context, req *netbox_goV1.GetDcimConsoleserverporttemplateByIDRequest) (*netbox_goV1.GetDcimConsoleserverporttemplateByIDReply, error) {
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

	data, err := convertDcimConsoleserverporttemplate(record)
	if err != nil {
		logger.Warn("convertDcimConsoleserverporttemplate error", logger.Err(err), logger.Any("dcimConsoleserverporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimConsoleserverporttemplate.Err()
	}

	return &netbox_goV1.GetDcimConsoleserverporttemplateByIDReply{DcimConsoleserverporttemplate: data}, nil
}

// List get a paginated list of dcimConsoleserverporttemplates by custom conditions
func (s *dcimConsoleserverporttemplate) List(ctx context.Context, req *netbox_goV1.ListDcimConsoleserverporttemplateRequest) (*netbox_goV1.ListDcimConsoleserverporttemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimConsoleserverporttemplate.Err()
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

	dcimConsoleserverporttemplates := []*netbox_goV1.DcimConsoleserverporttemplate{}
	for _, record := range records {
		data, err := convertDcimConsoleserverporttemplate(record)
		if err != nil {
			logger.Warn("convertDcimConsoleserverporttemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimConsoleserverporttemplates = append(dcimConsoleserverporttemplates, data)
	}

	return &netbox_goV1.ListDcimConsoleserverporttemplateReply{
		Total:                          total,
		DcimConsoleserverporttemplates: dcimConsoleserverporttemplates,
	}, nil
}

// DeleteByIDs batch delete dcimConsoleserverporttemplate by ids
func (s *dcimConsoleserverporttemplate) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimConsoleserverporttemplateByIDsRequest) (*netbox_goV1.DeleteDcimConsoleserverporttemplateByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimConsoleserverporttemplateByIDsReply{}, nil
}

// GetByCondition get a dcimConsoleserverporttemplate by custom condition
func (s *dcimConsoleserverporttemplate) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimConsoleserverporttemplateByConditionRequest) (*netbox_goV1.GetDcimConsoleserverporttemplateByConditionReply, error) {
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

	data, err := convertDcimConsoleserverporttemplate(record)
	if err != nil {
		logger.Warn("convertDcimConsoleserverporttemplate error", logger.Err(err), logger.Any("dcimConsoleserverporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimConsoleserverporttemplate.Err()
	}

	return &netbox_goV1.GetDcimConsoleserverporttemplateByConditionReply{
		DcimConsoleserverporttemplate: data,
	}, nil
}

// ListByIDs batch get dcimConsoleserverporttemplate by ids
func (s *dcimConsoleserverporttemplate) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimConsoleserverporttemplateByIDsRequest) (*netbox_goV1.ListDcimConsoleserverporttemplateByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimConsoleserverporttemplateMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimConsoleserverporttemplates := []*netbox_goV1.DcimConsoleserverporttemplate{}
	for _, id := range req.Ids {
		if v, ok := dcimConsoleserverporttemplateMap[id]; ok {
			record, err := convertDcimConsoleserverporttemplate(v)
			if err != nil {
				logger.Warn("convertDcimConsoleserverporttemplate error", logger.Err(err), logger.Any("dcimConsoleserverporttemplate", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimConsoleserverporttemplates = append(dcimConsoleserverporttemplates, record)
		}
	}

	return &netbox_goV1.ListDcimConsoleserverporttemplateByIDsReply{DcimConsoleserverporttemplates: dcimConsoleserverporttemplates}, nil
}

// ListByLastID get a paginated list of dcimConsoleserverporttemplates by last id
func (s *dcimConsoleserverporttemplate) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimConsoleserverporttemplateByLastIDRequest) (*netbox_goV1.ListDcimConsoleserverporttemplateByLastIDReply, error) {
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

	dcimConsoleserverporttemplates := []*netbox_goV1.DcimConsoleserverporttemplate{}
	for _, record := range records {
		data, err := convertDcimConsoleserverporttemplate(record)
		if err != nil {
			logger.Warn("convertDcimConsoleserverporttemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimConsoleserverporttemplates = append(dcimConsoleserverporttemplates, data)
	}

	return &netbox_goV1.ListDcimConsoleserverporttemplateByLastIDReply{
		DcimConsoleserverporttemplates: dcimConsoleserverporttemplates,
	}, nil
}

func convertDcimConsoleserverporttemplate(record *model.DcimConsoleserverporttemplate) (*netbox_goV1.DcimConsoleserverporttemplate, error) {
	value := &netbox_goV1.DcimConsoleserverporttemplate{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
