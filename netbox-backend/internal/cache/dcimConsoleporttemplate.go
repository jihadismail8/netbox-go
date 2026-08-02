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
	dcimConsoleporttemplateCachePrefixKey = "dcimConsoleporttemplate:"
	// DcimConsoleporttemplateExpireTime expire time
	DcimConsoleporttemplateExpireTime = 5 * time.Minute
)

var _ DcimConsoleporttemplateCache = (*dcimConsoleporttemplateCache)(nil)

// DcimConsoleporttemplateCache cache interface
type DcimConsoleporttemplateCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimConsoleporttemplate, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimConsoleporttemplate, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimConsoleporttemplate, error)
	MultiSet(ctx context.Context, data []*model.DcimConsoleporttemplate, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimConsoleporttemplateCache define a cache struct
type dcimConsoleporttemplateCache struct {
	cache cache.Cache
}

// NewDcimConsoleporttemplateCache new a cache
func NewDcimConsoleporttemplateCache(cacheType *database.CacheType) DcimConsoleporttemplateCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimConsoleporttemplate{}
		})
		return &dcimConsoleporttemplateCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimConsoleporttemplate{}
		})
		return &dcimConsoleporttemplateCache{cache: c}
	}

	return nil // no cache
}

// GetDcimConsoleporttemplateCacheKey cache key
func (c *dcimConsoleporttemplateCache) GetDcimConsoleporttemplateCacheKey(id uint64) string {
	return dcimConsoleporttemplateCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimConsoleporttemplateCache) Set(ctx context.Context, id uint64, data *model.DcimConsoleporttemplate, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimConsoleporttemplateCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimConsoleporttemplateCache) Get(ctx context.Context, id uint64) (*model.DcimConsoleporttemplate, error) {
	var data *model.DcimConsoleporttemplate
	cacheKey := c.GetDcimConsoleporttemplateCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimConsoleporttemplateCache) MultiSet(ctx context.Context, data []*model.DcimConsoleporttemplate, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimConsoleporttemplateCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimConsoleporttemplateCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimConsoleporttemplate, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimConsoleporttemplateCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimConsoleporttemplate)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimConsoleporttemplate)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimConsoleporttemplateCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimConsoleporttemplateCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimConsoleporttemplateCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimConsoleporttemplateCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimConsoleporttemplateCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimConsoleporttemplateCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
