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
	coreManagedfileCachePrefixKey = "coreManagedfile:"
	// CoreManagedfileExpireTime expire time
	CoreManagedfileExpireTime = 5 * time.Minute
)

var _ CoreManagedfileCache = (*coreManagedfileCache)(nil)

// CoreManagedfileCache cache interface
type CoreManagedfileCache interface {
	Set(ctx context.Context, id uint64, data *model.CoreManagedfile, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CoreManagedfile, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CoreManagedfile, error)
	MultiSet(ctx context.Context, data []*model.CoreManagedfile, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// coreManagedfileCache define a cache struct
type coreManagedfileCache struct {
	cache cache.Cache
}

// NewCoreManagedfileCache new a cache
func NewCoreManagedfileCache(cacheType *database.CacheType) CoreManagedfileCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreManagedfile{}
		})
		return &coreManagedfileCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreManagedfile{}
		})
		return &coreManagedfileCache{cache: c}
	}

	return nil // no cache
}

// GetCoreManagedfileCacheKey cache key
func (c *coreManagedfileCache) GetCoreManagedfileCacheKey(id uint64) string {
	return coreManagedfileCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *coreManagedfileCache) Set(ctx context.Context, id uint64, data *model.CoreManagedfile, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCoreManagedfileCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *coreManagedfileCache) Get(ctx context.Context, id uint64) (*model.CoreManagedfile, error) {
	var data *model.CoreManagedfile
	cacheKey := c.GetCoreManagedfileCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *coreManagedfileCache) MultiSet(ctx context.Context, data []*model.CoreManagedfile, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCoreManagedfileCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *coreManagedfileCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CoreManagedfile, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCoreManagedfileCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CoreManagedfile)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CoreManagedfile)
	for _, id := range ids {
		val, ok := itemMap[c.GetCoreManagedfileCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *coreManagedfileCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCoreManagedfileCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *coreManagedfileCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCoreManagedfileCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *coreManagedfileCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
