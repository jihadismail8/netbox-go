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
	wirelessWirelesslinkCachePrefixKey = "wirelessWirelesslink:"
	// WirelessWirelesslinkExpireTime expire time
	WirelessWirelesslinkExpireTime = 5 * time.Minute
)

var _ WirelessWirelesslinkCache = (*wirelessWirelesslinkCache)(nil)

// WirelessWirelesslinkCache cache interface
type WirelessWirelesslinkCache interface {
	Set(ctx context.Context, id uint64, data *model.WirelessWirelesslink, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.WirelessWirelesslink, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.WirelessWirelesslink, error)
	MultiSet(ctx context.Context, data []*model.WirelessWirelesslink, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// wirelessWirelesslinkCache define a cache struct
type wirelessWirelesslinkCache struct {
	cache cache.Cache
}

// NewWirelessWirelesslinkCache new a cache
func NewWirelessWirelesslinkCache(cacheType *database.CacheType) WirelessWirelesslinkCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.WirelessWirelesslink{}
		})
		return &wirelessWirelesslinkCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.WirelessWirelesslink{}
		})
		return &wirelessWirelesslinkCache{cache: c}
	}

	return nil // no cache
}

// GetWirelessWirelesslinkCacheKey cache key
func (c *wirelessWirelesslinkCache) GetWirelessWirelesslinkCacheKey(id uint64) string {
	return wirelessWirelesslinkCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *wirelessWirelesslinkCache) Set(ctx context.Context, id uint64, data *model.WirelessWirelesslink, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetWirelessWirelesslinkCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *wirelessWirelesslinkCache) Get(ctx context.Context, id uint64) (*model.WirelessWirelesslink, error) {
	var data *model.WirelessWirelesslink
	cacheKey := c.GetWirelessWirelesslinkCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *wirelessWirelesslinkCache) MultiSet(ctx context.Context, data []*model.WirelessWirelesslink, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetWirelessWirelesslinkCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *wirelessWirelesslinkCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.WirelessWirelesslink, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetWirelessWirelesslinkCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.WirelessWirelesslink)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.WirelessWirelesslink)
	for _, id := range ids {
		val, ok := itemMap[c.GetWirelessWirelesslinkCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *wirelessWirelesslinkCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetWirelessWirelesslinkCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *wirelessWirelesslinkCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetWirelessWirelesslinkCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *wirelessWirelesslinkCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
