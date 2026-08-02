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
	ipamFhrpgroupassignmentCachePrefixKey = "ipamFhrpgroupassignment:"
	// IpamFhrpgroupassignmentExpireTime expire time
	IpamFhrpgroupassignmentExpireTime = 5 * time.Minute
)

var _ IpamFhrpgroupassignmentCache = (*ipamFhrpgroupassignmentCache)(nil)

// IpamFhrpgroupassignmentCache cache interface
type IpamFhrpgroupassignmentCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamFhrpgroupassignment, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamFhrpgroupassignment, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamFhrpgroupassignment, error)
	MultiSet(ctx context.Context, data []*model.IpamFhrpgroupassignment, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamFhrpgroupassignmentCache define a cache struct
type ipamFhrpgroupassignmentCache struct {
	cache cache.Cache
}

// NewIpamFhrpgroupassignmentCache new a cache
func NewIpamFhrpgroupassignmentCache(cacheType *database.CacheType) IpamFhrpgroupassignmentCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamFhrpgroupassignment{}
		})
		return &ipamFhrpgroupassignmentCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamFhrpgroupassignment{}
		})
		return &ipamFhrpgroupassignmentCache{cache: c}
	}

	return nil // no cache
}

// GetIpamFhrpgroupassignmentCacheKey cache key
func (c *ipamFhrpgroupassignmentCache) GetIpamFhrpgroupassignmentCacheKey(id uint64) string {
	return ipamFhrpgroupassignmentCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamFhrpgroupassignmentCache) Set(ctx context.Context, id uint64, data *model.IpamFhrpgroupassignment, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamFhrpgroupassignmentCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamFhrpgroupassignmentCache) Get(ctx context.Context, id uint64) (*model.IpamFhrpgroupassignment, error) {
	var data *model.IpamFhrpgroupassignment
	cacheKey := c.GetIpamFhrpgroupassignmentCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamFhrpgroupassignmentCache) MultiSet(ctx context.Context, data []*model.IpamFhrpgroupassignment, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamFhrpgroupassignmentCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamFhrpgroupassignmentCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamFhrpgroupassignment, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamFhrpgroupassignmentCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamFhrpgroupassignment)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamFhrpgroupassignment)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamFhrpgroupassignmentCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamFhrpgroupassignmentCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamFhrpgroupassignmentCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamFhrpgroupassignmentCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamFhrpgroupassignmentCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamFhrpgroupassignmentCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
