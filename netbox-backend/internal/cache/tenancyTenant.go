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
	tenancyTenantCachePrefixKey = "tenancyTenant:"
	// TenancyTenantExpireTime expire time
	TenancyTenantExpireTime = 5 * time.Minute
)

var _ TenancyTenantCache = (*tenancyTenantCache)(nil)

// TenancyTenantCache cache interface
type TenancyTenantCache interface {
	Set(ctx context.Context, id uint64, data *model.TenancyTenant, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.TenancyTenant, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TenancyTenant, error)
	MultiSet(ctx context.Context, data []*model.TenancyTenant, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// tenancyTenantCache define a cache struct
type tenancyTenantCache struct {
	cache cache.Cache
}

// NewTenancyTenantCache new a cache
func NewTenancyTenantCache(cacheType *database.CacheType) TenancyTenantCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.TenancyTenant{}
		})
		return &tenancyTenantCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.TenancyTenant{}
		})
		return &tenancyTenantCache{cache: c}
	}

	return nil // no cache
}

// GetTenancyTenantCacheKey cache key
func (c *tenancyTenantCache) GetTenancyTenantCacheKey(id uint64) string {
	return tenancyTenantCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *tenancyTenantCache) Set(ctx context.Context, id uint64, data *model.TenancyTenant, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetTenancyTenantCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *tenancyTenantCache) Get(ctx context.Context, id uint64) (*model.TenancyTenant, error) {
	var data *model.TenancyTenant
	cacheKey := c.GetTenancyTenantCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *tenancyTenantCache) MultiSet(ctx context.Context, data []*model.TenancyTenant, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetTenancyTenantCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *tenancyTenantCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TenancyTenant, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetTenancyTenantCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.TenancyTenant)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.TenancyTenant)
	for _, id := range ids {
		val, ok := itemMap[c.GetTenancyTenantCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *tenancyTenantCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetTenancyTenantCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *tenancyTenantCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetTenancyTenantCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *tenancyTenantCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
