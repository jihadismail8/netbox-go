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
	socialAuthNonceCachePrefixKey = "socialAuthNonce:"
	// SocialAuthNonceExpireTime expire time
	SocialAuthNonceExpireTime = 5 * time.Minute
)

var _ SocialAuthNonceCache = (*socialAuthNonceCache)(nil)

// SocialAuthNonceCache cache interface
type SocialAuthNonceCache interface {
	Set(ctx context.Context, id uint64, data *model.SocialAuthNonce, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.SocialAuthNonce, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SocialAuthNonce, error)
	MultiSet(ctx context.Context, data []*model.SocialAuthNonce, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// socialAuthNonceCache define a cache struct
type socialAuthNonceCache struct {
	cache cache.Cache
}

// NewSocialAuthNonceCache new a cache
func NewSocialAuthNonceCache(cacheType *database.CacheType) SocialAuthNonceCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.SocialAuthNonce{}
		})
		return &socialAuthNonceCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.SocialAuthNonce{}
		})
		return &socialAuthNonceCache{cache: c}
	}

	return nil // no cache
}

// GetSocialAuthNonceCacheKey cache key
func (c *socialAuthNonceCache) GetSocialAuthNonceCacheKey(id uint64) string {
	return socialAuthNonceCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *socialAuthNonceCache) Set(ctx context.Context, id uint64, data *model.SocialAuthNonce, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetSocialAuthNonceCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *socialAuthNonceCache) Get(ctx context.Context, id uint64) (*model.SocialAuthNonce, error) {
	var data *model.SocialAuthNonce
	cacheKey := c.GetSocialAuthNonceCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *socialAuthNonceCache) MultiSet(ctx context.Context, data []*model.SocialAuthNonce, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetSocialAuthNonceCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *socialAuthNonceCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SocialAuthNonce, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetSocialAuthNonceCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.SocialAuthNonce)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.SocialAuthNonce)
	for _, id := range ids {
		val, ok := itemMap[c.GetSocialAuthNonceCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *socialAuthNonceCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetSocialAuthNonceCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *socialAuthNonceCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetSocialAuthNonceCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *socialAuthNonceCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
