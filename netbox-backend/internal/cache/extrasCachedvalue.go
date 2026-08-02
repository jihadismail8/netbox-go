package cache

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-dev-frame/sponge/pkg/cache"
	"github.com/go-dev-frame/sponge/pkg/encoding"

	"netbox-go/internal/database"
	"netbox-go/internal/model"
)

const (
	// cache prefix key, must end with a colon
	extrasCachedvalueCachePrefixKey = "extrasCachedvalue:"
	// ExtrasCachedvalueExpireTime expire time
	ExtrasCachedvalueExpireTime = 5 * time.Minute
)

var _ ExtrasCachedvalueCache = (*extrasCachedvalueCache)(nil)

// ExtrasCachedvalueCache cache interface
type ExtrasCachedvalueCache interface {
	Set(ctx context.Context, id string, data *model.ExtrasCachedvalue, duration time.Duration) error
	Get(ctx context.Context, id string) (*model.ExtrasCachedvalue, error)
	MultiGet(ctx context.Context, ids []string) (map[string]*model.ExtrasCachedvalue, error)
	MultiSet(ctx context.Context, data []*model.ExtrasCachedvalue, duration time.Duration) error
	Del(ctx context.Context, id string) error
	SetPlaceholder(ctx context.Context, id string) error
	IsPlaceholderErr(err error) bool
}

// extrasCachedvalueCache define a cache struct
type extrasCachedvalueCache struct {
	cache cache.Cache
}

// NewExtrasCachedvalueCache new a cache
func NewExtrasCachedvalueCache(cacheType *database.CacheType) ExtrasCachedvalueCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasCachedvalue{}
		})
		return &extrasCachedvalueCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasCachedvalue{}
		})
		return &extrasCachedvalueCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasCachedvalueCacheKey cache key
func (c *extrasCachedvalueCache) GetExtrasCachedvalueCacheKey(id string) string {
	return extrasCachedvalueCachePrefixKey + id
}

// Set write to cache
func (c *extrasCachedvalueCache) Set(ctx context.Context, id string, data *model.ExtrasCachedvalue, duration time.Duration) error {
	if data == nil {
		return nil
	}
	cacheKey := c.GetExtrasCachedvalueCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasCachedvalueCache) Get(ctx context.Context, id string) (*model.ExtrasCachedvalue, error) {
	var data *model.ExtrasCachedvalue
	cacheKey := c.GetExtrasCachedvalueCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasCachedvalueCache) MultiSet(ctx context.Context, data []*model.ExtrasCachedvalue, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasCachedvalueCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasCachedvalueCache) MultiGet(ctx context.Context, ids []string) (map[string]*model.ExtrasCachedvalue, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasCachedvalueCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasCachedvalue)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[string]*model.ExtrasCachedvalue)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasCachedvalueCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasCachedvalueCache) Del(ctx context.Context, id string) error {
	cacheKey := c.GetExtrasCachedvalueCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasCachedvalueCache) SetPlaceholder(ctx context.Context, id string) error {
	cacheKey := c.GetExtrasCachedvalueCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasCachedvalueCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
