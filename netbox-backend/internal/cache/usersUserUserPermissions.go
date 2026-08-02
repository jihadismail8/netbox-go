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
	usersUserUserPermissionsCachePrefixKey = "usersUserUserPermissions:"
	// UsersUserUserPermissionsExpireTime expire time
	UsersUserUserPermissionsExpireTime = 5 * time.Minute
)

var _ UsersUserUserPermissionsCache = (*usersUserUserPermissionsCache)(nil)

// UsersUserUserPermissionsCache cache interface
type UsersUserUserPermissionsCache interface {
	Set(ctx context.Context, id uint64, data *model.UsersUserUserPermissions, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.UsersUserUserPermissions, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersUserUserPermissions, error)
	MultiSet(ctx context.Context, data []*model.UsersUserUserPermissions, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// usersUserUserPermissionsCache define a cache struct
type usersUserUserPermissionsCache struct {
	cache cache.Cache
}

// NewUsersUserUserPermissionsCache new a cache
func NewUsersUserUserPermissionsCache(cacheType *database.CacheType) UsersUserUserPermissionsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersUserUserPermissions{}
		})
		return &usersUserUserPermissionsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersUserUserPermissions{}
		})
		return &usersUserUserPermissionsCache{cache: c}
	}

	return nil // no cache
}

// GetUsersUserUserPermissionsCacheKey cache key
func (c *usersUserUserPermissionsCache) GetUsersUserUserPermissionsCacheKey(id uint64) string {
	return usersUserUserPermissionsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *usersUserUserPermissionsCache) Set(ctx context.Context, id uint64, data *model.UsersUserUserPermissions, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetUsersUserUserPermissionsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *usersUserUserPermissionsCache) Get(ctx context.Context, id uint64) (*model.UsersUserUserPermissions, error) {
	var data *model.UsersUserUserPermissions
	cacheKey := c.GetUsersUserUserPermissionsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *usersUserUserPermissionsCache) MultiSet(ctx context.Context, data []*model.UsersUserUserPermissions, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetUsersUserUserPermissionsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *usersUserUserPermissionsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersUserUserPermissions, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetUsersUserUserPermissionsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.UsersUserUserPermissions)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.UsersUserUserPermissions)
	for _, id := range ids {
		val, ok := itemMap[c.GetUsersUserUserPermissionsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *usersUserUserPermissionsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersUserUserPermissionsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *usersUserUserPermissionsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersUserUserPermissionsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *usersUserUserPermissionsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
