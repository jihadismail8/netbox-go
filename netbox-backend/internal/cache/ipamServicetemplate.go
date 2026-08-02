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
	ipamServicetemplateCachePrefixKey = "ipamServicetemplate:"
	// IpamServicetemplateExpireTime expire time
	IpamServicetemplateExpireTime = 5 * time.Minute
)

var _ IpamServicetemplateCache = (*ipamServicetemplateCache)(nil)

// IpamServicetemplateCache cache interface
type IpamServicetemplateCache interface {
	Set(ctx context.Context, id uint64, data *model.IpamServicetemplate, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.IpamServicetemplate, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamServicetemplate, error)
	MultiSet(ctx context.Context, data []*model.IpamServicetemplate, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// ipamServicetemplateCache define a cache struct
type ipamServicetemplateCache struct {
	cache cache.Cache
}

// NewIpamServicetemplateCache new a cache
func NewIpamServicetemplateCache(cacheType *database.CacheType) IpamServicetemplateCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamServicetemplate{}
		})
		return &ipamServicetemplateCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.IpamServicetemplate{}
		})
		return &ipamServicetemplateCache{cache: c}
	}

	return nil // no cache
}

// GetIpamServicetemplateCacheKey cache key
func (c *ipamServicetemplateCache) GetIpamServicetemplateCacheKey(id uint64) string {
	return ipamServicetemplateCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *ipamServicetemplateCache) Set(ctx context.Context, id uint64, data *model.IpamServicetemplate, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetIpamServicetemplateCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *ipamServicetemplateCache) Get(ctx context.Context, id uint64) (*model.IpamServicetemplate, error) {
	var data *model.IpamServicetemplate
	cacheKey := c.GetIpamServicetemplateCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *ipamServicetemplateCache) MultiSet(ctx context.Context, data []*model.IpamServicetemplate, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetIpamServicetemplateCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *ipamServicetemplateCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.IpamServicetemplate, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetIpamServicetemplateCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.IpamServicetemplate)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.IpamServicetemplate)
	for _, id := range ids {
		val, ok := itemMap[c.GetIpamServicetemplateCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *ipamServicetemplateCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamServicetemplateCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *ipamServicetemplateCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetIpamServicetemplateCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *ipamServicetemplateCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
