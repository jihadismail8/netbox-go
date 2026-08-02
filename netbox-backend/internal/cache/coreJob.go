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
	coreJobCachePrefixKey = "coreJob:"
	// CoreJobExpireTime expire time
	CoreJobExpireTime = 5 * time.Minute
)

var _ CoreJobCache = (*coreJobCache)(nil)

// CoreJobCache cache interface
type CoreJobCache interface {
	Set(ctx context.Context, id uint64, data *model.CoreJob, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CoreJob, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CoreJob, error)
	MultiSet(ctx context.Context, data []*model.CoreJob, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// coreJobCache define a cache struct
type coreJobCache struct {
	cache cache.Cache
}

// NewCoreJobCache new a cache
func NewCoreJobCache(cacheType *database.CacheType) CoreJobCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreJob{}
		})
		return &coreJobCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreJob{}
		})
		return &coreJobCache{cache: c}
	}

	return nil // no cache
}

// GetCoreJobCacheKey cache key
func (c *coreJobCache) GetCoreJobCacheKey(id uint64) string {
	return coreJobCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *coreJobCache) Set(ctx context.Context, id uint64, data *model.CoreJob, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCoreJobCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *coreJobCache) Get(ctx context.Context, id uint64) (*model.CoreJob, error) {
	var data *model.CoreJob
	cacheKey := c.GetCoreJobCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *coreJobCache) MultiSet(ctx context.Context, data []*model.CoreJob, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCoreJobCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *coreJobCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CoreJob, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCoreJobCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CoreJob)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CoreJob)
	for _, id := range ids {
		val, ok := itemMap[c.GetCoreJobCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *coreJobCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCoreJobCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *coreJobCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCoreJobCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *coreJobCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
