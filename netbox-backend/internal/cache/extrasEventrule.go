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
	extrasEventruleCachePrefixKey = "extrasEventrule:"
	// ExtrasEventruleExpireTime expire time
	ExtrasEventruleExpireTime = 5 * time.Minute
)

var _ ExtrasEventruleCache = (*extrasEventruleCache)(nil)

// ExtrasEventruleCache cache interface
type ExtrasEventruleCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasEventrule, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasEventrule, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasEventrule, error)
	MultiSet(ctx context.Context, data []*model.ExtrasEventrule, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasEventruleCache define a cache struct
type extrasEventruleCache struct {
	cache cache.Cache
}

// NewExtrasEventruleCache new a cache
func NewExtrasEventruleCache(cacheType *database.CacheType) ExtrasEventruleCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasEventrule{}
		})
		return &extrasEventruleCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasEventrule{}
		})
		return &extrasEventruleCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasEventruleCacheKey cache key
func (c *extrasEventruleCache) GetExtrasEventruleCacheKey(id uint64) string {
	return extrasEventruleCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasEventruleCache) Set(ctx context.Context, id uint64, data *model.ExtrasEventrule, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasEventruleCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasEventruleCache) Get(ctx context.Context, id uint64) (*model.ExtrasEventrule, error) {
	var data *model.ExtrasEventrule
	cacheKey := c.GetExtrasEventruleCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasEventruleCache) MultiSet(ctx context.Context, data []*model.ExtrasEventrule, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasEventruleCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasEventruleCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasEventrule, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasEventruleCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasEventrule)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasEventrule)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasEventruleCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasEventruleCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasEventruleCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasEventruleCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasEventruleCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasEventruleCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
