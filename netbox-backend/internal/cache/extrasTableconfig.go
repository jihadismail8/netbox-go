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
	extrasTableconfigCachePrefixKey = "extrasTableconfig:"
	// ExtrasTableconfigExpireTime expire time
	ExtrasTableconfigExpireTime = 5 * time.Minute
)

var _ ExtrasTableconfigCache = (*extrasTableconfigCache)(nil)

// ExtrasTableconfigCache cache interface
type ExtrasTableconfigCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasTableconfig, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasTableconfig, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasTableconfig, error)
	MultiSet(ctx context.Context, data []*model.ExtrasTableconfig, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasTableconfigCache define a cache struct
type extrasTableconfigCache struct {
	cache cache.Cache
}

// NewExtrasTableconfigCache new a cache
func NewExtrasTableconfigCache(cacheType *database.CacheType) ExtrasTableconfigCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasTableconfig{}
		})
		return &extrasTableconfigCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasTableconfig{}
		})
		return &extrasTableconfigCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasTableconfigCacheKey cache key
func (c *extrasTableconfigCache) GetExtrasTableconfigCacheKey(id uint64) string {
	return extrasTableconfigCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasTableconfigCache) Set(ctx context.Context, id uint64, data *model.ExtrasTableconfig, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasTableconfigCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasTableconfigCache) Get(ctx context.Context, id uint64) (*model.ExtrasTableconfig, error) {
	var data *model.ExtrasTableconfig
	cacheKey := c.GetExtrasTableconfigCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasTableconfigCache) MultiSet(ctx context.Context, data []*model.ExtrasTableconfig, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasTableconfigCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasTableconfigCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasTableconfig, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasTableconfigCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasTableconfig)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasTableconfig)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasTableconfigCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasTableconfigCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasTableconfigCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasTableconfigCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasTableconfigCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasTableconfigCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
