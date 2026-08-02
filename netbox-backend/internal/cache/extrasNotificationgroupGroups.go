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
	extrasNotificationgroupGroupsCachePrefixKey = "extrasNotificationgroupGroups:"
	// ExtrasNotificationgroupGroupsExpireTime expire time
	ExtrasNotificationgroupGroupsExpireTime = 5 * time.Minute
)

var _ ExtrasNotificationgroupGroupsCache = (*extrasNotificationgroupGroupsCache)(nil)

// ExtrasNotificationgroupGroupsCache cache interface
type ExtrasNotificationgroupGroupsCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasNotificationgroupGroups, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasNotificationgroupGroups, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasNotificationgroupGroups, error)
	MultiSet(ctx context.Context, data []*model.ExtrasNotificationgroupGroups, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasNotificationgroupGroupsCache define a cache struct
type extrasNotificationgroupGroupsCache struct {
	cache cache.Cache
}

// NewExtrasNotificationgroupGroupsCache new a cache
func NewExtrasNotificationgroupGroupsCache(cacheType *database.CacheType) ExtrasNotificationgroupGroupsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasNotificationgroupGroups{}
		})
		return &extrasNotificationgroupGroupsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasNotificationgroupGroups{}
		})
		return &extrasNotificationgroupGroupsCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasNotificationgroupGroupsCacheKey cache key
func (c *extrasNotificationgroupGroupsCache) GetExtrasNotificationgroupGroupsCacheKey(id uint64) string {
	return extrasNotificationgroupGroupsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasNotificationgroupGroupsCache) Set(ctx context.Context, id uint64, data *model.ExtrasNotificationgroupGroups, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasNotificationgroupGroupsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasNotificationgroupGroupsCache) Get(ctx context.Context, id uint64) (*model.ExtrasNotificationgroupGroups, error) {
	var data *model.ExtrasNotificationgroupGroups
	cacheKey := c.GetExtrasNotificationgroupGroupsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasNotificationgroupGroupsCache) MultiSet(ctx context.Context, data []*model.ExtrasNotificationgroupGroups, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasNotificationgroupGroupsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasNotificationgroupGroupsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasNotificationgroupGroups, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasNotificationgroupGroupsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasNotificationgroupGroups)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasNotificationgroupGroups)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasNotificationgroupGroupsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasNotificationgroupGroupsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasNotificationgroupGroupsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasNotificationgroupGroupsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasNotificationgroupGroupsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasNotificationgroupGroupsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
