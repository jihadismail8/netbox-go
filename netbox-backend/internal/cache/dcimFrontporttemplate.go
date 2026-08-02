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
	dcimFrontporttemplateCachePrefixKey = "dcimFrontporttemplate:"
	// DcimFrontporttemplateExpireTime expire time
	DcimFrontporttemplateExpireTime = 5 * time.Minute
)

var _ DcimFrontporttemplateCache = (*dcimFrontporttemplateCache)(nil)

// DcimFrontporttemplateCache cache interface
type DcimFrontporttemplateCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimFrontporttemplate, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimFrontporttemplate, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimFrontporttemplate, error)
	MultiSet(ctx context.Context, data []*model.DcimFrontporttemplate, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimFrontporttemplateCache define a cache struct
type dcimFrontporttemplateCache struct {
	cache cache.Cache
}

// NewDcimFrontporttemplateCache new a cache
func NewDcimFrontporttemplateCache(cacheType *database.CacheType) DcimFrontporttemplateCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimFrontporttemplate{}
		})
		return &dcimFrontporttemplateCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimFrontporttemplate{}
		})
		return &dcimFrontporttemplateCache{cache: c}
	}

	return nil // no cache
}

// GetDcimFrontporttemplateCacheKey cache key
func (c *dcimFrontporttemplateCache) GetDcimFrontporttemplateCacheKey(id uint64) string {
	return dcimFrontporttemplateCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimFrontporttemplateCache) Set(ctx context.Context, id uint64, data *model.DcimFrontporttemplate, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimFrontporttemplateCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimFrontporttemplateCache) Get(ctx context.Context, id uint64) (*model.DcimFrontporttemplate, error) {
	var data *model.DcimFrontporttemplate
	cacheKey := c.GetDcimFrontporttemplateCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimFrontporttemplateCache) MultiSet(ctx context.Context, data []*model.DcimFrontporttemplate, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimFrontporttemplateCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimFrontporttemplateCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimFrontporttemplate, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimFrontporttemplateCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimFrontporttemplate)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimFrontporttemplate)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimFrontporttemplateCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimFrontporttemplateCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimFrontporttemplateCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimFrontporttemplateCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimFrontporttemplateCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimFrontporttemplateCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
