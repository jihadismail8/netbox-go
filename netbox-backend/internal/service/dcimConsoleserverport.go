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
		netbox_goV1.RegisterDcimConsoleserverportServer(server, NewDcimConsoleserverportServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimConsoleserverportServer = (*dcimConsoleserverport)(nil)
var _ time.Time

type dcimConsoleserverport struct {
	netbox_goV1.UnimplementedDcimConsoleserverportServer

	iDao dao.DcimConsoleserverportDao
}

// NewDcimConsoleserverportServer create a new service
func NewDcimConsoleserverportServer() netbox_goV1.DcimConsoleserverportServer {
	return &dcimConsoleserverport{
		iDao: dao.NewDcimConsoleserverportDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimConsoleserverportCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimConsoleserverport
func (s *dcimConsoleserverport) Create(ctx context.Context, req *netbox_goV1.CreateDcimConsoleserverportRequest) (*netbox_goV1.CreateDcimConsoleserverportReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimConsoleserverport{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimConsoleserverport.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimConsoleserverport", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimConsoleserverportReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimConsoleserverport by id
func (s *dcimConsoleserverport) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimConsoleserverportByIDRequest) (*netbox_goV1.DeleteDcimConsoleserverportByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimConsoleserverportByIDReply{}, nil
}

// UpdateByID update a dcimConsoleserverport by id
func (s *dcimConsoleserverport) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimConsoleserverportByIDRequest) (*netbox_goV1.UpdateDcimConsoleserverportByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimConsoleserverport{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimConsoleserverport.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimConsoleserverport", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimConsoleserverportByIDReply{}, nil
}

// GetByID get a dcimConsoleserverport by id
func (s *dcimConsoleserverport) GetByID(ctx context.Context, req *netbox_goV1.GetDcimConsoleserverportByIDRequest) (*netbox_goV1.GetDcimConsoleserverportByIDReply, error) {
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

	data, err := convertDcimConsoleserverport(record)
	if err != nil {
		logger.Warn("convertDcimConsoleserverport error", logger.Err(err), logger.Any("dcimConsoleserverport", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimConsoleserverport.Err()
	}

	return &netbox_goV1.GetDcimConsoleserverportByIDReply{DcimConsoleserverport: data}, nil
}

// List get a paginated list of dcimConsoleserverports by custom conditions
func (s *dcimConsoleserverport) List(ctx context.Context, req *netbox_goV1.ListDcimConsoleserverportRequest) (*netbox_goV1.ListDcimConsoleserverportReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimConsoleserverport.Err()
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

	dcimConsoleserverports := []*netbox_goV1.DcimConsoleserverport{}
	for _, record := range records {
		data, err := convertDcimConsoleserverport(record)
		if err != nil {
			logger.Warn("convertDcimConsoleserverport error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimConsoleserverports = append(dcimConsoleserverports, data)
	}

	return &netbox_goV1.ListDcimConsoleserverportReply{
		Total:                  total,
		DcimConsoleserverports: dcimConsoleserverports,
	}, nil
}

// DeleteByIDs batch delete dcimConsoleserverport by ids
func (s *dcimConsoleserverport) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimConsoleserverportByIDsRequest) (*netbox_goV1.DeleteDcimConsoleserverportByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimConsoleserverportByIDsReply{}, nil
}

// GetByCondition get a dcimConsoleserverport by custom condition
func (s *dcimConsoleserverport) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimConsoleserverportByConditionRequest) (*netbox_goV1.GetDcimConsoleserverportByConditionReply, error) {
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

	data, err := convertDcimConsoleserverport(record)
	if err != nil {
		logger.Warn("convertDcimConsoleserverport error", logger.Err(err), logger.Any("dcimConsoleserverport", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimConsoleserverport.Err()
	}

	return &netbox_goV1.GetDcimConsoleserverportByConditionReply{
		DcimConsoleserverport: data,
	}, nil
}

// ListByIDs batch get dcimConsoleserverport by ids
func (s *dcimConsoleserverport) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimConsoleserverportByIDsRequest) (*netbox_goV1.ListDcimConsoleserverportByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimConsoleserverportMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimConsoleserverports := []*netbox_goV1.DcimConsoleserverport{}
	for _, id := range req.Ids {
		if v, ok := dcimConsoleserverportMap[id]; ok {
			record, err := convertDcimConsoleserverport(v)
			if err != nil {
				logger.Warn("convertDcimConsoleserverport error", logger.Err(err), logger.Any("dcimConsoleserverport", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimConsoleserverports = append(dcimConsoleserverports, record)
		}
	}

	return &netbox_goV1.ListDcimConsoleserverportByIDsReply{DcimConsoleserverports: dcimConsoleserverports}, nil
}

// ListByLastID get a paginated list of dcimConsoleserverports by last id
func (s *dcimConsoleserverport) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimConsoleserverportByLastIDRequest) (*netbox_goV1.ListDcimConsoleserverportByLastIDReply, error) {
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

	dcimConsoleserverports := []*netbox_goV1.DcimConsoleserverport{}
	for _, record := range records {
		data, err := convertDcimConsoleserverport(record)
		if err != nil {
			logger.Warn("convertDcimConsoleserverport error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimConsoleserverports = append(dcimConsoleserverports, data)
	}

	return &netbox_goV1.ListDcimConsoleserverportByLastIDReply{
		DcimConsoleserverports: dcimConsoleserverports,
	}, nil
}

func convertDcimConsoleserverport(record *model.DcimConsoleserverport) (*netbox_goV1.DcimConsoleserverport, error) {
	value := &netbox_goV1.DcimConsoleserverport{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
