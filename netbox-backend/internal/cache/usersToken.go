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
	usersTokenCachePrefixKey = "usersToken:"
	// UsersTokenExpireTime expire time
	UsersTokenExpireTime = 5 * time.Minute
)

var _ UsersTokenCache = (*usersTokenCache)(nil)

// UsersTokenCache cache interface
type UsersTokenCache interface {
	Set(ctx context.Context, id uint64, data *model.UsersToken, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.UsersToken, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersToken, error)
	MultiSet(ctx context.Context, data []*model.UsersToken, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// usersTokenCache define a cache struct
type usersTokenCache struct {
	cache cache.Cache
}

// NewUsersTokenCache new a cache
func NewUsersTokenCache(cacheType *database.CacheType) UsersTokenCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersToken{}
		})
		return &usersTokenCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersToken{}
		})
		return &usersTokenCache{cache: c}
	}

	return nil // no cache
}

// GetUsersTokenCacheKey cache key
func (c *usersTokenCache) GetUsersTokenCacheKey(id uint64) string {
	return usersTokenCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *usersTokenCache) Set(ctx context.Context, id uint64, data *model.UsersToken, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetUsersTokenCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *usersTokenCache) Get(ctx context.Context, id uint64) (*model.UsersToken, error) {
	var data *model.UsersToken
	cacheKey := c.GetUsersTokenCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *usersTokenCache) MultiSet(ctx context.Context, data []*model.UsersToken, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetUsersTokenCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *usersTokenCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersToken, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetUsersTokenCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.UsersToken)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.UsersToken)
	for _, id := range ids {
		val, ok := itemMap[c.GetUsersTokenCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *usersTokenCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersTokenCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *usersTokenCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersTokenCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *usersTokenCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
