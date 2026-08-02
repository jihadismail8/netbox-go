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
	authPermissionCachePrefixKey = "authPermission:"
	// AuthPermissionExpireTime expire time
	AuthPermissionExpireTime = 5 * time.Minute
)

var _ AuthPermissionCache = (*authPermissionCache)(nil)

// AuthPermissionCache cache interface
type AuthPermissionCache interface {
	Set(ctx context.Context, id uint64, data *model.AuthPermission, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.AuthPermission, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.AuthPermission, error)
	MultiSet(ctx context.Context, data []*model.AuthPermission, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// authPermissionCache define a cache struct
type authPermissionCache struct {
	cache cache.Cache
}

// NewAuthPermissionCache new a cache
func NewAuthPermissionCache(cacheType *database.CacheType) AuthPermissionCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.AuthPermission{}
		})
		return &authPermissionCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.AuthPermission{}
		})
		return &authPermissionCache{cache: c}
	}

	return nil // no cache
}

// GetAuthPermissionCacheKey cache key
func (c *authPermissionCache) GetAuthPermissionCacheKey(id uint64) string {
	return authPermissionCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *authPermissionCache) Set(ctx context.Context, id uint64, data *model.AuthPermission, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetAuthPermissionCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *authPermissionCache) Get(ctx context.Context, id uint64) (*model.AuthPermission, error) {
	var data *model.AuthPermission
	cacheKey := c.GetAuthPermissionCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *authPermissionCache) MultiSet(ctx context.Context, data []*model.AuthPermission, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetAuthPermissionCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *authPermissionCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.AuthPermission, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetAuthPermissionCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.AuthPermission)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.AuthPermission)
	for _, id := range ids {
		val, ok := itemMap[c.GetAuthPermissionCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *authPermissionCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetAuthPermissionCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *authPermissionCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetAuthPermissionCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *authPermissionCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
