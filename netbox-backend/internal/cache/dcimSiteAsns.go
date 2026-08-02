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
	dcimSiteAsnsCachePrefixKey = "dcimSiteAsns:"
	// DcimSiteAsnsExpireTime expire time
	DcimSiteAsnsExpireTime = 5 * time.Minute
)

var _ DcimSiteAsnsCache = (*dcimSiteAsnsCache)(nil)

// DcimSiteAsnsCache cache interface
type DcimSiteAsnsCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimSiteAsns, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimSiteAsns, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimSiteAsns, error)
	MultiSet(ctx context.Context, data []*model.DcimSiteAsns, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimSiteAsnsCache define a cache struct
type dcimSiteAsnsCache struct {
	cache cache.Cache
}

// NewDcimSiteAsnsCache new a cache
func NewDcimSiteAsnsCache(cacheType *database.CacheType) DcimSiteAsnsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimSiteAsns{}
		})
		return &dcimSiteAsnsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimSiteAsns{}
		})
		return &dcimSiteAsnsCache{cache: c}
	}

	return nil // no cache
}

// GetDcimSiteAsnsCacheKey cache key
func (c *dcimSiteAsnsCache) GetDcimSiteAsnsCacheKey(id uint64) string {
	return dcimSiteAsnsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimSiteAsnsCache) Set(ctx context.Context, id uint64, data *model.DcimSiteAsns, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimSiteAsnsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimSiteAsnsCache) Get(ctx context.Context, id uint64) (*model.DcimSiteAsns, error) {
	var data *model.DcimSiteAsns
	cacheKey := c.GetDcimSiteAsnsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimSiteAsnsCache) MultiSet(ctx context.Context, data []*model.DcimSiteAsns, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimSiteAsnsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimSiteAsnsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimSiteAsns, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimSiteAsnsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimSiteAsns)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimSiteAsns)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimSiteAsnsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimSiteAsnsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimSiteAsnsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimSiteAsnsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimSiteAsnsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimSiteAsnsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
