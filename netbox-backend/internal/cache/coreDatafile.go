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
	coreDatafileCachePrefixKey = "coreDatafile:"
	// CoreDatafileExpireTime expire time
	CoreDatafileExpireTime = 5 * time.Minute
)

var _ CoreDatafileCache = (*coreDatafileCache)(nil)

// CoreDatafileCache cache interface
type CoreDatafileCache interface {
	Set(ctx context.Context, id uint64, data *model.CoreDatafile, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CoreDatafile, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CoreDatafile, error)
	MultiSet(ctx context.Context, data []*model.CoreDatafile, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// coreDatafileCache define a cache struct
type coreDatafileCache struct {
	cache cache.Cache
}

// NewCoreDatafileCache new a cache
func NewCoreDatafileCache(cacheType *database.CacheType) CoreDatafileCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreDatafile{}
		})
		return &coreDatafileCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreDatafile{}
		})
		return &coreDatafileCache{cache: c}
	}

	return nil // no cache
}

// GetCoreDatafileCacheKey cache key
func (c *coreDatafileCache) GetCoreDatafileCacheKey(id uint64) string {
	return coreDatafileCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *coreDatafileCache) Set(ctx context.Context, id uint64, data *model.CoreDatafile, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCoreDatafileCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *coreDatafileCache) Get(ctx context.Context, id uint64) (*model.CoreDatafile, error) {
	var data *model.CoreDatafile
	cacheKey := c.GetCoreDatafileCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *coreDatafileCache) MultiSet(ctx context.Context, data []*model.CoreDatafile, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCoreDatafileCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *coreDatafileCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CoreDatafile, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCoreDatafileCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CoreDatafile)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CoreDatafile)
	for _, id := range ids {
		val, ok := itemMap[c.GetCoreDatafileCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *coreDatafileCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCoreDatafileCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *coreDatafileCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCoreDatafileCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *coreDatafileCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
