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
	ipamAggregateCachePrefixKey = "ipamAggregate:"
	// IpamAggregateExpireTime expire time
	IpamAggregateExpireTime = 5 * time.Minute
)

var _ IpamAggregateCache = (*ipamAggregateCache)(nil)

// IpamAggregateCache cache interface
type IpamAggregateCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamAggregate, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamAggregate, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamAggregate, error)
	MultiSet(ctx context.Context, data []*model.IpamAggregate, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamAggregateCache define a cache struct
type ipamAggregateCache struct {
	cache cache.Cache
}

// NewIpamAggregateCache new a cache
func NewIpamAggregateCache(cacheType *database.CacheType) IpamAggregateCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamAggregate{}
		})
		return &ipamAggregateCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamAggregate{}
		})
		return &ipamAggregateCache{cache: c}
	}

	return nil // no cache
}

// GetIpamAggregateCacheKey cache key
func (c *ipamAggregateCache) GetIpamAggregateCacheKey(id uint64) string {
	return ipamAggregateCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamAggregateCache) Set(ctx context.Context, id uint64, data *model.IpamAggregate, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamAggregateCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamAggregateCache) Get(ctx context.Context, id uint64) (*model.IpamAggregate, error) {
	var data *model.IpamAggregate
	cacheKey := c.GetIpamAggregateCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamAggregateCache) MultiSet(ctx context.Context, data []*model.IpamAggregate, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamAggregateCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamAggregateCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamAggregate, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamAggregateCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamAggregate)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamAggregate)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamAggregateCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamAggregateCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamAggregateCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamAggregateCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamAggregateCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamAggregateCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
