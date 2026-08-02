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
		netbox_goV1.RegisterDcimPoweroutlettemplateServer(server, NewDcimPoweroutlettemplateServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimPoweroutlettemplateServer = (*dcimPoweroutlettemplate)(nil)
var _ time.Time

type dcimPoweroutlettemplate struct {
	netbox_goV1.UnimplementedDcimPoweroutlettemplateServer

	iDao dao.DcimPoweroutlettemplateDao
}

// NewDcimPoweroutlettemplateServer create a new service
func NewDcimPoweroutlettemplateServer() netbox_goV1.DcimPoweroutlettemplateServer {
	return &dcimPoweroutlettemplate{
		iDao: dao.NewDcimPoweroutlettemplateDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimPoweroutlettemplateCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimPoweroutlettemplate
func (s *dcimPoweroutlettemplate) Create(ctx context.Context, req *netbox_goV1.CreateDcimPoweroutlettemplateRequest) (*netbox_goV1.CreateDcimPoweroutlettemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimPoweroutlettemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimPoweroutlettemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimPoweroutlettemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimPoweroutlettemplateReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimPoweroutlettemplate by id
func (s *dcimPoweroutlettemplate) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimPoweroutlettemplateByIDRequest) (*netbox_goV1.DeleteDcimPoweroutlettemplateByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimPoweroutlettemplateByIDReply{}, nil
}

// UpdateByID update a dcimPoweroutlettemplate by id
func (s *dcimPoweroutlettemplate) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimPoweroutlettemplateByIDRequest) (*netbox_goV1.UpdateDcimPoweroutlettemplateByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimPoweroutlettemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimPoweroutlettemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimPoweroutlettemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimPoweroutlettemplateByIDReply{}, nil
}

// GetByID get a dcimPoweroutlettemplate by id
func (s *dcimPoweroutlettemplate) GetByID(ctx context.Context, req *netbox_goV1.GetDcimPoweroutlettemplateByIDRequest) (*netbox_goV1.GetDcimPoweroutlettemplateByIDReply, error) {
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

	data, err := convertDcimPoweroutlettemplate(record)
	if err != nil {
		logger.Warn("convertDcimPoweroutlettemplate error", logger.Err(err), logger.Any("dcimPoweroutlettemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimPoweroutlettemplate.Err()
	}

	return &netbox_goV1.GetDcimPoweroutlettemplateByIDReply{DcimPoweroutlettemplate: data}, nil
}

// List get a paginated list of dcimPoweroutlettemplates by custom conditions
func (s *dcimPoweroutlettemplate) List(ctx context.Context, req *netbox_goV1.ListDcimPoweroutlettemplateRequest) (*netbox_goV1.ListDcimPoweroutlettemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimPoweroutlettemplate.Err()
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

	dcimPoweroutlettemplates := []*netbox_goV1.DcimPoweroutlettemplate{}
	for _, record := range records {
		data, err := convertDcimPoweroutlettemplate(record)
		if err != nil {
			logger.Warn("convertDcimPoweroutlettemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimPoweroutlettemplates = append(dcimPoweroutlettemplates, data)
	}

	return &netbox_goV1.ListDcimPoweroutlettemplateReply{
		Total:                    total,
		DcimPoweroutlettemplates: dcimPoweroutlettemplates,
	}, nil
}

// DeleteByIDs batch delete dcimPoweroutlettemplate by ids
func (s *dcimPoweroutlettemplate) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimPoweroutlettemplateByIDsRequest) (*netbox_goV1.DeleteDcimPoweroutlettemplateByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimPoweroutlettemplateByIDsReply{}, nil
}

// GetByCondition get a dcimPoweroutlettemplate by custom condition
func (s *dcimPoweroutlettemplate) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimPoweroutlettemplateByConditionRequest) (*netbox_goV1.GetDcimPoweroutlettemplateByConditionReply, error) {
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

	data, err := convertDcimPoweroutlettemplate(record)
	if err != nil {
		logger.Warn("convertDcimPoweroutlettemplate error", logger.Err(err), logger.Any("dcimPoweroutlettemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimPoweroutlettemplate.Err()
	}

	return &netbox_goV1.GetDcimPoweroutlettemplateByConditionReply{
		DcimPoweroutlettemplate: data,
	}, nil
}

// ListByIDs batch get dcimPoweroutlettemplate by ids
func (s *dcimPoweroutlettemplate) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimPoweroutlettemplateByIDsRequest) (*netbox_goV1.ListDcimPoweroutlettemplateByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimPoweroutlettemplateMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimPoweroutlettemplates := []*netbox_goV1.DcimPoweroutlettemplate{}
	for _, id := range req.Ids {
		if v, ok := dcimPoweroutlettemplateMap[id]; ok {
			record, err := convertDcimPoweroutlettemplate(v)
			if err != nil {
				logger.Warn("convertDcimPoweroutlettemplate error", logger.Err(err), logger.Any("dcimPoweroutlettemplate", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimPoweroutlettemplates = append(dcimPoweroutlettemplates, record)
		}
	}

	return &netbox_goV1.ListDcimPoweroutlettemplateByIDsReply{DcimPoweroutlettemplates: dcimPoweroutlettemplates}, nil
}

// ListByLastID get a paginated list of dcimPoweroutlettemplates by last id
func (s *dcimPoweroutlettemplate) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimPoweroutlettemplateByLastIDRequest) (*netbox_goV1.ListDcimPoweroutlettemplateByLastIDReply, error) {
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

	dcimPoweroutlettemplates := []*netbox_goV1.DcimPoweroutlettemplate{}
	for _, record := range records {
		data, err := convertDcimPoweroutlettemplate(record)
		if err != nil {
			logger.Warn("convertDcimPoweroutlettemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimPoweroutlettemplates = append(dcimPoweroutlettemplates, data)
	}

	return &netbox_goV1.ListDcimPoweroutlettemplateByLastIDReply{
		DcimPoweroutlettemplates: dcimPoweroutlettemplates,
	}, nil
}

func convertDcimPoweroutlettemplate(record *model.DcimPoweroutlettemplate) (*netbox_goV1.DcimPoweroutlettemplate, error) {
	value := &netbox_goV1.DcimPoweroutlettemplate{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
