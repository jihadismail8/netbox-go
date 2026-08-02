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
	extrasTaggeditemCachePrefixKey = "extrasTaggeditem:"
	// ExtrasTaggeditemExpireTime expire time
	ExtrasTaggeditemExpireTime = 5 * time.Minute
)

var _ ExtrasTaggeditemCache = (*extrasTaggeditemCache)(nil)

// ExtrasTaggeditemCache cache interface
type ExtrasTaggeditemCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasTaggeditem, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasTaggeditem, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasTaggeditem, error)
	MultiSet(ctx context.Context, data []*model.ExtrasTaggeditem, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasTaggeditemCache define a cache struct
type extrasTaggeditemCache struct {
	cache cache.Cache
}

// NewExtrasTaggeditemCache new a cache
func NewExtrasTaggeditemCache(cacheType *database.CacheType) ExtrasTaggeditemCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasTaggeditem{}
		})
		return &extrasTaggeditemCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasTaggeditem{}
		})
		return &extrasTaggeditemCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasTaggeditemCacheKey cache key
func (c *extrasTaggeditemCache) GetExtrasTaggeditemCacheKey(id uint64) string {
	return extrasTaggeditemCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasTaggeditemCache) Set(ctx context.Context, id uint64, data *model.ExtrasTaggeditem, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasTaggeditemCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasTaggeditemCache) Get(ctx context.Context, id uint64) (*model.ExtrasTaggeditem, error) {
	var data *model.ExtrasTaggeditem
	cacheKey := c.GetExtrasTaggeditemCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasTaggeditemCache) MultiSet(ctx context.Context, data []*model.ExtrasTaggeditem, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasTaggeditemCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasTaggeditemCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasTaggeditem, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasTaggeditemCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasTaggeditem)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasTaggeditem)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasTaggeditemCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasTaggeditemCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasTaggeditemCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasTaggeditemCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasTaggeditemCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasTaggeditemCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
