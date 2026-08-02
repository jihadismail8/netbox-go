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
	dcimInventoryitemCachePrefixKey = "dcimInventoryitem:"
	// DcimInventoryitemExpireTime expire time
	DcimInventoryitemExpireTime = 5 * time.Minute
)

var _ DcimInventoryitemCache = (*dcimInventoryitemCache)(nil)

// DcimInventoryitemCache cache interface
type DcimInventoryitemCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimInventoryitem, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimInventoryitem, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimInventoryitem, error)
	MultiSet(ctx context.Context, data []*model.DcimInventoryitem, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimInventoryitemCache define a cache struct
type dcimInventoryitemCache struct {
	cache cache.Cache
}

// NewDcimInventoryitemCache new a cache
func NewDcimInventoryitemCache(cacheType *database.CacheType) DcimInventoryitemCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimInventoryitem{}
		})
		return &dcimInventoryitemCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimInventoryitem{}
		})
		return &dcimInventoryitemCache{cache: c}
	}

	return nil // no cache
}

// GetDcimInventoryitemCacheKey cache key
func (c *dcimInventoryitemCache) GetDcimInventoryitemCacheKey(id uint64) string {
	return dcimInventoryitemCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimInventoryitemCache) Set(ctx context.Context, id uint64, data *model.DcimInventoryitem, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimInventoryitemCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimInventoryitemCache) Get(ctx context.Context, id uint64) (*model.DcimInventoryitem, error) {
	var data *model.DcimInventoryitem
	cacheKey := c.GetDcimInventoryitemCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimInventoryitemCache) MultiSet(ctx context.Context, data []*model.DcimInventoryitem, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimInventoryitemCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimInventoryitemCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimInventoryitem, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimInventoryitemCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimInventoryitem)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimInventoryitem)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimInventoryitemCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimInventoryitemCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimInventoryitemCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimInventoryitemCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimInventoryitemCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimInventoryitemCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
