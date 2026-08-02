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
	circuitsProviderCachePrefixKey = "circuitsProvider:"
	// CircuitsProviderExpireTime expire time
	CircuitsProviderExpireTime = 5 * time.Minute
)

var _ CircuitsProviderCache = (*circuitsProviderCache)(nil)

// CircuitsProviderCache cache interface
type CircuitsProviderCache interface {
	Set(ctx context.Context, id uint64, data *model.CircuitsProvider, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CircuitsProvider, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsProvider, error)
	MultiSet(ctx context.Context, data []*model.CircuitsProvider, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// circuitsProviderCache define a cache struct
type circuitsProviderCache struct {
	cache cache.Cache
}

// NewCircuitsProviderCache new a cache
func NewCircuitsProviderCache(cacheType *database.CacheType) CircuitsProviderCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsProvider{}
		})
		return &circuitsProviderCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsProvider{}
		})
		return &circuitsProviderCache{cache: c}
	}

	return nil // no cache
}

// GetCircuitsProviderCacheKey cache key
func (c *circuitsProviderCache) GetCircuitsProviderCacheKey(id uint64) string {
	return circuitsProviderCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *circuitsProviderCache) Set(ctx context.Context, id uint64, data *model.CircuitsProvider, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCircuitsProviderCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *circuitsProviderCache) Get(ctx context.Context, id uint64) (*model.CircuitsProvider, error) {
	var data *model.CircuitsProvider
	cacheKey := c.GetCircuitsProviderCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *circuitsProviderCache) MultiSet(ctx context.Context, data []*model.CircuitsProvider, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCircuitsProviderCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *circuitsProviderCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsProvider, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCircuitsProviderCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CircuitsProvider)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CircuitsProvider)
	for _, id := range ids {
		val, ok := itemMap[c.GetCircuitsProviderCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *circuitsProviderCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsProviderCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *circuitsProviderCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsProviderCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *circuitsProviderCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
