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
	dcimModuletypeCachePrefixKey = "dcimModuletype:"
	// DcimModuletypeExpireTime expire time
	DcimModuletypeExpireTime = 5 * time.Minute
)

var _ DcimModuletypeCache = (*dcimModuletypeCache)(nil)

// DcimModuletypeCache cache interface
type DcimModuletypeCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimModuletype, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimModuletype, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimModuletype, error)
	MultiSet(ctx context.Context, data []*model.DcimModuletype, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimModuletypeCache define a cache struct
type dcimModuletypeCache struct {
	cache cache.Cache
}

// NewDcimModuletypeCache new a cache
func NewDcimModuletypeCache(cacheType *database.CacheType) DcimModuletypeCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimModuletype{}
		})
		return &dcimModuletypeCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimModuletype{}
		})
		return &dcimModuletypeCache{cache: c}
	}

	return nil // no cache
}

// GetDcimModuletypeCacheKey cache key
func (c *dcimModuletypeCache) GetDcimModuletypeCacheKey(id uint64) string {
	return dcimModuletypeCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimModuletypeCache) Set(ctx context.Context, id uint64, data *model.DcimModuletype, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimModuletypeCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimModuletypeCache) Get(ctx context.Context, id uint64) (*model.DcimModuletype, error) {
	var data *model.DcimModuletype
	cacheKey := c.GetDcimModuletypeCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimModuletypeCache) MultiSet(ctx context.Context, data []*model.DcimModuletype, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimModuletypeCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimModuletypeCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimModuletype, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimModuletypeCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimModuletype)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimModuletype)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimModuletypeCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimModuletypeCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimModuletypeCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimModuletypeCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimModuletypeCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimModuletypeCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
