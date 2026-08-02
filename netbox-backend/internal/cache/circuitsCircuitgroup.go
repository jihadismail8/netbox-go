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
	circuitsCircuitgroupCachePrefixKey = "circuitsCircuitgroup:"
	// CircuitsCircuitgroupExpireTime expire time
	CircuitsCircuitgroupExpireTime = 5 * time.Minute
)

var _ CircuitsCircuitgroupCache = (*circuitsCircuitgroupCache)(nil)

// CircuitsCircuitgroupCache cache interface
type CircuitsCircuitgroupCache interface {
	Set(ctx context.Context, id uint64, data *model.CircuitsCircuitgroup, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CircuitsCircuitgroup, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsCircuitgroup, error)
	MultiSet(ctx context.Context, data []*model.CircuitsCircuitgroup, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// circuitsCircuitgroupCache define a cache struct
type circuitsCircuitgroupCache struct {
	cache cache.Cache
}

// NewCircuitsCircuitgroupCache new a cache
func NewCircuitsCircuitgroupCache(cacheType *database.CacheType) CircuitsCircuitgroupCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsCircuitgroup{}
		})
		return &circuitsCircuitgroupCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsCircuitgroup{}
		})
		return &circuitsCircuitgroupCache{cache: c}
	}

	return nil // no cache
}

// GetCircuitsCircuitgroupCacheKey cache key
func (c *circuitsCircuitgroupCache) GetCircuitsCircuitgroupCacheKey(id uint64) string {
	return circuitsCircuitgroupCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *circuitsCircuitgroupCache) Set(ctx context.Context, id uint64, data *model.CircuitsCircuitgroup, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCircuitsCircuitgroupCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *circuitsCircuitgroupCache) Get(ctx context.Context, id uint64) (*model.CircuitsCircuitgroup, error) {
	var data *model.CircuitsCircuitgroup
	cacheKey := c.GetCircuitsCircuitgroupCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *circuitsCircuitgroupCache) MultiSet(ctx context.Context, data []*model.CircuitsCircuitgroup, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCircuitsCircuitgroupCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *circuitsCircuitgroupCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsCircuitgroup, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCircuitsCircuitgroupCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CircuitsCircuitgroup)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CircuitsCircuitgroup)
	for _, id := range ids {
		val, ok := itemMap[c.GetCircuitsCircuitgroupCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *circuitsCircuitgroupCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsCircuitgroupCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *circuitsCircuitgroupCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsCircuitgroupCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *circuitsCircuitgroupCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
