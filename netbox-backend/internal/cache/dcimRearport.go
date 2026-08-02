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
	dcimRearportCachePrefixKey = "dcimRearport:"
	// DcimRearportExpireTime expire time
	DcimRearportExpireTime = 5 * time.Minute
)

var _ DcimRearportCache = (*dcimRearportCache)(nil)

// DcimRearportCache cache interface
type DcimRearportCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimRearport, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimRearport, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimRearport, error)
	MultiSet(ctx context.Context, data []*model.DcimRearport, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimRearportCache define a cache struct
type dcimRearportCache struct {
	cache cache.Cache
}

// NewDcimRearportCache new a cache
func NewDcimRearportCache(cacheType *database.CacheType) DcimRearportCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimRearport{}
		})
		return &dcimRearportCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimRearport{}
		})
		return &dcimRearportCache{cache: c}
	}

	return nil // no cache
}

// GetDcimRearportCacheKey cache key
func (c *dcimRearportCache) GetDcimRearportCacheKey(id uint64) string {
	return dcimRearportCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimRearportCache) Set(ctx context.Context, id uint64, data *model.DcimRearport, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimRearportCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimRearportCache) Get(ctx context.Context, id uint64) (*model.DcimRearport, error) {
	var data *model.DcimRearport
	cacheKey := c.GetDcimRearportCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimRearportCache) MultiSet(ctx context.Context, data []*model.DcimRearport, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimRearportCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimRearportCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimRearport, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimRearportCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimRearport)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimRearport)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimRearportCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimRearportCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimRearportCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimRearportCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimRearportCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimRearportCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
