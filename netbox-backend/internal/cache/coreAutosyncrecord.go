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
	coreAutosyncrecordCachePrefixKey = "coreAutosyncrecord:"
	// CoreAutosyncrecordExpireTime expire time
	CoreAutosyncrecordExpireTime = 5 * time.Minute
)

var _ CoreAutosyncrecordCache = (*coreAutosyncrecordCache)(nil)

// CoreAutosyncrecordCache cache interface
type CoreAutosyncrecordCache interface {
	Set(ctx context.Context, id uint64, data *model.CoreAutosyncrecord, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.CoreAutosyncrecord, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CoreAutosyncrecord, error)
	MultiSet(ctx context.Context, data []*model.CoreAutosyncrecord, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// coreAutosyncrecordCache define a cache struct
type coreAutosyncrecordCache struct {
	cache cache.Cache
}

// NewCoreAutosyncrecordCache new a cache
func NewCoreAutosyncrecordCache(cacheType *database.CacheType) CoreAutosyncrecordCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreAutosyncrecord{}
		})
		return &coreAutosyncrecordCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.CoreAutosyncrecord{}
		})
		return &coreAutosyncrecordCache{cache: c}
	}

	return nil // no cache
}

// GetCoreAutosyncrecordCacheKey cache key
func (c *coreAutosyncrecordCache) GetCoreAutosyncrecordCacheKey(id uint64) string {
	return coreAutosyncrecordCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *coreAutosyncrecordCache) Set(ctx context.Context, id uint64, data *model.CoreAutosyncrecord, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetCoreAutosyncrecordCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *coreAutosyncrecordCache) Get(ctx context.Context, id uint64) (*model.CoreAutosyncrecord, error) {
	var data *model.CoreAutosyncrecord
	cacheKey := c.GetCoreAutosyncrecordCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *coreAutosyncrecordCache) MultiSet(ctx context.Context, data []*model.CoreAutosyncrecord, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetCoreAutosyncrecordCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *coreAutosyncrecordCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.CoreAutosyncrecord, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetCoreAutosyncrecordCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.CoreAutosyncrecord)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.CoreAutosyncrecord)
	for _, id := range ids {
		val, ok := itemMap[c.GetCoreAutosyncrecordCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *coreAutosyncrecordCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetCoreAutosyncrecordCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *coreAutosyncrecordCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetCoreAutosyncrecordCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *coreAutosyncrecordCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
