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
	extrasConfigcontextClusterGroupsCachePrefixKey = "extrasConfigcontextClusterGroups:"
	// ExtrasConfigcontextClusterGroupsExpireTime expire time
	ExtrasConfigcontextClusterGroupsExpireTime = 5 * time.Minute
)

var _ ExtrasConfigcontextClusterGroupsCache = (*extrasConfigcontextClusterGroupsCache)(nil)

// ExtrasConfigcontextClusterGroupsCache cache interface
type ExtrasConfigcontextClusterGroupsCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextClusterGroups, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextClusterGroups, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextClusterGroups, error)
	MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextClusterGroups, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasConfigcontextClusterGroupsCache define a cache struct
type extrasConfigcontextClusterGroupsCache struct {
	cache cache.Cache
}

// NewExtrasConfigcontextClusterGroupsCache new a cache
func NewExtrasConfigcontextClusterGroupsCache(cacheType *database.CacheType) ExtrasConfigcontextClusterGroupsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextClusterGroups{}
		})
		return &extrasConfigcontextClusterGroupsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextClusterGroups{}
		})
		return &extrasConfigcontextClusterGroupsCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasConfigcontextClusterGroupsCacheKey cache key
func (c *extrasConfigcontextClusterGroupsCache) GetExtrasConfigcontextClusterGroupsCacheKey(id uint64) string {
	return extrasConfigcontextClusterGroupsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasConfigcontextClusterGroupsCache) Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextClusterGroups, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasConfigcontextClusterGroupsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasConfigcontextClusterGroupsCache) Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextClusterGroups, error) {
	var data *model.ExtrasConfigcontextClusterGroups
	cacheKey := c.GetExtrasConfigcontextClusterGroupsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasConfigcontextClusterGroupsCache) MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextClusterGroups, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasConfigcontextClusterGroupsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasConfigcontextClusterGroupsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextClusterGroups, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasConfigcontextClusterGroupsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasConfigcontextClusterGroups)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasConfigcontextClusterGroups)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasConfigcontextClusterGroupsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasConfigcontextClusterGroupsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextClusterGroupsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasConfigcontextClusterGroupsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextClusterGroupsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasConfigcontextClusterGroupsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
