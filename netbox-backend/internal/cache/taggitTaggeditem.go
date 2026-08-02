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
	taggitTaggeditemCachePrefixKey = "taggitTaggeditem:"
	// TaggitTaggeditemExpireTime expire time
	TaggitTaggeditemExpireTime = 5 * time.Minute
)

var _ TaggitTaggeditemCache = (*taggitTaggeditemCache)(nil)

// TaggitTaggeditemCache cache interface
type TaggitTaggeditemCache interface {
	Set(ctx context.Context, id uint64, data *model.TaggitTaggeditem, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.TaggitTaggeditem, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TaggitTaggeditem, error)
	MultiSet(ctx context.Context, data []*model.TaggitTaggeditem, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// taggitTaggeditemCache define a cache struct
type taggitTaggeditemCache struct {
	cache cache.Cache
}

// NewTaggitTaggeditemCache new a cache
func NewTaggitTaggeditemCache(cacheType *database.CacheType) TaggitTaggeditemCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.TaggitTaggeditem{}
		})
		return &taggitTaggeditemCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.TaggitTaggeditem{}
		})
		return &taggitTaggeditemCache{cache: c}
	}

	return nil // no cache
}

// GetTaggitTaggeditemCacheKey cache key
func (c *taggitTaggeditemCache) GetTaggitTaggeditemCacheKey(id uint64) string {
	return taggitTaggeditemCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *taggitTaggeditemCache) Set(ctx context.Context, id uint64, data *model.TaggitTaggeditem, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetTaggitTaggeditemCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *taggitTaggeditemCache) Get(ctx context.Context, id uint64) (*model.TaggitTaggeditem, error) {
	var data *model.TaggitTaggeditem
	cacheKey := c.GetTaggitTaggeditemCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *taggitTaggeditemCache) MultiSet(ctx context.Context, data []*model.TaggitTaggeditem, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetTaggitTaggeditemCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *taggitTaggeditemCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TaggitTaggeditem, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetTaggitTaggeditemCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.TaggitTaggeditem)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.TaggitTaggeditem)
	for _, id := range ids {
		val, ok := itemMap[c.GetTaggitTaggeditemCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *taggitTaggeditemCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetTaggitTaggeditemCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *taggitTaggeditemCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetTaggitTaggeditemCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *taggitTaggeditemCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
