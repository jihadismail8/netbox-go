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
	dcimLocationCachePrefixKey = "dcimLocation:"
	// DcimLocationExpireTime expire time
	DcimLocationExpireTime = 5 * time.Minute
)

var _ DcimLocationCache = (*dcimLocationCache)(nil)

// DcimLocationCache cache interface
type DcimLocationCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimLocation, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimLocation, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimLocation, error)
	MultiSet(ctx context.Context, data []*model.DcimLocation, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimLocationCache define a cache struct
type dcimLocationCache struct {
	cache cache.Cache
}

// NewDcimLocationCache new a cache
func NewDcimLocationCache(cacheType *database.CacheType) DcimLocationCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimLocation{}
		})
		return &dcimLocationCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimLocation{}
		})
		return &dcimLocationCache{cache: c}
	}

	return nil // no cache
}

// GetDcimLocationCacheKey cache key
func (c *dcimLocationCache) GetDcimLocationCacheKey(id uint64) string {
	return dcimLocationCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimLocationCache) Set(ctx context.Context, id uint64, data *model.DcimLocation, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimLocationCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimLocationCache) Get(ctx context.Context, id uint64) (*model.DcimLocation, error) {
	var data *model.DcimLocation
	cacheKey := c.GetDcimLocationCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimLocationCache) MultiSet(ctx context.Context, data []*model.DcimLocation, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimLocationCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimLocationCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimLocation, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimLocationCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimLocation)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimLocation)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimLocationCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimLocationCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimLocationCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimLocationCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimLocationCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimLocationCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
