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
	ipamIprangeCachePrefixKey = "ipamIprange:"
	// IpamIprangeExpireTime expire time
	IpamIprangeExpireTime = 5 * time.Minute
)

var _ IpamIprangeCache = (*ipamIprangeCache)(nil)

// IpamIprangeCache cache interface
type IpamIprangeCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamIprange, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamIprange, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamIprange, error)
	MultiSet(ctx context.Context, data []*model.IpamIprange, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamIprangeCache define a cache struct
type ipamIprangeCache struct {
	cache cache.Cache
}

// NewIpamIprangeCache new a cache
func NewIpamIprangeCache(cacheType *database.CacheType) IpamIprangeCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamIprange{}
		})
		return &ipamIprangeCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamIprange{}
		})
		return &ipamIprangeCache{cache: c}
	}

	return nil // no cache
}

// GetIpamIprangeCacheKey cache key
func (c *ipamIprangeCache) GetIpamIprangeCacheKey(id uint64) string {
	return ipamIprangeCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamIprangeCache) Set(ctx context.Context, id uint64, data *model.IpamIprange, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamIprangeCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamIprangeCache) Get(ctx context.Context, id uint64) (*model.IpamIprange, error) {
	var data *model.IpamIprange
	cacheKey := c.GetIpamIprangeCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamIprangeCache) MultiSet(ctx context.Context, data []*model.IpamIprange, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamIprangeCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamIprangeCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamIprange, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamIprangeCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamIprange)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamIprange)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamIprangeCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamIprangeCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamIprangeCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamIprangeCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamIprangeCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamIprangeCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
