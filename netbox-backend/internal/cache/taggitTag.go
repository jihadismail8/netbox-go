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
	taggitTagCachePrefixKey = "taggitTag:"
	// TaggitTagExpireTime expire time
	TaggitTagExpireTime = 5 * time.Minute
)

var _ TaggitTagCache = (*taggitTagCache)(nil)

// TaggitTagCache cache interface
type TaggitTagCache interface {
	Set(ctx context.Context, id uint64, data *model.TaggitTag, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.TaggitTag, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TaggitTag, error)
	MultiSet(ctx context.Context, data []*model.TaggitTag, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// taggitTagCache define a cache struct
type taggitTagCache struct {
	cache cache.Cache
}

// NewTaggitTagCache new a cache
func NewTaggitTagCache(cacheType *database.CacheType) TaggitTagCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.TaggitTag{}
		})
		return &taggitTagCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.TaggitTag{}
		})
		return &taggitTagCache{cache: c}
	}

	return nil // no cache
}

// GetTaggitTagCacheKey cache key
func (c *taggitTagCache) GetTaggitTagCacheKey(id uint64) string {
	return taggitTagCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *taggitTagCache) Set(ctx context.Context, id uint64, data *model.TaggitTag, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetTaggitTagCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *taggitTagCache) Get(ctx context.Context, id uint64) (*model.TaggitTag, error) {
	var data *model.TaggitTag
	cacheKey := c.GetTaggitTagCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *taggitTagCache) MultiSet(ctx context.Context, data []*model.TaggitTag, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetTaggitTagCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *taggitTagCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TaggitTag, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetTaggitTagCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.TaggitTag)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.TaggitTag)
	for _, id := range ids {
		val, ok := itemMap[c.GetTaggitTagCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *taggitTagCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetTaggitTagCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *taggitTagCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetTaggitTagCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *taggitTagCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
