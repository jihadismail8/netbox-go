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
	socialAuthUsersocialauthCachePrefixKey = "socialAuthUsersocialauth:"
	// SocialAuthUsersocialauthExpireTime expire time
	SocialAuthUsersocialauthExpireTime = 5 * time.Minute
)

var _ SocialAuthUsersocialauthCache = (*socialAuthUsersocialauthCache)(nil)

// SocialAuthUsersocialauthCache cache interface
type SocialAuthUsersocialauthCache interface {
	Set(ctx context.Context, id uint64, data *model.SocialAuthUsersocialauth, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.SocialAuthUsersocialauth, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SocialAuthUsersocialauth, error)
	MultiSet(ctx context.Context, data []*model.SocialAuthUsersocialauth, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// socialAuthUsersocialauthCache define a cache struct
type socialAuthUsersocialauthCache struct {
	cache cache.Cache
}

// NewSocialAuthUsersocialauthCache new a cache
func NewSocialAuthUsersocialauthCache(cacheType *database.CacheType) SocialAuthUsersocialauthCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.SocialAuthUsersocialauth{}
		})
		return &socialAuthUsersocialauthCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.SocialAuthUsersocialauth{}
		})
		return &socialAuthUsersocialauthCache{cache: c}
	}

	return nil // no cache
}

// GetSocialAuthUsersocialauthCacheKey cache key
func (c *socialAuthUsersocialauthCache) GetSocialAuthUsersocialauthCacheKey(id uint64) string {
	return socialAuthUsersocialauthCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *socialAuthUsersocialauthCache) Set(ctx context.Context, id uint64, data *model.SocialAuthUsersocialauth, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetSocialAuthUsersocialauthCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *socialAuthUsersocialauthCache) Get(ctx context.Context, id uint64) (*model.SocialAuthUsersocialauth, error) {
	var data *model.SocialAuthUsersocialauth
	cacheKey := c.GetSocialAuthUsersocialauthCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *socialAuthUsersocialauthCache) MultiSet(ctx context.Context, data []*model.SocialAuthUsersocialauth, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetSocialAuthUsersocialauthCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *socialAuthUsersocialauthCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SocialAuthUsersocialauth, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetSocialAuthUsersocialauthCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.SocialAuthUsersocialauth)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.SocialAuthUsersocialauth)
	for _, id := range ids {
		val, ok := itemMap[c.GetSocialAuthUsersocialauthCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *socialAuthUsersocialauthCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetSocialAuthUsersocialauthCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *socialAuthUsersocialauthCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetSocialAuthUsersocialauthCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *socialAuthUsersocialauthCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
