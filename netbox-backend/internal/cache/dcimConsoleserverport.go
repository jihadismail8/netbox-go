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
	dcimConsoleserverportCachePrefixKey = "dcimConsoleserverport:"
	// DcimConsoleserverportExpireTime expire time
	DcimConsoleserverportExpireTime = 5 * time.Minute
)

var _ DcimConsoleserverportCache = (*dcimConsoleserverportCache)(nil)

// DcimConsoleserverportCache cache interface
type DcimConsoleserverportCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimConsoleserverport, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimConsoleserverport, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimConsoleserverport, error)
	MultiSet(ctx context.Context, data []*model.DcimConsoleserverport, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimConsoleserverportCache define a cache struct
type dcimConsoleserverportCache struct {
	cache cache.Cache
}

// NewDcimConsoleserverportCache new a cache
func NewDcimConsoleserverportCache(cacheType *database.CacheType) DcimConsoleserverportCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimConsoleserverport{}
		})
		return &dcimConsoleserverportCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimConsoleserverport{}
		})
		return &dcimConsoleserverportCache{cache: c}
	}

	return nil // no cache
}

// GetDcimConsoleserverportCacheKey cache key
func (c *dcimConsoleserverportCache) GetDcimConsoleserverportCacheKey(id uint64) string {
	return dcimConsoleserverportCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimConsoleserverportCache) Set(ctx context.Context, id uint64, data *model.DcimConsoleserverport, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimConsoleserverportCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimConsoleserverportCache) Get(ctx context.Context, id uint64) (*model.DcimConsoleserverport, error) {
	var data *model.DcimConsoleserverport
	cacheKey := c.GetDcimConsoleserverportCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimConsoleserverportCache) MultiSet(ctx context.Context, data []*model.DcimConsoleserverport, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimConsoleserverportCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimConsoleserverportCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimConsoleserverport, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimConsoleserverportCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimConsoleserverport)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimConsoleserverport)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimConsoleserverportCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimConsoleserverportCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimConsoleserverportCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimConsoleserverportCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimConsoleserverportCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimConsoleserverportCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
