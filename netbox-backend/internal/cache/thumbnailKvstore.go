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
	thumbnailKvstoreCachePrefixKey = "thumbnailKvstore:"
	// ThumbnailKvstoreExpireTime expire time
	ThumbnailKvstoreExpireTime = 5 * time.Minute
)

var _ ThumbnailKvstoreCache = (*thumbnailKvstoreCache)(nil)

// ThumbnailKvstoreCache cache interface
type ThumbnailKvstoreCache interface {
	Set(ctx context.Context, key string, data *model.ThumbnailKvstore, duration time.Duration) error
	Get(ctx context.Context, key string) (*model.ThumbnailKvstore, error)
	MultiGet(ctx context.Context, keys []string) (map[string]*model.ThumbnailKvstore, error)
	MultiSet(ctx context.Context, data []*model.ThumbnailKvstore, duration time.Duration) error
	Del(ctx context.Context, key string) error
	SetPlaceholder(ctx context.Context, key string) error
	IsPlaceholderErr(err error) bool
}

// thumbnailKvstoreCache define a cache struct
type thumbnailKvstoreCache struct {
	cache cache.Cache
}

// NewThumbnailKvstoreCache new a cache
func NewThumbnailKvstoreCache(cacheType *database.CacheType) ThumbnailKvstoreCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ThumbnailKvstore{}
		})
		return &thumbnailKvstoreCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ThumbnailKvstore{}
		})
		return &thumbnailKvstoreCache{cache: c}
	}

	return nil // no cache
}

// GetThumbnailKvstoreCacheKey cache key
func (c *thumbnailKvstoreCache) GetThumbnailKvstoreCacheKey(key string) string {
	return thumbnailKvstoreCachePrefixKey + key
}

// Set write to cache
func (c *thumbnailKvstoreCache) Set(ctx context.Context, key string, data *model.ThumbnailKvstore, duration time.Duration) error {
	if data == nil {
		return nil
	}
	cacheKey := c.GetThumbnailKvstoreCacheKey(key)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *thumbnailKvstoreCache) Get(ctx context.Context, key string) (*model.ThumbnailKvstore, error) {
	var data *model.ThumbnailKvstore
	cacheKey := c.GetThumbnailKvstoreCacheKey(key)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *thumbnailKvstoreCache) MultiSet(ctx context.Context, data []*model.ThumbnailKvstore, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetThumbnailKvstoreCacheKey(v.Key)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is key value
func (c *thumbnailKvstoreCache) MultiGet(ctx context.Context, keys []string) (map[string]*model.ThumbnailKvstore, error) {
	var cacheKeys []string
	for _, v := range keys {
		cacheKey := c.GetThumbnailKvstoreCacheKey(v)
		cacheKeys = append(cacheKeys, cacheKey)
	}

	itemMap := make(map[string]*model.ThumbnailKvstore)
	err := c.cache.MultiGet(ctx, cacheKeys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[string]*model.ThumbnailKvstore)
	for _, key := range keys {
		val, ok := itemMap[c.GetThumbnailKvstoreCacheKey(key)]
		if ok {
			retMap[key] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *thumbnailKvstoreCache) Del(ctx context.Context, key string) error {
	cacheKey := c.GetThumbnailKvstoreCacheKey(key)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *thumbnailKvstoreCache) SetPlaceholder(ctx context.Context, key string) error {
	cacheKey := c.GetThumbnailKvstoreCacheKey(key)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *thumbnailKvstoreCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
