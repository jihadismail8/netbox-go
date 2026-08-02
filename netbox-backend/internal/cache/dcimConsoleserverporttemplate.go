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
	dcimConsoleserverporttemplateCachePrefixKey = "dcimConsoleserverporttemplate:"
	// DcimConsoleserverporttemplateExpireTime expire time
	DcimConsoleserverporttemplateExpireTime = 5 * time.Minute
)

var _ DcimConsoleserverporttemplateCache = (*dcimConsoleserverporttemplateCache)(nil)

// DcimConsoleserverporttemplateCache cache interface
type DcimConsoleserverporttemplateCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimConsoleserverporttemplate, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimConsoleserverporttemplate, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimConsoleserverporttemplate, error)
	MultiSet(ctx context.Context, data []*model.DcimConsoleserverporttemplate, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimConsoleserverporttemplateCache define a cache struct
type dcimConsoleserverporttemplateCache struct {
	cache cache.Cache
}

// NewDcimConsoleserverporttemplateCache new a cache
func NewDcimConsoleserverporttemplateCache(cacheType *database.CacheType) DcimConsoleserverporttemplateCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimConsoleserverporttemplate{}
		})
		return &dcimConsoleserverporttemplateCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimConsoleserverporttemplate{}
		})
		return &dcimConsoleserverporttemplateCache{cache: c}
	}

	return nil // no cache
}

// GetDcimConsoleserverporttemplateCacheKey cache key
func (c *dcimConsoleserverporttemplateCache) GetDcimConsoleserverporttemplateCacheKey(id uint64) string {
	return dcimConsoleserverporttemplateCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimConsoleserverporttemplateCache) Set(ctx context.Context, id uint64, data *model.DcimConsoleserverporttemplate, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimConsoleserverporttemplateCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimConsoleserverporttemplateCache) Get(ctx context.Context, id uint64) (*model.DcimConsoleserverporttemplate, error) {
	var data *model.DcimConsoleserverporttemplate
	cacheKey := c.GetDcimConsoleserverporttemplateCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimConsoleserverporttemplateCache) MultiSet(ctx context.Context, data []*model.DcimConsoleserverporttemplate, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimConsoleserverporttemplateCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimConsoleserverporttemplateCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimConsoleserverporttemplate, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimConsoleserverporttemplateCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimConsoleserverporttemplate)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimConsoleserverporttemplate)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimConsoleserverporttemplateCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimConsoleserverporttemplateCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimConsoleserverporttemplateCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimConsoleserverporttemplateCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimConsoleserverporttemplateCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimConsoleserverporttemplateCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
