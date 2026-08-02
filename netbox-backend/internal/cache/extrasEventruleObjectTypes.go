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
	extrasEventruleObjectTypesCachePrefixKey = "extrasEventruleObjectTypes:"
	// ExtrasEventruleObjectTypesExpireTime expire time
	ExtrasEventruleObjectTypesExpireTime = 5 * time.Minute
)

var _ ExtrasEventruleObjectTypesCache = (*extrasEventruleObjectTypesCache)(nil)

// ExtrasEventruleObjectTypesCache cache interface
type ExtrasEventruleObjectTypesCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasEventruleObjectTypes, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasEventruleObjectTypes, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasEventruleObjectTypes, error)
	MultiSet(ctx context.Context, data []*model.ExtrasEventruleObjectTypes, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasEventruleObjectTypesCache define a cache struct
type extrasEventruleObjectTypesCache struct {
	cache cache.Cache
}

// NewExtrasEventruleObjectTypesCache new a cache
func NewExtrasEventruleObjectTypesCache(cacheType *database.CacheType) ExtrasEventruleObjectTypesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasEventruleObjectTypes{}
		})
		return &extrasEventruleObjectTypesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasEventruleObjectTypes{}
		})
		return &extrasEventruleObjectTypesCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasEventruleObjectTypesCacheKey cache key
func (c *extrasEventruleObjectTypesCache) GetExtrasEventruleObjectTypesCacheKey(id uint64) string {
	return extrasEventruleObjectTypesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasEventruleObjectTypesCache) Set(ctx context.Context, id uint64, data *model.ExtrasEventruleObjectTypes, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasEventruleObjectTypesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasEventruleObjectTypesCache) Get(ctx context.Context, id uint64) (*model.ExtrasEventruleObjectTypes, error) {
	var data *model.ExtrasEventruleObjectTypes
	cacheKey := c.GetExtrasEventruleObjectTypesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasEventruleObjectTypesCache) MultiSet(ctx context.Context, data []*model.ExtrasEventruleObjectTypes, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasEventruleObjectTypesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasEventruleObjectTypesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasEventruleObjectTypes, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasEventruleObjectTypesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasEventruleObjectTypes)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasEventruleObjectTypes)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasEventruleObjectTypesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasEventruleObjectTypesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasEventruleObjectTypesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasEventruleObjectTypesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasEventruleObjectTypesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasEventruleObjectTypesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
