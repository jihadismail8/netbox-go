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
	dcimPoweroutletCachePrefixKey = "dcimPoweroutlet:"
	// DcimPoweroutletExpireTime expire time
	DcimPoweroutletExpireTime = 5 * time.Minute
)

var _ DcimPoweroutletCache = (*dcimPoweroutletCache)(nil)

// DcimPoweroutletCache cache interface
type DcimPoweroutletCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimPoweroutlet, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimPoweroutlet, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimPoweroutlet, error)
	MultiSet(ctx context.Context, data []*model.DcimPoweroutlet, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimPoweroutletCache define a cache struct
type dcimPoweroutletCache struct {
	cache cache.Cache
}

// NewDcimPoweroutletCache new a cache
func NewDcimPoweroutletCache(cacheType *database.CacheType) DcimPoweroutletCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimPoweroutlet{}
		})
		return &dcimPoweroutletCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimPoweroutlet{}
		})
		return &dcimPoweroutletCache{cache: c}
	}

	return nil // no cache
}

// GetDcimPoweroutletCacheKey cache key
func (c *dcimPoweroutletCache) GetDcimPoweroutletCacheKey(id uint64) string {
	return dcimPoweroutletCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimPoweroutletCache) Set(ctx context.Context, id uint64, data *model.DcimPoweroutlet, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimPoweroutletCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimPoweroutletCache) Get(ctx context.Context, id uint64) (*model.DcimPoweroutlet, error) {
	var data *model.DcimPoweroutlet
	cacheKey := c.GetDcimPoweroutletCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimPoweroutletCache) MultiSet(ctx context.Context, data []*model.DcimPoweroutlet, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimPoweroutletCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimPoweroutletCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimPoweroutlet, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimPoweroutletCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimPoweroutlet)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimPoweroutlet)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimPoweroutletCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimPoweroutletCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimPoweroutletCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimPoweroutletCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimPoweroutletCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimPoweroutletCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
