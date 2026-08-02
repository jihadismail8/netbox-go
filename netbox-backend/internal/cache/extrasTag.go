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
	extrasTagCachePrefixKey = "extrasTag:"
	// ExtrasTagExpireTime expire time
	ExtrasTagExpireTime = 5 * time.Minute
)

var _ ExtrasTagCache = (*extrasTagCache)(nil)

// ExtrasTagCache cache interface
type ExtrasTagCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasTag, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasTag, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasTag, error)
	MultiSet(ctx context.Context, data []*model.ExtrasTag, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasTagCache define a cache struct
type extrasTagCache struct {
	cache cache.Cache
}

// NewExtrasTagCache new a cache
func NewExtrasTagCache(cacheType *database.CacheType) ExtrasTagCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasTag{}
		})
		return &extrasTagCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasTag{}
		})
		return &extrasTagCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasTagCacheKey cache key
func (c *extrasTagCache) GetExtrasTagCacheKey(id uint64) string {
	return extrasTagCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasTagCache) Set(ctx context.Context, id uint64, data *model.ExtrasTag, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasTagCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasTagCache) Get(ctx context.Context, id uint64) (*model.ExtrasTag, error) {
	var data *model.ExtrasTag
	cacheKey := c.GetExtrasTagCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasTagCache) MultiSet(ctx context.Context, data []*model.ExtrasTag, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasTagCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasTagCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasTag, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasTagCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasTag)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasTag)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasTagCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasTagCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasTagCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasTagCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasTagCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasTagCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
