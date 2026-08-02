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
		netbox_goV1.RegisterCoreObjecttypeServer(server, NewCoreObjecttypeServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CoreObjecttypeServer = (*coreObjecttype)(nil)
var _ time.Time

type coreObjecttype struct {
	netbox_goV1.UnimplementedCoreObjecttypeServer

	iDao dao.CoreObjecttypeDao
}

// NewCoreObjecttypeServer create a new service
func NewCoreObjecttypeServer() netbox_goV1.CoreObjecttypeServer {
	return &coreObjecttype{
		iDao: dao.NewCoreObjecttypeDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCoreObjecttypeCache(database.GetCacheType()),
		),
	}
}

// Create a new coreObjecttype
func (s *coreObjecttype) Create(ctx context.Context, req *netbox_goV1.CreateCoreObjecttypeRequest) (*netbox_goV1.CreateCoreObjecttypeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CoreObjecttype{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCoreObjecttype.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("coreObjecttype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCoreObjecttypeReply{ContenttypePtrID: int32(record.ContenttypePtrID)}, nil
}

// DeleteByContenttypePtrID delete a coreObjecttype by contenttypePtrID
func (s *coreObjecttype) DeleteByContenttypePtrID(ctx context.Context, req *netbox_goV1.DeleteCoreObjecttypeByContenttypePtrIDRequest) (*netbox_goV1.DeleteCoreObjecttypeByContenttypePtrIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	err = s.iDao.DeleteByContenttypePtrID(ctx, int(req.ContenttypePtrID))
	if err != nil {
		logger.Error("DeleteByContenttypePtrID error", logger.Err(err), logger.Any("contenttypePtrID", req.ContenttypePtrID), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.DeleteCoreObjecttypeByContenttypePtrIDReply{}, nil
}

// UpdateByContenttypePtrID update a coreObjecttype by contenttypePtrID
func (s *coreObjecttype) UpdateByContenttypePtrID(ctx context.Context, req *netbox_goV1.UpdateCoreObjecttypeByContenttypePtrIDRequest) (*netbox_goV1.UpdateCoreObjecttypeByContenttypePtrIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CoreObjecttype{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByContenttypePtrIDCoreObjecttype.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ContenttypePtrID = int(req.ContenttypePtrID)

	err = s.iDao.UpdateByContenttypePtrID(ctx, record)
	if err != nil {
		logger.Error("UpdateByContenttypePtrID error", logger.Err(err), logger.Any("coreObjecttype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCoreObjecttypeByContenttypePtrIDReply{}, nil
}

// GetByContenttypePtrID get a coreObjecttype by contenttypePtrID
func (s *coreObjecttype) GetByContenttypePtrID(ctx context.Context, req *netbox_goV1.GetCoreObjecttypeByContenttypePtrIDRequest) (*netbox_goV1.GetCoreObjecttypeByContenttypePtrIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record, err := s.iDao.GetByContenttypePtrID(ctx, int(req.ContenttypePtrID))
	if err != nil {
		if errors.Is(err, database.ErrRecordNotFound) {
			logger.Warn("GetByContenttypePtrID error", logger.Err(err), logger.Any("contenttypePtrID", req.ContenttypePtrID), interceptor.ServerCtxRequestIDField(ctx))
			return nil, ecode.StatusNotFound.Err()
		}
		logger.Error("GetByContenttypePtrID error", logger.Err(err), logger.Any("contenttypePtrID", req.ContenttypePtrID), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	data, err := convertCoreObjecttype(record)
	if err != nil {
		logger.Warn("convertCoreObjecttype error", logger.Err(err), logger.Any("coreObjecttype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByContenttypePtrIDCoreObjecttype.Err()
	}

	return &netbox_goV1.GetCoreObjecttypeByContenttypePtrIDReply{CoreObjecttype: data}, nil
}

// List get a paginated list of coreObjecttypes by custom conditions
func (s *coreObjecttype) List(ctx context.Context, req *netbox_goV1.ListCoreObjecttypeRequest) (*netbox_goV1.ListCoreObjecttypeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCoreObjecttype.Err()
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

	coreObjecttypes := []*netbox_goV1.CoreObjecttype{}
	for _, record := range records {
		data, err := convertCoreObjecttype(record)
		if err != nil {
			logger.Warn("convertCoreObjecttype error", logger.Err(err), logger.Any("contenttypePtrID", record.ContenttypePtrID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		coreObjecttypes = append(coreObjecttypes, data)
	}

	return &netbox_goV1.ListCoreObjecttypeReply{
		Total:           total,
		CoreObjecttypes: coreObjecttypes,
	}, nil
}

// DeleteByContenttypePtrIDs batch delete coreObjecttypes by contenttypePtrIDs
func (s *coreObjecttype) DeleteByContenttypePtrIDs(ctx context.Context, req *netbox_goV1.DeleteCoreObjecttypeByContenttypePtrIDsRequest) (*netbox_goV1.DeleteCoreObjecttypeByContenttypePtrIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ids := make([]int, len(req.ContenttypePtrIDs))
	for i, v := range req.ContenttypePtrIDs {
		ids[i] = int(v)
	}
	err = s.iDao.DeleteByContenttypePtrIDs(ctx, ids)
	if err != nil {
		logger.Error("DeleteByContenttypePtrID error", logger.Err(err), logger.Any("contenttypePtrIDs", req.ContenttypePtrIDs), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.DeleteCoreObjecttypeByContenttypePtrIDsReply{}, nil
}

// GetByCondition get a coreObjecttype by custom condition
func (s *coreObjecttype) GetByCondition(ctx context.Context, req *netbox_goV1.GetCoreObjecttypeByConditionRequest) (*netbox_goV1.GetCoreObjecttypeByConditionReply, error) {
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

	data, err := convertCoreObjecttype(record)
	if err != nil {
		logger.Warn("convertCoreObjecttype error", logger.Err(err), logger.Any("coreObjecttype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCoreObjecttype.Err()
	}

	return &netbox_goV1.GetCoreObjecttypeByConditionReply{
		CoreObjecttype: data,
	}, nil
}

// ListByContenttypePtrIDs batch get coreObjecttypes by contenttypePtrIDs
func (s *coreObjecttype) ListByContenttypePtrIDs(ctx context.Context, req *netbox_goV1.ListCoreObjecttypeByContenttypePtrIDsRequest) (*netbox_goV1.ListCoreObjecttypeByContenttypePtrIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ids := make([]int, len(req.ContenttypePtrIDs))
	for i, v := range req.ContenttypePtrIDs {
		ids[i] = int(v)
	}
	coreObjecttypeMap, err := s.iDao.GetByContenttypePtrIDs(ctx, ids)
	if err != nil {
		logger.Error("GetByContenttypePtrIDs error", logger.Err(err), logger.Any("contenttypePtrIDs", req.ContenttypePtrIDs), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	coreObjecttypes := []*netbox_goV1.CoreObjecttype{}
	for _, contenttypePtrID := range req.ContenttypePtrIDs {
		if v, ok := coreObjecttypeMap[int(contenttypePtrID)]; ok {
			record, err := convertCoreObjecttype(v)
			if err != nil {
				logger.Warn("convertCoreObjecttype error", logger.Err(err), logger.Any("coreObjecttype", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			coreObjecttypes = append(coreObjecttypes, record)
		}
	}

	return &netbox_goV1.ListCoreObjecttypeByContenttypePtrIDsReply{CoreObjecttypes: coreObjecttypes}, nil
}

// ListByLastContenttypePtrID get a paginated list of coreObjecttypes by last contenttypePtrID
func (s *coreObjecttype) ListByLastContenttypePtrID(ctx context.Context, req *netbox_goV1.ListCoreObjecttypeByLastContenttypePtrIDRequest) (*netbox_goV1.ListCoreObjecttypeByLastContenttypePtrIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.CtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}

	if req.LastContenttypePtrID == 0 {
		req.LastContenttypePtrID = math.MaxInt32
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	records, err := s.iDao.GetByLastContenttypePtrID(ctx, int(req.LastContenttypePtrID), int(req.Limit), req.Sort)
	if err != nil {
		logger.Error("ListByLastContenttypePtrID error", logger.Err(err), interceptor.CtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	coreObjecttypes := []*netbox_goV1.CoreObjecttype{}
	for _, record := range records {
		data, err := convertCoreObjecttype(record)
		if err != nil {
			logger.Warn("convertCoreObjecttype error", logger.Err(err), logger.Any("contenttypePtrID", record.ContenttypePtrID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		coreObjecttypes = append(coreObjecttypes, data)
	}

	return &netbox_goV1.ListCoreObjecttypeByLastContenttypePtrIDReply{
		CoreObjecttypes: coreObjecttypes,
	}, nil
}

func convertCoreObjecttype(record *model.CoreObjecttype) (*netbox_goV1.CoreObjecttype, error) {
	value := &netbox_goV1.CoreObjecttype{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
