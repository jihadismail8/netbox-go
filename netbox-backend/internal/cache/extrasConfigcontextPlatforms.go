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
	extrasConfigcontextPlatformsCachePrefixKey = "extrasConfigcontextPlatforms:"
	// ExtrasConfigcontextPlatformsExpireTime expire time
	ExtrasConfigcontextPlatformsExpireTime = 5 * time.Minute
)

var _ ExtrasConfigcontextPlatformsCache = (*extrasConfigcontextPlatformsCache)(nil)

// ExtrasConfigcontextPlatformsCache cache interface
type ExtrasConfigcontextPlatformsCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextPlatforms, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextPlatforms, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextPlatforms, error)
	MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextPlatforms, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasConfigcontextPlatformsCache define a cache struct
type extrasConfigcontextPlatformsCache struct {
	cache cache.Cache
}

// NewExtrasConfigcontextPlatformsCache new a cache
func NewExtrasConfigcontextPlatformsCache(cacheType *database.CacheType) ExtrasConfigcontextPlatformsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextPlatforms{}
		})
		return &extrasConfigcontextPlatformsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextPlatforms{}
		})
		return &extrasConfigcontextPlatformsCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasConfigcontextPlatformsCacheKey cache key
func (c *extrasConfigcontextPlatformsCache) GetExtrasConfigcontextPlatformsCacheKey(id uint64) string {
	return extrasConfigcontextPlatformsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasConfigcontextPlatformsCache) Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextPlatforms, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasConfigcontextPlatformsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasConfigcontextPlatformsCache) Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextPlatforms, error) {
	var data *model.ExtrasConfigcontextPlatforms
	cacheKey := c.GetExtrasConfigcontextPlatformsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasConfigcontextPlatformsCache) MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextPlatforms, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasConfigcontextPlatformsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasConfigcontextPlatformsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextPlatforms, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasConfigcontextPlatformsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasConfigcontextPlatforms)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasConfigcontextPlatforms)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasConfigcontextPlatformsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasConfigcontextPlatformsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextPlatformsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasConfigcontextPlatformsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextPlatformsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasConfigcontextPlatformsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
