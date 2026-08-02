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
	ipamAsnrangeCachePrefixKey = "ipamAsnrange:"
	// IpamAsnrangeExpireTime expire time
	IpamAsnrangeExpireTime = 5 * time.Minute
)

var _ IpamAsnrangeCache = (*ipamAsnrangeCache)(nil)

// IpamAsnrangeCache cache interface
type IpamAsnrangeCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamAsnrange, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamAsnrange, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamAsnrange, error)
	MultiSet(ctx context.Context, data []*model.IpamAsnrange, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamAsnrangeCache define a cache struct
type ipamAsnrangeCache struct {
	cache cache.Cache
}

// NewIpamAsnrangeCache new a cache
func NewIpamAsnrangeCache(cacheType *database.CacheType) IpamAsnrangeCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamAsnrange{}
		})
		return &ipamAsnrangeCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamAsnrange{}
		})
		return &ipamAsnrangeCache{cache: c}
	}

	return nil // no cache
}

// GetIpamAsnrangeCacheKey cache key
func (c *ipamAsnrangeCache) GetIpamAsnrangeCacheKey(id uint64) string {
	return ipamAsnrangeCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamAsnrangeCache) Set(ctx context.Context, id uint64, data *model.IpamAsnrange, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamAsnrangeCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamAsnrangeCache) Get(ctx context.Context, id uint64) (*model.IpamAsnrange, error) {
	var data *model.IpamAsnrange
	cacheKey := c.GetIpamAsnrangeCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamAsnrangeCache) MultiSet(ctx context.Context, data []*model.IpamAsnrange, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamAsnrangeCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamAsnrangeCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamAsnrange, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamAsnrangeCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamAsnrange)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamAsnrange)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamAsnrangeCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamAsnrangeCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamAsnrangeCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamAsnrangeCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamAsnrangeCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamAsnrangeCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
