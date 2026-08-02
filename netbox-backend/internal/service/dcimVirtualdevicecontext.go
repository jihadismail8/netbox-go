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
		netbox_goV1.RegisterDcimVirtualdevicecontextServer(server, NewDcimVirtualdevicecontextServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimVirtualdevicecontextServer = (*dcimVirtualdevicecontext)(nil)
var _ time.Time

type dcimVirtualdevicecontext struct {
	netbox_goV1.UnimplementedDcimVirtualdevicecontextServer

	iDao dao.DcimVirtualdevicecontextDao
}

// NewDcimVirtualdevicecontextServer create a new service
func NewDcimVirtualdevicecontextServer() netbox_goV1.DcimVirtualdevicecontextServer {
	return &dcimVirtualdevicecontext{
		iDao: dao.NewDcimVirtualdevicecontextDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimVirtualdevicecontextCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimVirtualdevicecontext
func (s *dcimVirtualdevicecontext) Create(ctx context.Context, req *netbox_goV1.CreateDcimVirtualdevicecontextRequest) (*netbox_goV1.CreateDcimVirtualdevicecontextReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimVirtualdevicecontext{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimVirtualdevicecontext.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimVirtualdevicecontext", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimVirtualdevicecontextReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimVirtualdevicecontext by id
func (s *dcimVirtualdevicecontext) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimVirtualdevicecontextByIDRequest) (*netbox_goV1.DeleteDcimVirtualdevicecontextByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimVirtualdevicecontextByIDReply{}, nil
}

// UpdateByID update a dcimVirtualdevicecontext by id
func (s *dcimVirtualdevicecontext) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimVirtualdevicecontextByIDRequest) (*netbox_goV1.UpdateDcimVirtualdevicecontextByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimVirtualdevicecontext{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimVirtualdevicecontext.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimVirtualdevicecontext", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimVirtualdevicecontextByIDReply{}, nil
}

// GetByID get a dcimVirtualdevicecontext by id
func (s *dcimVirtualdevicecontext) GetByID(ctx context.Context, req *netbox_goV1.GetDcimVirtualdevicecontextByIDRequest) (*netbox_goV1.GetDcimVirtualdevicecontextByIDReply, error) {
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

	data, err := convertDcimVirtualdevicecontext(record)
	if err != nil {
		logger.Warn("convertDcimVirtualdevicecontext error", logger.Err(err), logger.Any("dcimVirtualdevicecontext", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimVirtualdevicecontext.Err()
	}

	return &netbox_goV1.GetDcimVirtualdevicecontextByIDReply{DcimVirtualdevicecontext: data}, nil
}

// List get a paginated list of dcimVirtualdevicecontexts by custom conditions
func (s *dcimVirtualdevicecontext) List(ctx context.Context, req *netbox_goV1.ListDcimVirtualdevicecontextRequest) (*netbox_goV1.ListDcimVirtualdevicecontextReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimVirtualdevicecontext.Err()
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

	dcimVirtualdevicecontexts := []*netbox_goV1.DcimVirtualdevicecontext{}
	for _, record := range records {
		data, err := convertDcimVirtualdevicecontext(record)
		if err != nil {
			logger.Warn("convertDcimVirtualdevicecontext error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimVirtualdevicecontexts = append(dcimVirtualdevicecontexts, data)
	}

	return &netbox_goV1.ListDcimVirtualdevicecontextReply{
		Total:                     total,
		DcimVirtualdevicecontexts: dcimVirtualdevicecontexts,
	}, nil
}

// DeleteByIDs batch delete dcimVirtualdevicecontext by ids
func (s *dcimVirtualdevicecontext) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimVirtualdevicecontextByIDsRequest) (*netbox_goV1.DeleteDcimVirtualdevicecontextByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimVirtualdevicecontextByIDsReply{}, nil
}

// GetByCondition get a dcimVirtualdevicecontext by custom condition
func (s *dcimVirtualdevicecontext) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimVirtualdevicecontextByConditionRequest) (*netbox_goV1.GetDcimVirtualdevicecontextByConditionReply, error) {
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

	data, err := convertDcimVirtualdevicecontext(record)
	if err != nil {
		logger.Warn("convertDcimVirtualdevicecontext error", logger.Err(err), logger.Any("dcimVirtualdevicecontext", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimVirtualdevicecontext.Err()
	}

	return &netbox_goV1.GetDcimVirtualdevicecontextByConditionReply{
		DcimVirtualdevicecontext: data,
	}, nil
}

// ListByIDs batch get dcimVirtualdevicecontext by ids
func (s *dcimVirtualdevicecontext) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimVirtualdevicecontextByIDsRequest) (*netbox_goV1.ListDcimVirtualdevicecontextByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimVirtualdevicecontextMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimVirtualdevicecontexts := []*netbox_goV1.DcimVirtualdevicecontext{}
	for _, id := range req.Ids {
		if v, ok := dcimVirtualdevicecontextMap[id]; ok {
			record, err := convertDcimVirtualdevicecontext(v)
			if err != nil {
				logger.Warn("convertDcimVirtualdevicecontext error", logger.Err(err), logger.Any("dcimVirtualdevicecontext", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimVirtualdevicecontexts = append(dcimVirtualdevicecontexts, record)
		}
	}

	return &netbox_goV1.ListDcimVirtualdevicecontextByIDsReply{DcimVirtualdevicecontexts: dcimVirtualdevicecontexts}, nil
}

// ListByLastID get a paginated list of dcimVirtualdevicecontexts by last id
func (s *dcimVirtualdevicecontext) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimVirtualdevicecontextByLastIDRequest) (*netbox_goV1.ListDcimVirtualdevicecontextByLastIDReply, error) {
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

	dcimVirtualdevicecontexts := []*netbox_goV1.DcimVirtualdevicecontext{}
	for _, record := range records {
		data, err := convertDcimVirtualdevicecontext(record)
		if err != nil {
			logger.Warn("convertDcimVirtualdevicecontext error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimVirtualdevicecontexts = append(dcimVirtualdevicecontexts, data)
	}

	return &netbox_goV1.ListDcimVirtualdevicecontextByLastIDReply{
		DcimVirtualdevicecontexts: dcimVirtualdevicecontexts,
	}, nil
}

func convertDcimVirtualdevicecontext(record *model.DcimVirtualdevicecontext) (*netbox_goV1.DcimVirtualdevicecontext, error) {
	value := &netbox_goV1.DcimVirtualdevicecontext{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
