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
	tenancyContactCachePrefixKey = "tenancyContact:"
	// TenancyContactExpireTime expire time
	TenancyContactExpireTime = 5 * time.Minute
)

var _ TenancyContactCache = (*tenancyContactCache)(nil)

// TenancyContactCache cache interface
type TenancyContactCache interface {
	Set(ctx context.Context, id uint64, data *model.TenancyContact, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.TenancyContact, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TenancyContact, error)
	MultiSet(ctx context.Context, data []*model.TenancyContact, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// tenancyContactCache define a cache struct
type tenancyContactCache struct {
	cache cache.Cache
}

// NewTenancyContactCache new a cache
func NewTenancyContactCache(cacheType *database.CacheType) TenancyContactCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.TenancyContact{}
		})
		return &tenancyContactCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.TenancyContact{}
		})
		return &tenancyContactCache{cache: c}
	}

	return nil // no cache
}

// GetTenancyContactCacheKey cache key
func (c *tenancyContactCache) GetTenancyContactCacheKey(id uint64) string {
	return tenancyContactCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *tenancyContactCache) Set(ctx context.Context, id uint64, data *model.TenancyContact, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetTenancyContactCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *tenancyContactCache) Get(ctx context.Context, id uint64) (*model.TenancyContact, error) {
	var data *model.TenancyContact
	cacheKey := c.GetTenancyContactCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *tenancyContactCache) MultiSet(ctx context.Context, data []*model.TenancyContact, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetTenancyContactCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *tenancyContactCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TenancyContact, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetTenancyContactCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.TenancyContact)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.TenancyContact)
	for _, id := range ids {
		val, ok := itemMap[c.GetTenancyContactCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *tenancyContactCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetTenancyContactCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *tenancyContactCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetTenancyContactCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *tenancyContactCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
