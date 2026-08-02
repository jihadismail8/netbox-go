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
	extrasConfigcontextSitesCachePrefixKey = "extrasConfigcontextSites:"
	// ExtrasConfigcontextSitesExpireTime expire time
	ExtrasConfigcontextSitesExpireTime = 5 * time.Minute
)

var _ ExtrasConfigcontextSitesCache = (*extrasConfigcontextSitesCache)(nil)

// ExtrasConfigcontextSitesCache cache interface
type ExtrasConfigcontextSitesCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextSites, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextSites, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextSites, error)
	MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextSites, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasConfigcontextSitesCache define a cache struct
type extrasConfigcontextSitesCache struct {
	cache cache.Cache
}

// NewExtrasConfigcontextSitesCache new a cache
func NewExtrasConfigcontextSitesCache(cacheType *database.CacheType) ExtrasConfigcontextSitesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextSites{}
		})
		return &extrasConfigcontextSitesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextSites{}
		})
		return &extrasConfigcontextSitesCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasConfigcontextSitesCacheKey cache key
func (c *extrasConfigcontextSitesCache) GetExtrasConfigcontextSitesCacheKey(id uint64) string {
	return extrasConfigcontextSitesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasConfigcontextSitesCache) Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextSites, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasConfigcontextSitesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasConfigcontextSitesCache) Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextSites, error) {
	var data *model.ExtrasConfigcontextSites
	cacheKey := c.GetExtrasConfigcontextSitesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasConfigcontextSitesCache) MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextSites, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasConfigcontextSitesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasConfigcontextSitesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextSites, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasConfigcontextSitesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasConfigcontextSites)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasConfigcontextSites)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasConfigcontextSitesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasConfigcontextSitesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextSitesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasConfigcontextSitesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextSitesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasConfigcontextSitesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
