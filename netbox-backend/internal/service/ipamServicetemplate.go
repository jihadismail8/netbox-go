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
		netbox_goV1.RegisterIpamServicetemplateServer(server, NewIpamServicetemplateServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.IpamServicetemplateServer = (*ipamServicetemplate)(nil)
var _ time.Time

type ipamServicetemplate struct {
	netbox_goV1.UnimplementedIpamServicetemplateServer

	iDao dao.IpamServicetemplateDao
}

// NewIpamServicetemplateServer create a new service
func NewIpamServicetemplateServer() netbox_goV1.IpamServicetemplateServer {
	return &ipamServicetemplate{
		iDao: dao.NewIpamServicetemplateDao(
			database.GetDB(), // db driver is postgresql
			cache.NewIpamServicetemplateCache(database.GetCacheType()),
		),
	}
}

// Create a new ipamServicetemplate
func (s *ipamServicetemplate) Create(ctx context.Context, req *netbox_goV1.CreateIpamServicetemplateRequest) (*netbox_goV1.CreateIpamServicetemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamServicetemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateIpamServicetemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("ipamServicetemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateIpamServicetemplateReply{Id: record.ID}, nil
}

// DeleteByID delete a ipamServicetemplate by id
func (s *ipamServicetemplate) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamServicetemplateByIDRequest) (*netbox_goV1.DeleteIpamServicetemplateByIDReply, error) {
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

	return &netbox_goV1.DeleteIpamServicetemplateByIDReply{}, nil
}

// UpdateByID update a ipamServicetemplate by id
func (s *ipamServicetemplate) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamServicetemplateByIDRequest) (*netbox_goV1.UpdateIpamServicetemplateByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamServicetemplate{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDIpamServicetemplate.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("ipamServicetemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateIpamServicetemplateByIDReply{}, nil
}

// GetByID get a ipamServicetemplate by id
func (s *ipamServicetemplate) GetByID(ctx context.Context, req *netbox_goV1.GetIpamServicetemplateByIDRequest) (*netbox_goV1.GetIpamServicetemplateByIDReply, error) {
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

	data, err := convertIpamServicetemplate(record)
	if err != nil {
		logger.Warn("convertIpamServicetemplate error", logger.Err(err), logger.Any("ipamServicetemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDIpamServicetemplate.Err()
	}

	return &netbox_goV1.GetIpamServicetemplateByIDReply{IpamServicetemplate: data}, nil
}

// List get a paginated list of ipamServicetemplates by custom conditions
func (s *ipamServicetemplate) List(ctx context.Context, req *netbox_goV1.ListIpamServicetemplateRequest) (*netbox_goV1.ListIpamServicetemplateReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListIpamServicetemplate.Err()
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

	ipamServicetemplates := []*netbox_goV1.IpamServicetemplate{}
	for _, record := range records {
		data, err := convertIpamServicetemplate(record)
		if err != nil {
			logger.Warn("convertIpamServicetemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamServicetemplates = append(ipamServicetemplates, data)
	}

	return &netbox_goV1.ListIpamServicetemplateReply{
		Total:                total,
		IpamServicetemplates: ipamServicetemplates,
	}, nil
}

// DeleteByIDs batch delete ipamServicetemplate by ids
func (s *ipamServicetemplate) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamServicetemplateByIDsRequest) (*netbox_goV1.DeleteIpamServicetemplateByIDsReply, error) {
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

	return &netbox_goV1.DeleteIpamServicetemplateByIDsReply{}, nil
}

// GetByCondition get a ipamServicetemplate by custom condition
func (s *ipamServicetemplate) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamServicetemplateByConditionRequest) (*netbox_goV1.GetIpamServicetemplateByConditionReply, error) {
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

	data, err := convertIpamServicetemplate(record)
	if err != nil {
		logger.Warn("convertIpamServicetemplate error", logger.Err(err), logger.Any("ipamServicetemplate", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionIpamServicetemplate.Err()
	}

	return &netbox_goV1.GetIpamServicetemplateByConditionReply{
		IpamServicetemplate: data,
	}, nil
}

// ListByIDs batch get ipamServicetemplate by ids
func (s *ipamServicetemplate) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamServicetemplateByIDsRequest) (*netbox_goV1.ListIpamServicetemplateByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ipamServicetemplateMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	ipamServicetemplates := []*netbox_goV1.IpamServicetemplate{}
	for _, id := range req.Ids {
		if v, ok := ipamServicetemplateMap[id]; ok {
			record, err := convertIpamServicetemplate(v)
			if err != nil {
				logger.Warn("convertIpamServicetemplate error", logger.Err(err), logger.Any("ipamServicetemplate", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			ipamServicetemplates = append(ipamServicetemplates, record)
		}
	}

	return &netbox_goV1.ListIpamServicetemplateByIDsReply{IpamServicetemplates: ipamServicetemplates}, nil
}

// ListByLastID get a paginated list of ipamServicetemplates by last id
func (s *ipamServicetemplate) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamServicetemplateByLastIDRequest) (*netbox_goV1.ListIpamServicetemplateByLastIDReply, error) {
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

	ipamServicetemplates := []*netbox_goV1.IpamServicetemplate{}
	for _, record := range records {
		data, err := convertIpamServicetemplate(record)
		if err != nil {
			logger.Warn("convertIpamServicetemplate error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamServicetemplates = append(ipamServicetemplates, data)
	}

	return &netbox_goV1.ListIpamServicetemplateByLastIDReply{
		IpamServicetemplates: ipamServicetemplates,
	}, nil
}

func convertIpamServicetemplate(record *model.IpamServicetemplate) (*netbox_goV1.IpamServicetemplate, error) {
	value := &netbox_goV1.IpamServicetemplate{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
