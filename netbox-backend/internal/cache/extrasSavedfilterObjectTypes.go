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
	extrasSavedfilterObjectTypesCachePrefixKey = "extrasSavedfilterObjectTypes:"
	// ExtrasSavedfilterObjectTypesExpireTime expire time
	ExtrasSavedfilterObjectTypesExpireTime = 5 * time.Minute
)

var _ ExtrasSavedfilterObjectTypesCache = (*extrasSavedfilterObjectTypesCache)(nil)

// ExtrasSavedfilterObjectTypesCache cache interface
type ExtrasSavedfilterObjectTypesCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasSavedfilterObjectTypes, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasSavedfilterObjectTypes, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasSavedfilterObjectTypes, error)
	MultiSet(ctx context.Context, data []*model.ExtrasSavedfilterObjectTypes, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasSavedfilterObjectTypesCache define a cache struct
type extrasSavedfilterObjectTypesCache struct {
	cache cache.Cache
}

// NewExtrasSavedfilterObjectTypesCache new a cache
func NewExtrasSavedfilterObjectTypesCache(cacheType *database.CacheType) ExtrasSavedfilterObjectTypesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasSavedfilterObjectTypes{}
		})
		return &extrasSavedfilterObjectTypesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasSavedfilterObjectTypes{}
		})
		return &extrasSavedfilterObjectTypesCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasSavedfilterObjectTypesCacheKey cache key
func (c *extrasSavedfilterObjectTypesCache) GetExtrasSavedfilterObjectTypesCacheKey(id uint64) string {
	return extrasSavedfilterObjectTypesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasSavedfilterObjectTypesCache) Set(ctx context.Context, id uint64, data *model.ExtrasSavedfilterObjectTypes, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasSavedfilterObjectTypesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasSavedfilterObjectTypesCache) Get(ctx context.Context, id uint64) (*model.ExtrasSavedfilterObjectTypes, error) {
	var data *model.ExtrasSavedfilterObjectTypes
	cacheKey := c.GetExtrasSavedfilterObjectTypesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasSavedfilterObjectTypesCache) MultiSet(ctx context.Context, data []*model.ExtrasSavedfilterObjectTypes, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasSavedfilterObjectTypesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasSavedfilterObjectTypesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasSavedfilterObjectTypes, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasSavedfilterObjectTypesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasSavedfilterObjectTypes)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasSavedfilterObjectTypes)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasSavedfilterObjectTypesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasSavedfilterObjectTypesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasSavedfilterObjectTypesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasSavedfilterObjectTypesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasSavedfilterObjectTypesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasSavedfilterObjectTypesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
