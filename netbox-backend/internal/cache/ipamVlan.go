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
	ipamVlanCachePrefixKey = "ipamVlan:"
	// IpamVlanExpireTime expire time
	IpamVlanExpireTime = 5 * time.Minute
)

var _ IpamVlanCache = (*ipamVlanCache)(nil)

// IpamVlanCache cache interface
type IpamVlanCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamVlan, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamVlan, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamVlan, error)
	MultiSet(ctx context.Context, data []*model.IpamVlan, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamVlanCache define a cache struct
type ipamVlanCache struct {
	cache cache.Cache
}

// NewIpamVlanCache new a cache
func NewIpamVlanCache(cacheType *database.CacheType) IpamVlanCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamVlan{}
		})
		return &ipamVlanCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamVlan{}
		})
		return &ipamVlanCache{cache: c}
	}

	return nil // no cache
}

// GetIpamVlanCacheKey cache key
func (c *ipamVlanCache) GetIpamVlanCacheKey(id uint64) string {
	return ipamVlanCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamVlanCache) Set(ctx context.Context, id uint64, data *model.IpamVlan, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamVlanCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamVlanCache) Get(ctx context.Context, id uint64) (*model.IpamVlan, error) {
	var data *model.IpamVlan
	cacheKey := c.GetIpamVlanCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamVlanCache) MultiSet(ctx context.Context, data []*model.IpamVlan, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamVlanCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamVlanCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamVlan, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamVlanCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamVlan)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamVlan)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamVlanCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamVlanCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamVlanCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamVlanCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamVlanCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamVlanCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
