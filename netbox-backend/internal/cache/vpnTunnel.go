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
	vpnTunnelCachePrefixKey = "vpnTunnel:"
	// VpnTunnelExpireTime expire time
	VpnTunnelExpireTime = 5 * time.Minute
)

var _ VpnTunnelCache = (*vpnTunnelCache)(nil)

// VpnTunnelCache cache interface
type VpnTunnelCache interface {
	Set(ctx context.Context, id uint64, data *model.VpnTunnel, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VpnTunnel, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnTunnel, error)
	MultiSet(ctx context.Context, data []*model.VpnTunnel, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// vpnTunnelCache define a cache struct
type vpnTunnelCache struct {
	cache cache.Cache
}

// NewVpnTunnelCache new a cache
func NewVpnTunnelCache(cacheType *database.CacheType) VpnTunnelCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnTunnel{}
		})
		return &vpnTunnelCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnTunnel{}
		})
		return &vpnTunnelCache{cache: c}
	}

	return nil // no cache
}

// GetVpnTunnelCacheKey cache key
func (c *vpnTunnelCache) GetVpnTunnelCacheKey(id uint64) string {
	return vpnTunnelCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *vpnTunnelCache) Set(ctx context.Context, id uint64, data *model.VpnTunnel, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVpnTunnelCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *vpnTunnelCache) Get(ctx context.Context, id uint64) (*model.VpnTunnel, error) {
	var data *model.VpnTunnel
	cacheKey := c.GetVpnTunnelCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *vpnTunnelCache) MultiSet(ctx context.Context, data []*model.VpnTunnel, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVpnTunnelCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *vpnTunnelCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnTunnel, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVpnTunnelCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VpnTunnel)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VpnTunnel)
	for _, id := range ids {
		val, ok := itemMap[c.GetVpnTunnelCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *vpnTunnelCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnTunnelCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *vpnTunnelCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnTunnelCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *vpnTunnelCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
