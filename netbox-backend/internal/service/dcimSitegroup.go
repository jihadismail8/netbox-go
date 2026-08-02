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
		netbox_goV1.RegisterDcimSitegroupServer(server, NewDcimSitegroupServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimSitegroupServer = (*dcimSitegroup)(nil)
var _ time.Time

type dcimSitegroup struct {
	netbox_goV1.UnimplementedDcimSitegroupServer

	iDao dao.DcimSitegroupDao
}

// NewDcimSitegroupServer create a new service
func NewDcimSitegroupServer() netbox_goV1.DcimSitegroupServer {
	return &dcimSitegroup{
		iDao: dao.NewDcimSitegroupDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimSitegroupCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimSitegroup
func (s *dcimSitegroup) Create(ctx context.Context, req *netbox_goV1.CreateDcimSitegroupRequest) (*netbox_goV1.CreateDcimSitegroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimSitegroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimSitegroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimSitegroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimSitegroupReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimSitegroup by id
func (s *dcimSitegroup) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimSitegroupByIDRequest) (*netbox_goV1.DeleteDcimSitegroupByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimSitegroupByIDReply{}, nil
}

// UpdateByID update a dcimSitegroup by id
func (s *dcimSitegroup) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimSitegroupByIDRequest) (*netbox_goV1.UpdateDcimSitegroupByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimSitegroup{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimSitegroup.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimSitegroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimSitegroupByIDReply{}, nil
}

// GetByID get a dcimSitegroup by id
func (s *dcimSitegroup) GetByID(ctx context.Context, req *netbox_goV1.GetDcimSitegroupByIDRequest) (*netbox_goV1.GetDcimSitegroupByIDReply, error) {
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

	data, err := convertDcimSitegroup(record)
	if err != nil {
		logger.Warn("convertDcimSitegroup error", logger.Err(err), logger.Any("dcimSitegroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimSitegroup.Err()
	}

	return &netbox_goV1.GetDcimSitegroupByIDReply{DcimSitegroup: data}, nil
}

// List get a paginated list of dcimSitegroups by custom conditions
func (s *dcimSitegroup) List(ctx context.Context, req *netbox_goV1.ListDcimSitegroupRequest) (*netbox_goV1.ListDcimSitegroupReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimSitegroup.Err()
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

	dcimSitegroups := []*netbox_goV1.DcimSitegroup{}
	for _, record := range records {
		data, err := convertDcimSitegroup(record)
		if err != nil {
			logger.Warn("convertDcimSitegroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimSitegroups = append(dcimSitegroups, data)
	}

	return &netbox_goV1.ListDcimSitegroupReply{
		Total:          total,
		DcimSitegroups: dcimSitegroups,
	}, nil
}

// DeleteByIDs batch delete dcimSitegroup by ids
func (s *dcimSitegroup) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimSitegroupByIDsRequest) (*netbox_goV1.DeleteDcimSitegroupByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimSitegroupByIDsReply{}, nil
}

// GetByCondition get a dcimSitegroup by custom condition
func (s *dcimSitegroup) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimSitegroupByConditionRequest) (*netbox_goV1.GetDcimSitegroupByConditionReply, error) {
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

	data, err := convertDcimSitegroup(record)
	if err != nil {
		logger.Warn("convertDcimSitegroup error", logger.Err(err), logger.Any("dcimSitegroup", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimSitegroup.Err()
	}

	return &netbox_goV1.GetDcimSitegroupByConditionReply{
		DcimSitegroup: data,
	}, nil
}

// ListByIDs batch get dcimSitegroup by ids
func (s *dcimSitegroup) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimSitegroupByIDsRequest) (*netbox_goV1.ListDcimSitegroupByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimSitegroupMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimSitegroups := []*netbox_goV1.DcimSitegroup{}
	for _, id := range req.Ids {
		if v, ok := dcimSitegroupMap[id]; ok {
			record, err := convertDcimSitegroup(v)
			if err != nil {
				logger.Warn("convertDcimSitegroup error", logger.Err(err), logger.Any("dcimSitegroup", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimSitegroups = append(dcimSitegroups, record)
		}
	}

	return &netbox_goV1.ListDcimSitegroupByIDsReply{DcimSitegroups: dcimSitegroups}, nil
}

// ListByLastID get a paginated list of dcimSitegroups by last id
func (s *dcimSitegroup) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimSitegroupByLastIDRequest) (*netbox_goV1.ListDcimSitegroupByLastIDReply, error) {
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

	dcimSitegroups := []*netbox_goV1.DcimSitegroup{}
	for _, record := range records {
		data, err := convertDcimSitegroup(record)
		if err != nil {
			logger.Warn("convertDcimSitegroup error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimSitegroups = append(dcimSitegroups, data)
	}

	return &netbox_goV1.ListDcimSitegroupByLastIDReply{
		DcimSitegroups: dcimSitegroups,
	}, nil
}

func convertDcimSitegroup(record *model.DcimSitegroup) (*netbox_goV1.DcimSitegroup, error) {
	value := &netbox_goV1.DcimSitegroup{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
