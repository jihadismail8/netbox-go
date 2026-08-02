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
	dcimCableCachePrefixKey = "dcimCable:"
	// DcimCableExpireTime expire time
	DcimCableExpireTime = 5 * time.Minute
)

var _ DcimCableCache = (*dcimCableCache)(nil)

// DcimCableCache cache interface
type DcimCableCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimCable, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimCable, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimCable, error)
	MultiSet(ctx context.Context, data []*model.DcimCable, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimCableCache define a cache struct
type dcimCableCache struct {
	cache cache.Cache
}

// NewDcimCableCache new a cache
func NewDcimCableCache(cacheType *database.CacheType) DcimCableCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimCable{}
		})
		return &dcimCableCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimCable{}
		})
		return &dcimCableCache{cache: c}
	}

	return nil // no cache
}

// GetDcimCableCacheKey cache key
func (c *dcimCableCache) GetDcimCableCacheKey(id uint64) string {
	return dcimCableCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimCableCache) Set(ctx context.Context, id uint64, data *model.DcimCable, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimCableCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimCableCache) Get(ctx context.Context, id uint64) (*model.DcimCable, error) {
	var data *model.DcimCable
	cacheKey := c.GetDcimCableCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimCableCache) MultiSet(ctx context.Context, data []*model.DcimCable, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimCableCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimCableCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimCable, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimCableCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimCable)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimCable)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimCableCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimCableCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimCableCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimCableCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimCableCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimCableCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
