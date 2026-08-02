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
	extrasNotificationgroupUsersCachePrefixKey = "extrasNotificationgroupUsers:"
	// ExtrasNotificationgroupUsersExpireTime expire time
	ExtrasNotificationgroupUsersExpireTime = 5 * time.Minute
)

var _ ExtrasNotificationgroupUsersCache = (*extrasNotificationgroupUsersCache)(nil)

// ExtrasNotificationgroupUsersCache cache interface
type ExtrasNotificationgroupUsersCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasNotificationgroupUsers, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasNotificationgroupUsers, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasNotificationgroupUsers, error)
	MultiSet(ctx context.Context, data []*model.ExtrasNotificationgroupUsers, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasNotificationgroupUsersCache define a cache struct
type extrasNotificationgroupUsersCache struct {
	cache cache.Cache
}

// NewExtrasNotificationgroupUsersCache new a cache
func NewExtrasNotificationgroupUsersCache(cacheType *database.CacheType) ExtrasNotificationgroupUsersCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasNotificationgroupUsers{}
		})
		return &extrasNotificationgroupUsersCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasNotificationgroupUsers{}
		})
		return &extrasNotificationgroupUsersCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasNotificationgroupUsersCacheKey cache key
func (c *extrasNotificationgroupUsersCache) GetExtrasNotificationgroupUsersCacheKey(id uint64) string {
	return extrasNotificationgroupUsersCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasNotificationgroupUsersCache) Set(ctx context.Context, id uint64, data *model.ExtrasNotificationgroupUsers, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasNotificationgroupUsersCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasNotificationgroupUsersCache) Get(ctx context.Context, id uint64) (*model.ExtrasNotificationgroupUsers, error) {
	var data *model.ExtrasNotificationgroupUsers
	cacheKey := c.GetExtrasNotificationgroupUsersCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasNotificationgroupUsersCache) MultiSet(ctx context.Context, data []*model.ExtrasNotificationgroupUsers, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasNotificationgroupUsersCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasNotificationgroupUsersCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasNotificationgroupUsers, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasNotificationgroupUsersCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasNotificationgroupUsers)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasNotificationgroupUsers)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasNotificationgroupUsersCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasNotificationgroupUsersCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasNotificationgroupUsersCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasNotificationgroupUsersCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasNotificationgroupUsersCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasNotificationgroupUsersCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
