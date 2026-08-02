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
	vpnIpsecpolicyCachePrefixKey = "vpnIpsecpolicy:"
	// VpnIpsecpolicyExpireTime expire time
	VpnIpsecpolicyExpireTime = 5 * time.Minute
)

var _ VpnIpsecpolicyCache = (*vpnIpsecpolicyCache)(nil)

// VpnIpsecpolicyCache cache interface
type VpnIpsecpolicyCache interface {
	Set(ctx context.Context, id uint64, data *model.VpnIpsecpolicy, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VpnIpsecpolicy, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnIpsecpolicy, error)
	MultiSet(ctx context.Context, data []*model.VpnIpsecpolicy, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// vpnIpsecpolicyCache define a cache struct
type vpnIpsecpolicyCache struct {
	cache cache.Cache
}

// NewVpnIpsecpolicyCache new a cache
func NewVpnIpsecpolicyCache(cacheType *database.CacheType) VpnIpsecpolicyCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnIpsecpolicy{}
		})
		return &vpnIpsecpolicyCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnIpsecpolicy{}
		})
		return &vpnIpsecpolicyCache{cache: c}
	}

	return nil // no cache
}

// GetVpnIpsecpolicyCacheKey cache key
func (c *vpnIpsecpolicyCache) GetVpnIpsecpolicyCacheKey(id uint64) string {
	return vpnIpsecpolicyCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *vpnIpsecpolicyCache) Set(ctx context.Context, id uint64, data *model.VpnIpsecpolicy, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVpnIpsecpolicyCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *vpnIpsecpolicyCache) Get(ctx context.Context, id uint64) (*model.VpnIpsecpolicy, error) {
	var data *model.VpnIpsecpolicy
	cacheKey := c.GetVpnIpsecpolicyCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *vpnIpsecpolicyCache) MultiSet(ctx context.Context, data []*model.VpnIpsecpolicy, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVpnIpsecpolicyCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *vpnIpsecpolicyCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnIpsecpolicy, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVpnIpsecpolicyCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VpnIpsecpolicy)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VpnIpsecpolicy)
	for _, id := range ids {
		val, ok := itemMap[c.GetVpnIpsecpolicyCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *vpnIpsecpolicyCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnIpsecpolicyCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *vpnIpsecpolicyCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnIpsecpolicyCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *vpnIpsecpolicyCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
