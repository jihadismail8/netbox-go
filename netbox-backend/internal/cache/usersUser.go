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
	usersUserCachePrefixKey = "usersUser:"
	// UsersUserExpireTime expire time
	UsersUserExpireTime = 5 * time.Minute
)

var _ UsersUserCache = (*usersUserCache)(nil)

// UsersUserCache cache interface
type UsersUserCache interface {
	Set(ctx context.Context, id uint64, data *model.UsersUser, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.UsersUser, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersUser, error)
	MultiSet(ctx context.Context, data []*model.UsersUser, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// usersUserCache define a cache struct
type usersUserCache struct {
	cache cache.Cache
}

// NewUsersUserCache new a cache
func NewUsersUserCache(cacheType *database.CacheType) UsersUserCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersUser{}
		})
		return &usersUserCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersUser{}
		})
		return &usersUserCache{cache: c}
	}

	return nil // no cache
}

// GetUsersUserCacheKey cache key
func (c *usersUserCache) GetUsersUserCacheKey(id uint64) string {
	return usersUserCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *usersUserCache) Set(ctx context.Context, id uint64, data *model.UsersUser, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetUsersUserCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *usersUserCache) Get(ctx context.Context, id uint64) (*model.UsersUser, error) {
	var data *model.UsersUser
	cacheKey := c.GetUsersUserCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *usersUserCache) MultiSet(ctx context.Context, data []*model.UsersUser, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetUsersUserCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *usersUserCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersUser, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetUsersUserCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.UsersUser)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.UsersUser)
	for _, id := range ids {
		val, ok := itemMap[c.GetUsersUserCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *usersUserCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersUserCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *usersUserCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersUserCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *usersUserCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
