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
		netbox_goV1.RegisterVpnIpsecprofileServer(server, NewVpnIpsecprofileServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.VpnIpsecprofileServer = (*vpnIpsecprofile)(nil)
var _ time.Time

type vpnIpsecprofile struct {
	netbox_goV1.UnimplementedVpnIpsecprofileServer

	iDao dao.VpnIpsecprofileDao
}

// NewVpnIpsecprofileServer create a new service
func NewVpnIpsecprofileServer() netbox_goV1.VpnIpsecprofileServer {
	return &vpnIpsecprofile{
		iDao: dao.NewVpnIpsecprofileDao(
			database.GetDB(), // db driver is postgresql
			cache.NewVpnIpsecprofileCache(database.GetCacheType()),
		),
	}
}

// Create a new vpnIpsecprofile
func (s *vpnIpsecprofile) Create(ctx context.Context, req *netbox_goV1.CreateVpnIpsecprofileRequest) (*netbox_goV1.CreateVpnIpsecprofileReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnIpsecprofile{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateVpnIpsecprofile.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("vpnIpsecprofile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateVpnIpsecprofileReply{Id: record.ID}, nil
}

// DeleteByID delete a vpnIpsecprofile by id
func (s *vpnIpsecprofile) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteVpnIpsecprofileByIDRequest) (*netbox_goV1.DeleteVpnIpsecprofileByIDReply, error) {
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

	return &netbox_goV1.DeleteVpnIpsecprofileByIDReply{}, nil
}

// UpdateByID update a vpnIpsecprofile by id
func (s *vpnIpsecprofile) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateVpnIpsecprofileByIDRequest) (*netbox_goV1.UpdateVpnIpsecprofileByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.VpnIpsecprofile{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDVpnIpsecprofile.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("vpnIpsecprofile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateVpnIpsecprofileByIDReply{}, nil
}

// GetByID get a vpnIpsecprofile by id
func (s *vpnIpsecprofile) GetByID(ctx context.Context, req *netbox_goV1.GetVpnIpsecprofileByIDRequest) (*netbox_goV1.GetVpnIpsecprofileByIDReply, error) {
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

	data, err := convertVpnIpsecprofile(record)
	if err != nil {
		logger.Warn("convertVpnIpsecprofile error", logger.Err(err), logger.Any("vpnIpsecprofile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDVpnIpsecprofile.Err()
	}

	return &netbox_goV1.GetVpnIpsecprofileByIDReply{VpnIpsecprofile: data}, nil
}

// List get a paginated list of vpnIpsecprofiles by custom conditions
func (s *vpnIpsecprofile) List(ctx context.Context, req *netbox_goV1.ListVpnIpsecprofileRequest) (*netbox_goV1.ListVpnIpsecprofileReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListVpnIpsecprofile.Err()
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

	vpnIpsecprofiles := []*netbox_goV1.VpnIpsecprofile{}
	for _, record := range records {
		data, err := convertVpnIpsecprofile(record)
		if err != nil {
			logger.Warn("convertVpnIpsecprofile error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnIpsecprofiles = append(vpnIpsecprofiles, data)
	}

	return &netbox_goV1.ListVpnIpsecprofileReply{
		Total:            total,
		VpnIpsecprofiles: vpnIpsecprofiles,
	}, nil
}

// DeleteByIDs batch delete vpnIpsecprofile by ids
func (s *vpnIpsecprofile) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteVpnIpsecprofileByIDsRequest) (*netbox_goV1.DeleteVpnIpsecprofileByIDsReply, error) {
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

	return &netbox_goV1.DeleteVpnIpsecprofileByIDsReply{}, nil
}

// GetByCondition get a vpnIpsecprofile by custom condition
func (s *vpnIpsecprofile) GetByCondition(ctx context.Context, req *netbox_goV1.GetVpnIpsecprofileByConditionRequest) (*netbox_goV1.GetVpnIpsecprofileByConditionReply, error) {
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

	data, err := convertVpnIpsecprofile(record)
	if err != nil {
		logger.Warn("convertVpnIpsecprofile error", logger.Err(err), logger.Any("vpnIpsecprofile", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionVpnIpsecprofile.Err()
	}

	return &netbox_goV1.GetVpnIpsecprofileByConditionReply{
		VpnIpsecprofile: data,
	}, nil
}

// ListByIDs batch get vpnIpsecprofile by ids
func (s *vpnIpsecprofile) ListByIDs(ctx context.Context, req *netbox_goV1.ListVpnIpsecprofileByIDsRequest) (*netbox_goV1.ListVpnIpsecprofileByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	vpnIpsecprofileMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	vpnIpsecprofiles := []*netbox_goV1.VpnIpsecprofile{}
	for _, id := range req.Ids {
		if v, ok := vpnIpsecprofileMap[id]; ok {
			record, err := convertVpnIpsecprofile(v)
			if err != nil {
				logger.Warn("convertVpnIpsecprofile error", logger.Err(err), logger.Any("vpnIpsecprofile", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			vpnIpsecprofiles = append(vpnIpsecprofiles, record)
		}
	}

	return &netbox_goV1.ListVpnIpsecprofileByIDsReply{VpnIpsecprofiles: vpnIpsecprofiles}, nil
}

// ListByLastID get a paginated list of vpnIpsecprofiles by last id
func (s *vpnIpsecprofile) ListByLastID(ctx context.Context, req *netbox_goV1.ListVpnIpsecprofileByLastIDRequest) (*netbox_goV1.ListVpnIpsecprofileByLastIDReply, error) {
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

	vpnIpsecprofiles := []*netbox_goV1.VpnIpsecprofile{}
	for _, record := range records {
		data, err := convertVpnIpsecprofile(record)
		if err != nil {
			logger.Warn("convertVpnIpsecprofile error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		vpnIpsecprofiles = append(vpnIpsecprofiles, data)
	}

	return &netbox_goV1.ListVpnIpsecprofileByLastIDReply{
		VpnIpsecprofiles: vpnIpsecprofiles,
	}, nil
}

func convertVpnIpsecprofile(record *model.VpnIpsecprofile) (*netbox_goV1.VpnIpsecprofile, error) {
	value := &netbox_goV1.VpnIpsecprofile{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
