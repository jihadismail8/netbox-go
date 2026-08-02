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
		netbox_goV1.RegisterIpamVlantranslationruleServer(server, NewIpamVlantranslationruleServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.IpamVlantranslationruleServer = (*ipamVlantranslationrule)(nil)
var _ time.Time

type ipamVlantranslationrule struct {
	netbox_goV1.UnimplementedIpamVlantranslationruleServer

	iDao dao.IpamVlantranslationruleDao
}

// NewIpamVlantranslationruleServer create a new service
func NewIpamVlantranslationruleServer() netbox_goV1.IpamVlantranslationruleServer {
	return &ipamVlantranslationrule{
		iDao: dao.NewIpamVlantranslationruleDao(
			database.GetDB(), // db driver is postgresql
			cache.NewIpamVlantranslationruleCache(database.GetCacheType()),
		),
	}
}

// Create a new ipamVlantranslationrule
func (s *ipamVlantranslationrule) Create(ctx context.Context, req *netbox_goV1.CreateIpamVlantranslationruleRequest) (*netbox_goV1.CreateIpamVlantranslationruleReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamVlantranslationrule{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateIpamVlantranslationrule.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("ipamVlantranslationrule", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateIpamVlantranslationruleReply{Id: record.ID}, nil
}

// DeleteByID delete a ipamVlantranslationrule by id
func (s *ipamVlantranslationrule) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteIpamVlantranslationruleByIDRequest) (*netbox_goV1.DeleteIpamVlantranslationruleByIDReply, error) {
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

	return &netbox_goV1.DeleteIpamVlantranslationruleByIDReply{}, nil
}

// UpdateByID update a ipamVlantranslationrule by id
func (s *ipamVlantranslationrule) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateIpamVlantranslationruleByIDRequest) (*netbox_goV1.UpdateIpamVlantranslationruleByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.IpamVlantranslationrule{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDIpamVlantranslationrule.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("ipamVlantranslationrule", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateIpamVlantranslationruleByIDReply{}, nil
}

// GetByID get a ipamVlantranslationrule by id
func (s *ipamVlantranslationrule) GetByID(ctx context.Context, req *netbox_goV1.GetIpamVlantranslationruleByIDRequest) (*netbox_goV1.GetIpamVlantranslationruleByIDReply, error) {
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

	data, err := convertIpamVlantranslationrule(record)
	if err != nil {
		logger.Warn("convertIpamVlantranslationrule error", logger.Err(err), logger.Any("ipamVlantranslationrule", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDIpamVlantranslationrule.Err()
	}

	return &netbox_goV1.GetIpamVlantranslationruleByIDReply{IpamVlantranslationrule: data}, nil
}

// List get a paginated list of ipamVlantranslationrules by custom conditions
func (s *ipamVlantranslationrule) List(ctx context.Context, req *netbox_goV1.ListIpamVlantranslationruleRequest) (*netbox_goV1.ListIpamVlantranslationruleReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListIpamVlantranslationrule.Err()
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

	ipamVlantranslationrules := []*netbox_goV1.IpamVlantranslationrule{}
	for _, record := range records {
		data, err := convertIpamVlantranslationrule(record)
		if err != nil {
			logger.Warn("convertIpamVlantranslationrule error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamVlantranslationrules = append(ipamVlantranslationrules, data)
	}

	return &netbox_goV1.ListIpamVlantranslationruleReply{
		Total:                    total,
		IpamVlantranslationrules: ipamVlantranslationrules,
	}, nil
}

// DeleteByIDs batch delete ipamVlantranslationrule by ids
func (s *ipamVlantranslationrule) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteIpamVlantranslationruleByIDsRequest) (*netbox_goV1.DeleteIpamVlantranslationruleByIDsReply, error) {
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

	return &netbox_goV1.DeleteIpamVlantranslationruleByIDsReply{}, nil
}

// GetByCondition get a ipamVlantranslationrule by custom condition
func (s *ipamVlantranslationrule) GetByCondition(ctx context.Context, req *netbox_goV1.GetIpamVlantranslationruleByConditionRequest) (*netbox_goV1.GetIpamVlantranslationruleByConditionReply, error) {
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

	data, err := convertIpamVlantranslationrule(record)
	if err != nil {
		logger.Warn("convertIpamVlantranslationrule error", logger.Err(err), logger.Any("ipamVlantranslationrule", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionIpamVlantranslationrule.Err()
	}

	return &netbox_goV1.GetIpamVlantranslationruleByConditionReply{
		IpamVlantranslationrule: data,
	}, nil
}

// ListByIDs batch get ipamVlantranslationrule by ids
func (s *ipamVlantranslationrule) ListByIDs(ctx context.Context, req *netbox_goV1.ListIpamVlantranslationruleByIDsRequest) (*netbox_goV1.ListIpamVlantranslationruleByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	ipamVlantranslationruleMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	ipamVlantranslationrules := []*netbox_goV1.IpamVlantranslationrule{}
	for _, id := range req.Ids {
		if v, ok := ipamVlantranslationruleMap[id]; ok {
			record, err := convertIpamVlantranslationrule(v)
			if err != nil {
				logger.Warn("convertIpamVlantranslationrule error", logger.Err(err), logger.Any("ipamVlantranslationrule", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			ipamVlantranslationrules = append(ipamVlantranslationrules, record)
		}
	}

	return &netbox_goV1.ListIpamVlantranslationruleByIDsReply{IpamVlantranslationrules: ipamVlantranslationrules}, nil
}

// ListByLastID get a paginated list of ipamVlantranslationrules by last id
func (s *ipamVlantranslationrule) ListByLastID(ctx context.Context, req *netbox_goV1.ListIpamVlantranslationruleByLastIDRequest) (*netbox_goV1.ListIpamVlantranslationruleByLastIDReply, error) {
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

	ipamVlantranslationrules := []*netbox_goV1.IpamVlantranslationrule{}
	for _, record := range records {
		data, err := convertIpamVlantranslationrule(record)
		if err != nil {
			logger.Warn("convertIpamVlantranslationrule error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		ipamVlantranslationrules = append(ipamVlantranslationrules, data)
	}

	return &netbox_goV1.ListIpamVlantranslationruleByLastIDReply{
		IpamVlantranslationrules: ipamVlantranslationrules,
	}, nil
}

func convertIpamVlantranslationrule(record *model.IpamVlantranslationrule) (*netbox_goV1.IpamVlantranslationrule, error) {
	value := &netbox_goV1.IpamVlantranslationrule{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
