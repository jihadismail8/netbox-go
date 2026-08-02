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
	dcimPlatformCachePrefixKey = "dcimPlatform:"
	// DcimPlatformExpireTime expire time
	DcimPlatformExpireTime = 5 * time.Minute
)

var _ DcimPlatformCache = (*dcimPlatformCache)(nil)

// DcimPlatformCache cache interface
type DcimPlatformCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimPlatform, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimPlatform, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimPlatform, error)
	MultiSet(ctx context.Context, data []*model.DcimPlatform, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimPlatformCache define a cache struct
type dcimPlatformCache struct {
	cache cache.Cache
}

// NewDcimPlatformCache new a cache
func NewDcimPlatformCache(cacheType *database.CacheType) DcimPlatformCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimPlatform{}
		})
		return &dcimPlatformCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimPlatform{}
		})
		return &dcimPlatformCache{cache: c}
	}

	return nil // no cache
}

// GetDcimPlatformCacheKey cache key
func (c *dcimPlatformCache) GetDcimPlatformCacheKey(id uint64) string {
	return dcimPlatformCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimPlatformCache) Set(ctx context.Context, id uint64, data *model.DcimPlatform, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimPlatformCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimPlatformCache) Get(ctx context.Context, id uint64) (*model.DcimPlatform, error) {
	var data *model.DcimPlatform
	cacheKey := c.GetDcimPlatformCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimPlatformCache) MultiSet(ctx context.Context, data []*model.DcimPlatform, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimPlatformCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimPlatformCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimPlatform, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimPlatformCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimPlatform)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimPlatform)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimPlatformCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimPlatformCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimPlatformCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimPlatformCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimPlatformCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimPlatformCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
