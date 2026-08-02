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
	extrasConfigtemplateCachePrefixKey = "extrasConfigtemplate:"
	// ExtrasConfigtemplateExpireTime expire time
	ExtrasConfigtemplateExpireTime = 5 * time.Minute
)

var _ ExtrasConfigtemplateCache = (*extrasConfigtemplateCache)(nil)

// ExtrasConfigtemplateCache cache interface
type ExtrasConfigtemplateCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasConfigtemplate, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasConfigtemplate, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigtemplate, error)
	MultiSet(ctx context.Context, data []*model.ExtrasConfigtemplate, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasConfigtemplateCache define a cache struct
type extrasConfigtemplateCache struct {
	cache cache.Cache
}

// NewExtrasConfigtemplateCache new a cache
func NewExtrasConfigtemplateCache(cacheType *database.CacheType) ExtrasConfigtemplateCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigtemplate{}
		})
		return &extrasConfigtemplateCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigtemplate{}
		})
		return &extrasConfigtemplateCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasConfigtemplateCacheKey cache key
func (c *extrasConfigtemplateCache) GetExtrasConfigtemplateCacheKey(id uint64) string {
	return extrasConfigtemplateCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasConfigtemplateCache) Set(ctx context.Context, id uint64, data *model.ExtrasConfigtemplate, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasConfigtemplateCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasConfigtemplateCache) Get(ctx context.Context, id uint64) (*model.ExtrasConfigtemplate, error) {
	var data *model.ExtrasConfigtemplate
	cacheKey := c.GetExtrasConfigtemplateCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasConfigtemplateCache) MultiSet(ctx context.Context, data []*model.ExtrasConfigtemplate, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasConfigtemplateCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasConfigtemplateCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigtemplate, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasConfigtemplateCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasConfigtemplate)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasConfigtemplate)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasConfigtemplateCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasConfigtemplateCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigtemplateCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasConfigtemplateCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigtemplateCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasConfigtemplateCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
