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
		netbox_goV1.RegisterExtrasBookmarkServer(server, NewExtrasBookmarkServer()) // register service to the rpc service
	})
}

var _ netbox_goV1.ExtrasBookmarkServer = (*extrasBookmark)(nil)
var _ time.Time

type extrasBookmark struct {
	netbox_goV1.UnimplementedExtrasBookmarkServer

	iDao dao.ExtrasBookmarkDao
}

// NewExtrasBookmarkServer create a new service
func NewExtrasBookmarkServer() netbox_goV1.ExtrasBookmarkServer {
	return &extrasBookmark{
		iDao: dao.NewExtrasBookmarkDao(
			database.GetDB(), // db driver is postgresql
			cache.NewExtrasBookmarkCache(database.GetCacheType()),
		),
	}
}

// Create a new extrasBookmark
func (s *extrasBookmark) Create(ctx context.Context, req *netbox_goV1.CreateExtrasBookmarkRequest) (*netbox_goV1.CreateExtrasBookmarkReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasBookmark{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusCreateExtrasBookmark.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = s.iDao.Create(ctx, record)
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("extrasBookmark", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.CreateExtrasBookmarkReply{Id: record.ID}, nil
}

// DeleteByID delete a extrasBookmark by id
func (s *extrasBookmark) DeleteByID(ctx context.Context, req *netbox_goV1.DeleteExtrasBookmarkByIDRequest) (*netbox_goV1.DeleteExtrasBookmarkByIDReply, error) {
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

	return &netbox_goV1.DeleteExtrasBookmarkByIDReply{}, nil
}

// UpdateByID update a extrasBookmark by id
func (s *extrasBookmark) UpdateByID(ctx context.Context, req *netbox_goV1.UpdateExtrasBookmarkByIDRequest) (*netbox_goV1.UpdateExtrasBookmarkByIDReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	record := &model.ExtrasBookmark{}
	err = copier.Copy(record, req)
	if err != nil {
		return nil, ecode.StatusUpdateByIDExtrasBookmark.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	record.ID = req.Id

	err = s.iDao.UpdateByID(ctx, record)
	if err != nil {
		logger.Error("UpdateByID error", logger.Err(err), logger.Any("extrasBookmark", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	return &netbox_goV1.UpdateExtrasBookmarkByIDReply{}, nil
}

// GetByID get a extrasBookmark by id
func (s *extrasBookmark) GetByID(ctx context.Context, req *netbox_goV1.GetExtrasBookmarkByIDRequest) (*netbox_goV1.GetExtrasBookmarkByIDReply, error) {
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

	data, err := convertExtrasBookmark(record)
	if err != nil {
		logger.Warn("convertExtrasBookmark error", logger.Err(err), logger.Any("extrasBookmark", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByIDExtrasBookmark.Err()
	}

	return &netbox_goV1.GetExtrasBookmarkByIDReply{ExtrasBookmark: data}, nil
}

// List get a paginated list of extrasBookmarks by custom conditions
func (s *extrasBookmark) List(ctx context.Context, req *netbox_goV1.ListExtrasBookmarkRequest) (*netbox_goV1.ListExtrasBookmarkReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.StatusListExtrasBookmark.Err()
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

	extrasBookmarks := []*netbox_goV1.ExtrasBookmark{}
	for _, record := range records {
		data, err := convertExtrasBookmark(record)
		if err != nil {
			logger.Warn("convertExtrasBookmark error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasBookmarks = append(extrasBookmarks, data)
	}

	return &netbox_goV1.ListExtrasBookmarkReply{
		Total:           total,
		ExtrasBookmarks: extrasBookmarks,
	}, nil
}

// DeleteByIDs batch delete extrasBookmark by ids
func (s *extrasBookmark) DeleteByIDs(ctx context.Context, req *netbox_goV1.DeleteExtrasBookmarkByIDsRequest) (*netbox_goV1.DeleteExtrasBookmarkByIDsReply, error) {
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

	return &netbox_goV1.DeleteExtrasBookmarkByIDsReply{}, nil
}

// GetByCondition get a extrasBookmark by custom condition
func (s *extrasBookmark) GetByCondition(ctx context.Context, req *netbox_goV1.GetExtrasBookmarkByConditionRequest) (*netbox_goV1.GetExtrasBookmarkByConditionReply, error) {
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

	data, err := convertExtrasBookmark(record)
	if err != nil {
		logger.Warn("convertExtrasBookmark error", logger.Err(err), logger.Any("extrasBookmark", record), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusGetByConditionExtrasBookmark.Err()
	}

	return &netbox_goV1.GetExtrasBookmarkByConditionReply{
		ExtrasBookmark: data,
	}, nil
}

// ListByIDs batch get extrasBookmark by ids
func (s *extrasBookmark) ListByIDs(ctx context.Context, req *netbox_goV1.ListExtrasBookmarkByIDsRequest) (*netbox_goV1.ListExtrasBookmarkByIDsReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInvalidParams.Err()
	}
	ctx = interceptor.WrapServerCtx(ctx)

	extrasBookmarkMap, err := s.iDao.GetByIDs(ctx, req.Ids)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("ids", req.Ids), interceptor.ServerCtxRequestIDField(ctx))
		return nil, ecode.StatusInternalServerError.ToRPCErr()
	}

	extrasBookmarks := []*netbox_goV1.ExtrasBookmark{}
	for _, id := range req.Ids {
		if v, ok := extrasBookmarkMap[id]; ok {
			record, err := convertExtrasBookmark(v)
			if err != nil {
				logger.Warn("convertExtrasBookmark error", logger.Err(err), logger.Any("extrasBookmark", v), interceptor.ServerCtxRequestIDField(ctx))
				return nil, ecode.StatusInternalServerError.ToRPCErr()
			}
			extrasBookmarks = append(extrasBookmarks, record)
		}
	}

	return &netbox_goV1.ListExtrasBookmarkByIDsReply{ExtrasBookmarks: extrasBookmarks}, nil
}

// ListByLastID get a paginated list of extrasBookmarks by last id
func (s *extrasBookmark) ListByLastID(ctx context.Context, req *netbox_goV1.ListExtrasBookmarkByLastIDRequest) (*netbox_goV1.ListExtrasBookmarkByLastIDReply, error) {
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

	extrasBookmarks := []*netbox_goV1.ExtrasBookmark{}
	for _, record := range records {
		data, err := convertExtrasBookmark(record)
		if err != nil {
			logger.Warn("convertExtrasBookmark error", logger.Err(err), logger.Any("id", record.ID), interceptor.ServerCtxRequestIDField(ctx))
			continue
		}
		extrasBookmarks = append(extrasBookmarks, data)
	}

	return &netbox_goV1.ListExtrasBookmarkByLastIDReply{
		ExtrasBookmarks: extrasBookmarks,
	}, nil
}

func convertExtrasBookmark(record *model.ExtrasBookmark) (*netbox_goV1.ExtrasBookmark, error) {
	value := &netbox_goV1.ExtrasBookmark{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
