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
	extrasConfigcontextTenantsCachePrefixKey = "extrasConfigcontextTenants:"
	// ExtrasConfigcontextTenantsExpireTime expire time
	ExtrasConfigcontextTenantsExpireTime = 5 * time.Minute
)

var _ ExtrasConfigcontextTenantsCache = (*extrasConfigcontextTenantsCache)(nil)

// ExtrasConfigcontextTenantsCache cache interface
type ExtrasConfigcontextTenantsCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextTenants, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextTenants, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextTenants, error)
	MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextTenants, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasConfigcontextTenantsCache define a cache struct
type extrasConfigcontextTenantsCache struct {
	cache cache.Cache
}

// NewExtrasConfigcontextTenantsCache new a cache
func NewExtrasConfigcontextTenantsCache(cacheType *database.CacheType) ExtrasConfigcontextTenantsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextTenants{}
		})
		return &extrasConfigcontextTenantsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextTenants{}
		})
		return &extrasConfigcontextTenantsCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasConfigcontextTenantsCacheKey cache key
func (c *extrasConfigcontextTenantsCache) GetExtrasConfigcontextTenantsCacheKey(id uint64) string {
	return extrasConfigcontextTenantsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasConfigcontextTenantsCache) Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextTenants, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasConfigcontextTenantsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasConfigcontextTenantsCache) Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextTenants, error) {
	var data *model.ExtrasConfigcontextTenants
	cacheKey := c.GetExtrasConfigcontextTenantsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasConfigcontextTenantsCache) MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextTenants, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasConfigcontextTenantsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasConfigcontextTenantsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextTenants, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasConfigcontextTenantsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasConfigcontextTenants)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasConfigcontextTenants)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasConfigcontextTenantsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasConfigcontextTenantsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextTenantsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasConfigcontextTenantsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextTenantsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasConfigcontextTenantsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
