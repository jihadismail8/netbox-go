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
	dcimInterfaceWirelessLansCachePrefixKey = "dcimInterfaceWirelessLans:"
	// DcimInterfaceWirelessLansExpireTime expire time
	DcimInterfaceWirelessLansExpireTime = 5 * time.Minute
)

var _ DcimInterfaceWirelessLansCache = (*dcimInterfaceWirelessLansCache)(nil)

// DcimInterfaceWirelessLansCache cache interface
type DcimInterfaceWirelessLansCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimInterfaceWirelessLans, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimInterfaceWirelessLans, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimInterfaceWirelessLans, error)
	MultiSet(ctx context.Context, data []*model.DcimInterfaceWirelessLans, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimInterfaceWirelessLansCache define a cache struct
type dcimInterfaceWirelessLansCache struct {
	cache cache.Cache
}

// NewDcimInterfaceWirelessLansCache new a cache
func NewDcimInterfaceWirelessLansCache(cacheType *database.CacheType) DcimInterfaceWirelessLansCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimInterfaceWirelessLans{}
		})
		return &dcimInterfaceWirelessLansCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimInterfaceWirelessLans{}
		})
		return &dcimInterfaceWirelessLansCache{cache: c}
	}

	return nil // no cache
}

// GetDcimInterfaceWirelessLansCacheKey cache key
func (c *dcimInterfaceWirelessLansCache) GetDcimInterfaceWirelessLansCacheKey(id uint64) string {
	return dcimInterfaceWirelessLansCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimInterfaceWirelessLansCache) Set(ctx context.Context, id uint64, data *model.DcimInterfaceWirelessLans, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimInterfaceWirelessLansCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimInterfaceWirelessLansCache) Get(ctx context.Context, id uint64) (*model.DcimInterfaceWirelessLans, error) {
	var data *model.DcimInterfaceWirelessLans
	cacheKey := c.GetDcimInterfaceWirelessLansCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimInterfaceWirelessLansCache) MultiSet(ctx context.Context, data []*model.DcimInterfaceWirelessLans, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimInterfaceWirelessLansCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimInterfaceWirelessLansCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimInterfaceWirelessLans, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimInterfaceWirelessLansCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimInterfaceWirelessLans)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimInterfaceWirelessLans)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimInterfaceWirelessLansCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimInterfaceWirelessLansCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimInterfaceWirelessLansCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimInterfaceWirelessLansCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimInterfaceWirelessLansCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimInterfaceWirelessLansCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
