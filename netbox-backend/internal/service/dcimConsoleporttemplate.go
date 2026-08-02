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
		netbox_goV1.RegisterDcimConsoleporttemplateServer(server, NewDcimConsoleporttemplateServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimConsoleporttemplateServer = (*dcimConsoleporttemplate)(nil)
var _ time.Time

type dcimConsoleporttemplate struct {
	netbox_goV1.UnimplementedDcimConsoleporttemplateServer

	iDao dao.DcimConsoleporttemplateDao
}

// NewDcimConsoleporttemplateServer create a new service
func NewDcimConsoleporttemplateServer() netbox_goV1.DcimConsoleporttemplateServer {
	return &dcimConsoleporttemplate{
		iDao: dao.NewDcimConsoleporttemplateDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimConsoleporttemplateCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimConsoleporttemplate
func (s *dcimConsoleporttemplate) Create(ctx context.Context, req *netbox_goV1.CreateDcimConsoleporttemplateRequest) (*netbox_goV1.CreateDcimConsoleporttemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimConsoleporttemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimConsoleporttemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimConsoleporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimConsoleporttemplateReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimConsoleporttemplate by id
func (s *dcimConsoleporttemplate) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimConsoleporttemplateByIDRequest) (*netbox_goV1.DeleteDcimConsoleporttemplateByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimConsoleporttemplateByIDReply{}, nil
}

// UpdateByID update a dcimConsoleporttemplate by id
func (s *dcimConsoleporttemplate) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimConsoleporttemplateByIDRequest) (*netbox_goV1.UpdateDcimConsoleporttemplateByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimConsoleporttemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimConsoleporttemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimConsoleporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimConsoleporttemplateByIDReply{}, nil
}

// GetByID get a dcimConsoleporttemplate by id
func (s *dcimConsoleporttemplate) GetByID(ctx context.Context, req *netbox_goV1.GetDcimConsoleporttemplateByIDRequest) (*netbox_goV1.GetDcimConsoleporttemplateByIDReply, error) {
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

	data, err := convertDcimConsoleporttemplate(record)
	if err != nil {
		logger.Warn("convertDcimConsoleporttemplate error", logger.Err(err), logger.Any("dcimConsoleporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimConsoleporttemplate.Err()
	}

	return &netbox_goV1.GetDcimConsoleporttemplateByIDReply{DcimConsoleporttemplate: data}, nil
}

// List get a paginated list of dcimConsoleporttemplates by custom conditions
func (s *dcimConsoleporttemplate) List(ctx context.Context, req *netbox_goV1.ListDcimConsoleporttemplateRequest) (*netbox_goV1.ListDcimConsoleporttemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimConsoleporttemplate.Err()
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

	dcimConsoleporttemplates := []*netbox_goV1.DcimConsoleporttemplate{}
	for _, record := range records {
		data, err := convertDcimConsoleporttemplate(record)
		if err != nil {
			logger.Warn("convertDcimConsoleporttemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimConsoleporttemplates = append(dcimConsoleporttemplates, data)
	}

	return &netbox_goV1.ListDcimConsoleporttemplateReply{
		Total:                    total,
		DcimConsoleporttemplates: dcimConsoleporttemplates,
	}, nil
}

// DeleteByIDs batch delete dcimConsoleporttemplate by ids
func (s *dcimConsoleporttemplate) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimConsoleporttemplateByIDsRequest) (*netbox_goV1.DeleteDcimConsoleporttemplateByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimConsoleporttemplateByIDsReply{}, nil
}

// GetByCondition get a dcimConsoleporttemplate by custom condition
func (s *dcimConsoleporttemplate) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimConsoleporttemplateByConditionRequest) (*netbox_goV1.GetDcimConsoleporttemplateByConditionReply, error) {
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

	data, err := convertDcimConsoleporttemplate(record)
	if err != nil {
		logger.Warn("convertDcimConsoleporttemplate error", logger.Err(err), logger.Any("dcimConsoleporttemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimConsoleporttemplate.Err()
	}

	return &netbox_goV1.GetDcimConsoleporttemplateByConditionReply{
		DcimConsoleporttemplate: data,
	}, nil
}

// ListByIDs batch get dcimConsoleporttemplate by ids
func (s *dcimConsoleporttemplate) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimConsoleporttemplateByIDsRequest) (*netbox_goV1.ListDcimConsoleporttemplateByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimConsoleporttemplateMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimConsoleporttemplates := []*netbox_goV1.DcimConsoleporttemplate{}
	for _, id := range req.Ids {
		if v, ok := dcimConsoleporttemplateMap[id]; ok {
			record, err := convertDcimConsoleporttemplate(v)
			if err != nil {
				logger.Warn("convertDcimConsoleporttemplate error", logger.Err(err), logger.Any("dcimConsoleporttemplate", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimConsoleporttemplates = append(dcimConsoleporttemplates, record)
		}
	}

	return &netbox_goV1.ListDcimConsoleporttemplateByIDsReply{DcimConsoleporttemplates: dcimConsoleporttemplates}, nil
}

// ListByLastID get a paginated list of dcimConsoleporttemplates by last id
func (s *dcimConsoleporttemplate) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimConsoleporttemplateByLastIDRequest) (*netbox_goV1.ListDcimConsoleporttemplateByLastIDReply, error) {
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

	dcimConsoleporttemplates := []*netbox_goV1.DcimConsoleporttemplate{}
	for _, record := range records {
		data, err := convertDcimConsoleporttemplate(record)
		if err != nil {
			logger.Warn("convertDcimConsoleporttemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimConsoleporttemplates = append(dcimConsoleporttemplates, data)
	}

	return &netbox_goV1.ListDcimConsoleporttemplateByLastIDReply{
		DcimConsoleporttemplates: dcimConsoleporttemplates,
	}, nil
}

func convertDcimConsoleporttemplate(record *model.DcimConsoleporttemplate) (*netbox_goV1.DcimConsoleporttemplate, error) {
	value := &netbox_goV1.DcimConsoleporttemplate{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
