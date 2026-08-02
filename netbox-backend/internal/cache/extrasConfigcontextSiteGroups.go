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
	extrasConfigcontextSiteGroupsCachePrefixKey = "extrasConfigcontextSiteGroups:"
	// ExtrasConfigcontextSiteGroupsExpireTime expire time
	ExtrasConfigcontextSiteGroupsExpireTime = 5 * time.Minute
)

var _ ExtrasConfigcontextSiteGroupsCache = (*extrasConfigcontextSiteGroupsCache)(nil)

// ExtrasConfigcontextSiteGroupsCache cache interface
type ExtrasConfigcontextSiteGroupsCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextSiteGroups, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextSiteGroups, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextSiteGroups, error)
	MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextSiteGroups, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasConfigcontextSiteGroupsCache define a cache struct
type extrasConfigcontextSiteGroupsCache struct {
	cache cache.Cache
}

// NewExtrasConfigcontextSiteGroupsCache new a cache
func NewExtrasConfigcontextSiteGroupsCache(cacheType *database.CacheType) ExtrasConfigcontextSiteGroupsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextSiteGroups{}
		})
		return &extrasConfigcontextSiteGroupsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextSiteGroups{}
		})
		return &extrasConfigcontextSiteGroupsCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasConfigcontextSiteGroupsCacheKey cache key
func (c *extrasConfigcontextSiteGroupsCache) GetExtrasConfigcontextSiteGroupsCacheKey(id uint64) string {
	return extrasConfigcontextSiteGroupsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasConfigcontextSiteGroupsCache) Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextSiteGroups, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasConfigcontextSiteGroupsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasConfigcontextSiteGroupsCache) Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextSiteGroups, error) {
	var data *model.ExtrasConfigcontextSiteGroups
	cacheKey := c.GetExtrasConfigcontextSiteGroupsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasConfigcontextSiteGroupsCache) MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextSiteGroups, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasConfigcontextSiteGroupsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasConfigcontextSiteGroupsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextSiteGroups, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasConfigcontextSiteGroupsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasConfigcontextSiteGroups)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasConfigcontextSiteGroups)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasConfigcontextSiteGroupsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasConfigcontextSiteGroupsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextSiteGroupsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasConfigcontextSiteGroupsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextSiteGroupsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasConfigcontextSiteGroupsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
