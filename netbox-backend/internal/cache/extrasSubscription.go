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
	extrasSubscriptionCachePrefixKey = "extrasSubscription:"
	// ExtrasSubscriptionExpireTime expire time
	ExtrasSubscriptionExpireTime = 5 * time.Minute
)

var _ ExtrasSubscriptionCache = (*extrasSubscriptionCache)(nil)

// ExtrasSubscriptionCache cache interface
type ExtrasSubscriptionCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasSubscription, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasSubscription, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasSubscription, error)
	MultiSet(ctx context.Context, data []*model.ExtrasSubscription, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasSubscriptionCache define a cache struct
type extrasSubscriptionCache struct {
	cache cache.Cache
}

// NewExtrasSubscriptionCache new a cache
func NewExtrasSubscriptionCache(cacheType *database.CacheType) ExtrasSubscriptionCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasSubscription{}
		})
		return &extrasSubscriptionCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasSubscription{}
		})
		return &extrasSubscriptionCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasSubscriptionCacheKey cache key
func (c *extrasSubscriptionCache) GetExtrasSubscriptionCacheKey(id uint64) string {
	return extrasSubscriptionCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasSubscriptionCache) Set(ctx context.Context, id uint64, data *model.ExtrasSubscription, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasSubscriptionCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasSubscriptionCache) Get(ctx context.Context, id uint64) (*model.ExtrasSubscription, error) {
	var data *model.ExtrasSubscription
	cacheKey := c.GetExtrasSubscriptionCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasSubscriptionCache) MultiSet(ctx context.Context, data []*model.ExtrasSubscription, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasSubscriptionCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasSubscriptionCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasSubscription, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasSubscriptionCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasSubscription)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasSubscription)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasSubscriptionCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasSubscriptionCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasSubscriptionCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasSubscriptionCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasSubscriptionCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasSubscriptionCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
