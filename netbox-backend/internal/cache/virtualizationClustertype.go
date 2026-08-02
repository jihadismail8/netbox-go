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
	virtualizationClustertypeCachePrefixKey = "virtualizationClustertype:"
	// VirtualizationClustertypeExpireTime expire time
	VirtualizationClustertypeExpireTime = 5 * time.Minute
)

var _ VirtualizationClustertypeCache = (*virtualizationClustertypeCache)(nil)

// VirtualizationClustertypeCache cache interface
type VirtualizationClustertypeCache interface {
	Set(ctx context.Context, id uint64, data *model.VirtualizationClustertype, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VirtualizationClustertype, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VirtualizationClustertype, error)
	MultiSet(ctx context.Context, data []*model.VirtualizationClustertype, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// virtualizationClustertypeCache define a cache struct
type virtualizationClustertypeCache struct {
	cache cache.Cache
}

// NewVirtualizationClustertypeCache new a cache
func NewVirtualizationClustertypeCache(cacheType *database.CacheType) VirtualizationClustertypeCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VirtualizationClustertype{}
		})
		return &virtualizationClustertypeCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VirtualizationClustertype{}
		})
		return &virtualizationClustertypeCache{cache: c}
	}

	return nil // no cache
}

// GetVirtualizationClustertypeCacheKey cache key
func (c *virtualizationClustertypeCache) GetVirtualizationClustertypeCacheKey(id uint64) string {
	return virtualizationClustertypeCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *virtualizationClustertypeCache) Set(ctx context.Context, id uint64, data *model.VirtualizationClustertype, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVirtualizationClustertypeCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *virtualizationClustertypeCache) Get(ctx context.Context, id uint64) (*model.VirtualizationClustertype, error) {
	var data *model.VirtualizationClustertype
	cacheKey := c.GetVirtualizationClustertypeCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *virtualizationClustertypeCache) MultiSet(ctx context.Context, data []*model.VirtualizationClustertype, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVirtualizationClustertypeCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *virtualizationClustertypeCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VirtualizationClustertype, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVirtualizationClustertypeCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VirtualizationClustertype)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VirtualizationClustertype)
	for _, id := range ids {
		val, ok := itemMap[c.GetVirtualizationClustertypeCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *virtualizationClustertypeCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVirtualizationClustertypeCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *virtualizationClustertypeCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVirtualizationClustertypeCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *virtualizationClustertypeCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
