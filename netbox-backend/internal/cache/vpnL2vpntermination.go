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
	vpnL2VpnterminationCachePrefixKey = "vpnL2Vpntermination:"
	// VpnL2VpnterminationExpireTime expire time
	VpnL2VpnterminationExpireTime = 5 * time.Minute
)

var _ VpnL2VpnterminationCache = (*vpnL2VpnterminationCache)(nil)

// VpnL2VpnterminationCache cache interface
type VpnL2VpnterminationCache interface {
	Set(ctx context.Context, id uint64, data *model.VpnL2Vpntermination, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VpnL2Vpntermination, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnL2Vpntermination, error)
	MultiSet(ctx context.Context, data []*model.VpnL2Vpntermination, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// vpnL2VpnterminationCache define a cache struct
type vpnL2VpnterminationCache struct {
	cache cache.Cache
}

// NewVpnL2VpnterminationCache new a cache
func NewVpnL2VpnterminationCache(cacheType *database.CacheType) VpnL2VpnterminationCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnL2Vpntermination{}
		})
		return &vpnL2VpnterminationCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnL2Vpntermination{}
		})
		return &vpnL2VpnterminationCache{cache: c}
	}

	return nil // no cache
}

// GetVpnL2VpnterminationCacheKey cache key
func (c *vpnL2VpnterminationCache) GetVpnL2VpnterminationCacheKey(id uint64) string {
	return vpnL2VpnterminationCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *vpnL2VpnterminationCache) Set(ctx context.Context, id uint64, data *model.VpnL2Vpntermination, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVpnL2VpnterminationCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *vpnL2VpnterminationCache) Get(ctx context.Context, id uint64) (*model.VpnL2Vpntermination, error) {
	var data *model.VpnL2Vpntermination
	cacheKey := c.GetVpnL2VpnterminationCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *vpnL2VpnterminationCache) MultiSet(ctx context.Context, data []*model.VpnL2Vpntermination, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVpnL2VpnterminationCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *vpnL2VpnterminationCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnL2Vpntermination, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVpnL2VpnterminationCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VpnL2Vpntermination)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VpnL2Vpntermination)
	for _, id := range ids {
		val, ok := itemMap[c.GetVpnL2VpnterminationCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *vpnL2VpnterminationCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnL2VpnterminationCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *vpnL2VpnterminationCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnL2VpnterminationCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *vpnL2VpnterminationCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
