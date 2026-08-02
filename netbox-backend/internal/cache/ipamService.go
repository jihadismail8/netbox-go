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
	ipamServiceCachePrefixKey = "ipamService:"
	// IpamServiceExpireTime expire time
	IpamServiceExpireTime = 5 * time.Minute
)

var _ IpamServiceCache = (*ipamServiceCache)(nil)

// IpamServiceCache cache interface
type IpamServiceCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamService, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamService, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamService, error)
	MultiSet(ctx context.Context, data []*model.IpamService, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamServiceCache define a cache struct
type ipamServiceCache struct {
	cache cache.Cache
}

// NewIpamServiceCache new a cache
func NewIpamServiceCache(cacheType *database.CacheType) IpamServiceCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamService{}
		})
		return &ipamServiceCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamService{}
		})
		return &ipamServiceCache{cache: c}
	}

	return nil // no cache
}

// GetIpamServiceCacheKey cache key
func (c *ipamServiceCache) GetIpamServiceCacheKey(id uint64) string {
	return ipamServiceCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamServiceCache) Set(ctx context.Context, id uint64, data *model.IpamService, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamServiceCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamServiceCache) Get(ctx context.Context, id uint64) (*model.IpamService, error) {
	var data *model.IpamService
	cacheKey := c.GetIpamServiceCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamServiceCache) MultiSet(ctx context.Context, data []*model.IpamService, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamServiceCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamServiceCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamService, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamServiceCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamService)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamService)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamServiceCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamServiceCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamServiceCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamServiceCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamServiceCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamServiceCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
