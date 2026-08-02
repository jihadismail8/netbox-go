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
		netbox_goV1.RegisterIpamRirServer(server, NewIpamRirServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.IpamRirServer = (*ipamRir)(nil)
var _ time.Time

type ipamRir struct {
	netbox_goV1.UnimplementedIpamRirServer

	iDao dao.IpamRirDao
}

// NewIpamRirServer create a new service
func NewIpamRirServer() netbox_goV1.IpamRirServer {
	return &ipamRir{
		iDao: dao.NewIpamRirDao(
			database.GetDB(), // db driver is postgresql
			cache.NewIpamRirCache(database.GetCacheType()),
		),
	}
}

// Create a new ipamRir
func (s *ipamRir) Create(ctx context.Context, req *netbox_goV1.CreateIpamRirRequest) (*netbox_goV1.CreateIpamRirReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamRir{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateIpamRir.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("ipamRir", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateIpamRirReply{Id: record.ID}, nil
}

// DeleteByID delete a ipamRir by id
func (s *ipamRir) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamRirByIDRequest) (*netbox_goV1.DeleteIpamRirByIDReply, error) {
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

	return &netbox_goV1.DeleteIpamRirByIDReply{}, nil
}

// UpdateByID update a ipamRir by id
func (s *ipamRir) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamRirByIDRequest) (*netbox_goV1.UpdateIpamRirByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamRir{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDIpamRir.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("ipamRir", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateIpamRirByIDReply{}, nil
}

// GetByID get a ipamRir by id
func (s *ipamRir) GetByID(ctx context.Context, req *netbox_goV1.GetIpamRirByIDRequest) (*netbox_goV1.GetIpamRirByIDReply, error) {
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

	data, err := convertIpamRir(record)
	if err != nil {
		logger.Warn("convertIpamRir error", logger.Err(err), logger.Any("ipamRir", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDIpamRir.Err()
	}

	return &netbox_goV1.GetIpamRirByIDReply{IpamRir: data}, nil
}

// List get a paginated list of ipamRirs by custom conditions
func (s *ipamRir) List(ctx context.Context, req *netbox_goV1.ListIpamRirRequest) (*netbox_goV1.ListIpamRirReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListIpamRir.Err()
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

	ipamRirs := []*netbox_goV1.IpamRir{}
	for _, record := range records {
		data, err := convertIpamRir(record)
		if err != nil {
			logger.Warn("convertIpamRir error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamRirs = append(ipamRirs, data)
	}

	return &netbox_goV1.ListIpamRirReply{
		Total:    total,
		IpamRirs: ipamRirs,
	}, nil
}

// DeleteByIDs batch delete ipamRir by ids
func (s *ipamRir) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamRirByIDsRequest) (*netbox_goV1.DeleteIpamRirByIDsReply, error) {
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

	return &netbox_goV1.DeleteIpamRirByIDsReply{}, nil
}

// GetByCondition get a ipamRir by custom condition
func (s *ipamRir) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamRirByConditionRequest) (*netbox_goV1.GetIpamRirByConditionReply, error) {
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

	data, err := convertIpamRir(record)
	if err != nil {
		logger.Warn("convertIpamRir error", logger.Err(err), logger.Any("ipamRir", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionIpamRir.Err()
	}

	return &netbox_goV1.GetIpamRirByConditionReply{
		IpamRir: data,
	}, nil
}

// ListByIDs batch get ipamRir by ids
func (s *ipamRir) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamRirByIDsRequest) (*netbox_goV1.ListIpamRirByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ipamRirMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	ipamRirs := []*netbox_goV1.IpamRir{}
	for _, id := range req.Ids {
		if v, ok := ipamRirMap[id]; ok {
			record, err := convertIpamRir(v)
			if err != nil {
				logger.Warn("convertIpamRir error", logger.Err(err), logger.Any("ipamRir", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			ipamRirs = append(ipamRirs, record)
		}
	}

	return &netbox_goV1.ListIpamRirByIDsReply{IpamRirs: ipamRirs}, nil
}

// ListByLastID get a paginated list of ipamRirs by last id
func (s *ipamRir) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamRirByLastIDRequest) (*netbox_goV1.ListIpamRirByLastIDReply, error) {
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

	ipamRirs := []*netbox_goV1.IpamRir{}
	for _, record := range records {
		data, err := convertIpamRir(record)
		if err != nil {
			logger.Warn("convertIpamRir error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamRirs = append(ipamRirs, data)
	}

	return &netbox_goV1.ListIpamRirByLastIDReply{
		IpamRirs: ipamRirs,
	}, nil
}

func convertIpamRir(record *model.IpamRir) (*netbox_goV1.IpamRir, error) {
	value := &netbox_goV1.IpamRir{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
