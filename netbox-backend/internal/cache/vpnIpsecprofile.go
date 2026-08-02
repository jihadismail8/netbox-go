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
	vpnIpsecprofileCachePrefixKey = "vpnIpsecprofile:"
	// VpnIpsecprofileExpireTime expire time
	VpnIpsecprofileExpireTime = 5 * time.Minute
)

var _ VpnIpsecprofileCache = (*vpnIpsecprofileCache)(nil)

// VpnIpsecprofileCache cache interface
type VpnIpsecprofileCache interface {
	Set(ctx context.Context, id uint64, data *model.VpnIpsecprofile, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VpnIpsecprofile, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnIpsecprofile, error)
	MultiSet(ctx context.Context, data []*model.VpnIpsecprofile, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// vpnIpsecprofileCache define a cache struct
type vpnIpsecprofileCache struct {
	cache cache.Cache
}

// NewVpnIpsecprofileCache new a cache
func NewVpnIpsecprofileCache(cacheType *database.CacheType) VpnIpsecprofileCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnIpsecprofile{}
		})
		return &vpnIpsecprofileCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnIpsecprofile{}
		})
		return &vpnIpsecprofileCache{cache: c}
	}

	return nil // no cache
}

// GetVpnIpsecprofileCacheKey cache key
func (c *vpnIpsecprofileCache) GetVpnIpsecprofileCacheKey(id uint64) string {
	return vpnIpsecprofileCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *vpnIpsecprofileCache) Set(ctx context.Context, id uint64, data *model.VpnIpsecprofile, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVpnIpsecprofileCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *vpnIpsecprofileCache) Get(ctx context.Context, id uint64) (*model.VpnIpsecprofile, error) {
	var data *model.VpnIpsecprofile
	cacheKey := c.GetVpnIpsecprofileCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *vpnIpsecprofileCache) MultiSet(ctx context.Context, data []*model.VpnIpsecprofile, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVpnIpsecprofileCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *vpnIpsecprofileCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnIpsecprofile, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVpnIpsecprofileCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VpnIpsecprofile)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VpnIpsecprofile)
	for _, id := range ids {
		val, ok := itemMap[c.GetVpnIpsecprofileCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *vpnIpsecprofileCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnIpsecprofileCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *vpnIpsecprofileCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnIpsecprofileCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *vpnIpsecprofileCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
