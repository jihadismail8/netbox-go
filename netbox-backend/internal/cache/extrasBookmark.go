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
	extrasBookmarkCachePrefixKey = "extrasBookmark:"
	// ExtrasBookmarkExpireTime expire time
	ExtrasBookmarkExpireTime = 5 * time.Minute
)

var _ ExtrasBookmarkCache = (*extrasBookmarkCache)(nil)

// ExtrasBookmarkCache cache interface
type ExtrasBookmarkCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasBookmark, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasBookmark, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasBookmark, error)
	MultiSet(ctx context.Context, data []*model.ExtrasBookmark, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasBookmarkCache define a cache struct
type extrasBookmarkCache struct {
	cache cache.Cache
}

// NewExtrasBookmarkCache new a cache
func NewExtrasBookmarkCache(cacheType *database.CacheType) ExtrasBookmarkCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasBookmark{}
		})
		return &extrasBookmarkCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasBookmark{}
		})
		return &extrasBookmarkCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasBookmarkCacheKey cache key
func (c *extrasBookmarkCache) GetExtrasBookmarkCacheKey(id uint64) string {
	return extrasBookmarkCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasBookmarkCache) Set(ctx context.Context, id uint64, data *model.ExtrasBookmark, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasBookmarkCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasBookmarkCache) Get(ctx context.Context, id uint64) (*model.ExtrasBookmark, error) {
	var data *model.ExtrasBookmark
	cacheKey := c.GetExtrasBookmarkCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasBookmarkCache) MultiSet(ctx context.Context, data []*model.ExtrasBookmark, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasBookmarkCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasBookmarkCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasBookmark, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasBookmarkCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasBookmark)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasBookmark)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasBookmarkCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasBookmarkCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasBookmarkCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasBookmarkCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasBookmarkCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasBookmarkCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
