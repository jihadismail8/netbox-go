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
	dcimRackreservationCachePrefixKey = "dcimRackreservation:"
	// DcimRackreservationExpireTime expire time
	DcimRackreservationExpireTime = 5 * time.Minute
)

var _ DcimRackreservationCache = (*dcimRackreservationCache)(nil)

// DcimRackreservationCache cache interface
type DcimRackreservationCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimRackreservation, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimRackreservation, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimRackreservation, error)
	MultiSet(ctx context.Context, data []*model.DcimRackreservation, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimRackreservationCache define a cache struct
type dcimRackreservationCache struct {
	cache cache.Cache
}

// NewDcimRackreservationCache new a cache
func NewDcimRackreservationCache(cacheType *database.CacheType) DcimRackreservationCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimRackreservation{}
		})
		return &dcimRackreservationCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimRackreservation{}
		})
		return &dcimRackreservationCache{cache: c}
	}

	return nil // no cache
}

// GetDcimRackreservationCacheKey cache key
func (c *dcimRackreservationCache) GetDcimRackreservationCacheKey(id uint64) string {
	return dcimRackreservationCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimRackreservationCache) Set(ctx context.Context, id uint64, data *model.DcimRackreservation, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimRackreservationCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimRackreservationCache) Get(ctx context.Context, id uint64) (*model.DcimRackreservation, error) {
	var data *model.DcimRackreservation
	cacheKey := c.GetDcimRackreservationCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimRackreservationCache) MultiSet(ctx context.Context, data []*model.DcimRackreservation, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimRackreservationCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimRackreservationCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimRackreservation, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimRackreservationCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimRackreservation)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimRackreservation)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimRackreservationCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimRackreservationCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimRackreservationCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimRackreservationCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimRackreservationCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimRackreservationCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
