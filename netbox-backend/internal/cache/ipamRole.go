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
	ipamRoleCachePrefixKey = "ipamRole:"
	// IpamRoleExpireTime expire time
	IpamRoleExpireTime = 5 * time.Minute
)

var _ IpamRoleCache = (*ipamRoleCache)(nil)

// IpamRoleCache cache interface
type IpamRoleCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamRole, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamRole, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamRole, error)
	MultiSet(ctx context.Context, data []*model.IpamRole, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamRoleCache define a cache struct
type ipamRoleCache struct {
	cache cache.Cache
}

// NewIpamRoleCache new a cache
func NewIpamRoleCache(cacheType *database.CacheType) IpamRoleCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamRole{}
		})
		return &ipamRoleCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamRole{}
		})
		return &ipamRoleCache{cache: c}
	}

	return nil // no cache
}

// GetIpamRoleCacheKey cache key
func (c *ipamRoleCache) GetIpamRoleCacheKey(id uint64) string {
	return ipamRoleCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamRoleCache) Set(ctx context.Context, id uint64, data *model.IpamRole, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamRoleCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamRoleCache) Get(ctx context.Context, id uint64) (*model.IpamRole, error) {
	var data *model.IpamRole
	cacheKey := c.GetIpamRoleCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamRoleCache) MultiSet(ctx context.Context, data []*model.IpamRole, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamRoleCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamRoleCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamRole, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamRoleCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamRole)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamRole)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamRoleCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamRoleCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamRoleCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamRoleCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamRoleCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamRoleCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
