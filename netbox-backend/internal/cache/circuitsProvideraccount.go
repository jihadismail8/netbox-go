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
	circuitsProvideraccountCachePrefixKey = "circuitsProvideraccount:"
	// CircuitsProvideraccountExpireTime expire time
	CircuitsProvideraccountExpireTime = 5 * time.Minute
)

var _ CircuitsProvideraccountCache = (*circuitsProvideraccountCache)(nil)

// CircuitsProvideraccountCache cache interface
type CircuitsProvideraccountCache interface {
	Set(ctx context.Context, id uint64, data *model.CircuitsProvideraccount, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CircuitsProvideraccount, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsProvideraccount, error)
	MultiSet(ctx context.Context, data []*model.CircuitsProvideraccount, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// circuitsProvideraccountCache define a cache struct
type circuitsProvideraccountCache struct {
	cache cache.Cache
}

// NewCircuitsProvideraccountCache new a cache
func NewCircuitsProvideraccountCache(cacheType *database.CacheType) CircuitsProvideraccountCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsProvideraccount{}
		})
		return &circuitsProvideraccountCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CircuitsProvideraccount{}
		})
		return &circuitsProvideraccountCache{cache: c}
	}

	return nil // no cache
}

// GetCircuitsProvideraccountCacheKey cache key
func (c *circuitsProvideraccountCache) GetCircuitsProvideraccountCacheKey(id uint64) string {
	return circuitsProvideraccountCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *circuitsProvideraccountCache) Set(ctx context.Context, id uint64, data *model.CircuitsProvideraccount, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCircuitsProvideraccountCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *circuitsProvideraccountCache) Get(ctx context.Context, id uint64) (*model.CircuitsProvideraccount, error) {
	var data *model.CircuitsProvideraccount
	cacheKey := c.GetCircuitsProvideraccountCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *circuitsProvideraccountCache) MultiSet(ctx context.Context, data []*model.CircuitsProvideraccount, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCircuitsProvideraccountCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *circuitsProvideraccountCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsProvideraccount, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCircuitsProvideraccountCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CircuitsProvideraccount)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CircuitsProvideraccount)
	for _, id := range ids {
		val, ok := itemMap[c.GetCircuitsProvideraccountCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *circuitsProvideraccountCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsProvideraccountCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *circuitsProvideraccountCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCircuitsProvideraccountCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *circuitsProvideraccountCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
