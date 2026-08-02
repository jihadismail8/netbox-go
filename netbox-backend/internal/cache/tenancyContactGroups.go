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
	tenancyContactGroupsCachePrefixKey = "tenancyContactGroups:"
	// TenancyContactGroupsExpireTime expire time
	TenancyContactGroupsExpireTime = 5 * time.Minute
)

var _ TenancyContactGroupsCache = (*tenancyContactGroupsCache)(nil)

// TenancyContactGroupsCache cache interface
type TenancyContactGroupsCache interface {
	Set(ctx context.Context, id uint64, data *model.TenancyContactGroups, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.TenancyContactGroups, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TenancyContactGroups, error)
	MultiSet(ctx context.Context, data []*model.TenancyContactGroups, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// tenancyContactGroupsCache define a cache struct
type tenancyContactGroupsCache struct {
	cache cache.Cache
}

// NewTenancyContactGroupsCache new a cache
func NewTenancyContactGroupsCache(cacheType *database.CacheType) TenancyContactGroupsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.TenancyContactGroups{}
		})
		return &tenancyContactGroupsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.TenancyContactGroups{}
		})
		return &tenancyContactGroupsCache{cache: c}
	}

	return nil // no cache
}

// GetTenancyContactGroupsCacheKey cache key
func (c *tenancyContactGroupsCache) GetTenancyContactGroupsCacheKey(id uint64) string {
	return tenancyContactGroupsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *tenancyContactGroupsCache) Set(ctx context.Context, id uint64, data *model.TenancyContactGroups, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetTenancyContactGroupsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *tenancyContactGroupsCache) Get(ctx context.Context, id uint64) (*model.TenancyContactGroups, error) {
	var data *model.TenancyContactGroups
	cacheKey := c.GetTenancyContactGroupsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *tenancyContactGroupsCache) MultiSet(ctx context.Context, data []*model.TenancyContactGroups, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetTenancyContactGroupsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *tenancyContactGroupsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TenancyContactGroups, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetTenancyContactGroupsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.TenancyContactGroups)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.TenancyContactGroups)
	for _, id := range ids {
		val, ok := itemMap[c.GetTenancyContactGroupsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *tenancyContactGroupsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetTenancyContactGroupsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *tenancyContactGroupsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetTenancyContactGroupsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *tenancyContactGroupsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
