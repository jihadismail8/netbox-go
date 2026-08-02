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
		netbox_goV1.RegisterDcimPowerpanelServer(server, NewDcimPowerpanelServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimPowerpanelServer = (*dcimPowerpanel)(nil)
var _ time.Time

type dcimPowerpanel struct {
	netbox_goV1.UnimplementedDcimPowerpanelServer

	iDao dao.DcimPowerpanelDao
}

// NewDcimPowerpanelServer create a new service
func NewDcimPowerpanelServer() netbox_goV1.DcimPowerpanelServer {
	return &dcimPowerpanel{
		iDao: dao.NewDcimPowerpanelDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimPowerpanelCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimPowerpanel
func (s *dcimPowerpanel) Create(ctx context.Context, req *netbox_goV1.CreateDcimPowerpanelRequest) (*netbox_goV1.CreateDcimPowerpanelReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimPowerpanel{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimPowerpanel.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimPowerpanel", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimPowerpanelReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimPowerpanel by id
func (s *dcimPowerpanel) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimPowerpanelByIDRequest) (*netbox_goV1.DeleteDcimPowerpanelByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimPowerpanelByIDReply{}, nil
}

// UpdateByID update a dcimPowerpanel by id
func (s *dcimPowerpanel) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimPowerpanelByIDRequest) (*netbox_goV1.UpdateDcimPowerpanelByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimPowerpanel{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimPowerpanel.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimPowerpanel", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimPowerpanelByIDReply{}, nil
}

// GetByID get a dcimPowerpanel by id
func (s *dcimPowerpanel) GetByID(ctx context.Context, req *netbox_goV1.GetDcimPowerpanelByIDRequest) (*netbox_goV1.GetDcimPowerpanelByIDReply, error) {
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

	data, err := convertDcimPowerpanel(record)
	if err != nil {
		logger.Warn("convertDcimPowerpanel error", logger.Err(err), logger.Any("dcimPowerpanel", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimPowerpanel.Err()
	}

	return &netbox_goV1.GetDcimPowerpanelByIDReply{DcimPowerpanel: data}, nil
}

// List get a paginated list of dcimPowerpanels by custom conditions
func (s *dcimPowerpanel) List(ctx context.Context, req *netbox_goV1.ListDcimPowerpanelRequest) (*netbox_goV1.ListDcimPowerpanelReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimPowerpanel.Err()
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

	dcimPowerpanels := []*netbox_goV1.DcimPowerpanel{}
	for _, record := range records {
		data, err := convertDcimPowerpanel(record)
		if err != nil {
			logger.Warn("convertDcimPowerpanel error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimPowerpanels = append(dcimPowerpanels, data)
	}

	return &netbox_goV1.ListDcimPowerpanelReply{
		Total:           total,
		DcimPowerpanels: dcimPowerpanels,
	}, nil
}

// DeleteByIDs batch delete dcimPowerpanel by ids
func (s *dcimPowerpanel) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimPowerpanelByIDsRequest) (*netbox_goV1.DeleteDcimPowerpanelByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimPowerpanelByIDsReply{}, nil
}

// GetByCondition get a dcimPowerpanel by custom condition
func (s *dcimPowerpanel) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimPowerpanelByConditionRequest) (*netbox_goV1.GetDcimPowerpanelByConditionReply, error) {
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

	data, err := convertDcimPowerpanel(record)
	if err != nil {
		logger.Warn("convertDcimPowerpanel error", logger.Err(err), logger.Any("dcimPowerpanel", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimPowerpanel.Err()
	}

	return &netbox_goV1.GetDcimPowerpanelByConditionReply{
		DcimPowerpanel: data,
	}, nil
}

// ListByIDs batch get dcimPowerpanel by ids
func (s *dcimPowerpanel) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimPowerpanelByIDsRequest) (*netbox_goV1.ListDcimPowerpanelByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimPowerpanelMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimPowerpanels := []*netbox_goV1.DcimPowerpanel{}
	for _, id := range req.Ids {
		if v, ok := dcimPowerpanelMap[id]; ok {
			record, err := convertDcimPowerpanel(v)
			if err != nil {
				logger.Warn("convertDcimPowerpanel error", logger.Err(err), logger.Any("dcimPowerpanel", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimPowerpanels = append(dcimPowerpanels, record)
		}
	}

	return &netbox_goV1.ListDcimPowerpanelByIDsReply{DcimPowerpanels: dcimPowerpanels}, nil
}

// ListByLastID get a paginated list of dcimPowerpanels by last id
func (s *dcimPowerpanel) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimPowerpanelByLastIDRequest) (*netbox_goV1.ListDcimPowerpanelByLastIDReply, error) {
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

	dcimPowerpanels := []*netbox_goV1.DcimPowerpanel{}
	for _, record := range records {
		data, err := convertDcimPowerpanel(record)
		if err != nil {
			logger.Warn("convertDcimPowerpanel error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimPowerpanels = append(dcimPowerpanels, data)
	}

	return &netbox_goV1.ListDcimPowerpanelByLastIDReply{
		DcimPowerpanels: dcimPowerpanels,
	}, nil
}

func convertDcimPowerpanel(record *model.DcimPowerpanel) (*netbox_goV1.DcimPowerpanel, error) {
	value := &netbox_goV1.DcimPowerpanel{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
