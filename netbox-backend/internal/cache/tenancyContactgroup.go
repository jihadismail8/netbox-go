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
	tenancyContactgroupCachePrefixKey = "tenancyContactgroup:"
	// TenancyContactgroupExpireTime expire time
	TenancyContactgroupExpireTime = 5 * time.Minute
)

var _ TenancyContactgroupCache = (*tenancyContactgroupCache)(nil)

// TenancyContactgroupCache cache interface
type TenancyContactgroupCache interface {
	Set(ctx context.Context, id uint64, data *model.TenancyContactgroup, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.TenancyContactgroup, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TenancyContactgroup, error)
	MultiSet(ctx context.Context, data []*model.TenancyContactgroup, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// tenancyContactgroupCache define a cache struct
type tenancyContactgroupCache struct {
	cache cache.Cache
}

// NewTenancyContactgroupCache new a cache
func NewTenancyContactgroupCache(cacheType *database.CacheType) TenancyContactgroupCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.TenancyContactgroup{}
		})
		return &tenancyContactgroupCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.TenancyContactgroup{}
		})
		return &tenancyContactgroupCache{cache: c}
	}

	return nil // no cache
}

// GetTenancyContactgroupCacheKey cache key
func (c *tenancyContactgroupCache) GetTenancyContactgroupCacheKey(id uint64) string {
	return tenancyContactgroupCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *tenancyContactgroupCache) Set(ctx context.Context, id uint64, data *model.TenancyContactgroup, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetTenancyContactgroupCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *tenancyContactgroupCache) Get(ctx context.Context, id uint64) (*model.TenancyContactgroup, error) {
	var data *model.TenancyContactgroup
	cacheKey := c.GetTenancyContactgroupCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *tenancyContactgroupCache) MultiSet(ctx context.Context, data []*model.TenancyContactgroup, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetTenancyContactgroupCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *tenancyContactgroupCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TenancyContactgroup, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetTenancyContactgroupCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.TenancyContactgroup)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.TenancyContactgroup)
	for _, id := range ids {
		val, ok := itemMap[c.GetTenancyContactgroupCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *tenancyContactgroupCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetTenancyContactgroupCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *tenancyContactgroupCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetTenancyContactgroupCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *tenancyContactgroupCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
