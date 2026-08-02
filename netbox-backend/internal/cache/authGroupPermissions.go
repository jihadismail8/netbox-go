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
	authGroupPermissionsCachePrefixKey = "authGroupPermissions:"
	// AuthGroupPermissionsExpireTime expire time
	AuthGroupPermissionsExpireTime = 5 * time.Minute
)

var _ AuthGroupPermissionsCache = (*authGroupPermissionsCache)(nil)

// AuthGroupPermissionsCache cache interface
type AuthGroupPermissionsCache interface {
	Set(ctx context.Context, id uint64, data *model.AuthGroupPermissions, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.AuthGroupPermissions, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.AuthGroupPermissions, error)
	MultiSet(ctx context.Context, data []*model.AuthGroupPermissions, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// authGroupPermissionsCache define a cache struct
type authGroupPermissionsCache struct {
	cache cache.Cache
}

// NewAuthGroupPermissionsCache new a cache
func NewAuthGroupPermissionsCache(cacheType *database.CacheType) AuthGroupPermissionsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.AuthGroupPermissions{}
		})
		return &authGroupPermissionsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.AuthGroupPermissions{}
		})
		return &authGroupPermissionsCache{cache: c}
	}

	return nil // no cache
}

// GetAuthGroupPermissionsCacheKey cache key
func (c *authGroupPermissionsCache) GetAuthGroupPermissionsCacheKey(id uint64) string {
	return authGroupPermissionsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *authGroupPermissionsCache) Set(ctx context.Context, id uint64, data *model.AuthGroupPermissions, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetAuthGroupPermissionsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *authGroupPermissionsCache) Get(ctx context.Context, id uint64) (*model.AuthGroupPermissions, error) {
	var data *model.AuthGroupPermissions
	cacheKey := c.GetAuthGroupPermissionsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *authGroupPermissionsCache) MultiSet(ctx context.Context, data []*model.AuthGroupPermissions, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetAuthGroupPermissionsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *authGroupPermissionsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.AuthGroupPermissions, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetAuthGroupPermissionsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.AuthGroupPermissions)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.AuthGroupPermissions)
	for _, id := range ids {
		val, ok := itemMap[c.GetAuthGroupPermissionsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *authGroupPermissionsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetAuthGroupPermissionsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *authGroupPermissionsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetAuthGroupPermissionsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *authGroupPermissionsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
