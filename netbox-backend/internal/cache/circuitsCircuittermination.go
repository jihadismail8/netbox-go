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
	circuitsCircuitterminationCachePrefixKey = "circuitsCircuittermination:"
	// CircuitsCircuitterminationExpireTime expire time
	CircuitsCircuitterminationExpireTime = 5 * time.Minute
)

var _ CircuitsCircuitterminationCache = (*circuitsCircuitterminationCache)(nil)

// CircuitsCircuitterminationCache cache interface
type CircuitsCircuitterminationCache interface {
	Set(ctx context.Context, id uint64, data *model.CircuitsCircuittermination, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CircuitsCircuittermination, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsCircuittermination, error)
	MultiSet(ctx context.Context, data []*model.CircuitsCircuittermination, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// circuitsCircuitterminationCache define a cache struct
type circuitsCircuitterminationCache struct {
	cache cache.Cache
}

// NewCircuitsCircuitterminationCache new a cache
func NewCircuitsCircuitterminationCache(cacheType *database.CacheType) CircuitsCircuitterminationCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsCircuittermination{}
		})
		return &circuitsCircuitterminationCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsCircuittermination{}
		})
		return &circuitsCircuitterminationCache{cache: c}
	}

	return nil // no cache
}

// GetCircuitsCircuitterminationCacheKey cache key
func (c *circuitsCircuitterminationCache) GetCircuitsCircuitterminationCacheKey(id uint64) string {
	return circuitsCircuitterminationCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *circuitsCircuitterminationCache) Set(ctx context.Context, id uint64, data *model.CircuitsCircuittermination, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCircuitsCircuitterminationCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *circuitsCircuitterminationCache) Get(ctx context.Context, id uint64) (*model.CircuitsCircuittermination, error) {
	var data *model.CircuitsCircuittermination
	cacheKey := c.GetCircuitsCircuitterminationCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *circuitsCircuitterminationCache) MultiSet(ctx context.Context, data []*model.CircuitsCircuittermination, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCircuitsCircuitterminationCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *circuitsCircuitterminationCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsCircuittermination, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCircuitsCircuitterminationCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CircuitsCircuittermination)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CircuitsCircuittermination)
	for _, id := range ids {
		val, ok := itemMap[c.GetCircuitsCircuitterminationCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *circuitsCircuitterminationCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsCircuitterminationCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *circuitsCircuitterminationCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsCircuitterminationCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *circuitsCircuitterminationCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
