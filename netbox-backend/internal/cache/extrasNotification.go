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
	extrasNotificationCachePrefixKey = "extrasNotification:"
	// ExtrasNotificationExpireTime expire time
	ExtrasNotificationExpireTime = 5 * time.Minute
)

var _ ExtrasNotificationCache = (*extrasNotificationCache)(nil)

// ExtrasNotificationCache cache interface
type ExtrasNotificationCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasNotification, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasNotification, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasNotification, error)
	MultiSet(ctx context.Context, data []*model.ExtrasNotification, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasNotificationCache define a cache struct
type extrasNotificationCache struct {
	cache cache.Cache
}

// NewExtrasNotificationCache new a cache
func NewExtrasNotificationCache(cacheType *database.CacheType) ExtrasNotificationCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasNotification{}
		})
		return &extrasNotificationCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasNotification{}
		})
		return &extrasNotificationCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasNotificationCacheKey cache key
func (c *extrasNotificationCache) GetExtrasNotificationCacheKey(id uint64) string {
	return extrasNotificationCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasNotificationCache) Set(ctx context.Context, id uint64, data *model.ExtrasNotification, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasNotificationCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasNotificationCache) Get(ctx context.Context, id uint64) (*model.ExtrasNotification, error) {
	var data *model.ExtrasNotification
	cacheKey := c.GetExtrasNotificationCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasNotificationCache) MultiSet(ctx context.Context, data []*model.ExtrasNotification, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasNotificationCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasNotificationCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasNotification, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasNotificationCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasNotification)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasNotification)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasNotificationCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasNotificationCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasNotificationCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasNotificationCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasNotificationCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasNotificationCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
