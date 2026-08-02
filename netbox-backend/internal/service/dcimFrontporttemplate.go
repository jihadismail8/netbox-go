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
		netbox_goV1.RegisterDcimFrontporttemplateServer(server, NewDcimFrontporttemplateServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimFrontporttemplateServer = (*dcimFrontporttemplate)(nil)
var _ time.Time

type dcimFrontporttemplate struct {
	netbox_goV1.UnimplementedDcimFrontporttemplateServer

	iDao dao.DcimFrontporttemplateDao
}

// NewDcimFrontporttemplateServer create a new service
func NewDcimFrontporttemplateServer() netbox_goV1.DcimFrontporttemplateServer {
	return &dcimFrontporttemplate{
		iDao: dao.NewDcimFrontporttemplateDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimFrontporttemplateCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimFrontporttemplate
func (s *dcimFrontporttemplate) Create(ctx context.Context, req *netbox_goV1.CreateDcimFrontporttemplateRequest) (*netbox_goV1.CreateDcimFrontporttemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimFrontporttemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimFrontporttemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimFrontporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimFrontporttemplateReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimFrontporttemplate by id
func (s *dcimFrontporttemplate) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimFrontporttemplateByIDRequest) (*netbox_goV1.DeleteDcimFrontporttemplateByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimFrontporttemplateByIDReply{}, nil
}

// UpdateByID update a dcimFrontporttemplate by id
func (s *dcimFrontporttemplate) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimFrontporttemplateByIDRequest) (*netbox_goV1.UpdateDcimFrontporttemplateByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimFrontporttemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimFrontporttemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimFrontporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimFrontporttemplateByIDReply{}, nil
}

// GetByID get a dcimFrontporttemplate by id
func (s *dcimFrontporttemplate) GetByID(ctx context.Context, req *netbox_goV1.GetDcimFrontporttemplateByIDRequest) (*netbox_goV1.GetDcimFrontporttemplateByIDReply, error) {
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

	data, err := convertDcimFrontporttemplate(record)
	if err != nil {
		logger.Warn("convertDcimFrontporttemplate error", logger.Err(err), logger.Any("dcimFrontporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimFrontporttemplate.Err()
	}

	return &netbox_goV1.GetDcimFrontporttemplateByIDReply{DcimFrontporttemplate: data}, nil
}

// List get a paginated list of dcimFrontporttemplates by custom conditions
func (s *dcimFrontporttemplate) List(ctx context.Context, req *netbox_goV1.ListDcimFrontporttemplateRequest) (*netbox_goV1.ListDcimFrontporttemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimFrontporttemplate.Err()
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

	dcimFrontporttemplates := []*netbox_goV1.DcimFrontporttemplate{}
	for _, record := range records {
		data, err := convertDcimFrontporttemplate(record)
		if err != nil {
			logger.Warn("convertDcimFrontporttemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimFrontporttemplates = append(dcimFrontporttemplates, data)
	}

	return &netbox_goV1.ListDcimFrontporttemplateReply{
		Total:                  total,
		DcimFrontporttemplates: dcimFrontporttemplates,
	}, nil
}

// DeleteByIDs batch delete dcimFrontporttemplate by ids
func (s *dcimFrontporttemplate) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimFrontporttemplateByIDsRequest) (*netbox_goV1.DeleteDcimFrontporttemplateByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimFrontporttemplateByIDsReply{}, nil
}

// GetByCondition get a dcimFrontporttemplate by custom condition
func (s *dcimFrontporttemplate) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimFrontporttemplateByConditionRequest) (*netbox_goV1.GetDcimFrontporttemplateByConditionReply, error) {
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

	data, err := convertDcimFrontporttemplate(record)
	if err != nil {
		logger.Warn("convertDcimFrontporttemplate error", logger.Err(err), logger.Any("dcimFrontporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimFrontporttemplate.Err()
	}

	return &netbox_goV1.GetDcimFrontporttemplateByConditionReply{
		DcimFrontporttemplate: data,
	}, nil
}

// ListByIDs batch get dcimFrontporttemplate by ids
func (s *dcimFrontporttemplate) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimFrontporttemplateByIDsRequest) (*netbox_goV1.ListDcimFrontporttemplateByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimFrontporttemplateMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimFrontporttemplates := []*netbox_goV1.DcimFrontporttemplate{}
	for _, id := range req.Ids {
		if v, ok := dcimFrontporttemplateMap[id]; ok {
			record, err := convertDcimFrontporttemplate(v)
			if err != nil {
				logger.Warn("convertDcimFrontporttemplate error", logger.Err(err), logger.Any("dcimFrontporttemplate", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimFrontporttemplates = append(dcimFrontporttemplates, record)
		}
	}

	return &netbox_goV1.ListDcimFrontporttemplateByIDsReply{DcimFrontporttemplates: dcimFrontporttemplates}, nil
}

// ListByLastID get a paginated list of dcimFrontporttemplates by last id
func (s *dcimFrontporttemplate) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimFrontporttemplateByLastIDRequest) (*netbox_goV1.ListDcimFrontporttemplateByLastIDReply, error) {
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

	dcimFrontporttemplates := []*netbox_goV1.DcimFrontporttemplate{}
	for _, record := range records {
		data, err := convertDcimFrontporttemplate(record)
		if err != nil {
			logger.Warn("convertDcimFrontporttemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimFrontporttemplates = append(dcimFrontporttemplates, data)
	}

	return &netbox_goV1.ListDcimFrontporttemplateByLastIDReply{
		DcimFrontporttemplates: dcimFrontporttemplates,
	}, nil
}

func convertDcimFrontporttemplate(record *model.DcimFrontporttemplate) (*netbox_goV1.DcimFrontporttemplate, error) {
	value := &netbox_goV1.DcimFrontporttemplate{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
