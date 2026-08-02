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
	virtualizationVirtualmachineCachePrefixKey = "virtualizationVirtualmachine:"
	// VirtualizationVirtualmachineExpireTime expire time
	VirtualizationVirtualmachineExpireTime = 5 * time.Minute
)

var _ VirtualizationVirtualmachineCache = (*virtualizationVirtualmachineCache)(nil)

// VirtualizationVirtualmachineCache cache interface
type VirtualizationVirtualmachineCache interface {
	Set(ctx context.Context, id uint64, data *model.VirtualizationVirtualmachine, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VirtualizationVirtualmachine, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VirtualizationVirtualmachine, error)
	MultiSet(ctx context.Context, data []*model.VirtualizationVirtualmachine, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// virtualizationVirtualmachineCache define a cache struct
type virtualizationVirtualmachineCache struct {
	cache cache.Cache
}

// NewVirtualizationVirtualmachineCache new a cache
func NewVirtualizationVirtualmachineCache(cacheType *database.CacheType) VirtualizationVirtualmachineCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VirtualizationVirtualmachine{}
		})
		return &virtualizationVirtualmachineCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VirtualizationVirtualmachine{}
		})
		return &virtualizationVirtualmachineCache{cache: c}
	}

	return nil // no cache
}

// GetVirtualizationVirtualmachineCacheKey cache key
func (c *virtualizationVirtualmachineCache) GetVirtualizationVirtualmachineCacheKey(id uint64) string {
	return virtualizationVirtualmachineCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *virtualizationVirtualmachineCache) Set(ctx context.Context, id uint64, data *model.VirtualizationVirtualmachine, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVirtualizationVirtualmachineCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *virtualizationVirtualmachineCache) Get(ctx context.Context, id uint64) (*model.VirtualizationVirtualmachine, error) {
	var data *model.VirtualizationVirtualmachine
	cacheKey := c.GetVirtualizationVirtualmachineCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *virtualizationVirtualmachineCache) MultiSet(ctx context.Context, data []*model.VirtualizationVirtualmachine, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVirtualizationVirtualmachineCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *virtualizationVirtualmachineCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VirtualizationVirtualmachine, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVirtualizationVirtualmachineCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VirtualizationVirtualmachine)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VirtualizationVirtualmachine)
	for _, id := range ids {
		val, ok := itemMap[c.GetVirtualizationVirtualmachineCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *virtualizationVirtualmachineCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVirtualizationVirtualmachineCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *virtualizationVirtualmachineCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVirtualizationVirtualmachineCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *virtualizationVirtualmachineCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
