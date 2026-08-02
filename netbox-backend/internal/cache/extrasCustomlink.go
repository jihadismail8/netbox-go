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
	extrasCustomlinkCachePrefixKey = "extrasCustomlink:"
	// ExtrasCustomlinkExpireTime expire time
	ExtrasCustomlinkExpireTime = 5 * time.Minute
)

var _ ExtrasCustomlinkCache = (*extrasCustomlinkCache)(nil)

// ExtrasCustomlinkCache cache interface
type ExtrasCustomlinkCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasCustomlink, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasCustomlink, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasCustomlink, error)
	MultiSet(ctx context.Context, data []*model.ExtrasCustomlink, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasCustomlinkCache define a cache struct
type extrasCustomlinkCache struct {
	cache cache.Cache
}

// NewExtrasCustomlinkCache new a cache
func NewExtrasCustomlinkCache(cacheType *database.CacheType) ExtrasCustomlinkCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasCustomlink{}
		})
		return &extrasCustomlinkCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasCustomlink{}
		})
		return &extrasCustomlinkCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasCustomlinkCacheKey cache key
func (c *extrasCustomlinkCache) GetExtrasCustomlinkCacheKey(id uint64) string {
	return extrasCustomlinkCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasCustomlinkCache) Set(ctx context.Context, id uint64, data *model.ExtrasCustomlink, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasCustomlinkCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasCustomlinkCache) Get(ctx context.Context, id uint64) (*model.ExtrasCustomlink, error) {
	var data *model.ExtrasCustomlink
	cacheKey := c.GetExtrasCustomlinkCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasCustomlinkCache) MultiSet(ctx context.Context, data []*model.ExtrasCustomlink, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasCustomlinkCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasCustomlinkCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasCustomlink, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasCustomlinkCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasCustomlink)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasCustomlink)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasCustomlinkCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasCustomlinkCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasCustomlinkCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasCustomlinkCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasCustomlinkCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasCustomlinkCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
