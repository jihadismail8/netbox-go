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
	authGroupCachePrefixKey = "authGroup:"
	// AuthGroupExpireTime expire time
	AuthGroupExpireTime = 5 * time.Minute
)

var _ AuthGroupCache = (*authGroupCache)(nil)

// AuthGroupCache cache interface
type AuthGroupCache interface {
	Set(ctx context.Context, id uint64, data *model.AuthGroup, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.AuthGroup, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.AuthGroup, error)
	MultiSet(ctx context.Context, data []*model.AuthGroup, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// authGroupCache define a cache struct
type authGroupCache struct {
	cache cache.Cache
}

// NewAuthGroupCache new a cache
func NewAuthGroupCache(cacheType *database.CacheType) AuthGroupCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.AuthGroup{}
		})
		return &authGroupCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.AuthGroup{}
		})
		return &authGroupCache{cache: c}
	}

	return nil // no cache
}

// GetAuthGroupCacheKey cache key
func (c *authGroupCache) GetAuthGroupCacheKey(id uint64) string {
	return authGroupCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *authGroupCache) Set(ctx context.Context, id uint64, data *model.AuthGroup, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetAuthGroupCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *authGroupCache) Get(ctx context.Context, id uint64) (*model.AuthGroup, error) {
	var data *model.AuthGroup
	cacheKey := c.GetAuthGroupCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *authGroupCache) MultiSet(ctx context.Context, data []*model.AuthGroup, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetAuthGroupCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *authGroupCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.AuthGroup, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetAuthGroupCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.AuthGroup)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.AuthGroup)
	for _, id := range ids {
		val, ok := itemMap[c.GetAuthGroupCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *authGroupCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetAuthGroupCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *authGroupCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetAuthGroupCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *authGroupCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
