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
	extrasConfigcontextClusterTypesCachePrefixKey = "extrasConfigcontextClusterTypes:"
	// ExtrasConfigcontextClusterTypesExpireTime expire time
	ExtrasConfigcontextClusterTypesExpireTime = 5 * time.Minute
)

var _ ExtrasConfigcontextClusterTypesCache = (*extrasConfigcontextClusterTypesCache)(nil)

// ExtrasConfigcontextClusterTypesCache cache interface
type ExtrasConfigcontextClusterTypesCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextClusterTypes, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextClusterTypes, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextClusterTypes, error)
	MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextClusterTypes, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasConfigcontextClusterTypesCache define a cache struct
type extrasConfigcontextClusterTypesCache struct {
	cache cache.Cache
}

// NewExtrasConfigcontextClusterTypesCache new a cache
func NewExtrasConfigcontextClusterTypesCache(cacheType *database.CacheType) ExtrasConfigcontextClusterTypesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextClusterTypes{}
		})
		return &extrasConfigcontextClusterTypesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextClusterTypes{}
		})
		return &extrasConfigcontextClusterTypesCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasConfigcontextClusterTypesCacheKey cache key
func (c *extrasConfigcontextClusterTypesCache) GetExtrasConfigcontextClusterTypesCacheKey(id uint64) string {
	return extrasConfigcontextClusterTypesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasConfigcontextClusterTypesCache) Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextClusterTypes, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasConfigcontextClusterTypesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasConfigcontextClusterTypesCache) Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextClusterTypes, error) {
	var data *model.ExtrasConfigcontextClusterTypes
	cacheKey := c.GetExtrasConfigcontextClusterTypesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasConfigcontextClusterTypesCache) MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextClusterTypes, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasConfigcontextClusterTypesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasConfigcontextClusterTypesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextClusterTypes, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasConfigcontextClusterTypesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasConfigcontextClusterTypes)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasConfigcontextClusterTypes)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasConfigcontextClusterTypesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasConfigcontextClusterTypesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextClusterTypesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasConfigcontextClusterTypesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextClusterTypesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasConfigcontextClusterTypesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
