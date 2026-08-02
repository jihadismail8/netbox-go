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
	usersGroupObjectPermissionsCachePrefixKey = "usersGroupObjectPermissions:"
	// UsersGroupObjectPermissionsExpireTime expire time
	UsersGroupObjectPermissionsExpireTime = 5 * time.Minute
)

var _ UsersGroupObjectPermissionsCache = (*usersGroupObjectPermissionsCache)(nil)

// UsersGroupObjectPermissionsCache cache interface
type UsersGroupObjectPermissionsCache interface {
	Set(ctx context.Context, id uint64, data *model.UsersGroupObjectPermissions, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.UsersGroupObjectPermissions, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersGroupObjectPermissions, error)
	MultiSet(ctx context.Context, data []*model.UsersGroupObjectPermissions, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// usersGroupObjectPermissionsCache define a cache struct
type usersGroupObjectPermissionsCache struct {
	cache cache.Cache
}

// NewUsersGroupObjectPermissionsCache new a cache
func NewUsersGroupObjectPermissionsCache(cacheType *database.CacheType) UsersGroupObjectPermissionsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersGroupObjectPermissions{}
		})
		return &usersGroupObjectPermissionsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersGroupObjectPermissions{}
		})
		return &usersGroupObjectPermissionsCache{cache: c}
	}

	return nil // no cache
}

// GetUsersGroupObjectPermissionsCacheKey cache key
func (c *usersGroupObjectPermissionsCache) GetUsersGroupObjectPermissionsCacheKey(id uint64) string {
	return usersGroupObjectPermissionsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *usersGroupObjectPermissionsCache) Set(ctx context.Context, id uint64, data *model.UsersGroupObjectPermissions, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetUsersGroupObjectPermissionsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *usersGroupObjectPermissionsCache) Get(ctx context.Context, id uint64) (*model.UsersGroupObjectPermissions, error) {
	var data *model.UsersGroupObjectPermissions
	cacheKey := c.GetUsersGroupObjectPermissionsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *usersGroupObjectPermissionsCache) MultiSet(ctx context.Context, data []*model.UsersGroupObjectPermissions, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetUsersGroupObjectPermissionsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *usersGroupObjectPermissionsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersGroupObjectPermissions, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetUsersGroupObjectPermissionsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.UsersGroupObjectPermissions)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.UsersGroupObjectPermissions)
	for _, id := range ids {
		val, ok := itemMap[c.GetUsersGroupObjectPermissionsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *usersGroupObjectPermissionsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersGroupObjectPermissionsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *usersGroupObjectPermissionsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersGroupObjectPermissionsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *usersGroupObjectPermissionsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
