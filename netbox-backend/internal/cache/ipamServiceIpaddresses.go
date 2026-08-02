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
	ipamServiceIpaddressesCachePrefixKey = "ipamServiceIpaddresses:"
	// IpamServiceIpaddressesExpireTime expire time
	IpamServiceIpaddressesExpireTime = 5 * time.Minute
)

var _ IpamServiceIpaddressesCache = (*ipamServiceIpaddressesCache)(nil)

// IpamServiceIpaddressesCache cache interface
type IpamServiceIpaddressesCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamServiceIpaddresses, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamServiceIpaddresses, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamServiceIpaddresses, error)
	MultiSet(ctx context.Context, data []*model.IpamServiceIpaddresses, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamServiceIpaddressesCache define a cache struct
type ipamServiceIpaddressesCache struct {
	cache cache.Cache
}

// NewIpamServiceIpaddressesCache new a cache
func NewIpamServiceIpaddressesCache(cacheType *database.CacheType) IpamServiceIpaddressesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamServiceIpaddresses{}
		})
		return &ipamServiceIpaddressesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamServiceIpaddresses{}
		})
		return &ipamServiceIpaddressesCache{cache: c}
	}

	return nil // no cache
}

// GetIpamServiceIpaddressesCacheKey cache key
func (c *ipamServiceIpaddressesCache) GetIpamServiceIpaddressesCacheKey(id uint64) string {
	return ipamServiceIpaddressesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamServiceIpaddressesCache) Set(ctx context.Context, id uint64, data *model.IpamServiceIpaddresses, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamServiceIpaddressesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamServiceIpaddressesCache) Get(ctx context.Context, id uint64) (*model.IpamServiceIpaddresses, error) {
	var data *model.IpamServiceIpaddresses
	cacheKey := c.GetIpamServiceIpaddressesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamServiceIpaddressesCache) MultiSet(ctx context.Context, data []*model.IpamServiceIpaddresses, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamServiceIpaddressesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamServiceIpaddressesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamServiceIpaddresses, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamServiceIpaddressesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamServiceIpaddresses)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamServiceIpaddresses)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamServiceIpaddressesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamServiceIpaddressesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamServiceIpaddressesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamServiceIpaddressesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamServiceIpaddressesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamServiceIpaddressesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
