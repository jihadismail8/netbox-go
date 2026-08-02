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
	vpnL2VpnExportTargetsCachePrefixKey = "vpnL2VpnExportTargets:"
	// VpnL2VpnExportTargetsExpireTime expire time
	VpnL2VpnExportTargetsExpireTime = 5 * time.Minute
)

var _ VpnL2VpnExportTargetsCache = (*vpnL2VpnExportTargetsCache)(nil)

// VpnL2VpnExportTargetsCache cache interface
type VpnL2VpnExportTargetsCache interface {
	Set(ctx context.Context, id uint64, data *model.VpnL2VpnExportTargets, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VpnL2VpnExportTargets, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnL2VpnExportTargets, error)
	MultiSet(ctx context.Context, data []*model.VpnL2VpnExportTargets, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// vpnL2VpnExportTargetsCache define a cache struct
type vpnL2VpnExportTargetsCache struct {
	cache cache.Cache
}

// NewVpnL2VpnExportTargetsCache new a cache
func NewVpnL2VpnExportTargetsCache(cacheType *database.CacheType) VpnL2VpnExportTargetsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnL2VpnExportTargets{}
		})
		return &vpnL2VpnExportTargetsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnL2VpnExportTargets{}
		})
		return &vpnL2VpnExportTargetsCache{cache: c}
	}

	return nil // no cache
}

// GetVpnL2VpnExportTargetsCacheKey cache key
func (c *vpnL2VpnExportTargetsCache) GetVpnL2VpnExportTargetsCacheKey(id uint64) string {
	return vpnL2VpnExportTargetsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *vpnL2VpnExportTargetsCache) Set(ctx context.Context, id uint64, data *model.VpnL2VpnExportTargets, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVpnL2VpnExportTargetsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *vpnL2VpnExportTargetsCache) Get(ctx context.Context, id uint64) (*model.VpnL2VpnExportTargets, error) {
	var data *model.VpnL2VpnExportTargets
	cacheKey := c.GetVpnL2VpnExportTargetsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *vpnL2VpnExportTargetsCache) MultiSet(ctx context.Context, data []*model.VpnL2VpnExportTargets, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVpnL2VpnExportTargetsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *vpnL2VpnExportTargetsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnL2VpnExportTargets, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVpnL2VpnExportTargetsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VpnL2VpnExportTargets)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VpnL2VpnExportTargets)
	for _, id := range ids {
		val, ok := itemMap[c.GetVpnL2VpnExportTargetsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *vpnL2VpnExportTargetsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnL2VpnExportTargetsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *vpnL2VpnExportTargetsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnL2VpnExportTargetsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *vpnL2VpnExportTargetsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
