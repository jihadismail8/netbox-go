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
		netbox_goV1.RegisterDcimPowerporttemplateServer(server, NewDcimPowerporttemplateServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimPowerporttemplateServer = (*dcimPowerporttemplate)(nil)
var _ time.Time

type dcimPowerporttemplate struct {
	netbox_goV1.UnimplementedDcimPowerporttemplateServer

	iDao dao.DcimPowerporttemplateDao
}

// NewDcimPowerporttemplateServer create a new service
func NewDcimPowerporttemplateServer() netbox_goV1.DcimPowerporttemplateServer {
	return &dcimPowerporttemplate{
		iDao: dao.NewDcimPowerporttemplateDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimPowerporttemplateCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimPowerporttemplate
func (s *dcimPowerporttemplate) Create(ctx context.Context, req *netbox_goV1.CreateDcimPowerporttemplateRequest) (*netbox_goV1.CreateDcimPowerporttemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimPowerporttemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimPowerporttemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimPowerporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimPowerporttemplateReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimPowerporttemplate by id
func (s *dcimPowerporttemplate) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimPowerporttemplateByIDRequest) (*netbox_goV1.DeleteDcimPowerporttemplateByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimPowerporttemplateByIDReply{}, nil
}

// UpdateByID update a dcimPowerporttemplate by id
func (s *dcimPowerporttemplate) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimPowerporttemplateByIDRequest) (*netbox_goV1.UpdateDcimPowerporttemplateByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimPowerporttemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimPowerporttemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimPowerporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimPowerporttemplateByIDReply{}, nil
}

// GetByID get a dcimPowerporttemplate by id
func (s *dcimPowerporttemplate) GetByID(ctx context.Context, req *netbox_goV1.GetDcimPowerporttemplateByIDRequest) (*netbox_goV1.GetDcimPowerporttemplateByIDReply, error) {
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

	data, err := convertDcimPowerporttemplate(record)
	if err != nil {
		logger.Warn("convertDcimPowerporttemplate error", logger.Err(err), logger.Any("dcimPowerporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimPowerporttemplate.Err()
	}

	return &netbox_goV1.GetDcimPowerporttemplateByIDReply{DcimPowerporttemplate: data}, nil
}

// List get a paginated list of dcimPowerporttemplates by custom conditions
func (s *dcimPowerporttemplate) List(ctx context.Context, req *netbox_goV1.ListDcimPowerporttemplateRequest) (*netbox_goV1.ListDcimPowerporttemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimPowerporttemplate.Err()
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

	dcimPowerporttemplates := []*netbox_goV1.DcimPowerporttemplate{}
	for _, record := range records {
		data, err := convertDcimPowerporttemplate(record)
		if err != nil {
			logger.Warn("convertDcimPowerporttemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimPowerporttemplates = append(dcimPowerporttemplates, data)
	}

	return &netbox_goV1.ListDcimPowerporttemplateReply{
		Total:                  total,
		DcimPowerporttemplates: dcimPowerporttemplates,
	}, nil
}

// DeleteByIDs batch delete dcimPowerporttemplate by ids
func (s *dcimPowerporttemplate) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimPowerporttemplateByIDsRequest) (*netbox_goV1.DeleteDcimPowerporttemplateByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimPowerporttemplateByIDsReply{}, nil
}

// GetByCondition get a dcimPowerporttemplate by custom condition
func (s *dcimPowerporttemplate) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimPowerporttemplateByConditionRequest) (*netbox_goV1.GetDcimPowerporttemplateByConditionReply, error) {
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

	data, err := convertDcimPowerporttemplate(record)
	if err != nil {
		logger.Warn("convertDcimPowerporttemplate error", logger.Err(err), logger.Any("dcimPowerporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimPowerporttemplate.Err()
	}

	return &netbox_goV1.GetDcimPowerporttemplateByConditionReply{
		DcimPowerporttemplate: data,
	}, nil
}

// ListByIDs batch get dcimPowerporttemplate by ids
func (s *dcimPowerporttemplate) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimPowerporttemplateByIDsRequest) (*netbox_goV1.ListDcimPowerporttemplateByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimPowerporttemplateMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimPowerporttemplates := []*netbox_goV1.DcimPowerporttemplate{}
	for _, id := range req.Ids {
		if v, ok := dcimPowerporttemplateMap[id]; ok {
			record, err := convertDcimPowerporttemplate(v)
			if err != nil {
				logger.Warn("convertDcimPowerporttemplate error", logger.Err(err), logger.Any("dcimPowerporttemplate", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimPowerporttemplates = append(dcimPowerporttemplates, record)
		}
	}

	return &netbox_goV1.ListDcimPowerporttemplateByIDsReply{DcimPowerporttemplates: dcimPowerporttemplates}, nil
}

// ListByLastID get a paginated list of dcimPowerporttemplates by last id
func (s *dcimPowerporttemplate) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimPowerporttemplateByLastIDRequest) (*netbox_goV1.ListDcimPowerporttemplateByLastIDReply, error) {
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

	dcimPowerporttemplates := []*netbox_goV1.DcimPowerporttemplate{}
	for _, record := range records {
		data, err := convertDcimPowerporttemplate(record)
		if err != nil {
			logger.Warn("convertDcimPowerporttemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimPowerporttemplates = append(dcimPowerporttemplates, data)
	}

	return &netbox_goV1.ListDcimPowerporttemplateByLastIDReply{
		DcimPowerporttemplates: dcimPowerporttemplates,
	}, nil
}

func convertDcimPowerporttemplate(record *model.DcimPowerporttemplate) (*netbox_goV1.DcimPowerporttemplate, error) {
	value := &netbox_goV1.DcimPowerporttemplate{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
