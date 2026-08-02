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
	circuitsProviderAsnsCachePrefixKey = "circuitsProviderAsns:"
	// CircuitsProviderAsnsExpireTime expire time
	CircuitsProviderAsnsExpireTime = 5 * time.Minute
)

var _ CircuitsProviderAsnsCache = (*circuitsProviderAsnsCache)(nil)

// CircuitsProviderAsnsCache cache interface
type CircuitsProviderAsnsCache interface {
	Set(ctx context.Context, id uint64, data *model.CircuitsProviderAsns, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CircuitsProviderAsns, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsProviderAsns, error)
	MultiSet(ctx context.Context, data []*model.CircuitsProviderAsns, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// circuitsProviderAsnsCache define a cache struct
type circuitsProviderAsnsCache struct {
	cache cache.Cache
}

// NewCircuitsProviderAsnsCache new a cache
func NewCircuitsProviderAsnsCache(cacheType *database.CacheType) CircuitsProviderAsnsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsProviderAsns{}
		})
		return &circuitsProviderAsnsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsProviderAsns{}
		})
		return &circuitsProviderAsnsCache{cache: c}
	}

	return nil // no cache
}

// GetCircuitsProviderAsnsCacheKey cache key
func (c *circuitsProviderAsnsCache) GetCircuitsProviderAsnsCacheKey(id uint64) string {
	return circuitsProviderAsnsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *circuitsProviderAsnsCache) Set(ctx context.Context, id uint64, data *model.CircuitsProviderAsns, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCircuitsProviderAsnsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *circuitsProviderAsnsCache) Get(ctx context.Context, id uint64) (*model.CircuitsProviderAsns, error) {
	var data *model.CircuitsProviderAsns
	cacheKey := c.GetCircuitsProviderAsnsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *circuitsProviderAsnsCache) MultiSet(ctx context.Context, data []*model.CircuitsProviderAsns, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCircuitsProviderAsnsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *circuitsProviderAsnsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsProviderAsns, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCircuitsProviderAsnsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CircuitsProviderAsns)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CircuitsProviderAsns)
	for _, id := range ids {
		val, ok := itemMap[c.GetCircuitsProviderAsnsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *circuitsProviderAsnsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsProviderAsnsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *circuitsProviderAsnsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsProviderAsnsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *circuitsProviderAsnsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
