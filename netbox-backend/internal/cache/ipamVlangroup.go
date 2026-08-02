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
	ipamVlangroupCachePrefixKey = "ipamVlangroup:"
	// IpamVlangroupExpireTime expire time
	IpamVlangroupExpireTime = 5 * time.Minute
)

var _ IpamVlangroupCache = (*ipamVlangroupCache)(nil)

// IpamVlangroupCache cache interface
type IpamVlangroupCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamVlangroup, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamVlangroup, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamVlangroup, error)
	MultiSet(ctx context.Context, data []*model.IpamVlangroup, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamVlangroupCache define a cache struct
type ipamVlangroupCache struct {
	cache cache.Cache
}

// NewIpamVlangroupCache new a cache
func NewIpamVlangroupCache(cacheType *database.CacheType) IpamVlangroupCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamVlangroup{}
		})
		return &ipamVlangroupCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamVlangroup{}
		})
		return &ipamVlangroupCache{cache: c}
	}

	return nil // no cache
}

// GetIpamVlangroupCacheKey cache key
func (c *ipamVlangroupCache) GetIpamVlangroupCacheKey(id uint64) string {
	return ipamVlangroupCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamVlangroupCache) Set(ctx context.Context, id uint64, data *model.IpamVlangroup, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamVlangroupCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamVlangroupCache) Get(ctx context.Context, id uint64) (*model.IpamVlangroup, error) {
	var data *model.IpamVlangroup
	cacheKey := c.GetIpamVlangroupCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamVlangroupCache) MultiSet(ctx context.Context, data []*model.IpamVlangroup, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamVlangroupCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamVlangroupCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamVlangroup, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamVlangroupCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamVlangroup)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamVlangroup)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamVlangroupCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamVlangroupCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamVlangroupCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamVlangroupCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamVlangroupCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamVlangroupCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
