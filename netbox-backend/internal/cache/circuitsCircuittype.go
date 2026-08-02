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
	circuitsCircuittypeCachePrefixKey = "circuitsCircuittype:"
	// CircuitsCircuittypeExpireTime expire time
	CircuitsCircuittypeExpireTime = 5 * time.Minute
)

var _ CircuitsCircuittypeCache = (*circuitsCircuittypeCache)(nil)

// CircuitsCircuittypeCache cache interface
type CircuitsCircuittypeCache interface {
	Set(ctx context.Context, id uint64, data *model.CircuitsCircuittype, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CircuitsCircuittype, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsCircuittype, error)
	MultiSet(ctx context.Context, data []*model.CircuitsCircuittype, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// circuitsCircuittypeCache define a cache struct
type circuitsCircuittypeCache struct {
	cache cache.Cache
}

// NewCircuitsCircuittypeCache new a cache
func NewCircuitsCircuittypeCache(cacheType *database.CacheType) CircuitsCircuittypeCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsCircuittype{}
		})
		return &circuitsCircuittypeCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsCircuittype{}
		})
		return &circuitsCircuittypeCache{cache: c}
	}

	return nil // no cache
}

// GetCircuitsCircuittypeCacheKey cache key
func (c *circuitsCircuittypeCache) GetCircuitsCircuittypeCacheKey(id uint64) string {
	return circuitsCircuittypeCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *circuitsCircuittypeCache) Set(ctx context.Context, id uint64, data *model.CircuitsCircuittype, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCircuitsCircuittypeCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *circuitsCircuittypeCache) Get(ctx context.Context, id uint64) (*model.CircuitsCircuittype, error) {
	var data *model.CircuitsCircuittype
	cacheKey := c.GetCircuitsCircuittypeCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *circuitsCircuittypeCache) MultiSet(ctx context.Context, data []*model.CircuitsCircuittype, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCircuitsCircuittypeCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *circuitsCircuittypeCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsCircuittype, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCircuitsCircuittypeCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CircuitsCircuittype)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CircuitsCircuittype)
	for _, id := range ids {
		val, ok := itemMap[c.GetCircuitsCircuittypeCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *circuitsCircuittypeCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsCircuittypeCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *circuitsCircuittypeCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsCircuittypeCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *circuitsCircuittypeCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
