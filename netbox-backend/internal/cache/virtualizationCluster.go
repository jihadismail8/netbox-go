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
	virtualizationClusterCachePrefixKey = "virtualizationCluster:"
	// VirtualizationClusterExpireTime expire time
	VirtualizationClusterExpireTime = 5 * time.Minute
)

var _ VirtualizationClusterCache = (*virtualizationClusterCache)(nil)

// VirtualizationClusterCache cache interface
type VirtualizationClusterCache interface {
	Set(ctx context.Context, id uint64, data *model.VirtualizationCluster, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VirtualizationCluster, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VirtualizationCluster, error)
	MultiSet(ctx context.Context, data []*model.VirtualizationCluster, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// virtualizationClusterCache define a cache struct
type virtualizationClusterCache struct {
	cache cache.Cache
}

// NewVirtualizationClusterCache new a cache
func NewVirtualizationClusterCache(cacheType *database.CacheType) VirtualizationClusterCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VirtualizationCluster{}
		})
		return &virtualizationClusterCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VirtualizationCluster{}
		})
		return &virtualizationClusterCache{cache: c}
	}

	return nil // no cache
}

// GetVirtualizationClusterCacheKey cache key
func (c *virtualizationClusterCache) GetVirtualizationClusterCacheKey(id uint64) string {
	return virtualizationClusterCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *virtualizationClusterCache) Set(ctx context.Context, id uint64, data *model.VirtualizationCluster, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVirtualizationClusterCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *virtualizationClusterCache) Get(ctx context.Context, id uint64) (*model.VirtualizationCluster, error) {
	var data *model.VirtualizationCluster
	cacheKey := c.GetVirtualizationClusterCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *virtualizationClusterCache) MultiSet(ctx context.Context, data []*model.VirtualizationCluster, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVirtualizationClusterCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *virtualizationClusterCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VirtualizationCluster, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVirtualizationClusterCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VirtualizationCluster)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VirtualizationCluster)
	for _, id := range ids {
		val, ok := itemMap[c.GetVirtualizationClusterCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *virtualizationClusterCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVirtualizationClusterCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *virtualizationClusterCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVirtualizationClusterCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *virtualizationClusterCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
