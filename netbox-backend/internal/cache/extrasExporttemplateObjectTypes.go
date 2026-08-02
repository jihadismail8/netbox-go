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
	extrasExporttemplateObjectTypesCachePrefixKey = "extrasExporttemplateObjectTypes:"
	// ExtrasExporttemplateObjectTypesExpireTime expire time
	ExtrasExporttemplateObjectTypesExpireTime = 5 * time.Minute
)

var _ ExtrasExporttemplateObjectTypesCache = (*extrasExporttemplateObjectTypesCache)(nil)

// ExtrasExporttemplateObjectTypesCache cache interface
type ExtrasExporttemplateObjectTypesCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasExporttemplateObjectTypes, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasExporttemplateObjectTypes, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasExporttemplateObjectTypes, error)
	MultiSet(ctx context.Context, data []*model.ExtrasExporttemplateObjectTypes, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasExporttemplateObjectTypesCache define a cache struct
type extrasExporttemplateObjectTypesCache struct {
	cache cache.Cache
}

// NewExtrasExporttemplateObjectTypesCache new a cache
func NewExtrasExporttemplateObjectTypesCache(cacheType *database.CacheType) ExtrasExporttemplateObjectTypesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasExporttemplateObjectTypes{}
		})
		return &extrasExporttemplateObjectTypesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasExporttemplateObjectTypes{}
		})
		return &extrasExporttemplateObjectTypesCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasExporttemplateObjectTypesCacheKey cache key
func (c *extrasExporttemplateObjectTypesCache) GetExtrasExporttemplateObjectTypesCacheKey(id uint64) string {
	return extrasExporttemplateObjectTypesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasExporttemplateObjectTypesCache) Set(ctx context.Context, id uint64, data *model.ExtrasExporttemplateObjectTypes, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasExporttemplateObjectTypesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasExporttemplateObjectTypesCache) Get(ctx context.Context, id uint64) (*model.ExtrasExporttemplateObjectTypes, error) {
	var data *model.ExtrasExporttemplateObjectTypes
	cacheKey := c.GetExtrasExporttemplateObjectTypesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasExporttemplateObjectTypesCache) MultiSet(ctx context.Context, data []*model.ExtrasExporttemplateObjectTypes, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasExporttemplateObjectTypesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasExporttemplateObjectTypesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasExporttemplateObjectTypes, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasExporttemplateObjectTypesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasExporttemplateObjectTypes)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasExporttemplateObjectTypes)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasExporttemplateObjectTypesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasExporttemplateObjectTypesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasExporttemplateObjectTypesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasExporttemplateObjectTypesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasExporttemplateObjectTypesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasExporttemplateObjectTypesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
