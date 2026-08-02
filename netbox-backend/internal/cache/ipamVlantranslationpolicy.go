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
	ipamVlantranslationpolicyCachePrefixKey = "ipamVlantranslationpolicy:"
	// IpamVlantranslationpolicyExpireTime expire time
	IpamVlantranslationpolicyExpireTime = 5 * time.Minute
)

var _ IpamVlantranslationpolicyCache = (*ipamVlantranslationpolicyCache)(nil)

// IpamVlantranslationpolicyCache cache interface
type IpamVlantranslationpolicyCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamVlantranslationpolicy, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamVlantranslationpolicy, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamVlantranslationpolicy, error)
	MultiSet(ctx context.Context, data []*model.IpamVlantranslationpolicy, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamVlantranslationpolicyCache define a cache struct
type ipamVlantranslationpolicyCache struct {
	cache cache.Cache
}

// NewIpamVlantranslationpolicyCache new a cache
func NewIpamVlantranslationpolicyCache(cacheType *database.CacheType) IpamVlantranslationpolicyCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamVlantranslationpolicy{}
		})
		return &ipamVlantranslationpolicyCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamVlantranslationpolicy{}
		})
		return &ipamVlantranslationpolicyCache{cache: c}
	}

	return nil // no cache
}

// GetIpamVlantranslationpolicyCacheKey cache key
func (c *ipamVlantranslationpolicyCache) GetIpamVlantranslationpolicyCacheKey(id uint64) string {
	return ipamVlantranslationpolicyCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamVlantranslationpolicyCache) Set(ctx context.Context, id uint64, data *model.IpamVlantranslationpolicy, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamVlantranslationpolicyCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamVlantranslationpolicyCache) Get(ctx context.Context, id uint64) (*model.IpamVlantranslationpolicy, error) {
	var data *model.IpamVlantranslationpolicy
	cacheKey := c.GetIpamVlantranslationpolicyCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamVlantranslationpolicyCache) MultiSet(ctx context.Context, data []*model.IpamVlantranslationpolicy, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamVlantranslationpolicyCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamVlantranslationpolicyCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamVlantranslationpolicy, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamVlantranslationpolicyCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamVlantranslationpolicy)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamVlantranslationpolicy)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamVlantranslationpolicyCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamVlantranslationpolicyCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamVlantranslationpolicyCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamVlantranslationpolicyCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamVlantranslationpolicyCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamVlantranslationpolicyCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
