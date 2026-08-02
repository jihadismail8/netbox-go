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
	circuitsCircuitgroupassignmentCachePrefixKey = "circuitsCircuitgroupassignment:"
	// CircuitsCircuitgroupassignmentExpireTime expire time
	CircuitsCircuitgroupassignmentExpireTime = 5 * time.Minute
)

var _ CircuitsCircuitgroupassignmentCache = (*circuitsCircuitgroupassignmentCache)(nil)

// CircuitsCircuitgroupassignmentCache cache interface
type CircuitsCircuitgroupassignmentCache interface {
	Set(ctx context.Context, id uint64, data *model.CircuitsCircuitgroupassignment, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CircuitsCircuitgroupassignment, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsCircuitgroupassignment, error)
	MultiSet(ctx context.Context, data []*model.CircuitsCircuitgroupassignment, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// circuitsCircuitgroupassignmentCache define a cache struct
type circuitsCircuitgroupassignmentCache struct {
	cache cache.Cache
}

// NewCircuitsCircuitgroupassignmentCache new a cache
func NewCircuitsCircuitgroupassignmentCache(cacheType *database.CacheType) CircuitsCircuitgroupassignmentCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsCircuitgroupassignment{}
		})
		return &circuitsCircuitgroupassignmentCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsCircuitgroupassignment{}
		})
		return &circuitsCircuitgroupassignmentCache{cache: c}
	}

	return nil // no cache
}

// GetCircuitsCircuitgroupassignmentCacheKey cache key
func (c *circuitsCircuitgroupassignmentCache) GetCircuitsCircuitgroupassignmentCacheKey(id uint64) string {
	return circuitsCircuitgroupassignmentCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *circuitsCircuitgroupassignmentCache) Set(ctx context.Context, id uint64, data *model.CircuitsCircuitgroupassignment, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCircuitsCircuitgroupassignmentCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *circuitsCircuitgroupassignmentCache) Get(ctx context.Context, id uint64) (*model.CircuitsCircuitgroupassignment, error) {
	var data *model.CircuitsCircuitgroupassignment
	cacheKey := c.GetCircuitsCircuitgroupassignmentCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *circuitsCircuitgroupassignmentCache) MultiSet(ctx context.Context, data []*model.CircuitsCircuitgroupassignment, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCircuitsCircuitgroupassignmentCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *circuitsCircuitgroupassignmentCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsCircuitgroupassignment, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCircuitsCircuitgroupassignmentCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CircuitsCircuitgroupassignment)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CircuitsCircuitgroupassignment)
	for _, id := range ids {
		val, ok := itemMap[c.GetCircuitsCircuitgroupassignmentCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *circuitsCircuitgroupassignmentCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsCircuitgroupassignmentCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *circuitsCircuitgroupassignmentCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsCircuitgroupassignmentCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *circuitsCircuitgroupassignmentCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
