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
	circuitsVirtualcircuitCachePrefixKey = "circuitsVirtualcircuit:"
	// CircuitsVirtualcircuitExpireTime expire time
	CircuitsVirtualcircuitExpireTime = 5 * time.Minute
)

var _ CircuitsVirtualcircuitCache = (*circuitsVirtualcircuitCache)(nil)

// CircuitsVirtualcircuitCache cache interface
type CircuitsVirtualcircuitCache interface {
	Set(ctx context.Context, id uint64, data *model.CircuitsVirtualcircuit, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CircuitsVirtualcircuit, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsVirtualcircuit, error)
	MultiSet(ctx context.Context, data []*model.CircuitsVirtualcircuit, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// circuitsVirtualcircuitCache define a cache struct
type circuitsVirtualcircuitCache struct {
	cache cache.Cache
}

// NewCircuitsVirtualcircuitCache new a cache
func NewCircuitsVirtualcircuitCache(cacheType *database.CacheType) CircuitsVirtualcircuitCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsVirtualcircuit{}
		})
		return &circuitsVirtualcircuitCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsVirtualcircuit{}
		})
		return &circuitsVirtualcircuitCache{cache: c}
	}

	return nil // no cache
}

// GetCircuitsVirtualcircuitCacheKey cache key
func (c *circuitsVirtualcircuitCache) GetCircuitsVirtualcircuitCacheKey(id uint64) string {
	return circuitsVirtualcircuitCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *circuitsVirtualcircuitCache) Set(ctx context.Context, id uint64, data *model.CircuitsVirtualcircuit, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCircuitsVirtualcircuitCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *circuitsVirtualcircuitCache) Get(ctx context.Context, id uint64) (*model.CircuitsVirtualcircuit, error) {
	var data *model.CircuitsVirtualcircuit
	cacheKey := c.GetCircuitsVirtualcircuitCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *circuitsVirtualcircuitCache) MultiSet(ctx context.Context, data []*model.CircuitsVirtualcircuit, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCircuitsVirtualcircuitCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *circuitsVirtualcircuitCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsVirtualcircuit, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCircuitsVirtualcircuitCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CircuitsVirtualcircuit)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CircuitsVirtualcircuit)
	for _, id := range ids {
		val, ok := itemMap[c.GetCircuitsVirtualcircuitCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *circuitsVirtualcircuitCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsVirtualcircuitCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *circuitsVirtualcircuitCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsVirtualcircuitCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *circuitsVirtualcircuitCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
