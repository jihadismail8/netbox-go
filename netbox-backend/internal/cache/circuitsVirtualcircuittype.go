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
	circuitsVirtualcircuittypeCachePrefixKey = "circuitsVirtualcircuittype:"
	// CircuitsVirtualcircuittypeExpireTime expire time
	CircuitsVirtualcircuittypeExpireTime = 5 * time.Minute
)

var _ CircuitsVirtualcircuittypeCache = (*circuitsVirtualcircuittypeCache)(nil)

// CircuitsVirtualcircuittypeCache cache interface
type CircuitsVirtualcircuittypeCache interface {
	Set(ctx context.Context, id uint64, data *model.CircuitsVirtualcircuittype, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CircuitsVirtualcircuittype, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsVirtualcircuittype, error)
	MultiSet(ctx context.Context, data []*model.CircuitsVirtualcircuittype, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// circuitsVirtualcircuittypeCache define a cache struct
type circuitsVirtualcircuittypeCache struct {
	cache cache.Cache
}

// NewCircuitsVirtualcircuittypeCache new a cache
func NewCircuitsVirtualcircuittypeCache(cacheType *database.CacheType) CircuitsVirtualcircuittypeCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsVirtualcircuittype{}
		})
		return &circuitsVirtualcircuittypeCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsVirtualcircuittype{}
		})
		return &circuitsVirtualcircuittypeCache{cache: c}
	}

	return nil // no cache
}

// GetCircuitsVirtualcircuittypeCacheKey cache key
func (c *circuitsVirtualcircuittypeCache) GetCircuitsVirtualcircuittypeCacheKey(id uint64) string {
	return circuitsVirtualcircuittypeCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *circuitsVirtualcircuittypeCache) Set(ctx context.Context, id uint64, data *model.CircuitsVirtualcircuittype, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCircuitsVirtualcircuittypeCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *circuitsVirtualcircuittypeCache) Get(ctx context.Context, id uint64) (*model.CircuitsVirtualcircuittype, error) {
	var data *model.CircuitsVirtualcircuittype
	cacheKey := c.GetCircuitsVirtualcircuittypeCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *circuitsVirtualcircuittypeCache) MultiSet(ctx context.Context, data []*model.CircuitsVirtualcircuittype, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCircuitsVirtualcircuittypeCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *circuitsVirtualcircuittypeCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsVirtualcircuittype, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCircuitsVirtualcircuittypeCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CircuitsVirtualcircuittype)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CircuitsVirtualcircuittype)
	for _, id := range ids {
		val, ok := itemMap[c.GetCircuitsVirtualcircuittypeCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *circuitsVirtualcircuittypeCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsVirtualcircuittypeCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *circuitsVirtualcircuittypeCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsVirtualcircuittypeCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *circuitsVirtualcircuittypeCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
