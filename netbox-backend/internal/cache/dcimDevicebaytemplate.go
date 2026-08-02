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
	dcimDevicebaytemplateCachePrefixKey = "dcimDevicebaytemplate:"
	// DcimDevicebaytemplateExpireTime expire time
	DcimDevicebaytemplateExpireTime = 5 * time.Minute
)

var _ DcimDevicebaytemplateCache = (*dcimDevicebaytemplateCache)(nil)

// DcimDevicebaytemplateCache cache interface
type DcimDevicebaytemplateCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimDevicebaytemplate, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimDevicebaytemplate, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimDevicebaytemplate, error)
	MultiSet(ctx context.Context, data []*model.DcimDevicebaytemplate, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimDevicebaytemplateCache define a cache struct
type dcimDevicebaytemplateCache struct {
	cache cache.Cache
}

// NewDcimDevicebaytemplateCache new a cache
func NewDcimDevicebaytemplateCache(cacheType *database.CacheType) DcimDevicebaytemplateCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimDevicebaytemplate{}
		})
		return &dcimDevicebaytemplateCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimDevicebaytemplate{}
		})
		return &dcimDevicebaytemplateCache{cache: c}
	}

	return nil // no cache
}

// GetDcimDevicebaytemplateCacheKey cache key
func (c *dcimDevicebaytemplateCache) GetDcimDevicebaytemplateCacheKey(id uint64) string {
	return dcimDevicebaytemplateCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimDevicebaytemplateCache) Set(ctx context.Context, id uint64, data *model.DcimDevicebaytemplate, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimDevicebaytemplateCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimDevicebaytemplateCache) Get(ctx context.Context, id uint64) (*model.DcimDevicebaytemplate, error) {
	var data *model.DcimDevicebaytemplate
	cacheKey := c.GetDcimDevicebaytemplateCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimDevicebaytemplateCache) MultiSet(ctx context.Context, data []*model.DcimDevicebaytemplate, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimDevicebaytemplateCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimDevicebaytemplateCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimDevicebaytemplate, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimDevicebaytemplateCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimDevicebaytemplate)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimDevicebaytemplate)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimDevicebaytemplateCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimDevicebaytemplateCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimDevicebaytemplateCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimDevicebaytemplateCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimDevicebaytemplateCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimDevicebaytemplateCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
