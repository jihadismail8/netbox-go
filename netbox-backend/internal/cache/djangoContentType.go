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
	djangoContentTypeCachePrefixKey = "djangoContentType:"
	// DjangoContentTypeExpireTime expire time
	DjangoContentTypeExpireTime = 5 * time.Minute
)

var _ DjangoContentTypeCache = (*djangoContentTypeCache)(nil)

// DjangoContentTypeCache cache interface
type DjangoContentTypeCache interface {
	Set(ctx context.Context, id uint64, data *model.DjangoContentType, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DjangoContentType, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DjangoContentType, error)
	MultiSet(ctx context.Context, data []*model.DjangoContentType, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// djangoContentTypeCache define a cache struct
type djangoContentTypeCache struct {
	cache cache.Cache
}

// NewDjangoContentTypeCache new a cache
func NewDjangoContentTypeCache(cacheType *database.CacheType) DjangoContentTypeCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DjangoContentType{}
		})
		return &djangoContentTypeCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DjangoContentType{}
		})
		return &djangoContentTypeCache{cache: c}
	}

	return nil // no cache
}

// GetDjangoContentTypeCacheKey cache key
func (c *djangoContentTypeCache) GetDjangoContentTypeCacheKey(id uint64) string {
	return djangoContentTypeCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *djangoContentTypeCache) Set(ctx context.Context, id uint64, data *model.DjangoContentType, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDjangoContentTypeCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *djangoContentTypeCache) Get(ctx context.Context, id uint64) (*model.DjangoContentType, error) {
	var data *model.DjangoContentType
	cacheKey := c.GetDjangoContentTypeCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *djangoContentTypeCache) MultiSet(ctx context.Context, data []*model.DjangoContentType, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDjangoContentTypeCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *djangoContentTypeCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DjangoContentType, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDjangoContentTypeCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DjangoContentType)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DjangoContentType)
	for _, id := range ids {
		val, ok := itemMap[c.GetDjangoContentTypeCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *djangoContentTypeCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDjangoContentTypeCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *djangoContentTypeCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDjangoContentTypeCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *djangoContentTypeCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
