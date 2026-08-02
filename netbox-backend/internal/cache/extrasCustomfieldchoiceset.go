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
	extrasCustomfieldchoicesetCachePrefixKey = "extrasCustomfieldchoiceset:"
	// ExtrasCustomfieldchoicesetExpireTime expire time
	ExtrasCustomfieldchoicesetExpireTime = 5 * time.Minute
)

var _ ExtrasCustomfieldchoicesetCache = (*extrasCustomfieldchoicesetCache)(nil)

// ExtrasCustomfieldchoicesetCache cache interface
type ExtrasCustomfieldchoicesetCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasCustomfieldchoiceset, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasCustomfieldchoiceset, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasCustomfieldchoiceset, error)
	MultiSet(ctx context.Context, data []*model.ExtrasCustomfieldchoiceset, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasCustomfieldchoicesetCache define a cache struct
type extrasCustomfieldchoicesetCache struct {
	cache cache.Cache
}

// NewExtrasCustomfieldchoicesetCache new a cache
func NewExtrasCustomfieldchoicesetCache(cacheType *database.CacheType) ExtrasCustomfieldchoicesetCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasCustomfieldchoiceset{}
		})
		return &extrasCustomfieldchoicesetCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasCustomfieldchoiceset{}
		})
		return &extrasCustomfieldchoicesetCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasCustomfieldchoicesetCacheKey cache key
func (c *extrasCustomfieldchoicesetCache) GetExtrasCustomfieldchoicesetCacheKey(id uint64) string {
	return extrasCustomfieldchoicesetCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasCustomfieldchoicesetCache) Set(ctx context.Context, id uint64, data *model.ExtrasCustomfieldchoiceset, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasCustomfieldchoicesetCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasCustomfieldchoicesetCache) Get(ctx context.Context, id uint64) (*model.ExtrasCustomfieldchoiceset, error) {
	var data *model.ExtrasCustomfieldchoiceset
	cacheKey := c.GetExtrasCustomfieldchoicesetCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasCustomfieldchoicesetCache) MultiSet(ctx context.Context, data []*model.ExtrasCustomfieldchoiceset, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasCustomfieldchoicesetCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasCustomfieldchoicesetCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasCustomfieldchoiceset, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasCustomfieldchoicesetCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasCustomfieldchoiceset)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasCustomfieldchoiceset)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasCustomfieldchoicesetCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasCustomfieldchoicesetCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasCustomfieldchoicesetCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasCustomfieldchoicesetCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasCustomfieldchoicesetCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasCustomfieldchoicesetCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
