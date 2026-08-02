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
	djangoSessionCachePrefixKey = "djangoSession:"
	// DjangoSessionExpireTime expire time
	DjangoSessionExpireTime = 5 * time.Minute
)

var _ DjangoSessionCache = (*djangoSessionCache)(nil)

// DjangoSessionCache cache interface
type DjangoSessionCache interface {
	Set(ctx context.Context, sessionKey string, data *model.DjangoSession, duration time.Duration) error
	Get(ctx context.Context, sessionKey string) (*model.DjangoSession, error)
	MultiGet(ctx context.Context, sessionKeys []string) (map[string]*model.DjangoSession, error)
	MultiSet(ctx context.Context, data []*model.DjangoSession, duration time.Duration) error
	Del(ctx context.Context, sessionKey string) error
	SetPlaceholder(ctx context.Context, sessionKey string) error
	IsPlaceholderErr(err error) bool
}

// djangoSessionCache define a cache struct
type djangoSessionCache struct {
	cache cache.Cache
}

// NewDjangoSessionCache new a cache
func NewDjangoSessionCache(cacheType *database.CacheType) DjangoSessionCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DjangoSession{}
		})
		return &djangoSessionCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DjangoSession{}
		})
		return &djangoSessionCache{cache: c}
	}

	return nil // no cache
}

// GetDjangoSessionCacheKey cache key
func (c *djangoSessionCache) GetDjangoSessionCacheKey(sessionKey string) string {
	return djangoSessionCachePrefixKey + sessionKey
}

// Set write to cache
func (c *djangoSessionCache) Set(ctx context.Context, sessionKey string, data *model.DjangoSession, duration time.Duration) error {
	if data == nil {
		return nil
	}
	cacheKey := c.GetDjangoSessionCacheKey(sessionKey)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *djangoSessionCache) Get(ctx context.Context, sessionKey string) (*model.DjangoSession, error) {
	var data *model.DjangoSession
	cacheKey := c.GetDjangoSessionCacheKey(sessionKey)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *djangoSessionCache) MultiSet(ctx context.Context, data []*model.DjangoSession, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDjangoSessionCacheKey(v.SessionKey)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is sessionKey value
func (c *djangoSessionCache) MultiGet(ctx context.Context, sessionKeys []string) (map[string]*model.DjangoSession, error) {
	var keys []string
	for _, v := range sessionKeys {
		cacheKey := c.GetDjangoSessionCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DjangoSession)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[string]*model.DjangoSession)
	for _, sessionKey := range sessionKeys {
		val, ok := itemMap[c.GetDjangoSessionCacheKey(sessionKey)]
		if ok {
			retMap[sessionKey] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *djangoSessionCache) Del(ctx context.Context, sessionKey string) error {
	cacheKey := c.GetDjangoSessionCacheKey(sessionKey)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *djangoSessionCache) SetPlaceholder(ctx context.Context, sessionKey string) error {
	cacheKey := c.GetDjangoSessionCacheKey(sessionKey)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *djangoSessionCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
