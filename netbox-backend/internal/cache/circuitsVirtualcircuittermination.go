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
	circuitsVirtualcircuitterminationCachePrefixKey = "circuitsVirtualcircuittermination:"
	// CircuitsVirtualcircuitterminationExpireTime expire time
	CircuitsVirtualcircuitterminationExpireTime = 5 * time.Minute
)

var _ CircuitsVirtualcircuitterminationCache = (*circuitsVirtualcircuitterminationCache)(nil)

// CircuitsVirtualcircuitterminationCache cache interface
type CircuitsVirtualcircuitterminationCache interface {
	Set(ctx context.Context, id uint64, data *model.CircuitsVirtualcircuittermination, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CircuitsVirtualcircuittermination, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsVirtualcircuittermination, error)
	MultiSet(ctx context.Context, data []*model.CircuitsVirtualcircuittermination, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// circuitsVirtualcircuitterminationCache define a cache struct
type circuitsVirtualcircuitterminationCache struct {
	cache cache.Cache
}

// NewCircuitsVirtualcircuitterminationCache new a cache
func NewCircuitsVirtualcircuitterminationCache(cacheType *database.CacheType) CircuitsVirtualcircuitterminationCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsVirtualcircuittermination{}
		})
		return &circuitsVirtualcircuitterminationCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsVirtualcircuittermination{}
		})
		return &circuitsVirtualcircuitterminationCache{cache: c}
	}

	return nil // no cache
}

// GetCircuitsVirtualcircuitterminationCacheKey cache key
func (c *circuitsVirtualcircuitterminationCache) GetCircuitsVirtualcircuitterminationCacheKey(id uint64) string {
	return circuitsVirtualcircuitterminationCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *circuitsVirtualcircuitterminationCache) Set(ctx context.Context, id uint64, data *model.CircuitsVirtualcircuittermination, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCircuitsVirtualcircuitterminationCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *circuitsVirtualcircuitterminationCache) Get(ctx context.Context, id uint64) (*model.CircuitsVirtualcircuittermination, error) {
	var data *model.CircuitsVirtualcircuittermination
	cacheKey := c.GetCircuitsVirtualcircuitterminationCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *circuitsVirtualcircuitterminationCache) MultiSet(ctx context.Context, data []*model.CircuitsVirtualcircuittermination, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCircuitsVirtualcircuitterminationCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *circuitsVirtualcircuitterminationCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsVirtualcircuittermination, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCircuitsVirtualcircuitterminationCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CircuitsVirtualcircuittermination)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CircuitsVirtualcircuittermination)
	for _, id := range ids {
		val, ok := itemMap[c.GetCircuitsVirtualcircuitterminationCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *circuitsVirtualcircuitterminationCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsVirtualcircuitterminationCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *circuitsVirtualcircuitterminationCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsVirtualcircuitterminationCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *circuitsVirtualcircuitterminationCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
