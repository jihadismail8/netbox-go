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
	virtualizationVirtualdiskCachePrefixKey = "virtualizationVirtualdisk:"
	// VirtualizationVirtualdiskExpireTime expire time
	VirtualizationVirtualdiskExpireTime = 5 * time.Minute
)

var _ VirtualizationVirtualdiskCache = (*virtualizationVirtualdiskCache)(nil)

// VirtualizationVirtualdiskCache cache interface
type VirtualizationVirtualdiskCache interface {
	Set(ctx context.Context, id uint64, data *model.VirtualizationVirtualdisk, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VirtualizationVirtualdisk, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VirtualizationVirtualdisk, error)
	MultiSet(ctx context.Context, data []*model.VirtualizationVirtualdisk, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// virtualizationVirtualdiskCache define a cache struct
type virtualizationVirtualdiskCache struct {
	cache cache.Cache
}

// NewVirtualizationVirtualdiskCache new a cache
func NewVirtualizationVirtualdiskCache(cacheType *database.CacheType) VirtualizationVirtualdiskCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VirtualizationVirtualdisk{}
		})
		return &virtualizationVirtualdiskCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VirtualizationVirtualdisk{}
		})
		return &virtualizationVirtualdiskCache{cache: c}
	}

	return nil // no cache
}

// GetVirtualizationVirtualdiskCacheKey cache key
func (c *virtualizationVirtualdiskCache) GetVirtualizationVirtualdiskCacheKey(id uint64) string {
	return virtualizationVirtualdiskCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *virtualizationVirtualdiskCache) Set(ctx context.Context, id uint64, data *model.VirtualizationVirtualdisk, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVirtualizationVirtualdiskCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *virtualizationVirtualdiskCache) Get(ctx context.Context, id uint64) (*model.VirtualizationVirtualdisk, error) {
	var data *model.VirtualizationVirtualdisk
	cacheKey := c.GetVirtualizationVirtualdiskCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *virtualizationVirtualdiskCache) MultiSet(ctx context.Context, data []*model.VirtualizationVirtualdisk, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVirtualizationVirtualdiskCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *virtualizationVirtualdiskCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VirtualizationVirtualdisk, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVirtualizationVirtualdiskCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VirtualizationVirtualdisk)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VirtualizationVirtualdisk)
	for _, id := range ids {
		val, ok := itemMap[c.GetVirtualizationVirtualdiskCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *virtualizationVirtualdiskCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVirtualizationVirtualdiskCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *virtualizationVirtualdiskCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVirtualizationVirtualdiskCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *virtualizationVirtualdiskCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
