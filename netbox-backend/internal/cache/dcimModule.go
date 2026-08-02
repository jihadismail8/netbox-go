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
	dcimModuleCachePrefixKey = "dcimModule:"
	// DcimModuleExpireTime expire time
	DcimModuleExpireTime = 5 * time.Minute
)

var _ DcimModuleCache = (*dcimModuleCache)(nil)

// DcimModuleCache cache interface
type DcimModuleCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimModule, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimModule, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimModule, error)
	MultiSet(ctx context.Context, data []*model.DcimModule, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimModuleCache define a cache struct
type dcimModuleCache struct {
	cache cache.Cache
}

// NewDcimModuleCache new a cache
func NewDcimModuleCache(cacheType *database.CacheType) DcimModuleCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimModule{}
		})
		return &dcimModuleCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimModule{}
		})
		return &dcimModuleCache{cache: c}
	}

	return nil // no cache
}

// GetDcimModuleCacheKey cache key
func (c *dcimModuleCache) GetDcimModuleCacheKey(id uint64) string {
	return dcimModuleCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimModuleCache) Set(ctx context.Context, id uint64, data *model.DcimModule, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimModuleCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimModuleCache) Get(ctx context.Context, id uint64) (*model.DcimModule, error) {
	var data *model.DcimModule
	cacheKey := c.GetDcimModuleCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimModuleCache) MultiSet(ctx context.Context, data []*model.DcimModule, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimModuleCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimModuleCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimModule, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimModuleCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimModule)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimModule)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimModuleCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimModuleCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimModuleCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimModuleCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimModuleCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimModuleCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
