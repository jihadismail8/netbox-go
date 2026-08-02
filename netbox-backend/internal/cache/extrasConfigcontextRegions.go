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
	extrasConfigcontextRegionsCachePrefixKey = "extrasConfigcontextRegions:"
	// ExtrasConfigcontextRegionsExpireTime expire time
	ExtrasConfigcontextRegionsExpireTime = 5 * time.Minute
)

var _ ExtrasConfigcontextRegionsCache = (*extrasConfigcontextRegionsCache)(nil)

// ExtrasConfigcontextRegionsCache cache interface
type ExtrasConfigcontextRegionsCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextRegions, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextRegions, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextRegions, error)
	MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextRegions, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasConfigcontextRegionsCache define a cache struct
type extrasConfigcontextRegionsCache struct {
	cache cache.Cache
}

// NewExtrasConfigcontextRegionsCache new a cache
func NewExtrasConfigcontextRegionsCache(cacheType *database.CacheType) ExtrasConfigcontextRegionsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextRegions{}
		})
		return &extrasConfigcontextRegionsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextRegions{}
		})
		return &extrasConfigcontextRegionsCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasConfigcontextRegionsCacheKey cache key
func (c *extrasConfigcontextRegionsCache) GetExtrasConfigcontextRegionsCacheKey(id uint64) string {
	return extrasConfigcontextRegionsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasConfigcontextRegionsCache) Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextRegions, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasConfigcontextRegionsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasConfigcontextRegionsCache) Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextRegions, error) {
	var data *model.ExtrasConfigcontextRegions
	cacheKey := c.GetExtrasConfigcontextRegionsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasConfigcontextRegionsCache) MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextRegions, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasConfigcontextRegionsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasConfigcontextRegionsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextRegions, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasConfigcontextRegionsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasConfigcontextRegions)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasConfigcontextRegions)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasConfigcontextRegionsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasConfigcontextRegionsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextRegionsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasConfigcontextRegionsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextRegionsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasConfigcontextRegionsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
