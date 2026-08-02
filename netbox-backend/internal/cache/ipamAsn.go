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
	ipamAsnCachePrefixKey = "ipamAsn:"
	// IpamAsnExpireTime expire time
	IpamAsnExpireTime = 5 * time.Minute
)

var _ IpamAsnCache = (*ipamAsnCache)(nil)

// IpamAsnCache cache interface
type IpamAsnCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamAsn, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamAsn, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamAsn, error)
	MultiSet(ctx context.Context, data []*model.IpamAsn, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamAsnCache define a cache struct
type ipamAsnCache struct {
	cache cache.Cache
}

// NewIpamAsnCache new a cache
func NewIpamAsnCache(cacheType *database.CacheType) IpamAsnCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamAsn{}
		})
		return &ipamAsnCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamAsn{}
		})
		return &ipamAsnCache{cache: c}
	}

	return nil // no cache
}

// GetIpamAsnCacheKey cache key
func (c *ipamAsnCache) GetIpamAsnCacheKey(id uint64) string {
	return ipamAsnCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamAsnCache) Set(ctx context.Context, id uint64, data *model.IpamAsn, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamAsnCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamAsnCache) Get(ctx context.Context, id uint64) (*model.IpamAsn, error) {
	var data *model.IpamAsn
	cacheKey := c.GetIpamAsnCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamAsnCache) MultiSet(ctx context.Context, data []*model.IpamAsn, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamAsnCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamAsnCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamAsn, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamAsnCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamAsn)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamAsn)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamAsnCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamAsnCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamAsnCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamAsnCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamAsnCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamAsnCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
