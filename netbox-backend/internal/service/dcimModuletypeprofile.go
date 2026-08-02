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
		netbox_goV1.RegisterDcimModuletypeprofileServer(server, NewDcimModuletypeprofileServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.DcimModuletypeprofileServer = (*dcimModuletypeprofile)(nil)
var _ time.Time

type dcimModuletypeprofile struct {
	netbox_goV1.UnimplementedDcimModuletypeprofileServer

	iDao dao.DcimModuletypeprofileDao
}

// NewDcimModuletypeprofileServer create a new service
func NewDcimModuletypeprofileServer() netbox_goV1.DcimModuletypeprofileServer {
	return &dcimModuletypeprofile{
		iDao: dao.NewDcimModuletypeprofileDao(
			database.GetDB(), // db driver is postgresql
			cache.NewDcimModuletypeprofileCache(database.GetCacheType()),
		),
	}
}

// Create a new dcimModuletypeprofile
func (s *dcimModuletypeprofile) Create(ctx context.Context, req *netbox_goV1.CreateDcimModuletypeprofileRequest) (*netbox_goV1.CreateDcimModuletypeprofileReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimModuletypeprofile{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateDcimModuletypeprofile.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("dcimModuletypeprofile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateDcimModuletypeprofileReply{Id: record.ID}, nil
}

// DeleteByID delete a dcimModuletypeprofile by id
func (s *dcimModuletypeprofile) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteDcimModuletypeprofileByIDRequest) (*netbox_goV1.DeleteDcimModuletypeprofileByIDReply, error) {
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

	return &netbox_goV1.DeleteDcimModuletypeprofileByIDReply{}, nil
}

// UpdateByID update a dcimModuletypeprofile by id
func (s *dcimModuletypeprofile) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateDcimModuletypeprofileByIDRequest) (*netbox_goV1.UpdateDcimModuletypeprofileByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.DcimModuletypeprofile{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDDcimModuletypeprofile.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("dcimModuletypeprofile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateDcimModuletypeprofileByIDReply{}, nil
}

// GetByID get a dcimModuletypeprofile by id
func (s *dcimModuletypeprofile) GetByID(ctx context.Context, req *netbox_goV1.GetDcimModuletypeprofileByIDRequest) (*netbox_goV1.GetDcimModuletypeprofileByIDReply, error) {
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

	data, err := convertDcimModuletypeprofile(record)
	if err != nil {
		logger.Warn("convertDcimModuletypeprofile error", logger.Err(err), logger.Any("dcimModuletypeprofile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDDcimModuletypeprofile.Err()
	}

	return &netbox_goV1.GetDcimModuletypeprofileByIDReply{DcimModuletypeprofile: data}, nil
}

// List get a paginated list of dcimModuletypeprofiles by custom conditions
func (s *dcimModuletypeprofile) List(ctx context.Context, req *netbox_goV1.ListDcimModuletypeprofileRequest) (*netbox_goV1.ListDcimModuletypeprofileReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListDcimModuletypeprofile.Err()
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

	dcimModuletypeprofiles := []*netbox_goV1.DcimModuletypeprofile{}
	for _, record := range records {
		data, err := convertDcimModuletypeprofile(record)
		if err != nil {
			logger.Warn("convertDcimModuletypeprofile error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimModuletypeprofiles = append(dcimModuletypeprofiles, data)
	}

	return &netbox_goV1.ListDcimModuletypeprofileReply{
		Total:                  total,
		DcimModuletypeprofiles: dcimModuletypeprofiles,
	}, nil
}

// DeleteByIDs batch delete dcimModuletypeprofile by ids
func (s *dcimModuletypeprofile) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteDcimModuletypeprofileByIDsRequest) (*netbox_goV1.DeleteDcimModuletypeprofileByIDsReply, error) {
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

	return &netbox_goV1.DeleteDcimModuletypeprofileByIDsReply{}, nil
}

// GetByCondition get a dcimModuletypeprofile by custom condition
func (s *dcimModuletypeprofile) GetByCondition(ctx context.Context, req *netbox_goV1.GetDcimModuletypeprofileByConditionRequest) (*netbox_goV1.GetDcimModuletypeprofileByConditionReply, error) {
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

	data, err := convertDcimModuletypeprofile(record)
	if err != nil {
		logger.Warn("convertDcimModuletypeprofile error", logger.Err(err), logger.Any("dcimModuletypeprofile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionDcimModuletypeprofile.Err()
	}

	return &netbox_goV1.GetDcimModuletypeprofileByConditionReply{
		DcimModuletypeprofile: data,
	}, nil
}

// ListByIDs batch get dcimModuletypeprofile by ids
func (s *dcimModuletypeprofile) ListByIDs(ctx context.Context, req *netbox_goV1.ListDcimModuletypeprofileByIDsRequest) (*netbox_goV1.ListDcimModuletypeprofileByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	dcimModuletypeprofileMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	dcimModuletypeprofiles := []*netbox_goV1.DcimModuletypeprofile{}
	for _, id := range req.Ids {
		if v, ok := dcimModuletypeprofileMap[id]; ok {
			record, err := convertDcimModuletypeprofile(v)
			if err != nil {
				logger.Warn("convertDcimModuletypeprofile error", logger.Err(err), logger.Any("dcimModuletypeprofile", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			dcimModuletypeprofiles = append(dcimModuletypeprofiles, record)
		}
	}

	return &netbox_goV1.ListDcimModuletypeprofileByIDsReply{DcimModuletypeprofiles: dcimModuletypeprofiles}, nil
}

// ListByLastID get a paginated list of dcimModuletypeprofiles by last id
func (s *dcimModuletypeprofile) ListByLastID(ctx context.Context, req *netbox_goV1.ListDcimModuletypeprofileByLastIDRequest) (*netbox_goV1.ListDcimModuletypeprofileByLastIDReply, error) {
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

	dcimModuletypeprofiles := []*netbox_goV1.DcimModuletypeprofile{}
	for _, record := range records {
		data, err := convertDcimModuletypeprofile(record)
		if err != nil {
			logger.Warn("convertDcimModuletypeprofile error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		dcimModuletypeprofiles = append(dcimModuletypeprofiles, data)
	}

	return &netbox_goV1.ListDcimModuletypeprofileByLastIDReply{
		DcimModuletypeprofiles: dcimModuletypeprofiles,
	}, nil
}

func convertDcimModuletypeprofile(record *model.DcimModuletypeprofile) (*netbox_goV1.DcimModuletypeprofile, error) {
	value := &netbox_goV1.DcimModuletypeprofile{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
