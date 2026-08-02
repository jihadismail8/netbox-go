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
	tenancyTenantgroupCachePrefixKey = "tenancyTenantgroup:"
	// TenancyTenantgroupExpireTime expire time
	TenancyTenantgroupExpireTime = 5 * time.Minute
)

var _ TenancyTenantgroupCache = (*tenancyTenantgroupCache)(nil)

// TenancyTenantgroupCache cache interface
type TenancyTenantgroupCache interface {
	Set(ctx context.Context, id uint64, data *model.TenancyTenantgroup, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.TenancyTenantgroup, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TenancyTenantgroup, error)
	MultiSet(ctx context.Context, data []*model.TenancyTenantgroup, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// tenancyTenantgroupCache define a cache struct
type tenancyTenantgroupCache struct {
	cache cache.Cache
}

// NewTenancyTenantgroupCache new a cache
func NewTenancyTenantgroupCache(cacheType *database.CacheType) TenancyTenantgroupCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.TenancyTenantgroup{}
		})
		return &tenancyTenantgroupCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.TenancyTenantgroup{}
		})
		return &tenancyTenantgroupCache{cache: c}
	}

	return nil // no cache
}

// GetTenancyTenantgroupCacheKey cache key
func (c *tenancyTenantgroupCache) GetTenancyTenantgroupCacheKey(id uint64) string {
	return tenancyTenantgroupCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *tenancyTenantgroupCache) Set(ctx context.Context, id uint64, data *model.TenancyTenantgroup, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetTenancyTenantgroupCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *tenancyTenantgroupCache) Get(ctx context.Context, id uint64) (*model.TenancyTenantgroup, error) {
	var data *model.TenancyTenantgroup
	cacheKey := c.GetTenancyTenantgroupCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *tenancyTenantgroupCache) MultiSet(ctx context.Context, data []*model.TenancyTenantgroup, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetTenancyTenantgroupCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *tenancyTenantgroupCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TenancyTenantgroup, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetTenancyTenantgroupCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.TenancyTenantgroup)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.TenancyTenantgroup)
	for _, id := range ids {
		val, ok := itemMap[c.GetTenancyTenantgroupCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *tenancyTenantgroupCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetTenancyTenantgroupCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *tenancyTenantgroupCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetTenancyTenantgroupCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *tenancyTenantgroupCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
