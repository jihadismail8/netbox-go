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
	coreObjectchangeCachePrefixKey = "coreObjectchange:"
	// CoreObjectchangeExpireTime expire time
	CoreObjectchangeExpireTime = 5 * time.Minute
)

var _ CoreObjectchangeCache = (*coreObjectchangeCache)(nil)

// CoreObjectchangeCache cache interface
type CoreObjectchangeCache interface {
	Set(ctx context.Context, id uint64, data *model.CoreObjectchange, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CoreObjectchange, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CoreObjectchange, error)
	MultiSet(ctx context.Context, data []*model.CoreObjectchange, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// coreObjectchangeCache define a cache struct
type coreObjectchangeCache struct {
	cache cache.Cache
}

// NewCoreObjectchangeCache new a cache
func NewCoreObjectchangeCache(cacheType *database.CacheType) CoreObjectchangeCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreObjectchange{}
		})
		return &coreObjectchangeCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreObjectchange{}
		})
		return &coreObjectchangeCache{cache: c}
	}

	return nil // no cache
}

// GetCoreObjectchangeCacheKey cache key
func (c *coreObjectchangeCache) GetCoreObjectchangeCacheKey(id uint64) string {
	return coreObjectchangeCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *coreObjectchangeCache) Set(ctx context.Context, id uint64, data *model.CoreObjectchange, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCoreObjectchangeCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *coreObjectchangeCache) Get(ctx context.Context, id uint64) (*model.CoreObjectchange, error) {
	var data *model.CoreObjectchange
	cacheKey := c.GetCoreObjectchangeCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *coreObjectchangeCache) MultiSet(ctx context.Context, data []*model.CoreObjectchange, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCoreObjectchangeCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *coreObjectchangeCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CoreObjectchange, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCoreObjectchangeCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CoreObjectchange)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CoreObjectchange)
	for _, id := range ids {
		val, ok := itemMap[c.GetCoreObjectchangeCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *coreObjectchangeCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCoreObjectchangeCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *coreObjectchangeCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCoreObjectchangeCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *coreObjectchangeCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
