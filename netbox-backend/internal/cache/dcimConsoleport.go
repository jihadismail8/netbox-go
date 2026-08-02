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
	dcimConsoleportCachePrefixKey = "dcimConsoleport:"
	// DcimConsoleportExpireTime expire time
	DcimConsoleportExpireTime = 5 * time.Minute
)

var _ DcimConsoleportCache = (*dcimConsoleportCache)(nil)

// DcimConsoleportCache cache interface
type DcimConsoleportCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimConsoleport, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimConsoleport, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimConsoleport, error)
	MultiSet(ctx context.Context, data []*model.DcimConsoleport, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimConsoleportCache define a cache struct
type dcimConsoleportCache struct {
	cache cache.Cache
}

// NewDcimConsoleportCache new a cache
func NewDcimConsoleportCache(cacheType *database.CacheType) DcimConsoleportCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimConsoleport{}
		})
		return &dcimConsoleportCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimConsoleport{}
		})
		return &dcimConsoleportCache{cache: c}
	}

	return nil // no cache
}

// GetDcimConsoleportCacheKey cache key
func (c *dcimConsoleportCache) GetDcimConsoleportCacheKey(id uint64) string {
	return dcimConsoleportCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimConsoleportCache) Set(ctx context.Context, id uint64, data *model.DcimConsoleport, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimConsoleportCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimConsoleportCache) Get(ctx context.Context, id uint64) (*model.DcimConsoleport, error) {
	var data *model.DcimConsoleport
	cacheKey := c.GetDcimConsoleportCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimConsoleportCache) MultiSet(ctx context.Context, data []*model.DcimConsoleport, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimConsoleportCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimConsoleportCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimConsoleport, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimConsoleportCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimConsoleport)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimConsoleport)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimConsoleportCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimConsoleportCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimConsoleportCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimConsoleportCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimConsoleportCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimConsoleportCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
