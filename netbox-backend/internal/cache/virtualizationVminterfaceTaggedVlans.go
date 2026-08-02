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
	virtualizationVminterfaceTaggedVlansCachePrefixKey = "virtualizationVminterfaceTaggedVlans:"
	// VirtualizationVminterfaceTaggedVlansExpireTime expire time
	VirtualizationVminterfaceTaggedVlansExpireTime = 5 * time.Minute
)

var _ VirtualizationVminterfaceTaggedVlansCache = (*virtualizationVminterfaceTaggedVlansCache)(nil)

// VirtualizationVminterfaceTaggedVlansCache cache interface
type VirtualizationVminterfaceTaggedVlansCache interface {
	Set(ctx context.Context, id uint64, data *model.VirtualizationVminterfaceTaggedVlans, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VirtualizationVminterfaceTaggedVlans, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VirtualizationVminterfaceTaggedVlans, error)
	MultiSet(ctx context.Context, data []*model.VirtualizationVminterfaceTaggedVlans, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// virtualizationVminterfaceTaggedVlansCache define a cache struct
type virtualizationVminterfaceTaggedVlansCache struct {
	cache cache.Cache
}

// NewVirtualizationVminterfaceTaggedVlansCache new a cache
func NewVirtualizationVminterfaceTaggedVlansCache(cacheType *database.CacheType) VirtualizationVminterfaceTaggedVlansCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VirtualizationVminterfaceTaggedVlans{}
		})
		return &virtualizationVminterfaceTaggedVlansCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VirtualizationVminterfaceTaggedVlans{}
		})
		return &virtualizationVminterfaceTaggedVlansCache{cache: c}
	}

	return nil // no cache
}

// GetVirtualizationVminterfaceTaggedVlansCacheKey cache key
func (c *virtualizationVminterfaceTaggedVlansCache) GetVirtualizationVminterfaceTaggedVlansCacheKey(id uint64) string {
	return virtualizationVminterfaceTaggedVlansCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *virtualizationVminterfaceTaggedVlansCache) Set(ctx context.Context, id uint64, data *model.VirtualizationVminterfaceTaggedVlans, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVirtualizationVminterfaceTaggedVlansCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *virtualizationVminterfaceTaggedVlansCache) Get(ctx context.Context, id uint64) (*model.VirtualizationVminterfaceTaggedVlans, error) {
	var data *model.VirtualizationVminterfaceTaggedVlans
	cacheKey := c.GetVirtualizationVminterfaceTaggedVlansCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *virtualizationVminterfaceTaggedVlansCache) MultiSet(ctx context.Context, data []*model.VirtualizationVminterfaceTaggedVlans, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVirtualizationVminterfaceTaggedVlansCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *virtualizationVminterfaceTaggedVlansCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VirtualizationVminterfaceTaggedVlans, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVirtualizationVminterfaceTaggedVlansCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VirtualizationVminterfaceTaggedVlans)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VirtualizationVminterfaceTaggedVlans)
	for _, id := range ids {
		val, ok := itemMap[c.GetVirtualizationVminterfaceTaggedVlansCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *virtualizationVminterfaceTaggedVlansCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVirtualizationVminterfaceTaggedVlansCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *virtualizationVminterfaceTaggedVlansCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVirtualizationVminterfaceTaggedVlansCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *virtualizationVminterfaceTaggedVlansCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
