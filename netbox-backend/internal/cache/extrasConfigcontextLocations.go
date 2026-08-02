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
	extrasConfigcontextLocationsCachePrefixKey = "extrasConfigcontextLocations:"
	// ExtrasConfigcontextLocationsExpireTime expire time
	ExtrasConfigcontextLocationsExpireTime = 5 * time.Minute
)

var _ ExtrasConfigcontextLocationsCache = (*extrasConfigcontextLocationsCache)(nil)

// ExtrasConfigcontextLocationsCache cache interface
type ExtrasConfigcontextLocationsCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextLocations, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextLocations, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextLocations, error)
	MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextLocations, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasConfigcontextLocationsCache define a cache struct
type extrasConfigcontextLocationsCache struct {
	cache cache.Cache
}

// NewExtrasConfigcontextLocationsCache new a cache
func NewExtrasConfigcontextLocationsCache(cacheType *database.CacheType) ExtrasConfigcontextLocationsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextLocations{}
		})
		return &extrasConfigcontextLocationsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextLocations{}
		})
		return &extrasConfigcontextLocationsCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasConfigcontextLocationsCacheKey cache key
func (c *extrasConfigcontextLocationsCache) GetExtrasConfigcontextLocationsCacheKey(id uint64) string {
	return extrasConfigcontextLocationsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasConfigcontextLocationsCache) Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextLocations, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasConfigcontextLocationsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasConfigcontextLocationsCache) Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextLocations, error) {
	var data *model.ExtrasConfigcontextLocations
	cacheKey := c.GetExtrasConfigcontextLocationsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasConfigcontextLocationsCache) MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextLocations, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasConfigcontextLocationsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasConfigcontextLocationsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextLocations, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasConfigcontextLocationsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasConfigcontextLocations)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasConfigcontextLocations)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasConfigcontextLocationsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasConfigcontextLocationsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextLocationsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasConfigcontextLocationsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextLocationsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasConfigcontextLocationsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
