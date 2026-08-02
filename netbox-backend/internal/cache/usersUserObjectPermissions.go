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
	usersUserObjectPermissionsCachePrefixKey = "usersUserObjectPermissions:"
	// UsersUserObjectPermissionsExpireTime expire time
	UsersUserObjectPermissionsExpireTime = 5 * time.Minute
)

var _ UsersUserObjectPermissionsCache = (*usersUserObjectPermissionsCache)(nil)

// UsersUserObjectPermissionsCache cache interface
type UsersUserObjectPermissionsCache interface {
	Set(ctx context.Context, id uint64, data *model.UsersUserObjectPermissions, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.UsersUserObjectPermissions, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersUserObjectPermissions, error)
	MultiSet(ctx context.Context, data []*model.UsersUserObjectPermissions, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// usersUserObjectPermissionsCache define a cache struct
type usersUserObjectPermissionsCache struct {
	cache cache.Cache
}

// NewUsersUserObjectPermissionsCache new a cache
func NewUsersUserObjectPermissionsCache(cacheType *database.CacheType) UsersUserObjectPermissionsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersUserObjectPermissions{}
		})
		return &usersUserObjectPermissionsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersUserObjectPermissions{}
		})
		return &usersUserObjectPermissionsCache{cache: c}
	}

	return nil // no cache
}

// GetUsersUserObjectPermissionsCacheKey cache key
func (c *usersUserObjectPermissionsCache) GetUsersUserObjectPermissionsCacheKey(id uint64) string {
	return usersUserObjectPermissionsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *usersUserObjectPermissionsCache) Set(ctx context.Context, id uint64, data *model.UsersUserObjectPermissions, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetUsersUserObjectPermissionsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *usersUserObjectPermissionsCache) Get(ctx context.Context, id uint64) (*model.UsersUserObjectPermissions, error) {
	var data *model.UsersUserObjectPermissions
	cacheKey := c.GetUsersUserObjectPermissionsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *usersUserObjectPermissionsCache) MultiSet(ctx context.Context, data []*model.UsersUserObjectPermissions, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetUsersUserObjectPermissionsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *usersUserObjectPermissionsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersUserObjectPermissions, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetUsersUserObjectPermissionsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.UsersUserObjectPermissions)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.UsersUserObjectPermissions)
	for _, id := range ids {
		val, ok := itemMap[c.GetUsersUserObjectPermissionsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *usersUserObjectPermissionsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersUserObjectPermissionsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *usersUserObjectPermissionsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersUserObjectPermissionsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *usersUserObjectPermissionsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
