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
	virtualizationClustergroupCachePrefixKey = "virtualizationClustergroup:"
	// VirtualizationClustergroupExpireTime expire time
	VirtualizationClustergroupExpireTime = 5 * time.Minute
)

var _ VirtualizationClustergroupCache = (*virtualizationClustergroupCache)(nil)

// VirtualizationClustergroupCache cache interface
type VirtualizationClustergroupCache interface {
	Set(ctx context.Context, id uint64, data *model.VirtualizationClustergroup, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VirtualizationClustergroup, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VirtualizationClustergroup, error)
	MultiSet(ctx context.Context, data []*model.VirtualizationClustergroup, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// virtualizationClustergroupCache define a cache struct
type virtualizationClustergroupCache struct {
	cache cache.Cache
}

// NewVirtualizationClustergroupCache new a cache
func NewVirtualizationClustergroupCache(cacheType *database.CacheType) VirtualizationClustergroupCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VirtualizationClustergroup{}
		})
		return &virtualizationClustergroupCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VirtualizationClustergroup{}
		})
		return &virtualizationClustergroupCache{cache: c}
	}

	return nil // no cache
}

// GetVirtualizationClustergroupCacheKey cache key
func (c *virtualizationClustergroupCache) GetVirtualizationClustergroupCacheKey(id uint64) string {
	return virtualizationClustergroupCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *virtualizationClustergroupCache) Set(ctx context.Context, id uint64, data *model.VirtualizationClustergroup, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVirtualizationClustergroupCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *virtualizationClustergroupCache) Get(ctx context.Context, id uint64) (*model.VirtualizationClustergroup, error) {
	var data *model.VirtualizationClustergroup
	cacheKey := c.GetVirtualizationClustergroupCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *virtualizationClustergroupCache) MultiSet(ctx context.Context, data []*model.VirtualizationClustergroup, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVirtualizationClustergroupCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *virtualizationClustergroupCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VirtualizationClustergroup, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVirtualizationClustergroupCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VirtualizationClustergroup)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VirtualizationClustergroup)
	for _, id := range ids {
		val, ok := itemMap[c.GetVirtualizationClustergroupCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *virtualizationClustergroupCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVirtualizationClustergroupCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *virtualizationClustergroupCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVirtualizationClustergroupCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *virtualizationClustergroupCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
