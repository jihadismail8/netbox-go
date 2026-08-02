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
		netbox_goV1.RegisterDcimDevicebaytemplateServer(server, NewDcimDevicebaytemplateServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimDevicebaytemplateServer = (*dcimDevicebaytemplate)(nil)
var _ time.Time

type dcimDevicebaytemplate struct {
	netbox_goV1.UnimplementedDcimDevicebaytemplateServer

	iDao dao.DcimDevicebaytemplateDao
}

// NewDcimDevicebaytemplateServer create a new service
func NewDcimDevicebaytemplateServer() netbox_goV1.DcimDevicebaytemplateServer {
	return &dcimDevicebaytemplate{
		iDao: dao.NewDcimDevicebaytemplateDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimDevicebaytemplateCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimDevicebaytemplate
func (s *dcimDevicebaytemplate) Create(ctx context.Context, req *netbox_goV1.CreateDcimDevicebaytemplateRequest) (*netbox_goV1.CreateDcimDevicebaytemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimDevicebaytemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimDevicebaytemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimDevicebaytemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimDevicebaytemplateReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimDevicebaytemplate by id
func (s *dcimDevicebaytemplate) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimDevicebaytemplateByIDRequest) (*netbox_goV1.DeleteDcimDevicebaytemplateByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimDevicebaytemplateByIDReply{}, nil
}

// UpdateByID update a dcimDevicebaytemplate by id
func (s *dcimDevicebaytemplate) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimDevicebaytemplateByIDRequest) (*netbox_goV1.UpdateDcimDevicebaytemplateByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimDevicebaytemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimDevicebaytemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimDevicebaytemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimDevicebaytemplateByIDReply{}, nil
}

// GetByID get a dcimDevicebaytemplate by id
func (s *dcimDevicebaytemplate) GetByID(ctx context.Context, req *netbox_goV1.GetDcimDevicebaytemplateByIDRequest) (*netbox_goV1.GetDcimDevicebaytemplateByIDReply, error) {
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

	data, err := convertDcimDevicebaytemplate(record)
	if err != nil {
		logger.Warn("convertDcimDevicebaytemplate error", logger.Err(err), logger.Any("dcimDevicebaytemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimDevicebaytemplate.Err()
	}

	return &netbox_goV1.GetDcimDevicebaytemplateByIDReply{DcimDevicebaytemplate: data}, nil
}

// List get a paginated list of dcimDevicebaytemplates by custom conditions
func (s *dcimDevicebaytemplate) List(ctx context.Context, req *netbox_goV1.ListDcimDevicebaytemplateRequest) (*netbox_goV1.ListDcimDevicebaytemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimDevicebaytemplate.Err()
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

	dcimDevicebaytemplates := []*netbox_goV1.DcimDevicebaytemplate{}
	for _, record := range records {
		data, err := convertDcimDevicebaytemplate(record)
		if err != nil {
			logger.Warn("convertDcimDevicebaytemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimDevicebaytemplates = append(dcimDevicebaytemplates, data)
	}

	return &netbox_goV1.ListDcimDevicebaytemplateReply{
		Total:                  total,
		DcimDevicebaytemplates: dcimDevicebaytemplates,
	}, nil
}

// DeleteByIDs batch delete dcimDevicebaytemplate by ids
func (s *dcimDevicebaytemplate) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimDevicebaytemplateByIDsRequest) (*netbox_goV1.DeleteDcimDevicebaytemplateByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimDevicebaytemplateByIDsReply{}, nil
}

// GetByCondition get a dcimDevicebaytemplate by custom condition
func (s *dcimDevicebaytemplate) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimDevicebaytemplateByConditionRequest) (*netbox_goV1.GetDcimDevicebaytemplateByConditionReply, error) {
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

	data, err := convertDcimDevicebaytemplate(record)
	if err != nil {
		logger.Warn("convertDcimDevicebaytemplate error", logger.Err(err), logger.Any("dcimDevicebaytemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimDevicebaytemplate.Err()
	}

	return &netbox_goV1.GetDcimDevicebaytemplateByConditionReply{
		DcimDevicebaytemplate: data,
	}, nil
}

// ListByIDs batch get dcimDevicebaytemplate by ids
func (s *dcimDevicebaytemplate) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimDevicebaytemplateByIDsRequest) (*netbox_goV1.ListDcimDevicebaytemplateByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimDevicebaytemplateMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimDevicebaytemplates := []*netbox_goV1.DcimDevicebaytemplate{}
	for _, id := range req.Ids {
		if v, ok := dcimDevicebaytemplateMap[id]; ok {
			record, err := convertDcimDevicebaytemplate(v)
			if err != nil {
				logger.Warn("convertDcimDevicebaytemplate error", logger.Err(err), logger.Any("dcimDevicebaytemplate", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimDevicebaytemplates = append(dcimDevicebaytemplates, record)
		}
	}

	return &netbox_goV1.ListDcimDevicebaytemplateByIDsReply{DcimDevicebaytemplates: dcimDevicebaytemplates}, nil
}

// ListByLastID get a paginated list of dcimDevicebaytemplates by last id
func (s *dcimDevicebaytemplate) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimDevicebaytemplateByLastIDRequest) (*netbox_goV1.ListDcimDevicebaytemplateByLastIDReply, error) {
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

	dcimDevicebaytemplates := []*netbox_goV1.DcimDevicebaytemplate{}
	for _, record := range records {
		data, err := convertDcimDevicebaytemplate(record)
		if err != nil {
			logger.Warn("convertDcimDevicebaytemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimDevicebaytemplates = append(dcimDevicebaytemplates, data)
	}

	return &netbox_goV1.ListDcimDevicebaytemplateByLastIDReply{
		DcimDevicebaytemplates: dcimDevicebaytemplates,
	}, nil
}

func convertDcimDevicebaytemplate(record *model.DcimDevicebaytemplate) (*netbox_goV1.DcimDevicebaytemplate, error) {
	value := &netbox_goV1.DcimDevicebaytemplate{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
