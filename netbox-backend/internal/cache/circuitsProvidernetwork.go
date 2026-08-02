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
	circuitsProvidernetworkCachePrefixKey = "circuitsProvidernetwork:"
	// CircuitsProvidernetworkExpireTime expire time
	CircuitsProvidernetworkExpireTime = 5 * time.Minute
)

var _ CircuitsProvidernetworkCache = (*circuitsProvidernetworkCache)(nil)

// CircuitsProvidernetworkCache cache interface
type CircuitsProvidernetworkCache interface {
	Set(ctx context.Context, id uint64, data *model.CircuitsProvidernetwork, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CircuitsProvidernetwork, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsProvidernetwork, error)
	MultiSet(ctx context.Context, data []*model.CircuitsProvidernetwork, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// circuitsProvidernetworkCache define a cache struct
type circuitsProvidernetworkCache struct {
	cache cache.Cache
}

// NewCircuitsProvidernetworkCache new a cache
func NewCircuitsProvidernetworkCache(cacheType *database.CacheType) CircuitsProvidernetworkCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsProvidernetwork{}
		})
		return &circuitsProvidernetworkCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsProvidernetwork{}
		})
		return &circuitsProvidernetworkCache{cache: c}
	}

	return nil // no cache
}

// GetCircuitsProvidernetworkCacheKey cache key
func (c *circuitsProvidernetworkCache) GetCircuitsProvidernetworkCacheKey(id uint64) string {
	return circuitsProvidernetworkCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *circuitsProvidernetworkCache) Set(ctx context.Context, id uint64, data *model.CircuitsProvidernetwork, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCircuitsProvidernetworkCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *circuitsProvidernetworkCache) Get(ctx context.Context, id uint64) (*model.CircuitsProvidernetwork, error) {
	var data *model.CircuitsProvidernetwork
	cacheKey := c.GetCircuitsProvidernetworkCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *circuitsProvidernetworkCache) MultiSet(ctx context.Context, data []*model.CircuitsProvidernetwork, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCircuitsProvidernetworkCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *circuitsProvidernetworkCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsProvidernetwork, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCircuitsProvidernetworkCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CircuitsProvidernetwork)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CircuitsProvidernetwork)
	for _, id := range ids {
		val, ok := itemMap[c.GetCircuitsProvidernetworkCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *circuitsProvidernetworkCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsProvidernetworkCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *circuitsProvidernetworkCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsProvidernetworkCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *circuitsProvidernetworkCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
