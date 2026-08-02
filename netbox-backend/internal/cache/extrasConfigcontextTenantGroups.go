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
	extrasConfigcontextTenantGroupsCachePrefixKey = "extrasConfigcontextTenantGroups:"
	// ExtrasConfigcontextTenantGroupsExpireTime expire time
	ExtrasConfigcontextTenantGroupsExpireTime = 5 * time.Minute
)

var _ ExtrasConfigcontextTenantGroupsCache = (*extrasConfigcontextTenantGroupsCache)(nil)

// ExtrasConfigcontextTenantGroupsCache cache interface
type ExtrasConfigcontextTenantGroupsCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextTenantGroups, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextTenantGroups, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextTenantGroups, error)
	MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextTenantGroups, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasConfigcontextTenantGroupsCache define a cache struct
type extrasConfigcontextTenantGroupsCache struct {
	cache cache.Cache
}

// NewExtrasConfigcontextTenantGroupsCache new a cache
func NewExtrasConfigcontextTenantGroupsCache(cacheType *database.CacheType) ExtrasConfigcontextTenantGroupsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextTenantGroups{}
		})
		return &extrasConfigcontextTenantGroupsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextTenantGroups{}
		})
		return &extrasConfigcontextTenantGroupsCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasConfigcontextTenantGroupsCacheKey cache key
func (c *extrasConfigcontextTenantGroupsCache) GetExtrasConfigcontextTenantGroupsCacheKey(id uint64) string {
	return extrasConfigcontextTenantGroupsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasConfigcontextTenantGroupsCache) Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextTenantGroups, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasConfigcontextTenantGroupsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasConfigcontextTenantGroupsCache) Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextTenantGroups, error) {
	var data *model.ExtrasConfigcontextTenantGroups
	cacheKey := c.GetExtrasConfigcontextTenantGroupsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasConfigcontextTenantGroupsCache) MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextTenantGroups, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasConfigcontextTenantGroupsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasConfigcontextTenantGroupsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextTenantGroups, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasConfigcontextTenantGroupsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasConfigcontextTenantGroups)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasConfigcontextTenantGroups)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasConfigcontextTenantGroupsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasConfigcontextTenantGroupsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextTenantGroupsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasConfigcontextTenantGroupsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextTenantGroupsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasConfigcontextTenantGroupsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
