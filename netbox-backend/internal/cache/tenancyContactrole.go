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
	tenancyContactroleCachePrefixKey = "tenancyContactrole:"
	// TenancyContactroleExpireTime expire time
	TenancyContactroleExpireTime = 5 * time.Minute
)

var _ TenancyContactroleCache = (*tenancyContactroleCache)(nil)

// TenancyContactroleCache cache interface
type TenancyContactroleCache interface {
	Set(ctx context.Context, id uint64, data *model.TenancyContactrole, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.TenancyContactrole, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TenancyContactrole, error)
	MultiSet(ctx context.Context, data []*model.TenancyContactrole, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// tenancyContactroleCache define a cache struct
type tenancyContactroleCache struct {
	cache cache.Cache
}

// NewTenancyContactroleCache new a cache
func NewTenancyContactroleCache(cacheType *database.CacheType) TenancyContactroleCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.TenancyContactrole{}
		})
		return &tenancyContactroleCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.TenancyContactrole{}
		})
		return &tenancyContactroleCache{cache: c}
	}

	return nil // no cache
}

// GetTenancyContactroleCacheKey cache key
func (c *tenancyContactroleCache) GetTenancyContactroleCacheKey(id uint64) string {
	return tenancyContactroleCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *tenancyContactroleCache) Set(ctx context.Context, id uint64, data *model.TenancyContactrole, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetTenancyContactroleCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *tenancyContactroleCache) Get(ctx context.Context, id uint64) (*model.TenancyContactrole, error) {
	var data *model.TenancyContactrole
	cacheKey := c.GetTenancyContactroleCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *tenancyContactroleCache) MultiSet(ctx context.Context, data []*model.TenancyContactrole, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetTenancyContactroleCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *tenancyContactroleCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TenancyContactrole, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetTenancyContactroleCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.TenancyContactrole)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.TenancyContactrole)
	for _, id := range ids {
		val, ok := itemMap[c.GetTenancyContactroleCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *tenancyContactroleCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetTenancyContactroleCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *tenancyContactroleCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetTenancyContactroleCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *tenancyContactroleCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
