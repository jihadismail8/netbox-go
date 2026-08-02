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
	usersGroupPermissionsCachePrefixKey = "usersGroupPermissions:"
	// UsersGroupPermissionsExpireTime expire time
	UsersGroupPermissionsExpireTime = 5 * time.Minute
)

var _ UsersGroupPermissionsCache = (*usersGroupPermissionsCache)(nil)

// UsersGroupPermissionsCache cache interface
type UsersGroupPermissionsCache interface {
	Set(ctx context.Context, id uint64, data *model.UsersGroupPermissions, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.UsersGroupPermissions, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersGroupPermissions, error)
	MultiSet(ctx context.Context, data []*model.UsersGroupPermissions, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// usersGroupPermissionsCache define a cache struct
type usersGroupPermissionsCache struct {
	cache cache.Cache
}

// NewUsersGroupPermissionsCache new a cache
func NewUsersGroupPermissionsCache(cacheType *database.CacheType) UsersGroupPermissionsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersGroupPermissions{}
		})
		return &usersGroupPermissionsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersGroupPermissions{}
		})
		return &usersGroupPermissionsCache{cache: c}
	}

	return nil // no cache
}

// GetUsersGroupPermissionsCacheKey cache key
func (c *usersGroupPermissionsCache) GetUsersGroupPermissionsCacheKey(id uint64) string {
	return usersGroupPermissionsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *usersGroupPermissionsCache) Set(ctx context.Context, id uint64, data *model.UsersGroupPermissions, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetUsersGroupPermissionsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *usersGroupPermissionsCache) Get(ctx context.Context, id uint64) (*model.UsersGroupPermissions, error) {
	var data *model.UsersGroupPermissions
	cacheKey := c.GetUsersGroupPermissionsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *usersGroupPermissionsCache) MultiSet(ctx context.Context, data []*model.UsersGroupPermissions, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetUsersGroupPermissionsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *usersGroupPermissionsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersGroupPermissions, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetUsersGroupPermissionsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.UsersGroupPermissions)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.UsersGroupPermissions)
	for _, id := range ids {
		val, ok := itemMap[c.GetUsersGroupPermissionsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *usersGroupPermissionsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersGroupPermissionsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *usersGroupPermissionsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersGroupPermissionsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *usersGroupPermissionsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
