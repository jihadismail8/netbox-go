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
	extrasConfigcontextTagsCachePrefixKey = "extrasConfigcontextTags:"
	// ExtrasConfigcontextTagsExpireTime expire time
	ExtrasConfigcontextTagsExpireTime = 5 * time.Minute
)

var _ ExtrasConfigcontextTagsCache = (*extrasConfigcontextTagsCache)(nil)

// ExtrasConfigcontextTagsCache cache interface
type ExtrasConfigcontextTagsCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextTags, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextTags, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextTags, error)
	MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextTags, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasConfigcontextTagsCache define a cache struct
type extrasConfigcontextTagsCache struct {
	cache cache.Cache
}

// NewExtrasConfigcontextTagsCache new a cache
func NewExtrasConfigcontextTagsCache(cacheType *database.CacheType) ExtrasConfigcontextTagsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextTags{}
		})
		return &extrasConfigcontextTagsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextTags{}
		})
		return &extrasConfigcontextTagsCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasConfigcontextTagsCacheKey cache key
func (c *extrasConfigcontextTagsCache) GetExtrasConfigcontextTagsCacheKey(id uint64) string {
	return extrasConfigcontextTagsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasConfigcontextTagsCache) Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextTags, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasConfigcontextTagsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasConfigcontextTagsCache) Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextTags, error) {
	var data *model.ExtrasConfigcontextTags
	cacheKey := c.GetExtrasConfigcontextTagsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasConfigcontextTagsCache) MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextTags, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasConfigcontextTagsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasConfigcontextTagsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextTags, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasConfigcontextTagsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasConfigcontextTags)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasConfigcontextTags)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasConfigcontextTagsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasConfigcontextTagsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextTagsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasConfigcontextTagsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextTagsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasConfigcontextTagsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
