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
		netbox_goV1.RegisterCoreDatafileServer(server, NewCoreDatafileServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CoreDatafileServer = (*coreDatafile)(nil)
var _ time.Time

type coreDatafile struct {
	netbox_goV1.UnimplementedCoreDatafileServer

	iDao dao.CoreDatafileDao
}

// NewCoreDatafileServer create a new service
func NewCoreDatafileServer() netbox_goV1.CoreDatafileServer {
	return &coreDatafile{
		iDao: dao.NewCoreDatafileDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCoreDatafileCache(database.GetCacheType()),
		),
	}
}

// Create a new coreDatafile
func (s *coreDatafile) Create(ctx context.Context, req *netbox_goV1.CreateCoreDatafileRequest) (*netbox_goV1.CreateCoreDatafileReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CoreDatafile{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCoreDatafile.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("coreDatafile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCoreDatafileReply{Id: record.ID}, nil
}

// DeleteByID delete a coreDatafile by id
func (s *coreDatafile) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCoreDatafileByIDRequest) (*netbox_goV1.DeleteCoreDatafileByIDReply, error) {
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

	return &netbox_goV1.DeleteCoreDatafileByIDReply{}, nil
}

// UpdateByID update a coreDatafile by id
func (s *coreDatafile) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCoreDatafileByIDRequest) (*netbox_goV1.UpdateCoreDatafileByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CoreDatafile{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDCoreDatafile.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("coreDatafile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCoreDatafileByIDReply{}, nil
}

// GetByID get a coreDatafile by id
func (s *coreDatafile) GetByID(ctx context.Context, req *netbox_goV1.GetCoreDatafileByIDRequest) (*netbox_goV1.GetCoreDatafileByIDReply, error) {
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

	data, err := convertCoreDatafile(record)
	if err != nil {
		logger.Warn("convertCoreDatafile error", logger.Err(err), logger.Any("coreDatafile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDCoreDatafile.Err()
	}

	return &netbox_goV1.GetCoreDatafileByIDReply{CoreDatafile: data}, nil
}

// List get a paginated list of coreDatafiles by custom conditions
func (s *coreDatafile) List(ctx context.Context, req *netbox_goV1.ListCoreDatafileRequest) (*netbox_goV1.ListCoreDatafileReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCoreDatafile.Err()
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

	coreDatafiles := []*netbox_goV1.CoreDatafile{}
	for _, record := range records {
		data, err := convertCoreDatafile(record)
		if err != nil {
			logger.Warn("convertCoreDatafile error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		coreDatafiles = append(coreDatafiles, data)
	}

	return &netbox_goV1.ListCoreDatafileReply{
		Total:         total,
		CoreDatafiles: coreDatafiles,
	}, nil
}

// DeleteByIDs batch delete coreDatafile by ids
func (s *coreDatafile) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCoreDatafileByIDsRequest) (*netbox_goV1.DeleteCoreDatafileByIDsReply, error) {
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

	return &netbox_goV1.DeleteCoreDatafileByIDsReply{}, nil
}

// GetByCondition get a coreDatafile by custom condition
func (s *coreDatafile) GetByCondition(ctx context.Context, req *netbox_goV1.GetCoreDatafileByConditionRequest) (*netbox_goV1.GetCoreDatafileByConditionReply, error) {
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

	data, err := convertCoreDatafile(record)
	if err != nil {
		logger.Warn("convertCoreDatafile error", logger.Err(err), logger.Any("coreDatafile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCoreDatafile.Err()
	}

	return &netbox_goV1.GetCoreDatafileByConditionReply{
		CoreDatafile: data,
	}, nil
}

// ListByIDs batch get coreDatafile by ids
func (s *coreDatafile) ListByIDs(ctx context.Context, req *netbox_goV1.ListCoreDatafileByIDsRequest) (*netbox_goV1.ListCoreDatafileByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	coreDatafileMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	coreDatafiles := []*netbox_goV1.CoreDatafile{}
	for _, id := range req.Ids {
		if v, ok := coreDatafileMap[id]; ok {
			record, err := convertCoreDatafile(v)
			if err != nil {
				logger.Warn("convertCoreDatafile error", logger.Err(err), logger.Any("coreDatafile", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			coreDatafiles = append(coreDatafiles, record)
		}
	}

	return &netbox_goV1.ListCoreDatafileByIDsReply{CoreDatafiles: coreDatafiles}, nil
}

// ListByLastID get a paginated list of coreDatafiles by last id
func (s *coreDatafile) ListByLastID(ctx context.Context, req *netbox_goV1.ListCoreDatafileByLastIDRequest) (*netbox_goV1.ListCoreDatafileByLastIDReply, error) {
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

	coreDatafiles := []*netbox_goV1.CoreDatafile{}
	for _, record := range records {
		data, err := convertCoreDatafile(record)
		if err != nil {
			logger.Warn("convertCoreDatafile error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		coreDatafiles = append(coreDatafiles, data)
	}

	return &netbox_goV1.ListCoreDatafileByLastIDReply{
		CoreDatafiles: coreDatafiles,
	}, nil
}

func convertCoreDatafile(record *model.CoreDatafile) (*netbox_goV1.CoreDatafile, error) {
	value := &netbox_goV1.CoreDatafile{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
