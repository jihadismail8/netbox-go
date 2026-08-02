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
	dcimCableterminationCachePrefixKey = "dcimCabletermination:"
	// DcimCableterminationExpireTime expire time
	DcimCableterminationExpireTime = 5 * time.Minute
)

var _ DcimCableterminationCache = (*dcimCableterminationCache)(nil)

// DcimCableterminationCache cache interface
type DcimCableterminationCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimCabletermination, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimCabletermination, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimCabletermination, error)
	MultiSet(ctx context.Context, data []*model.DcimCabletermination, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimCableterminationCache define a cache struct
type dcimCableterminationCache struct {
	cache cache.Cache
}

// NewDcimCableterminationCache new a cache
func NewDcimCableterminationCache(cacheType *database.CacheType) DcimCableterminationCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimCabletermination{}
		})
		return &dcimCableterminationCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimCabletermination{}
		})
		return &dcimCableterminationCache{cache: c}
	}

	return nil // no cache
}

// GetDcimCableterminationCacheKey cache key
func (c *dcimCableterminationCache) GetDcimCableterminationCacheKey(id uint64) string {
	return dcimCableterminationCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimCableterminationCache) Set(ctx context.Context, id uint64, data *model.DcimCabletermination, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimCableterminationCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimCableterminationCache) Get(ctx context.Context, id uint64) (*model.DcimCabletermination, error) {
	var data *model.DcimCabletermination
	cacheKey := c.GetDcimCableterminationCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimCableterminationCache) MultiSet(ctx context.Context, data []*model.DcimCabletermination, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimCableterminationCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimCableterminationCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimCabletermination, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimCableterminationCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimCabletermination)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimCabletermination)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimCableterminationCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimCableterminationCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimCableterminationCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimCableterminationCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimCableterminationCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimCableterminationCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
