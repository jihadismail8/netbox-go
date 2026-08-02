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
	socialAuthAssociationCachePrefixKey = "socialAuthAssociation:"
	// SocialAuthAssociationExpireTime expire time
	SocialAuthAssociationExpireTime = 5 * time.Minute
)

var _ SocialAuthAssociationCache = (*socialAuthAssociationCache)(nil)

// SocialAuthAssociationCache cache interface
type SocialAuthAssociationCache interface {
	Set(ctx context.Context, id uint64, data *model.SocialAuthAssociation, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.SocialAuthAssociation, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SocialAuthAssociation, error)
	MultiSet(ctx context.Context, data []*model.SocialAuthAssociation, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// socialAuthAssociationCache define a cache struct
type socialAuthAssociationCache struct {
	cache cache.Cache
}

// NewSocialAuthAssociationCache new a cache
func NewSocialAuthAssociationCache(cacheType *database.CacheType) SocialAuthAssociationCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.SocialAuthAssociation{}
		})
		return &socialAuthAssociationCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.SocialAuthAssociation{}
		})
		return &socialAuthAssociationCache{cache: c}
	}

	return nil // no cache
}

// GetSocialAuthAssociationCacheKey cache key
func (c *socialAuthAssociationCache) GetSocialAuthAssociationCacheKey(id uint64) string {
	return socialAuthAssociationCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *socialAuthAssociationCache) Set(ctx context.Context, id uint64, data *model.SocialAuthAssociation, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetSocialAuthAssociationCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *socialAuthAssociationCache) Get(ctx context.Context, id uint64) (*model.SocialAuthAssociation, error) {
	var data *model.SocialAuthAssociation
	cacheKey := c.GetSocialAuthAssociationCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *socialAuthAssociationCache) MultiSet(ctx context.Context, data []*model.SocialAuthAssociation, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetSocialAuthAssociationCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *socialAuthAssociationCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SocialAuthAssociation, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetSocialAuthAssociationCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.SocialAuthAssociation)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.SocialAuthAssociation)
	for _, id := range ids {
		val, ok := itemMap[c.GetSocialAuthAssociationCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *socialAuthAssociationCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetSocialAuthAssociationCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *socialAuthAssociationCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetSocialAuthAssociationCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *socialAuthAssociationCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
