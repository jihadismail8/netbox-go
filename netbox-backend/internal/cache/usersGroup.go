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
	usersGroupCachePrefixKey = "usersGroup:"
	// UsersGroupExpireTime expire time
	UsersGroupExpireTime = 5 * time.Minute
)

var _ UsersGroupCache = (*usersGroupCache)(nil)

// UsersGroupCache cache interface
type UsersGroupCache interface {
	Set(ctx context.Context, id uint64, data *model.UsersGroup, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.UsersGroup, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersGroup, error)
	MultiSet(ctx context.Context, data []*model.UsersGroup, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// usersGroupCache define a cache struct
type usersGroupCache struct {
	cache cache.Cache
}

// NewUsersGroupCache new a cache
func NewUsersGroupCache(cacheType *database.CacheType) UsersGroupCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersGroup{}
		})
		return &usersGroupCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersGroup{}
		})
		return &usersGroupCache{cache: c}
	}

	return nil // no cache
}

// GetUsersGroupCacheKey cache key
func (c *usersGroupCache) GetUsersGroupCacheKey(id uint64) string {
	return usersGroupCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *usersGroupCache) Set(ctx context.Context, id uint64, data *model.UsersGroup, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetUsersGroupCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *usersGroupCache) Get(ctx context.Context, id uint64) (*model.UsersGroup, error) {
	var data *model.UsersGroup
	cacheKey := c.GetUsersGroupCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *usersGroupCache) MultiSet(ctx context.Context, data []*model.UsersGroup, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetUsersGroupCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *usersGroupCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersGroup, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetUsersGroupCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.UsersGroup)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.UsersGroup)
	for _, id := range ids {
		val, ok := itemMap[c.GetUsersGroupCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *usersGroupCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersGroupCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *usersGroupCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersGroupCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *usersGroupCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
