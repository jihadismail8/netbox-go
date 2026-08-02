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
	wirelessWirelesslangroupCachePrefixKey = "wirelessWirelesslangroup:"
	// WirelessWirelesslangroupExpireTime expire time
	WirelessWirelesslangroupExpireTime = 5 * time.Minute
)

var _ WirelessWirelesslangroupCache = (*wirelessWirelesslangroupCache)(nil)

// WirelessWirelesslangroupCache cache interface
type WirelessWirelesslangroupCache interface {
	Set(ctx context.Context, id uint64, data *model.WirelessWirelesslangroup, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.WirelessWirelesslangroup, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.WirelessWirelesslangroup, error)
	MultiSet(ctx context.Context, data []*model.WirelessWirelesslangroup, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// wirelessWirelesslangroupCache define a cache struct
type wirelessWirelesslangroupCache struct {
	cache cache.Cache
}

// NewWirelessWirelesslangroupCache new a cache
func NewWirelessWirelesslangroupCache(cacheType *database.CacheType) WirelessWirelesslangroupCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.WirelessWirelesslangroup{}
		})
		return &wirelessWirelesslangroupCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.WirelessWirelesslangroup{}
		})
		return &wirelessWirelesslangroupCache{cache: c}
	}

	return nil // no cache
}

// GetWirelessWirelesslangroupCacheKey cache key
func (c *wirelessWirelesslangroupCache) GetWirelessWirelesslangroupCacheKey(id uint64) string {
	return wirelessWirelesslangroupCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *wirelessWirelesslangroupCache) Set(ctx context.Context, id uint64, data *model.WirelessWirelesslangroup, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetWirelessWirelesslangroupCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *wirelessWirelesslangroupCache) Get(ctx context.Context, id uint64) (*model.WirelessWirelesslangroup, error) {
	var data *model.WirelessWirelesslangroup
	cacheKey := c.GetWirelessWirelesslangroupCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *wirelessWirelesslangroupCache) MultiSet(ctx context.Context, data []*model.WirelessWirelesslangroup, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetWirelessWirelesslangroupCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *wirelessWirelesslangroupCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.WirelessWirelesslangroup, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetWirelessWirelesslangroupCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.WirelessWirelesslangroup)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.WirelessWirelesslangroup)
	for _, id := range ids {
		val, ok := itemMap[c.GetWirelessWirelesslangroupCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *wirelessWirelesslangroupCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetWirelessWirelesslangroupCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *wirelessWirelesslangroupCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetWirelessWirelesslangroupCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *wirelessWirelesslangroupCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
