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
	ipamVlantranslationruleCachePrefixKey = "ipamVlantranslationrule:"
	// IpamVlantranslationruleExpireTime expire time
	IpamVlantranslationruleExpireTime = 5 * time.Minute
)

var _ IpamVlantranslationruleCache = (*ipamVlantranslationruleCache)(nil)

// IpamVlantranslationruleCache cache interface
type IpamVlantranslationruleCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamVlantranslationrule, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamVlantranslationrule, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamVlantranslationrule, error)
	MultiSet(ctx context.Context, data []*model.IpamVlantranslationrule, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamVlantranslationruleCache define a cache struct
type ipamVlantranslationruleCache struct {
	cache cache.Cache
}

// NewIpamVlantranslationruleCache new a cache
func NewIpamVlantranslationruleCache(cacheType *database.CacheType) IpamVlantranslationruleCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamVlantranslationrule{}
		})
		return &ipamVlantranslationruleCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamVlantranslationrule{}
		})
		return &ipamVlantranslationruleCache{cache: c}
	}

	return nil // no cache
}

// GetIpamVlantranslationruleCacheKey cache key
func (c *ipamVlantranslationruleCache) GetIpamVlantranslationruleCacheKey(id uint64) string {
	return ipamVlantranslationruleCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamVlantranslationruleCache) Set(ctx context.Context, id uint64, data *model.IpamVlantranslationrule, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamVlantranslationruleCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamVlantranslationruleCache) Get(ctx context.Context, id uint64) (*model.IpamVlantranslationrule, error) {
	var data *model.IpamVlantranslationrule
	cacheKey := c.GetIpamVlantranslationruleCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamVlantranslationruleCache) MultiSet(ctx context.Context, data []*model.IpamVlantranslationrule, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamVlantranslationruleCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamVlantranslationruleCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamVlantranslationrule, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamVlantranslationruleCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamVlantranslationrule)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamVlantranslationrule)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamVlantranslationruleCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamVlantranslationruleCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamVlantranslationruleCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamVlantranslationruleCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamVlantranslationruleCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamVlantranslationruleCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
