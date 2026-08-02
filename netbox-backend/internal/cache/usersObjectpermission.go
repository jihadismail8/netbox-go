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
	usersObjectpermissionCachePrefixKey = "usersObjectpermission:"
	// UsersObjectpermissionExpireTime expire time
	UsersObjectpermissionExpireTime = 5 * time.Minute
)

var _ UsersObjectpermissionCache = (*usersObjectpermissionCache)(nil)

// UsersObjectpermissionCache cache interface
type UsersObjectpermissionCache interface {
	Set(ctx context.Context, id uint64, data *model.UsersObjectpermission, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.UsersObjectpermission, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersObjectpermission, error)
	MultiSet(ctx context.Context, data []*model.UsersObjectpermission, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// usersObjectpermissionCache define a cache struct
type usersObjectpermissionCache struct {
	cache cache.Cache
}

// NewUsersObjectpermissionCache new a cache
func NewUsersObjectpermissionCache(cacheType *database.CacheType) UsersObjectpermissionCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersObjectpermission{}
		})
		return &usersObjectpermissionCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersObjectpermission{}
		})
		return &usersObjectpermissionCache{cache: c}
	}

	return nil // no cache
}

// GetUsersObjectpermissionCacheKey cache key
func (c *usersObjectpermissionCache) GetUsersObjectpermissionCacheKey(id uint64) string {
	return usersObjectpermissionCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *usersObjectpermissionCache) Set(ctx context.Context, id uint64, data *model.UsersObjectpermission, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetUsersObjectpermissionCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *usersObjectpermissionCache) Get(ctx context.Context, id uint64) (*model.UsersObjectpermission, error) {
	var data *model.UsersObjectpermission
	cacheKey := c.GetUsersObjectpermissionCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *usersObjectpermissionCache) MultiSet(ctx context.Context, data []*model.UsersObjectpermission, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetUsersObjectpermissionCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *usersObjectpermissionCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersObjectpermission, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetUsersObjectpermissionCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.UsersObjectpermission)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.UsersObjectpermission)
	for _, id := range ids {
		val, ok := itemMap[c.GetUsersObjectpermissionCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *usersObjectpermissionCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersObjectpermissionCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *usersObjectpermissionCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersObjectpermissionCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *usersObjectpermissionCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
