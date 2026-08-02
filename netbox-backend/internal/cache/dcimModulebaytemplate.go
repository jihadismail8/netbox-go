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
	dcimModulebaytemplateCachePrefixKey = "dcimModulebaytemplate:"
	// DcimModulebaytemplateExpireTime expire time
	DcimModulebaytemplateExpireTime = 5 * time.Minute
)

var _ DcimModulebaytemplateCache = (*dcimModulebaytemplateCache)(nil)

// DcimModulebaytemplateCache cache interface
type DcimModulebaytemplateCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimModulebaytemplate, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimModulebaytemplate, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimModulebaytemplate, error)
	MultiSet(ctx context.Context, data []*model.DcimModulebaytemplate, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimModulebaytemplateCache define a cache struct
type dcimModulebaytemplateCache struct {
	cache cache.Cache
}

// NewDcimModulebaytemplateCache new a cache
func NewDcimModulebaytemplateCache(cacheType *database.CacheType) DcimModulebaytemplateCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimModulebaytemplate{}
		})
		return &dcimModulebaytemplateCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimModulebaytemplate{}
		})
		return &dcimModulebaytemplateCache{cache: c}
	}

	return nil // no cache
}

// GetDcimModulebaytemplateCacheKey cache key
func (c *dcimModulebaytemplateCache) GetDcimModulebaytemplateCacheKey(id uint64) string {
	return dcimModulebaytemplateCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimModulebaytemplateCache) Set(ctx context.Context, id uint64, data *model.DcimModulebaytemplate, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimModulebaytemplateCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimModulebaytemplateCache) Get(ctx context.Context, id uint64) (*model.DcimModulebaytemplate, error) {
	var data *model.DcimModulebaytemplate
	cacheKey := c.GetDcimModulebaytemplateCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimModulebaytemplateCache) MultiSet(ctx context.Context, data []*model.DcimModulebaytemplate, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimModulebaytemplateCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimModulebaytemplateCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimModulebaytemplate, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimModulebaytemplateCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimModulebaytemplate)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimModulebaytemplate)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimModulebaytemplateCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimModulebaytemplateCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimModulebaytemplateCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimModulebaytemplateCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimModulebaytemplateCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimModulebaytemplateCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
