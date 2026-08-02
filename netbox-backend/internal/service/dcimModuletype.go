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
		netbox_goV1.RegisterDcimModuletypeServer(server, NewDcimModuletypeServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimModuletypeServer = (*dcimModuletype)(nil)
var _ time.Time

type dcimModuletype struct {
	netbox_goV1.UnimplementedDcimModuletypeServer

	iDao dao.DcimModuletypeDao
}

// NewDcimModuletypeServer create a new service
func NewDcimModuletypeServer() netbox_goV1.DcimModuletypeServer {
	return &dcimModuletype{
		iDao: dao.NewDcimModuletypeDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimModuletypeCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimModuletype
func (s *dcimModuletype) Create(ctx context.Context, req *netbox_goV1.CreateDcimModuletypeRequest) (*netbox_goV1.CreateDcimModuletypeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimModuletype{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimModuletype.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimModuletype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimModuletypeReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimModuletype by id
func (s *dcimModuletype) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimModuletypeByIDRequest) (*netbox_goV1.DeleteDcimModuletypeByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimModuletypeByIDReply{}, nil
}

// UpdateByID update a dcimModuletype by id
func (s *dcimModuletype) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimModuletypeByIDRequest) (*netbox_goV1.UpdateDcimModuletypeByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimModuletype{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimModuletype.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimModuletype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimModuletypeByIDReply{}, nil
}

// GetByID get a dcimModuletype by id
func (s *dcimModuletype) GetByID(ctx context.Context, req *netbox_goV1.GetDcimModuletypeByIDRequest) (*netbox_goV1.GetDcimModuletypeByIDReply, error) {
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

	data, err := convertDcimModuletype(record)
	if err != nil {
		logger.Warn("convertDcimModuletype error", logger.Err(err), logger.Any("dcimModuletype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimModuletype.Err()
	}

	return &netbox_goV1.GetDcimModuletypeByIDReply{DcimModuletype: data}, nil
}

// List get a paginated list of dcimModuletypes by custom conditions
func (s *dcimModuletype) List(ctx context.Context, req *netbox_goV1.ListDcimModuletypeRequest) (*netbox_goV1.ListDcimModuletypeReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimModuletype.Err()
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

	dcimModuletypes := []*netbox_goV1.DcimModuletype{}
	for _, record := range records {
		data, err := convertDcimModuletype(record)
		if err != nil {
			logger.Warn("convertDcimModuletype error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimModuletypes = append(dcimModuletypes, data)
	}

	return &netbox_goV1.ListDcimModuletypeReply{
		Total:           total,
		DcimModuletypes: dcimModuletypes,
	}, nil
}

// DeleteByIDs batch delete dcimModuletype by ids
func (s *dcimModuletype) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimModuletypeByIDsRequest) (*netbox_goV1.DeleteDcimModuletypeByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimModuletypeByIDsReply{}, nil
}

// GetByCondition get a dcimModuletype by custom condition
func (s *dcimModuletype) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimModuletypeByConditionRequest) (*netbox_goV1.GetDcimModuletypeByConditionReply, error) {
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

	data, err := convertDcimModuletype(record)
	if err != nil {
		logger.Warn("convertDcimModuletype error", logger.Err(err), logger.Any("dcimModuletype", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimModuletype.Err()
	}

	return &netbox_goV1.GetDcimModuletypeByConditionReply{
		DcimModuletype: data,
	}, nil
}

// ListByIDs batch get dcimModuletype by ids
func (s *dcimModuletype) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimModuletypeByIDsRequest) (*netbox_goV1.ListDcimModuletypeByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimModuletypeMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimModuletypes := []*netbox_goV1.DcimModuletype{}
	for _, id := range req.Ids {
		if v, ok := dcimModuletypeMap[id]; ok {
			record, err := convertDcimModuletype(v)
			if err != nil {
				logger.Warn("convertDcimModuletype error", logger.Err(err), logger.Any("dcimModuletype", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimModuletypes = append(dcimModuletypes, record)
		}
	}

	return &netbox_goV1.ListDcimModuletypeByIDsReply{DcimModuletypes: dcimModuletypes}, nil
}

// ListByLastID get a paginated list of dcimModuletypes by last id
func (s *dcimModuletype) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimModuletypeByLastIDRequest) (*netbox_goV1.ListDcimModuletypeByLastIDReply, error) {
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

	dcimModuletypes := []*netbox_goV1.DcimModuletype{}
	for _, record := range records {
		data, err := convertDcimModuletype(record)
		if err != nil {
			logger.Warn("convertDcimModuletype error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimModuletypes = append(dcimModuletypes, data)
	}

	return &netbox_goV1.ListDcimModuletypeByLastIDReply{
		DcimModuletypes: dcimModuletypes,
	}, nil
}

func convertDcimModuletype(record *model.DcimModuletype) (*netbox_goV1.DcimModuletype, error) {
	value := &netbox_goV1.DcimModuletype{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
