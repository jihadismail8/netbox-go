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
	extrasConfigcontextCachePrefixKey = "extrasConfigcontext:"
	// ExtrasConfigcontextExpireTime expire time
	ExtrasConfigcontextExpireTime = 5 * time.Minute
)

var _ ExtrasConfigcontextCache = (*extrasConfigcontextCache)(nil)

// ExtrasConfigcontextCache cache interface
type ExtrasConfigcontextCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontext, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontext, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontext, error)
	MultiSet(ctx context.Context, data []*model.ExtrasConfigcontext, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasConfigcontextCache define a cache struct
type extrasConfigcontextCache struct {
	cache cache.Cache
}

// NewExtrasConfigcontextCache new a cache
func NewExtrasConfigcontextCache(cacheType *database.CacheType) ExtrasConfigcontextCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontext{}
		})
		return &extrasConfigcontextCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontext{}
		})
		return &extrasConfigcontextCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasConfigcontextCacheKey cache key
func (c *extrasConfigcontextCache) GetExtrasConfigcontextCacheKey(id uint64) string {
	return extrasConfigcontextCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasConfigcontextCache) Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontext, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasConfigcontextCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasConfigcontextCache) Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontext, error) {
	var data *model.ExtrasConfigcontext
	cacheKey := c.GetExtrasConfigcontextCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasConfigcontextCache) MultiSet(ctx context.Context, data []*model.ExtrasConfigcontext, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasConfigcontextCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasConfigcontextCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontext, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasConfigcontextCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasConfigcontext)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasConfigcontext)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasConfigcontextCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasConfigcontextCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasConfigcontextCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasConfigcontextCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
