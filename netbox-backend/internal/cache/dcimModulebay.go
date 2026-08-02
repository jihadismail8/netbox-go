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
	dcimModulebayCachePrefixKey = "dcimModulebay:"
	// DcimModulebayExpireTime expire time
	DcimModulebayExpireTime = 5 * time.Minute
)

var _ DcimModulebayCache = (*dcimModulebayCache)(nil)

// DcimModulebayCache cache interface
type DcimModulebayCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimModulebay, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimModulebay, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimModulebay, error)
	MultiSet(ctx context.Context, data []*model.DcimModulebay, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimModulebayCache define a cache struct
type dcimModulebayCache struct {
	cache cache.Cache
}

// NewDcimModulebayCache new a cache
func NewDcimModulebayCache(cacheType *database.CacheType) DcimModulebayCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimModulebay{}
		})
		return &dcimModulebayCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimModulebay{}
		})
		return &dcimModulebayCache{cache: c}
	}

	return nil // no cache
}

// GetDcimModulebayCacheKey cache key
func (c *dcimModulebayCache) GetDcimModulebayCacheKey(id uint64) string {
	return dcimModulebayCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimModulebayCache) Set(ctx context.Context, id uint64, data *model.DcimModulebay, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimModulebayCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimModulebayCache) Get(ctx context.Context, id uint64) (*model.DcimModulebay, error) {
	var data *model.DcimModulebay
	cacheKey := c.GetDcimModulebayCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimModulebayCache) MultiSet(ctx context.Context, data []*model.DcimModulebay, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimModulebayCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimModulebayCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimModulebay, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimModulebayCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimModulebay)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimModulebay)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimModulebayCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimModulebayCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimModulebayCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimModulebayCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimModulebayCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimModulebayCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
