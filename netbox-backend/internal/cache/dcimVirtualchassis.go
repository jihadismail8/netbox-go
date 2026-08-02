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
	dcimVirtualchassisCachePrefixKey = "dcimVirtualchassis:"
	// DcimVirtualchassisExpireTime expire time
	DcimVirtualchassisExpireTime = 5 * time.Minute
)

var _ DcimVirtualchassisCache = (*dcimVirtualchassisCache)(nil)

// DcimVirtualchassisCache cache interface
type DcimVirtualchassisCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimVirtualchassis, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimVirtualchassis, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimVirtualchassis, error)
	MultiSet(ctx context.Context, data []*model.DcimVirtualchassis, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimVirtualchassisCache define a cache struct
type dcimVirtualchassisCache struct {
	cache cache.Cache
}

// NewDcimVirtualchassisCache new a cache
func NewDcimVirtualchassisCache(cacheType *database.CacheType) DcimVirtualchassisCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimVirtualchassis{}
		})
		return &dcimVirtualchassisCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimVirtualchassis{}
		})
		return &dcimVirtualchassisCache{cache: c}
	}

	return nil // no cache
}

// GetDcimVirtualchassisCacheKey cache key
func (c *dcimVirtualchassisCache) GetDcimVirtualchassisCacheKey(id uint64) string {
	return dcimVirtualchassisCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimVirtualchassisCache) Set(ctx context.Context, id uint64, data *model.DcimVirtualchassis, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimVirtualchassisCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimVirtualchassisCache) Get(ctx context.Context, id uint64) (*model.DcimVirtualchassis, error) {
	var data *model.DcimVirtualchassis
	cacheKey := c.GetDcimVirtualchassisCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimVirtualchassisCache) MultiSet(ctx context.Context, data []*model.DcimVirtualchassis, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimVirtualchassisCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimVirtualchassisCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimVirtualchassis, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimVirtualchassisCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimVirtualchassis)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimVirtualchassis)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimVirtualchassisCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimVirtualchassisCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimVirtualchassisCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimVirtualchassisCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimVirtualchassisCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimVirtualchassisCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
