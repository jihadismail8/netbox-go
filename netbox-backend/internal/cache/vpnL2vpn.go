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
	vpnL2VpnCachePrefixKey = "vpnL2Vpn:"
	// VpnL2VpnExpireTime expire time
	VpnL2VpnExpireTime = 5 * time.Minute
)

var _ VpnL2VpnCache = (*vpnL2VpnCache)(nil)

// VpnL2VpnCache cache interface
type VpnL2VpnCache interface {
	Set(ctx context.Context, id uint64, data *model.VpnL2Vpn, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VpnL2Vpn, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnL2Vpn, error)
	MultiSet(ctx context.Context, data []*model.VpnL2Vpn, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// vpnL2VpnCache define a cache struct
type vpnL2VpnCache struct {
	cache cache.Cache
}

// NewVpnL2VpnCache new a cache
func NewVpnL2VpnCache(cacheType *database.CacheType) VpnL2VpnCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnL2Vpn{}
		})
		return &vpnL2VpnCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnL2Vpn{}
		})
		return &vpnL2VpnCache{cache: c}
	}

	return nil // no cache
}

// GetVpnL2VpnCacheKey cache key
func (c *vpnL2VpnCache) GetVpnL2VpnCacheKey(id uint64) string {
	return vpnL2VpnCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *vpnL2VpnCache) Set(ctx context.Context, id uint64, data *model.VpnL2Vpn, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVpnL2VpnCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *vpnL2VpnCache) Get(ctx context.Context, id uint64) (*model.VpnL2Vpn, error) {
	var data *model.VpnL2Vpn
	cacheKey := c.GetVpnL2VpnCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *vpnL2VpnCache) MultiSet(ctx context.Context, data []*model.VpnL2Vpn, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVpnL2VpnCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *vpnL2VpnCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnL2Vpn, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVpnL2VpnCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VpnL2Vpn)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VpnL2Vpn)
	for _, id := range ids {
		val, ok := itemMap[c.GetVpnL2VpnCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *vpnL2VpnCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnL2VpnCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *vpnL2VpnCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnL2VpnCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *vpnL2VpnCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
