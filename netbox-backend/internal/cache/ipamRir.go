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
	ipamRirCachePrefixKey = "ipamRir:"
	// IpamRirExpireTime expire time
	IpamRirExpireTime = 5 * time.Minute
)

var _ IpamRirCache = (*ipamRirCache)(nil)

// IpamRirCache cache interface
type IpamRirCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamRir, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamRir, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamRir, error)
	MultiSet(ctx context.Context, data []*model.IpamRir, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamRirCache define a cache struct
type ipamRirCache struct {
	cache cache.Cache
}

// NewIpamRirCache new a cache
func NewIpamRirCache(cacheType *database.CacheType) IpamRirCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamRir{}
		})
		return &ipamRirCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamRir{}
		})
		return &ipamRirCache{cache: c}
	}

	return nil // no cache
}

// GetIpamRirCacheKey cache key
func (c *ipamRirCache) GetIpamRirCacheKey(id uint64) string {
	return ipamRirCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamRirCache) Set(ctx context.Context, id uint64, data *model.IpamRir, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamRirCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamRirCache) Get(ctx context.Context, id uint64) (*model.IpamRir, error) {
	var data *model.IpamRir
	cacheKey := c.GetIpamRirCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamRirCache) MultiSet(ctx context.Context, data []*model.IpamRir, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamRirCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamRirCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamRir, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamRirCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamRir)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamRir)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamRirCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamRirCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamRirCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamRirCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamRirCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamRirCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
