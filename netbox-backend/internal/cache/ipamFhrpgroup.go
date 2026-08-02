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
	ipamFhrpgroupCachePrefixKey = "ipamFhrpgroup:"
	// IpamFhrpgroupExpireTime expire time
	IpamFhrpgroupExpireTime = 5 * time.Minute
)

var _ IpamFhrpgroupCache = (*ipamFhrpgroupCache)(nil)

// IpamFhrpgroupCache cache interface
type IpamFhrpgroupCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamFhrpgroup, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamFhrpgroup, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamFhrpgroup, error)
	MultiSet(ctx context.Context, data []*model.IpamFhrpgroup, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamFhrpgroupCache define a cache struct
type ipamFhrpgroupCache struct {
	cache cache.Cache
}

// NewIpamFhrpgroupCache new a cache
func NewIpamFhrpgroupCache(cacheType *database.CacheType) IpamFhrpgroupCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamFhrpgroup{}
		})
		return &ipamFhrpgroupCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamFhrpgroup{}
		})
		return &ipamFhrpgroupCache{cache: c}
	}

	return nil // no cache
}

// GetIpamFhrpgroupCacheKey cache key
func (c *ipamFhrpgroupCache) GetIpamFhrpgroupCacheKey(id uint64) string {
	return ipamFhrpgroupCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamFhrpgroupCache) Set(ctx context.Context, id uint64, data *model.IpamFhrpgroup, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamFhrpgroupCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamFhrpgroupCache) Get(ctx context.Context, id uint64) (*model.IpamFhrpgroup, error) {
	var data *model.IpamFhrpgroup
	cacheKey := c.GetIpamFhrpgroupCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamFhrpgroupCache) MultiSet(ctx context.Context, data []*model.IpamFhrpgroup, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamFhrpgroupCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamFhrpgroupCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamFhrpgroup, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamFhrpgroupCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamFhrpgroup)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamFhrpgroup)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamFhrpgroupCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamFhrpgroupCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamFhrpgroupCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamFhrpgroupCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamFhrpgroupCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamFhrpgroupCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
