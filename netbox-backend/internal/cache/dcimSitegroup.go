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
	dcimSitegroupCachePrefixKey = "dcimSitegroup:"
	// DcimSitegroupExpireTime expire time
	DcimSitegroupExpireTime = 5 * time.Minute
)

var _ DcimSitegroupCache = (*dcimSitegroupCache)(nil)

// DcimSitegroupCache cache interface
type DcimSitegroupCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimSitegroup, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimSitegroup, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimSitegroup, error)
	MultiSet(ctx context.Context, data []*model.DcimSitegroup, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimSitegroupCache define a cache struct
type dcimSitegroupCache struct {
	cache cache.Cache
}

// NewDcimSitegroupCache new a cache
func NewDcimSitegroupCache(cacheType *database.CacheType) DcimSitegroupCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimSitegroup{}
		})
		return &dcimSitegroupCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimSitegroup{}
		})
		return &dcimSitegroupCache{cache: c}
	}

	return nil // no cache
}

// GetDcimSitegroupCacheKey cache key
func (c *dcimSitegroupCache) GetDcimSitegroupCacheKey(id uint64) string {
	return dcimSitegroupCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimSitegroupCache) Set(ctx context.Context, id uint64, data *model.DcimSitegroup, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimSitegroupCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimSitegroupCache) Get(ctx context.Context, id uint64) (*model.DcimSitegroup, error) {
	var data *model.DcimSitegroup
	cacheKey := c.GetDcimSitegroupCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimSitegroupCache) MultiSet(ctx context.Context, data []*model.DcimSitegroup, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimSitegroupCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimSitegroupCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimSitegroup, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimSitegroupCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimSitegroup)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimSitegroup)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimSitegroupCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimSitegroupCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimSitegroupCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimSitegroupCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimSitegroupCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimSitegroupCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
