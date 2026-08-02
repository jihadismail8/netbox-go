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
	extrasExporttemplateCachePrefixKey = "extrasExporttemplate:"
	// ExtrasExporttemplateExpireTime expire time
	ExtrasExporttemplateExpireTime = 5 * time.Minute
)

var _ ExtrasExporttemplateCache = (*extrasExporttemplateCache)(nil)

// ExtrasExporttemplateCache cache interface
type ExtrasExporttemplateCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasExporttemplate, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasExporttemplate, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasExporttemplate, error)
	MultiSet(ctx context.Context, data []*model.ExtrasExporttemplate, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasExporttemplateCache define a cache struct
type extrasExporttemplateCache struct {
	cache cache.Cache
}

// NewExtrasExporttemplateCache new a cache
func NewExtrasExporttemplateCache(cacheType *database.CacheType) ExtrasExporttemplateCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasExporttemplate{}
		})
		return &extrasExporttemplateCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasExporttemplate{}
		})
		return &extrasExporttemplateCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasExporttemplateCacheKey cache key
func (c *extrasExporttemplateCache) GetExtrasExporttemplateCacheKey(id uint64) string {
	return extrasExporttemplateCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasExporttemplateCache) Set(ctx context.Context, id uint64, data *model.ExtrasExporttemplate, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasExporttemplateCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasExporttemplateCache) Get(ctx context.Context, id uint64) (*model.ExtrasExporttemplate, error) {
	var data *model.ExtrasExporttemplate
	cacheKey := c.GetExtrasExporttemplateCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasExporttemplateCache) MultiSet(ctx context.Context, data []*model.ExtrasExporttemplate, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasExporttemplateCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasExporttemplateCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasExporttemplate, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasExporttemplateCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasExporttemplate)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasExporttemplate)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasExporttemplateCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasExporttemplateCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasExporttemplateCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasExporttemplateCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasExporttemplateCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasExporttemplateCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
