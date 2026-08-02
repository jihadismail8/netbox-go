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
	extrasNotificationgroupCachePrefixKey = "extrasNotificationgroup:"
	// ExtrasNotificationgroupExpireTime expire time
	ExtrasNotificationgroupExpireTime = 5 * time.Minute
)

var _ ExtrasNotificationgroupCache = (*extrasNotificationgroupCache)(nil)

// ExtrasNotificationgroupCache cache interface
type ExtrasNotificationgroupCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasNotificationgroup, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasNotificationgroup, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasNotificationgroup, error)
	MultiSet(ctx context.Context, data []*model.ExtrasNotificationgroup, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasNotificationgroupCache define a cache struct
type extrasNotificationgroupCache struct {
	cache cache.Cache
}

// NewExtrasNotificationgroupCache new a cache
func NewExtrasNotificationgroupCache(cacheType *database.CacheType) ExtrasNotificationgroupCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasNotificationgroup{}
		})
		return &extrasNotificationgroupCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasNotificationgroup{}
		})
		return &extrasNotificationgroupCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasNotificationgroupCacheKey cache key
func (c *extrasNotificationgroupCache) GetExtrasNotificationgroupCacheKey(id uint64) string {
	return extrasNotificationgroupCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasNotificationgroupCache) Set(ctx context.Context, id uint64, data *model.ExtrasNotificationgroup, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasNotificationgroupCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasNotificationgroupCache) Get(ctx context.Context, id uint64) (*model.ExtrasNotificationgroup, error) {
	var data *model.ExtrasNotificationgroup
	cacheKey := c.GetExtrasNotificationgroupCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasNotificationgroupCache) MultiSet(ctx context.Context, data []*model.ExtrasNotificationgroup, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasNotificationgroupCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasNotificationgroupCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasNotificationgroup, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasNotificationgroupCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasNotificationgroup)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasNotificationgroup)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasNotificationgroupCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasNotificationgroupCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasNotificationgroupCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasNotificationgroupCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasNotificationgroupCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasNotificationgroupCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
