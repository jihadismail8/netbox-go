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
		netbox_goV1.RegisterCoreJobServer(server, NewCoreJobServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.CoreJobServer = (*coreJob)(nil)
var _ time.Time

type coreJob struct {
	netbox_goV1.UnimplementedCoreJobServer

	iDao dao.CoreJobDao
}

// NewCoreJobServer create a new service
func NewCoreJobServer() netbox_goV1.CoreJobServer {
	return &coreJob{
		iDao: dao.NewCoreJobDao(
			database.GetDB(), // db driver is postgresql
			cache.NewCoreJobCache(database.GetCacheType()),
		),
	}
}

// Create a new coreJob
func (s *coreJob) Create(ctx context.Context, req *netbox_goV1.CreateCoreJobRequest) (*netbox_goV1.CreateCoreJobReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CoreJob{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateCoreJob.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("coreJob", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateCoreJobReply{Id: record.ID}, nil
}

// DeleteByID delete a coreJob by id
func (s *coreJob) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteCoreJobByIDRequest) (*netbox_goV1.DeleteCoreJobByIDReply, error) {
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

	return &netbox_goV1.DeleteCoreJobByIDReply{}, nil
}

// UpdateByID update a coreJob by id
func (s *coreJob) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateCoreJobByIDRequest) (*netbox_goV1.UpdateCoreJobByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.CoreJob{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDCoreJob.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("coreJob", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateCoreJobByIDReply{}, nil
}

// GetByID get a coreJob by id
func (s *coreJob) GetByID(ctx context.Context, req *netbox_goV1.GetCoreJobByIDRequest) (*netbox_goV1.GetCoreJobByIDReply, error) {
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

	data, err := convertCoreJob(record)
	if err != nil {
		logger.Warn("convertCoreJob error", logger.Err(err), logger.Any("coreJob", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDCoreJob.Err()
	}

	return &netbox_goV1.GetCoreJobByIDReply{CoreJob: data}, nil
}

// List get a paginated list of coreJobs by custom conditions
func (s *coreJob) List(ctx context.Context, req *netbox_goV1.ListCoreJobRequest) (*netbox_goV1.ListCoreJobReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListCoreJob.Err()
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

	coreJobs := []*netbox_goV1.CoreJob{}
	for _, record := range records {
		data, err := convertCoreJob(record)
		if err != nil {
			logger.Warn("convertCoreJob error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		coreJobs = append(coreJobs, data)
	}

	return &netbox_goV1.ListCoreJobReply{
		Total:    total,
		CoreJobs: coreJobs,
	}, nil
}

// DeleteByIDs batch delete coreJob by ids
func (s *coreJob) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteCoreJobByIDsRequest) (*netbox_goV1.DeleteCoreJobByIDsReply, error) {
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

	return &netbox_goV1.DeleteCoreJobByIDsReply{}, nil
}

// GetByCondition get a coreJob by custom condition
func (s *coreJob) GetByCondition(ctx context.Context, req *netbox_goV1.GetCoreJobByConditionRequest) (*netbox_goV1.GetCoreJobByConditionReply, error) {
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

	data, err := convertCoreJob(record)
	if err != nil {
		logger.Warn("convertCoreJob error", logger.Err(err), logger.Any("coreJob", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionCoreJob.Err()
	}

	return &netbox_goV1.GetCoreJobByConditionReply{
		CoreJob: data,
	}, nil
}

// ListByIDs batch get coreJob by ids
func (s *coreJob) ListByIDs(ctx context.Context, req *netbox_goV1.ListCoreJobByIDsRequest) (*netbox_goV1.ListCoreJobByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	coreJobMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	coreJobs := []*netbox_goV1.CoreJob{}
	for _, id := range req.Ids {
		if v, ok := coreJobMap[id]; ok {
			record, err := convertCoreJob(v)
			if err != nil {
				logger.Warn("convertCoreJob error", logger.Err(err), logger.Any("coreJob", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			coreJobs = append(coreJobs, record)
		}
	}

	return &netbox_goV1.ListCoreJobByIDsReply{CoreJobs: coreJobs}, nil
}

// ListByLastID get a paginated list of coreJobs by last id
func (s *coreJob) ListByLastID(ctx context.Context, req *netbox_goV1.ListCoreJobByLastIDRequest) (*netbox_goV1.ListCoreJobByLastIDReply, error) {
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

	coreJobs := []*netbox_goV1.CoreJob{}
	for _, record := range records {
		data, err := convertCoreJob(record)
		if err != nil {
			logger.Warn("convertCoreJob error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		coreJobs = append(coreJobs, data)
	}

	return &netbox_goV1.ListCoreJobByLastIDReply{
		CoreJobs: coreJobs,
	}, nil
}

func convertCoreJob(record *model.CoreJob) (*netbox_goV1.CoreJob, error) {
	value := &netbox_goV1.CoreJob{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
