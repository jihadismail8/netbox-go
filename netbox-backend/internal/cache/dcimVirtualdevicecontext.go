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
	dcimVirtualdevicecontextCachePrefixKey = "dcimVirtualdevicecontext:"
	// DcimVirtualdevicecontextExpireTime expire time
	DcimVirtualdevicecontextExpireTime = 5 * time.Minute
)

var _ DcimVirtualdevicecontextCache = (*dcimVirtualdevicecontextCache)(nil)

// DcimVirtualdevicecontextCache cache interface
type DcimVirtualdevicecontextCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimVirtualdevicecontext, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimVirtualdevicecontext, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimVirtualdevicecontext, error)
	MultiSet(ctx context.Context, data []*model.DcimVirtualdevicecontext, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimVirtualdevicecontextCache define a cache struct
type dcimVirtualdevicecontextCache struct {
	cache cache.Cache
}

// NewDcimVirtualdevicecontextCache new a cache
func NewDcimVirtualdevicecontextCache(cacheType *database.CacheType) DcimVirtualdevicecontextCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimVirtualdevicecontext{}
		})
		return &dcimVirtualdevicecontextCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimVirtualdevicecontext{}
		})
		return &dcimVirtualdevicecontextCache{cache: c}
	}

	return nil // no cache
}

// GetDcimVirtualdevicecontextCacheKey cache key
func (c *dcimVirtualdevicecontextCache) GetDcimVirtualdevicecontextCacheKey(id uint64) string {
	return dcimVirtualdevicecontextCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimVirtualdevicecontextCache) Set(ctx context.Context, id uint64, data *model.DcimVirtualdevicecontext, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimVirtualdevicecontextCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimVirtualdevicecontextCache) Get(ctx context.Context, id uint64) (*model.DcimVirtualdevicecontext, error) {
	var data *model.DcimVirtualdevicecontext
	cacheKey := c.GetDcimVirtualdevicecontextCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimVirtualdevicecontextCache) MultiSet(ctx context.Context, data []*model.DcimVirtualdevicecontext, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimVirtualdevicecontextCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimVirtualdevicecontextCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimVirtualdevicecontext, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimVirtualdevicecontextCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimVirtualdevicecontext)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimVirtualdevicecontext)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimVirtualdevicecontextCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimVirtualdevicecontextCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimVirtualdevicecontextCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimVirtualdevicecontextCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimVirtualdevicecontextCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimVirtualdevicecontextCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
