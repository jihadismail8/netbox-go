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
	socialAuthCodeCachePrefixKey = "socialAuthCode:"
	// SocialAuthCodeExpireTime expire time
	SocialAuthCodeExpireTime = 5 * time.Minute
)

var _ SocialAuthCodeCache = (*socialAuthCodeCache)(nil)

// SocialAuthCodeCache cache interface
type SocialAuthCodeCache interface {
	Set(ctx context.Context, id uint64, data *model.SocialAuthCode, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.SocialAuthCode, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SocialAuthCode, error)
	MultiSet(ctx context.Context, data []*model.SocialAuthCode, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// socialAuthCodeCache define a cache struct
type socialAuthCodeCache struct {
	cache cache.Cache
}

// NewSocialAuthCodeCache new a cache
func NewSocialAuthCodeCache(cacheType *database.CacheType) SocialAuthCodeCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.SocialAuthCode{}
		})
		return &socialAuthCodeCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.SocialAuthCode{}
		})
		return &socialAuthCodeCache{cache: c}
	}

	return nil // no cache
}

// GetSocialAuthCodeCacheKey cache key
func (c *socialAuthCodeCache) GetSocialAuthCodeCacheKey(id uint64) string {
	return socialAuthCodeCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *socialAuthCodeCache) Set(ctx context.Context, id uint64, data *model.SocialAuthCode, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetSocialAuthCodeCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *socialAuthCodeCache) Get(ctx context.Context, id uint64) (*model.SocialAuthCode, error) {
	var data *model.SocialAuthCode
	cacheKey := c.GetSocialAuthCodeCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *socialAuthCodeCache) MultiSet(ctx context.Context, data []*model.SocialAuthCode, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetSocialAuthCodeCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *socialAuthCodeCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SocialAuthCode, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetSocialAuthCodeCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.SocialAuthCode)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.SocialAuthCode)
	for _, id := range ids {
		val, ok := itemMap[c.GetSocialAuthCodeCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *socialAuthCodeCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetSocialAuthCodeCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *socialAuthCodeCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetSocialAuthCodeCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *socialAuthCodeCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
