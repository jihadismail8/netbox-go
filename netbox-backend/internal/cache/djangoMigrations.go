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
	djangoMigrationsCachePrefixKey = "djangoMigrations:"
	// DjangoMigrationsExpireTime expire time
	DjangoMigrationsExpireTime = 5 * time.Minute
)

var _ DjangoMigrationsCache = (*djangoMigrationsCache)(nil)

// DjangoMigrationsCache cache interface
type DjangoMigrationsCache interface {
	Set(ctx context.Context, id uint64, data *model.DjangoMigrations, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DjangoMigrations, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DjangoMigrations, error)
	MultiSet(ctx context.Context, data []*model.DjangoMigrations, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// djangoMigrationsCache define a cache struct
type djangoMigrationsCache struct {
	cache cache.Cache
}

// NewDjangoMigrationsCache new a cache
func NewDjangoMigrationsCache(cacheType *database.CacheType) DjangoMigrationsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DjangoMigrations{}
		})
		return &djangoMigrationsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DjangoMigrations{}
		})
		return &djangoMigrationsCache{cache: c}
	}

	return nil // no cache
}

// GetDjangoMigrationsCacheKey cache key
func (c *djangoMigrationsCache) GetDjangoMigrationsCacheKey(id uint64) string {
	return djangoMigrationsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *djangoMigrationsCache) Set(ctx context.Context, id uint64, data *model.DjangoMigrations, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDjangoMigrationsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *djangoMigrationsCache) Get(ctx context.Context, id uint64) (*model.DjangoMigrations, error) {
	var data *model.DjangoMigrations
	cacheKey := c.GetDjangoMigrationsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *djangoMigrationsCache) MultiSet(ctx context.Context, data []*model.DjangoMigrations, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDjangoMigrationsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *djangoMigrationsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DjangoMigrations, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDjangoMigrationsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DjangoMigrations)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DjangoMigrations)
	for _, id := range ids {
		val, ok := itemMap[c.GetDjangoMigrationsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *djangoMigrationsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDjangoMigrationsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *djangoMigrationsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDjangoMigrationsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *djangoMigrationsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
