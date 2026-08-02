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
	dcimPowerpanelCachePrefixKey = "dcimPowerpanel:"
	// DcimPowerpanelExpireTime expire time
	DcimPowerpanelExpireTime = 5 * time.Minute
)

var _ DcimPowerpanelCache = (*dcimPowerpanelCache)(nil)

// DcimPowerpanelCache cache interface
type DcimPowerpanelCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimPowerpanel, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimPowerpanel, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimPowerpanel, error)
	MultiSet(ctx context.Context, data []*model.DcimPowerpanel, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimPowerpanelCache define a cache struct
type dcimPowerpanelCache struct {
	cache cache.Cache
}

// NewDcimPowerpanelCache new a cache
func NewDcimPowerpanelCache(cacheType *database.CacheType) DcimPowerpanelCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimPowerpanel{}
		})
		return &dcimPowerpanelCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimPowerpanel{}
		})
		return &dcimPowerpanelCache{cache: c}
	}

	return nil // no cache
}

// GetDcimPowerpanelCacheKey cache key
func (c *dcimPowerpanelCache) GetDcimPowerpanelCacheKey(id uint64) string {
	return dcimPowerpanelCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimPowerpanelCache) Set(ctx context.Context, id uint64, data *model.DcimPowerpanel, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimPowerpanelCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimPowerpanelCache) Get(ctx context.Context, id uint64) (*model.DcimPowerpanel, error) {
	var data *model.DcimPowerpanel
	cacheKey := c.GetDcimPowerpanelCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimPowerpanelCache) MultiSet(ctx context.Context, data []*model.DcimPowerpanel, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimPowerpanelCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimPowerpanelCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimPowerpanel, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimPowerpanelCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimPowerpanel)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimPowerpanel)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimPowerpanelCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimPowerpanelCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimPowerpanelCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimPowerpanelCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimPowerpanelCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimPowerpanelCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
