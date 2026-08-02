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
		netbox_goV1.RegisterIpamServiceServer(server, NewIpamServiceServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.IpamServiceServer = (*ipamService)(nil)
var _ time.Time

type ipamService struct {
	netbox_goV1.UnimplementedIpamServiceServer

	iDao dao.IpamServiceDao
}

// NewIpamServiceServer create a new service
func NewIpamServiceServer() netbox_goV1.IpamServiceServer {
	return &ipamService{
		iDao: dao.NewIpamServiceDao(
			database.GetDB(), // db driver is postgresql
			cache.NewIpamServiceCache(database.GetCacheType()),
		),
	}
}

// Create a new ipamService
func (s *ipamService) Create(ctx context.Context, req *netbox_goV1.CreateIpamServiceRequest) (*netbox_goV1.CreateIpamServiceReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamService{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateIpamService.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("ipamService", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateIpamServiceReply{Id: record.ID}, nil
}

// DeleteByID delete a ipamService by id
func (s *ipamService) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamServiceByIDRequest) (*netbox_goV1.DeleteIpamServiceByIDReply, error) {
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

	return &netbox_goV1.DeleteIpamServiceByIDReply{}, nil
}

// UpdateByID update a ipamService by id
func (s *ipamService) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamServiceByIDRequest) (*netbox_goV1.UpdateIpamServiceByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamService{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDIpamService.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("ipamService", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateIpamServiceByIDReply{}, nil
}

// GetByID get a ipamService by id
func (s *ipamService) GetByID(ctx context.Context, req *netbox_goV1.GetIpamServiceByIDRequest) (*netbox_goV1.GetIpamServiceByIDReply, error) {
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

	data, err := convertIpamService(record)
	if err != nil {
		logger.Warn("convertIpamService error", logger.Err(err), logger.Any("ipamService", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDIpamService.Err()
	}

	return &netbox_goV1.GetIpamServiceByIDReply{IpamService: data}, nil
}

// List get a paginated list of ipamServices by custom conditions
func (s *ipamService) List(ctx context.Context, req *netbox_goV1.ListIpamServiceRequest) (*netbox_goV1.ListIpamServiceReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListIpamService.Err()
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

	ipamServices := []*netbox_goV1.IpamService{}
	for _, record := range records {
		data, err := convertIpamService(record)
		if err != nil {
			logger.Warn("convertIpamService error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamServices = append(ipamServices, data)
	}

	return &netbox_goV1.ListIpamServiceReply{
		Total:        total,
		IpamServices: ipamServices,
	}, nil
}

// DeleteByIDs batch delete ipamService by ids
func (s *ipamService) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamServiceByIDsRequest) (*netbox_goV1.DeleteIpamServiceByIDsReply, error) {
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

	return &netbox_goV1.DeleteIpamServiceByIDsReply{}, nil
}

// GetByCondition get a ipamService by custom condition
func (s *ipamService) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamServiceByConditionRequest) (*netbox_goV1.GetIpamServiceByConditionReply, error) {
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

	data, err := convertIpamService(record)
	if err != nil {
		logger.Warn("convertIpamService error", logger.Err(err), logger.Any("ipamService", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionIpamService.Err()
	}

	return &netbox_goV1.GetIpamServiceByConditionReply{
		IpamService: data,
	}, nil
}

// ListByIDs batch get ipamService by ids
func (s *ipamService) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamServiceByIDsRequest) (*netbox_goV1.ListIpamServiceByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ipamServiceMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	ipamServices := []*netbox_goV1.IpamService{}
	for _, id := range req.Ids {
		if v, ok := ipamServiceMap[id]; ok {
			record, err := convertIpamService(v)
			if err != nil {
				logger.Warn("convertIpamService error", logger.Err(err), logger.Any("ipamService", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			ipamServices = append(ipamServices, record)
		}
	}

	return &netbox_goV1.ListIpamServiceByIDsReply{IpamServices: ipamServices}, nil
}

// ListByLastID get a paginated list of ipamServices by last id
func (s *ipamService) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamServiceByLastIDRequest) (*netbox_goV1.ListIpamServiceByLastIDReply, error) {
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

	ipamServices := []*netbox_goV1.IpamService{}
	for _, record := range records {
		data, err := convertIpamService(record)
		if err != nil {
			logger.Warn("convertIpamService error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamServices = append(ipamServices, data)
	}

	return &netbox_goV1.ListIpamServiceByLastIDReply{
		IpamServices: ipamServices,
	}, nil
}

func convertIpamService(record *model.IpamService) (*netbox_goV1.IpamService, error) {
	value := &netbox_goV1.IpamService{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
