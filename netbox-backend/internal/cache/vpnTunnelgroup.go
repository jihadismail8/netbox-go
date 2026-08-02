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
	vpnTunnelgroupCachePrefixKey = "vpnTunnelgroup:"
	// VpnTunnelgroupExpireTime expire time
	VpnTunnelgroupExpireTime = 5 * time.Minute
)

var _ VpnTunnelgroupCache = (*vpnTunnelgroupCache)(nil)

// VpnTunnelgroupCache cache interface
type VpnTunnelgroupCache interface {
	Set(ctx context.Context, id uint64, data *model.VpnTunnelgroup, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VpnTunnelgroup, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnTunnelgroup, error)
	MultiSet(ctx context.Context, data []*model.VpnTunnelgroup, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// vpnTunnelgroupCache define a cache struct
type vpnTunnelgroupCache struct {
	cache cache.Cache
}

// NewVpnTunnelgroupCache new a cache
func NewVpnTunnelgroupCache(cacheType *database.CacheType) VpnTunnelgroupCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnTunnelgroup{}
		})
		return &vpnTunnelgroupCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnTunnelgroup{}
		})
		return &vpnTunnelgroupCache{cache: c}
	}

	return nil // no cache
}

// GetVpnTunnelgroupCacheKey cache key
func (c *vpnTunnelgroupCache) GetVpnTunnelgroupCacheKey(id uint64) string {
	return vpnTunnelgroupCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *vpnTunnelgroupCache) Set(ctx context.Context, id uint64, data *model.VpnTunnelgroup, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVpnTunnelgroupCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *vpnTunnelgroupCache) Get(ctx context.Context, id uint64) (*model.VpnTunnelgroup, error) {
	var data *model.VpnTunnelgroup
	cacheKey := c.GetVpnTunnelgroupCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *vpnTunnelgroupCache) MultiSet(ctx context.Context, data []*model.VpnTunnelgroup, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVpnTunnelgroupCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *vpnTunnelgroupCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnTunnelgroup, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVpnTunnelgroupCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VpnTunnelgroup)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VpnTunnelgroup)
	for _, id := range ids {
		val, ok := itemMap[c.GetVpnTunnelgroupCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *vpnTunnelgroupCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnTunnelgroupCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *vpnTunnelgroupCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnTunnelgroupCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *vpnTunnelgroupCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
