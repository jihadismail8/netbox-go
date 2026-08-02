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
	vpnL2VpnImportTargetsCachePrefixKey = "vpnL2VpnImportTargets:"
	// VpnL2VpnImportTargetsExpireTime expire time
	VpnL2VpnImportTargetsExpireTime = 5 * time.Minute
)

var _ VpnL2VpnImportTargetsCache = (*vpnL2VpnImportTargetsCache)(nil)

// VpnL2VpnImportTargetsCache cache interface
type VpnL2VpnImportTargetsCache interface {
	Set(ctx context.Context, id uint64, data *model.VpnL2VpnImportTargets, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VpnL2VpnImportTargets, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnL2VpnImportTargets, error)
	MultiSet(ctx context.Context, data []*model.VpnL2VpnImportTargets, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// vpnL2VpnImportTargetsCache define a cache struct
type vpnL2VpnImportTargetsCache struct {
	cache cache.Cache
}

// NewVpnL2VpnImportTargetsCache new a cache
func NewVpnL2VpnImportTargetsCache(cacheType *database.CacheType) VpnL2VpnImportTargetsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnL2VpnImportTargets{}
		})
		return &vpnL2VpnImportTargetsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnL2VpnImportTargets{}
		})
		return &vpnL2VpnImportTargetsCache{cache: c}
	}

	return nil // no cache
}

// GetVpnL2VpnImportTargetsCacheKey cache key
func (c *vpnL2VpnImportTargetsCache) GetVpnL2VpnImportTargetsCacheKey(id uint64) string {
	return vpnL2VpnImportTargetsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *vpnL2VpnImportTargetsCache) Set(ctx context.Context, id uint64, data *model.VpnL2VpnImportTargets, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVpnL2VpnImportTargetsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *vpnL2VpnImportTargetsCache) Get(ctx context.Context, id uint64) (*model.VpnL2VpnImportTargets, error) {
	var data *model.VpnL2VpnImportTargets
	cacheKey := c.GetVpnL2VpnImportTargetsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *vpnL2VpnImportTargetsCache) MultiSet(ctx context.Context, data []*model.VpnL2VpnImportTargets, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVpnL2VpnImportTargetsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *vpnL2VpnImportTargetsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnL2VpnImportTargets, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVpnL2VpnImportTargetsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VpnL2VpnImportTargets)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VpnL2VpnImportTargets)
	for _, id := range ids {
		val, ok := itemMap[c.GetVpnL2VpnImportTargetsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *vpnL2VpnImportTargetsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnL2VpnImportTargetsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *vpnL2VpnImportTargetsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnL2VpnImportTargetsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *vpnL2VpnImportTargetsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
