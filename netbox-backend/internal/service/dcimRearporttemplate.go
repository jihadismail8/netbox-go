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
		netbox_goV1.RegisterDcimRearporttemplateServer(server, NewDcimRearporttemplateServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimRearporttemplateServer = (*dcimRearporttemplate)(nil)
var _ time.Time

type dcimRearporttemplate struct {
	netbox_goV1.UnimplementedDcimRearporttemplateServer

	iDao dao.DcimRearporttemplateDao
}

// NewDcimRearporttemplateServer create a new service
func NewDcimRearporttemplateServer() netbox_goV1.DcimRearporttemplateServer {
	return &dcimRearporttemplate{
		iDao: dao.NewDcimRearporttemplateDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimRearporttemplateCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimRearporttemplate
func (s *dcimRearporttemplate) Create(ctx context.Context, req *netbox_goV1.CreateDcimRearporttemplateRequest) (*netbox_goV1.CreateDcimRearporttemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimRearporttemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimRearporttemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimRearporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimRearporttemplateReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimRearporttemplate by id
func (s *dcimRearporttemplate) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimRearporttemplateByIDRequest) (*netbox_goV1.DeleteDcimRearporttemplateByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimRearporttemplateByIDReply{}, nil
}

// UpdateByID update a dcimRearporttemplate by id
func (s *dcimRearporttemplate) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimRearporttemplateByIDRequest) (*netbox_goV1.UpdateDcimRearporttemplateByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimRearporttemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimRearporttemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimRearporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimRearporttemplateByIDReply{}, nil
}

// GetByID get a dcimRearporttemplate by id
func (s *dcimRearporttemplate) GetByID(ctx context.Context, req *netbox_goV1.GetDcimRearporttemplateByIDRequest) (*netbox_goV1.GetDcimRearporttemplateByIDReply, error) {
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

	data, err := convertDcimRearporttemplate(record)
	if err != nil {
		logger.Warn("convertDcimRearporttemplate error", logger.Err(err), logger.Any("dcimRearporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimRearporttemplate.Err()
	}

	return &netbox_goV1.GetDcimRearporttemplateByIDReply{DcimRearporttemplate: data}, nil
}

// List get a paginated list of dcimRearporttemplates by custom conditions
func (s *dcimRearporttemplate) List(ctx context.Context, req *netbox_goV1.ListDcimRearporttemplateRequest) (*netbox_goV1.ListDcimRearporttemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimRearporttemplate.Err()
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

	dcimRearporttemplates := []*netbox_goV1.DcimRearporttemplate{}
	for _, record := range records {
		data, err := convertDcimRearporttemplate(record)
		if err != nil {
			logger.Warn("convertDcimRearporttemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimRearporttemplates = append(dcimRearporttemplates, data)
	}

	return &netbox_goV1.ListDcimRearporttemplateReply{
		Total:                 total,
		DcimRearporttemplates: dcimRearporttemplates,
	}, nil
}

// DeleteByIDs batch delete dcimRearporttemplate by ids
func (s *dcimRearporttemplate) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimRearporttemplateByIDsRequest) (*netbox_goV1.DeleteDcimRearporttemplateByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimRearporttemplateByIDsReply{}, nil
}

// GetByCondition get a dcimRearporttemplate by custom condition
func (s *dcimRearporttemplate) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimRearporttemplateByConditionRequest) (*netbox_goV1.GetDcimRearporttemplateByConditionReply, error) {
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

	data, err := convertDcimRearporttemplate(record)
	if err != nil {
		logger.Warn("convertDcimRearporttemplate error", logger.Err(err), logger.Any("dcimRearporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimRearporttemplate.Err()
	}

	return &netbox_goV1.GetDcimRearporttemplateByConditionReply{
		DcimRearporttemplate: data,
	}, nil
}

// ListByIDs batch get dcimRearporttemplate by ids
func (s *dcimRearporttemplate) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimRearporttemplateByIDsRequest) (*netbox_goV1.ListDcimRearporttemplateByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimRearporttemplateMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimRearporttemplates := []*netbox_goV1.DcimRearporttemplate{}
	for _, id := range req.Ids {
		if v, ok := dcimRearporttemplateMap[id]; ok {
			record, err := convertDcimRearporttemplate(v)
			if err != nil {
				logger.Warn("convertDcimRearporttemplate error", logger.Err(err), logger.Any("dcimRearporttemplate", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimRearporttemplates = append(dcimRearporttemplates, record)
		}
	}

	return &netbox_goV1.ListDcimRearporttemplateByIDsReply{DcimRearporttemplates: dcimRearporttemplates}, nil
}

// ListByLastID get a paginated list of dcimRearporttemplates by last id
func (s *dcimRearporttemplate) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimRearporttemplateByLastIDRequest) (*netbox_goV1.ListDcimRearporttemplateByLastIDReply, error) {
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

	dcimRearporttemplates := []*netbox_goV1.DcimRearporttemplate{}
	for _, record := range records {
		data, err := convertDcimRearporttemplate(record)
		if err != nil {
			logger.Warn("convertDcimRearporttemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimRearporttemplates = append(dcimRearporttemplates, data)
	}

	return &netbox_goV1.ListDcimRearporttemplateByLastIDReply{
		DcimRearporttemplates: dcimRearporttemplates,
	}, nil
}

func convertDcimRearporttemplate(record *model.DcimRearporttemplate) (*netbox_goV1.DcimRearporttemplate, error) {
	value := &netbox_goV1.DcimRearporttemplate{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
