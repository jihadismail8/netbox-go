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
	extrasTagObjectTypesCachePrefixKey = "extrasTagObjectTypes:"
	// ExtrasTagObjectTypesExpireTime expire time
	ExtrasTagObjectTypesExpireTime = 5 * time.Minute
)

var _ ExtrasTagObjectTypesCache = (*extrasTagObjectTypesCache)(nil)

// ExtrasTagObjectTypesCache cache interface
type ExtrasTagObjectTypesCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasTagObjectTypes, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasTagObjectTypes, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasTagObjectTypes, error)
	MultiSet(ctx context.Context, data []*model.ExtrasTagObjectTypes, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasTagObjectTypesCache define a cache struct
type extrasTagObjectTypesCache struct {
	cache cache.Cache
}

// NewExtrasTagObjectTypesCache new a cache
func NewExtrasTagObjectTypesCache(cacheType *database.CacheType) ExtrasTagObjectTypesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasTagObjectTypes{}
		})
		return &extrasTagObjectTypesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasTagObjectTypes{}
		})
		return &extrasTagObjectTypesCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasTagObjectTypesCacheKey cache key
func (c *extrasTagObjectTypesCache) GetExtrasTagObjectTypesCacheKey(id uint64) string {
	return extrasTagObjectTypesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasTagObjectTypesCache) Set(ctx context.Context, id uint64, data *model.ExtrasTagObjectTypes, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasTagObjectTypesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasTagObjectTypesCache) Get(ctx context.Context, id uint64) (*model.ExtrasTagObjectTypes, error) {
	var data *model.ExtrasTagObjectTypes
	cacheKey := c.GetExtrasTagObjectTypesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasTagObjectTypesCache) MultiSet(ctx context.Context, data []*model.ExtrasTagObjectTypes, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasTagObjectTypesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasTagObjectTypesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasTagObjectTypes, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasTagObjectTypesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasTagObjectTypes)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasTagObjectTypes)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasTagObjectTypesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasTagObjectTypesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasTagObjectTypesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasTagObjectTypesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasTagObjectTypesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasTagObjectTypesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
