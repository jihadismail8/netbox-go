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
	dcimRegionCachePrefixKey = "dcimRegion:"
	// DcimRegionExpireTime expire time
	DcimRegionExpireTime = 5 * time.Minute
)

var _ DcimRegionCache = (*dcimRegionCache)(nil)

// DcimRegionCache cache interface
type DcimRegionCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimRegion, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimRegion, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimRegion, error)
	MultiSet(ctx context.Context, data []*model.DcimRegion, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimRegionCache define a cache struct
type dcimRegionCache struct {
	cache cache.Cache
}

// NewDcimRegionCache new a cache
func NewDcimRegionCache(cacheType *database.CacheType) DcimRegionCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimRegion{}
		})
		return &dcimRegionCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimRegion{}
		})
		return &dcimRegionCache{cache: c}
	}

	return nil // no cache
}

// GetDcimRegionCacheKey cache key
func (c *dcimRegionCache) GetDcimRegionCacheKey(id uint64) string {
	return dcimRegionCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimRegionCache) Set(ctx context.Context, id uint64, data *model.DcimRegion, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimRegionCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimRegionCache) Get(ctx context.Context, id uint64) (*model.DcimRegion, error) {
	var data *model.DcimRegion
	cacheKey := c.GetDcimRegionCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimRegionCache) MultiSet(ctx context.Context, data []*model.DcimRegion, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimRegionCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimRegionCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimRegion, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimRegionCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimRegion)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimRegion)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimRegionCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimRegionCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimRegionCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimRegionCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimRegionCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimRegionCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
