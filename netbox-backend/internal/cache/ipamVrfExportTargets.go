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
	ipamVrfExportTargetsCachePrefixKey = "ipamVrfExportTargets:"
	// IpamVrfExportTargetsExpireTime expire time
	IpamVrfExportTargetsExpireTime = 5 * time.Minute
)

var _ IpamVrfExportTargetsCache = (*ipamVrfExportTargetsCache)(nil)

// IpamVrfExportTargetsCache cache interface
type IpamVrfExportTargetsCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamVrfExportTargets, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamVrfExportTargets, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamVrfExportTargets, error)
	MultiSet(ctx context.Context, data []*model.IpamVrfExportTargets, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamVrfExportTargetsCache define a cache struct
type ipamVrfExportTargetsCache struct {
	cache cache.Cache
}

// NewIpamVrfExportTargetsCache new a cache
func NewIpamVrfExportTargetsCache(cacheType *database.CacheType) IpamVrfExportTargetsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamVrfExportTargets{}
		})
		return &ipamVrfExportTargetsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamVrfExportTargets{}
		})
		return &ipamVrfExportTargetsCache{cache: c}
	}

	return nil // no cache
}

// GetIpamVrfExportTargetsCacheKey cache key
func (c *ipamVrfExportTargetsCache) GetIpamVrfExportTargetsCacheKey(id uint64) string {
	return ipamVrfExportTargetsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamVrfExportTargetsCache) Set(ctx context.Context, id uint64, data *model.IpamVrfExportTargets, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamVrfExportTargetsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamVrfExportTargetsCache) Get(ctx context.Context, id uint64) (*model.IpamVrfExportTargets, error) {
	var data *model.IpamVrfExportTargets
	cacheKey := c.GetIpamVrfExportTargetsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamVrfExportTargetsCache) MultiSet(ctx context.Context, data []*model.IpamVrfExportTargets, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamVrfExportTargetsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamVrfExportTargetsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamVrfExportTargets, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamVrfExportTargetsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamVrfExportTargets)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamVrfExportTargets)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamVrfExportTargetsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamVrfExportTargetsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamVrfExportTargetsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamVrfExportTargetsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamVrfExportTargetsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamVrfExportTargetsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
