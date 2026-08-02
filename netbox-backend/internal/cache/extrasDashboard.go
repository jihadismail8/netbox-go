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
	extrasDashboardCachePrefixKey = "extrasDashboard:"
	// ExtrasDashboardExpireTime expire time
	ExtrasDashboardExpireTime = 5 * time.Minute
)

var _ ExtrasDashboardCache = (*extrasDashboardCache)(nil)

// ExtrasDashboardCache cache interface
type ExtrasDashboardCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasDashboard, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasDashboard, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasDashboard, error)
	MultiSet(ctx context.Context, data []*model.ExtrasDashboard, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasDashboardCache define a cache struct
type extrasDashboardCache struct {
	cache cache.Cache
}

// NewExtrasDashboardCache new a cache
func NewExtrasDashboardCache(cacheType *database.CacheType) ExtrasDashboardCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasDashboard{}
		})
		return &extrasDashboardCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasDashboard{}
		})
		return &extrasDashboardCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasDashboardCacheKey cache key
func (c *extrasDashboardCache) GetExtrasDashboardCacheKey(id uint64) string {
	return extrasDashboardCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasDashboardCache) Set(ctx context.Context, id uint64, data *model.ExtrasDashboard, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasDashboardCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasDashboardCache) Get(ctx context.Context, id uint64) (*model.ExtrasDashboard, error) {
	var data *model.ExtrasDashboard
	cacheKey := c.GetExtrasDashboardCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasDashboardCache) MultiSet(ctx context.Context, data []*model.ExtrasDashboard, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasDashboardCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasDashboardCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasDashboard, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasDashboardCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasDashboard)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasDashboard)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasDashboardCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasDashboardCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasDashboardCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasDashboardCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasDashboardCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasDashboardCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
