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
	virtualizationVminterfaceCachePrefixKey = "virtualizationVminterface:"
	// VirtualizationVminterfaceExpireTime expire time
	VirtualizationVminterfaceExpireTime = 5 * time.Minute
)

var _ VirtualizationVminterfaceCache = (*virtualizationVminterfaceCache)(nil)

// VirtualizationVminterfaceCache cache interface
type VirtualizationVminterfaceCache interface {
	Set(ctx context.Context, id uint64, data *model.VirtualizationVminterface, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VirtualizationVminterface, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VirtualizationVminterface, error)
	MultiSet(ctx context.Context, data []*model.VirtualizationVminterface, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// virtualizationVminterfaceCache define a cache struct
type virtualizationVminterfaceCache struct {
	cache cache.Cache
}

// NewVirtualizationVminterfaceCache new a cache
func NewVirtualizationVminterfaceCache(cacheType *database.CacheType) VirtualizationVminterfaceCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VirtualizationVminterface{}
		})
		return &virtualizationVminterfaceCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VirtualizationVminterface{}
		})
		return &virtualizationVminterfaceCache{cache: c}
	}

	return nil // no cache
}

// GetVirtualizationVminterfaceCacheKey cache key
func (c *virtualizationVminterfaceCache) GetVirtualizationVminterfaceCacheKey(id uint64) string {
	return virtualizationVminterfaceCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *virtualizationVminterfaceCache) Set(ctx context.Context, id uint64, data *model.VirtualizationVminterface, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVirtualizationVminterfaceCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *virtualizationVminterfaceCache) Get(ctx context.Context, id uint64) (*model.VirtualizationVminterface, error) {
	var data *model.VirtualizationVminterface
	cacheKey := c.GetVirtualizationVminterfaceCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *virtualizationVminterfaceCache) MultiSet(ctx context.Context, data []*model.VirtualizationVminterface, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVirtualizationVminterfaceCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *virtualizationVminterfaceCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VirtualizationVminterface, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVirtualizationVminterfaceCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VirtualizationVminterface)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VirtualizationVminterface)
	for _, id := range ids {
		val, ok := itemMap[c.GetVirtualizationVminterfaceCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *virtualizationVminterfaceCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVirtualizationVminterfaceCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *virtualizationVminterfaceCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVirtualizationVminterfaceCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *virtualizationVminterfaceCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
