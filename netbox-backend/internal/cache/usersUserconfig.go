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
	usersUserconfigCachePrefixKey = "usersUserconfig:"
	// UsersUserconfigExpireTime expire time
	UsersUserconfigExpireTime = 5 * time.Minute
)

var _ UsersUserconfigCache = (*usersUserconfigCache)(nil)

// UsersUserconfigCache cache interface
type UsersUserconfigCache interface {
	Set(ctx context.Context, id uint64, data *model.UsersUserconfig, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.UsersUserconfig, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersUserconfig, error)
	MultiSet(ctx context.Context, data []*model.UsersUserconfig, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// usersUserconfigCache define a cache struct
type usersUserconfigCache struct {
	cache cache.Cache
}

// NewUsersUserconfigCache new a cache
func NewUsersUserconfigCache(cacheType *database.CacheType) UsersUserconfigCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersUserconfig{}
		})
		return &usersUserconfigCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersUserconfig{}
		})
		return &usersUserconfigCache{cache: c}
	}

	return nil // no cache
}

// GetUsersUserconfigCacheKey cache key
func (c *usersUserconfigCache) GetUsersUserconfigCacheKey(id uint64) string {
	return usersUserconfigCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *usersUserconfigCache) Set(ctx context.Context, id uint64, data *model.UsersUserconfig, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetUsersUserconfigCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *usersUserconfigCache) Get(ctx context.Context, id uint64) (*model.UsersUserconfig, error) {
	var data *model.UsersUserconfig
	cacheKey := c.GetUsersUserconfigCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *usersUserconfigCache) MultiSet(ctx context.Context, data []*model.UsersUserconfig, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetUsersUserconfigCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *usersUserconfigCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersUserconfig, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetUsersUserconfigCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.UsersUserconfig)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.UsersUserconfig)
	for _, id := range ids {
		val, ok := itemMap[c.GetUsersUserconfigCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *usersUserconfigCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersUserconfigCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *usersUserconfigCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersUserconfigCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *usersUserconfigCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
