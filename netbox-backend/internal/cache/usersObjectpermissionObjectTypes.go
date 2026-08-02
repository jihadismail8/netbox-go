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
	usersObjectpermissionObjectTypesCachePrefixKey = "usersObjectpermissionObjectTypes:"
	// UsersObjectpermissionObjectTypesExpireTime expire time
	UsersObjectpermissionObjectTypesExpireTime = 5 * time.Minute
)

var _ UsersObjectpermissionObjectTypesCache = (*usersObjectpermissionObjectTypesCache)(nil)

// UsersObjectpermissionObjectTypesCache cache interface
type UsersObjectpermissionObjectTypesCache interface {
	Set(ctx context.Context, id uint64, data *model.UsersObjectpermissionObjectTypes, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.UsersObjectpermissionObjectTypes, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersObjectpermissionObjectTypes, error)
	MultiSet(ctx context.Context, data []*model.UsersObjectpermissionObjectTypes, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// usersObjectpermissionObjectTypesCache define a cache struct
type usersObjectpermissionObjectTypesCache struct {
	cache cache.Cache
}

// NewUsersObjectpermissionObjectTypesCache new a cache
func NewUsersObjectpermissionObjectTypesCache(cacheType *database.CacheType) UsersObjectpermissionObjectTypesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersObjectpermissionObjectTypes{}
		})
		return &usersObjectpermissionObjectTypesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.UsersObjectpermissionObjectTypes{}
		})
		return &usersObjectpermissionObjectTypesCache{cache: c}
	}

	return nil // no cache
}

// GetUsersObjectpermissionObjectTypesCacheKey cache key
func (c *usersObjectpermissionObjectTypesCache) GetUsersObjectpermissionObjectTypesCacheKey(id uint64) string {
	return usersObjectpermissionObjectTypesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *usersObjectpermissionObjectTypesCache) Set(ctx context.Context, id uint64, data *model.UsersObjectpermissionObjectTypes, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetUsersObjectpermissionObjectTypesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *usersObjectpermissionObjectTypesCache) Get(ctx context.Context, id uint64) (*model.UsersObjectpermissionObjectTypes, error) {
	var data *model.UsersObjectpermissionObjectTypes
	cacheKey := c.GetUsersObjectpermissionObjectTypesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *usersObjectpermissionObjectTypesCache) MultiSet(ctx context.Context, data []*model.UsersObjectpermissionObjectTypes, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetUsersObjectpermissionObjectTypesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *usersObjectpermissionObjectTypesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.UsersObjectpermissionObjectTypes, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetUsersObjectpermissionObjectTypesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.UsersObjectpermissionObjectTypes)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.UsersObjectpermissionObjectTypes)
	for _, id := range ids {
		val, ok := itemMap[c.GetUsersObjectpermissionObjectTypesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *usersObjectpermissionObjectTypesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersObjectpermissionObjectTypesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *usersObjectpermissionObjectTypesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetUsersObjectpermissionObjectTypesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *usersObjectpermissionObjectTypesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
