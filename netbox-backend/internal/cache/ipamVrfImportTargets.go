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
	ipamVrfImportTargetsCachePrefixKey = "ipamVrfImportTargets:"
	// IpamVrfImportTargetsExpireTime expire time
	IpamVrfImportTargetsExpireTime = 5 * time.Minute
)

var _ IpamVrfImportTargetsCache = (*ipamVrfImportTargetsCache)(nil)

// IpamVrfImportTargetsCache cache interface
type IpamVrfImportTargetsCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamVrfImportTargets, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamVrfImportTargets, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamVrfImportTargets, error)
	MultiSet(ctx context.Context, data []*model.IpamVrfImportTargets, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamVrfImportTargetsCache define a cache struct
type ipamVrfImportTargetsCache struct {
	cache cache.Cache
}

// NewIpamVrfImportTargetsCache new a cache
func NewIpamVrfImportTargetsCache(cacheType *database.CacheType) IpamVrfImportTargetsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamVrfImportTargets{}
		})
		return &ipamVrfImportTargetsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamVrfImportTargets{}
		})
		return &ipamVrfImportTargetsCache{cache: c}
	}

	return nil // no cache
}

// GetIpamVrfImportTargetsCacheKey cache key
func (c *ipamVrfImportTargetsCache) GetIpamVrfImportTargetsCacheKey(id uint64) string {
	return ipamVrfImportTargetsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamVrfImportTargetsCache) Set(ctx context.Context, id uint64, data *model.IpamVrfImportTargets, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamVrfImportTargetsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamVrfImportTargetsCache) Get(ctx context.Context, id uint64) (*model.IpamVrfImportTargets, error) {
	var data *model.IpamVrfImportTargets
	cacheKey := c.GetIpamVrfImportTargetsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamVrfImportTargetsCache) MultiSet(ctx context.Context, data []*model.IpamVrfImportTargets, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamVrfImportTargetsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamVrfImportTargetsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamVrfImportTargets, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamVrfImportTargetsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamVrfImportTargets)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamVrfImportTargets)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamVrfImportTargetsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamVrfImportTargetsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamVrfImportTargetsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamVrfImportTargetsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamVrfImportTargetsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamVrfImportTargetsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
