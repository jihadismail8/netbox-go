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
	vpnIkepolicyCachePrefixKey = "vpnIkepolicy:"
	// VpnIkepolicyExpireTime expire time
	VpnIkepolicyExpireTime = 5 * time.Minute
)

var _ VpnIkepolicyCache = (*vpnIkepolicyCache)(nil)

// VpnIkepolicyCache cache interface
type VpnIkepolicyCache interface {
	Set(ctx context.Context, id uint64, data *model.VpnIkepolicy, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VpnIkepolicy, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnIkepolicy, error)
	MultiSet(ctx context.Context, data []*model.VpnIkepolicy, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// vpnIkepolicyCache define a cache struct
type vpnIkepolicyCache struct {
	cache cache.Cache
}

// NewVpnIkepolicyCache new a cache
func NewVpnIkepolicyCache(cacheType *database.CacheType) VpnIkepolicyCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnIkepolicy{}
		})
		return &vpnIkepolicyCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnIkepolicy{}
		})
		return &vpnIkepolicyCache{cache: c}
	}

	return nil // no cache
}

// GetVpnIkepolicyCacheKey cache key
func (c *vpnIkepolicyCache) GetVpnIkepolicyCacheKey(id uint64) string {
	return vpnIkepolicyCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *vpnIkepolicyCache) Set(ctx context.Context, id uint64, data *model.VpnIkepolicy, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVpnIkepolicyCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *vpnIkepolicyCache) Get(ctx context.Context, id uint64) (*model.VpnIkepolicy, error) {
	var data *model.VpnIkepolicy
	cacheKey := c.GetVpnIkepolicyCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *vpnIkepolicyCache) MultiSet(ctx context.Context, data []*model.VpnIkepolicy, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVpnIkepolicyCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *vpnIkepolicyCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnIkepolicy, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVpnIkepolicyCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VpnIkepolicy)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VpnIkepolicy)
	for _, id := range ids {
		val, ok := itemMap[c.GetVpnIkepolicyCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *vpnIkepolicyCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnIkepolicyCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *vpnIkepolicyCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnIkepolicyCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *vpnIkepolicyCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
