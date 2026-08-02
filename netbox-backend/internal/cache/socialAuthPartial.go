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
	socialAuthPartialCachePrefixKey = "socialAuthPartial:"
	// SocialAuthPartialExpireTime expire time
	SocialAuthPartialExpireTime = 5 * time.Minute
)

var _ SocialAuthPartialCache = (*socialAuthPartialCache)(nil)

// SocialAuthPartialCache cache interface
type SocialAuthPartialCache interface {
	Set(ctx context.Context, id uint64, data *model.SocialAuthPartial, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.SocialAuthPartial, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SocialAuthPartial, error)
	MultiSet(ctx context.Context, data []*model.SocialAuthPartial, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// socialAuthPartialCache define a cache struct
type socialAuthPartialCache struct {
	cache cache.Cache
}

// NewSocialAuthPartialCache new a cache
func NewSocialAuthPartialCache(cacheType *database.CacheType) SocialAuthPartialCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.SocialAuthPartial{}
		})
		return &socialAuthPartialCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.SocialAuthPartial{}
		})
		return &socialAuthPartialCache{cache: c}
	}

	return nil // no cache
}

// GetSocialAuthPartialCacheKey cache key
func (c *socialAuthPartialCache) GetSocialAuthPartialCacheKey(id uint64) string {
	return socialAuthPartialCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *socialAuthPartialCache) Set(ctx context.Context, id uint64, data *model.SocialAuthPartial, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetSocialAuthPartialCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *socialAuthPartialCache) Get(ctx context.Context, id uint64) (*model.SocialAuthPartial, error) {
	var data *model.SocialAuthPartial
	cacheKey := c.GetSocialAuthPartialCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *socialAuthPartialCache) MultiSet(ctx context.Context, data []*model.SocialAuthPartial, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetSocialAuthPartialCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *socialAuthPartialCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SocialAuthPartial, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetSocialAuthPartialCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.SocialAuthPartial)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.SocialAuthPartial)
	for _, id := range ids {
		val, ok := itemMap[c.GetSocialAuthPartialCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *socialAuthPartialCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetSocialAuthPartialCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *socialAuthPartialCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetSocialAuthPartialCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *socialAuthPartialCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
