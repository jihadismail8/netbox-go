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
	coreObjecttypeCachePrefixKey = "coreObjecttype:"
	// CoreObjecttypeExpireTime expire time
	CoreObjecttypeExpireTime = 5 * time.Minute
)

var _ CoreObjecttypeCache = (*coreObjecttypeCache)(nil)

// CoreObjecttypeCache cache interface
type CoreObjecttypeCache interface {
	Set(ctx context.Context, contenttypePtrID int, data *model.CoreObjecttype, duration time.Duration) error
	Get(ctx context.Context, contenttypePtrID int) (*model.CoreObjecttype, error)
	MultiGet(ctx context.Context, contenttypePtrIDs []int) (map[int]*model.CoreObjecttype, error)
	MultiSet(ctx context.Context, data []*model.CoreObjecttype, duration time.Duration) error
	Del(ctx context.Context, contenttypePtrID int) error
	SetPlaceholder(ctx context.Context, contenttypePtrID int) error
	IsPlaceholderErr(err error) bool
}

// coreObjecttypeCache define a cache struct
type coreObjecttypeCache struct {
	cache cache.Cache
}

// NewCoreObjecttypeCache new a cache
func NewCoreObjecttypeCache(cacheType *database.CacheType) CoreObjecttypeCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreObjecttype{}
		})
		return &coreObjecttypeCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreObjecttype{}
		})
		return &coreObjecttypeCache{cache: c}
	}

	return nil // no cache
}

// GetCoreObjecttypeCacheKey cache key
func (c *coreObjecttypeCache) GetCoreObjecttypeCacheKey(contenttypePtrID int) string {
	return coreObjecttypeCachePrefixKey + utils.IntToStr(contenttypePtrID)
}

// Set write to cache
func (c *coreObjecttypeCache) Set(ctx context.Context, contenttypePtrID int, data *model.CoreObjecttype, duration time.Duration) error {
	if data == nil {
		return nil
	}
	cacheKey := c.GetCoreObjecttypeCacheKey(contenttypePtrID)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *coreObjecttypeCache) Get(ctx context.Context, contenttypePtrID int) (*model.CoreObjecttype, error) {
	var data *model.CoreObjecttype
	cacheKey := c.GetCoreObjecttypeCacheKey(contenttypePtrID)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *coreObjecttypeCache) MultiSet(ctx context.Context, data []*model.CoreObjecttype, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCoreObjecttypeCacheKey(v.ContenttypePtrID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is contenttypePtrID value
func (c *coreObjecttypeCache) MultiGet(ctx context.Context, contenttypePtrIDs []int) (map[int]*model.CoreObjecttype, error) {
	var keys []string
	for _, v := range contenttypePtrIDs {
		cacheKey := c.GetCoreObjecttypeCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CoreObjecttype)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[int]*model.CoreObjecttype)
	for _, contenttypePtrID := range contenttypePtrIDs {
		val, ok := itemMap[c.GetCoreObjecttypeCacheKey(contenttypePtrID)]
		if ok {
			retMap[contenttypePtrID] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *coreObjecttypeCache) Del(ctx context.Context, contenttypePtrID int) error {
	cacheKey := c.GetCoreObjecttypeCacheKey(contenttypePtrID)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *coreObjecttypeCache) SetPlaceholder(ctx context.Context, contenttypePtrID int) error {
	cacheKey := c.GetCoreObjecttypeCacheKey(contenttypePtrID)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *coreObjecttypeCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
