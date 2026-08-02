package cache

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-dev-frame/sponge/pkg/cache"
	"github.com/go-dev-frame/sponge/pkg/encoding"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"netbox-go/internal/database"
	"netbox-go/internal/model"
)

const (
	// cache prefix key, must end with a colon
	coreDatasourceCachePrefixKey = "coreDatasource:"
	// CoreDatasourceExpireTime expire time
	CoreDatasourceExpireTime = 5 * time.Minute
)

var _ CoreDatasourceCache = (*coreDatasourceCache)(nil)

// CoreDatasourceCache cache interface
type CoreDatasourceCache interface {
	Set(ctx context.Context, id uint64, data *model.CoreDatasource, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CoreDatasource, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CoreDatasource, error)
	MultiSet(ctx context.Context, data []*model.CoreDatasource, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// coreDatasourceCache define a cache struct
type coreDatasourceCache struct {
	cache cache.Cache
}

// NewCoreDatasourceCache new a cache
func NewCoreDatasourceCache(cacheType *database.CacheType) CoreDatasourceCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreDatasource{}
		})
		return &coreDatasourceCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreDatasource{}
		})
		return &coreDatasourceCache{cache: c}
	}

	return nil // no cache
}

// GetCoreDatasourceCacheKey cache key
func (c *coreDatasourceCache) GetCoreDatasourceCacheKey(id uint64) string {
	return coreDatasourceCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *coreDatasourceCache) Set(ctx context.Context, id uint64, data *model.CoreDatasource, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCoreDatasourceCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *coreDatasourceCache) Get(ctx context.Context, id uint64) (*model.CoreDatasource, error) {
	var data *model.CoreDatasource
	cacheKey := c.GetCoreDatasourceCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *coreDatasourceCache) MultiSet(ctx context.Context, data []*model.CoreDatasource, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCoreDatasourceCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *coreDatasourceCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CoreDatasource, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCoreDatasourceCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CoreDatasource)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CoreDatasource)
	for _, id := range ids {
		val, ok := itemMap[c.GetCoreDatasourceCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *coreDatasourceCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCoreDatasourceCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *coreDatasourceCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCoreDatasourceCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *coreDatasourceCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
