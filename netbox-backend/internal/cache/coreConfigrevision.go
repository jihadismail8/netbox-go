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
	coreConfigrevisionCachePrefixKey = "coreConfigrevision:"
	// CoreConfigrevisionExpireTime expire time
	CoreConfigrevisionExpireTime = 5 * time.Minute
)

var _ CoreConfigrevisionCache = (*coreConfigrevisionCache)(nil)

// CoreConfigrevisionCache cache interface
type CoreConfigrevisionCache interface {
	Set(ctx context.Context, id uint64, data *model.CoreConfigrevision, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CoreConfigrevision, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CoreConfigrevision, error)
	MultiSet(ctx context.Context, data []*model.CoreConfigrevision, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// coreConfigrevisionCache define a cache struct
type coreConfigrevisionCache struct {
	cache cache.Cache
}

// NewCoreConfigrevisionCache new a cache
func NewCoreConfigrevisionCache(cacheType *database.CacheType) CoreConfigrevisionCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreConfigrevision{}
		})
		return &coreConfigrevisionCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreConfigrevision{}
		})
		return &coreConfigrevisionCache{cache: c}
	}

	return nil // no cache
}

// GetCoreConfigrevisionCacheKey cache key
func (c *coreConfigrevisionCache) GetCoreConfigrevisionCacheKey(id uint64) string {
	return coreConfigrevisionCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *coreConfigrevisionCache) Set(ctx context.Context, id uint64, data *model.CoreConfigrevision, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCoreConfigrevisionCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *coreConfigrevisionCache) Get(ctx context.Context, id uint64) (*model.CoreConfigrevision, error) {
	var data *model.CoreConfigrevision
	cacheKey := c.GetCoreConfigrevisionCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *coreConfigrevisionCache) MultiSet(ctx context.Context, data []*model.CoreConfigrevision, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCoreConfigrevisionCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *coreConfigrevisionCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CoreConfigrevision, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCoreConfigrevisionCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CoreConfigrevision)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CoreConfigrevision)
	for _, id := range ids {
		val, ok := itemMap[c.GetCoreConfigrevisionCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *coreConfigrevisionCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCoreConfigrevisionCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *coreConfigrevisionCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCoreConfigrevisionCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *coreConfigrevisionCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
