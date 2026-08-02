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
	extrasImageattachmentCachePrefixKey = "extrasImageattachment:"
	// ExtrasImageattachmentExpireTime expire time
	ExtrasImageattachmentExpireTime = 5 * time.Minute
)

var _ ExtrasImageattachmentCache = (*extrasImageattachmentCache)(nil)

// ExtrasImageattachmentCache cache interface
type ExtrasImageattachmentCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasImageattachment, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasImageattachment, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasImageattachment, error)
	MultiSet(ctx context.Context, data []*model.ExtrasImageattachment, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasImageattachmentCache define a cache struct
type extrasImageattachmentCache struct {
	cache cache.Cache
}

// NewExtrasImageattachmentCache new a cache
func NewExtrasImageattachmentCache(cacheType *database.CacheType) ExtrasImageattachmentCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasImageattachment{}
		})
		return &extrasImageattachmentCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasImageattachment{}
		})
		return &extrasImageattachmentCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasImageattachmentCacheKey cache key
func (c *extrasImageattachmentCache) GetExtrasImageattachmentCacheKey(id uint64) string {
	return extrasImageattachmentCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasImageattachmentCache) Set(ctx context.Context, id uint64, data *model.ExtrasImageattachment, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasImageattachmentCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasImageattachmentCache) Get(ctx context.Context, id uint64) (*model.ExtrasImageattachment, error) {
	var data *model.ExtrasImageattachment
	cacheKey := c.GetExtrasImageattachmentCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasImageattachmentCache) MultiSet(ctx context.Context, data []*model.ExtrasImageattachment, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasImageattachmentCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasImageattachmentCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasImageattachment, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasImageattachmentCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasImageattachment)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasImageattachment)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasImageattachmentCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasImageattachmentCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasImageattachmentCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasImageattachmentCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasImageattachmentCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasImageattachmentCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
