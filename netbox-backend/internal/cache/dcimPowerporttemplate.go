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
	dcimPowerporttemplateCachePrefixKey = "dcimPowerporttemplate:"
	// DcimPowerporttemplateExpireTime expire time
	DcimPowerporttemplateExpireTime = 5 * time.Minute
)

var _ DcimPowerporttemplateCache = (*dcimPowerporttemplateCache)(nil)

// DcimPowerporttemplateCache cache interface
type DcimPowerporttemplateCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimPowerporttemplate, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimPowerporttemplate, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimPowerporttemplate, error)
	MultiSet(ctx context.Context, data []*model.DcimPowerporttemplate, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimPowerporttemplateCache define a cache struct
type dcimPowerporttemplateCache struct {
	cache cache.Cache
}

// NewDcimPowerporttemplateCache new a cache
func NewDcimPowerporttemplateCache(cacheType *database.CacheType) DcimPowerporttemplateCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimPowerporttemplate{}
		})
		return &dcimPowerporttemplateCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimPowerporttemplate{}
		})
		return &dcimPowerporttemplateCache{cache: c}
	}

	return nil // no cache
}

// GetDcimPowerporttemplateCacheKey cache key
func (c *dcimPowerporttemplateCache) GetDcimPowerporttemplateCacheKey(id uint64) string {
	return dcimPowerporttemplateCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimPowerporttemplateCache) Set(ctx context.Context, id uint64, data *model.DcimPowerporttemplate, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimPowerporttemplateCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimPowerporttemplateCache) Get(ctx context.Context, id uint64) (*model.DcimPowerporttemplate, error) {
	var data *model.DcimPowerporttemplate
	cacheKey := c.GetDcimPowerporttemplateCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimPowerporttemplateCache) MultiSet(ctx context.Context, data []*model.DcimPowerporttemplate, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimPowerporttemplateCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimPowerporttemplateCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimPowerporttemplate, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimPowerporttemplateCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimPowerporttemplate)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimPowerporttemplate)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimPowerporttemplateCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimPowerporttemplateCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimPowerporttemplateCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimPowerporttemplateCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimPowerporttemplateCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimPowerporttemplateCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
