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
	wirelessWirelesslanCachePrefixKey = "wirelessWirelesslan:"
	// WirelessWirelesslanExpireTime expire time
	WirelessWirelesslanExpireTime = 5 * time.Minute
)

var _ WirelessWirelesslanCache = (*wirelessWirelesslanCache)(nil)

// WirelessWirelesslanCache cache interface
type WirelessWirelesslanCache interface {
	Set(ctx context.Context, id uint64, data *model.WirelessWirelesslan, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.WirelessWirelesslan, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.WirelessWirelesslan, error)
	MultiSet(ctx context.Context, data []*model.WirelessWirelesslan, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// wirelessWirelesslanCache define a cache struct
type wirelessWirelesslanCache struct {
	cache cache.Cache
}

// NewWirelessWirelesslanCache new a cache
func NewWirelessWirelesslanCache(cacheType *database.CacheType) WirelessWirelesslanCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.WirelessWirelesslan{}
		})
		return &wirelessWirelesslanCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.WirelessWirelesslan{}
		})
		return &wirelessWirelesslanCache{cache: c}
	}

	return nil // no cache
}

// GetWirelessWirelesslanCacheKey cache key
func (c *wirelessWirelesslanCache) GetWirelessWirelesslanCacheKey(id uint64) string {
	return wirelessWirelesslanCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *wirelessWirelesslanCache) Set(ctx context.Context, id uint64, data *model.WirelessWirelesslan, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetWirelessWirelesslanCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *wirelessWirelesslanCache) Get(ctx context.Context, id uint64) (*model.WirelessWirelesslan, error) {
	var data *model.WirelessWirelesslan
	cacheKey := c.GetWirelessWirelesslanCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *wirelessWirelesslanCache) MultiSet(ctx context.Context, data []*model.WirelessWirelesslan, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetWirelessWirelesslanCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *wirelessWirelesslanCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.WirelessWirelesslan, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetWirelessWirelesslanCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.WirelessWirelesslan)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.WirelessWirelesslan)
	for _, id := range ids {
		val, ok := itemMap[c.GetWirelessWirelesslanCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *wirelessWirelesslanCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetWirelessWirelesslanCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *wirelessWirelesslanCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetWirelessWirelesslanCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *wirelessWirelesslanCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
