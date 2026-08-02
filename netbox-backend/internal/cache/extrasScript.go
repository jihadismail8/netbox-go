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
	extrasScriptCachePrefixKey = "extrasScript:"
	// ExtrasScriptExpireTime expire time
	ExtrasScriptExpireTime = 5 * time.Minute
)

var _ ExtrasScriptCache = (*extrasScriptCache)(nil)

// ExtrasScriptCache cache interface
type ExtrasScriptCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasScript, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasScript, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasScript, error)
	MultiSet(ctx context.Context, data []*model.ExtrasScript, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasScriptCache define a cache struct
type extrasScriptCache struct {
	cache cache.Cache
}

// NewExtrasScriptCache new a cache
func NewExtrasScriptCache(cacheType *database.CacheType) ExtrasScriptCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasScript{}
		})
		return &extrasScriptCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasScript{}
		})
		return &extrasScriptCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasScriptCacheKey cache key
func (c *extrasScriptCache) GetExtrasScriptCacheKey(id uint64) string {
	return extrasScriptCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasScriptCache) Set(ctx context.Context, id uint64, data *model.ExtrasScript, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasScriptCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasScriptCache) Get(ctx context.Context, id uint64) (*model.ExtrasScript, error) {
	var data *model.ExtrasScript
	cacheKey := c.GetExtrasScriptCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasScriptCache) MultiSet(ctx context.Context, data []*model.ExtrasScript, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasScriptCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasScriptCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasScript, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasScriptCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasScript)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasScript)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasScriptCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasScriptCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasScriptCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasScriptCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasScriptCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasScriptCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
