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
	extrasJournalentryCachePrefixKey = "extrasJournalentry:"
	// ExtrasJournalentryExpireTime expire time
	ExtrasJournalentryExpireTime = 5 * time.Minute
)

var _ ExtrasJournalentryCache = (*extrasJournalentryCache)(nil)

// ExtrasJournalentryCache cache interface
type ExtrasJournalentryCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasJournalentry, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasJournalentry, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasJournalentry, error)
	MultiSet(ctx context.Context, data []*model.ExtrasJournalentry, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasJournalentryCache define a cache struct
type extrasJournalentryCache struct {
	cache cache.Cache
}

// NewExtrasJournalentryCache new a cache
func NewExtrasJournalentryCache(cacheType *database.CacheType) ExtrasJournalentryCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasJournalentry{}
		})
		return &extrasJournalentryCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasJournalentry{}
		})
		return &extrasJournalentryCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasJournalentryCacheKey cache key
func (c *extrasJournalentryCache) GetExtrasJournalentryCacheKey(id uint64) string {
	return extrasJournalentryCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasJournalentryCache) Set(ctx context.Context, id uint64, data *model.ExtrasJournalentry, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasJournalentryCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasJournalentryCache) Get(ctx context.Context, id uint64) (*model.ExtrasJournalentry, error) {
	var data *model.ExtrasJournalentry
	cacheKey := c.GetExtrasJournalentryCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasJournalentryCache) MultiSet(ctx context.Context, data []*model.ExtrasJournalentry, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasJournalentryCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasJournalentryCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasJournalentry, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasJournalentryCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasJournalentry)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasJournalentry)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasJournalentryCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasJournalentryCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasJournalentryCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasJournalentryCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasJournalentryCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasJournalentryCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
