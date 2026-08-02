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
	vpnTunnelterminationCachePrefixKey = "vpnTunneltermination:"
	// VpnTunnelterminationExpireTime expire time
	VpnTunnelterminationExpireTime = 5 * time.Minute
)

var _ VpnTunnelterminationCache = (*vpnTunnelterminationCache)(nil)

// VpnTunnelterminationCache cache interface
type VpnTunnelterminationCache interface {
	Set(ctx context.Context, id uint64, data *model.VpnTunneltermination, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VpnTunneltermination, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnTunneltermination, error)
	MultiSet(ctx context.Context, data []*model.VpnTunneltermination, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// vpnTunnelterminationCache define a cache struct
type vpnTunnelterminationCache struct {
	cache cache.Cache
}

// NewVpnTunnelterminationCache new a cache
func NewVpnTunnelterminationCache(cacheType *database.CacheType) VpnTunnelterminationCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnTunneltermination{}
		})
		return &vpnTunnelterminationCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnTunneltermination{}
		})
		return &vpnTunnelterminationCache{cache: c}
	}

	return nil // no cache
}

// GetVpnTunnelterminationCacheKey cache key
func (c *vpnTunnelterminationCache) GetVpnTunnelterminationCacheKey(id uint64) string {
	return vpnTunnelterminationCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *vpnTunnelterminationCache) Set(ctx context.Context, id uint64, data *model.VpnTunneltermination, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVpnTunnelterminationCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *vpnTunnelterminationCache) Get(ctx context.Context, id uint64) (*model.VpnTunneltermination, error) {
	var data *model.VpnTunneltermination
	cacheKey := c.GetVpnTunnelterminationCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *vpnTunnelterminationCache) MultiSet(ctx context.Context, data []*model.VpnTunneltermination, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVpnTunnelterminationCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *vpnTunnelterminationCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnTunneltermination, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVpnTunnelterminationCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VpnTunneltermination)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VpnTunneltermination)
	for _, id := range ids {
		val, ok := itemMap[c.GetVpnTunnelterminationCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *vpnTunnelterminationCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnTunnelterminationCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *vpnTunnelterminationCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnTunnelterminationCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *vpnTunnelterminationCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
