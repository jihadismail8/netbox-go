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
	extrasConfigcontextRolesCachePrefixKey = "extrasConfigcontextRoles:"
	// ExtrasConfigcontextRolesExpireTime expire time
	ExtrasConfigcontextRolesExpireTime = 5 * time.Minute
)

var _ ExtrasConfigcontextRolesCache = (*extrasConfigcontextRolesCache)(nil)

// ExtrasConfigcontextRolesCache cache interface
type ExtrasConfigcontextRolesCache interface {
	Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextRoles, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextRoles, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextRoles, error)
	MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextRoles, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// extrasConfigcontextRolesCache define a cache struct
type extrasConfigcontextRolesCache struct {
	cache cache.Cache
}

// NewExtrasConfigcontextRolesCache new a cache
func NewExtrasConfigcontextRolesCache(cacheType *database.CacheType) ExtrasConfigcontextRolesCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextRoles{}
		})
		return &extrasConfigcontextRolesCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ExtrasConfigcontextRoles{}
		})
		return &extrasConfigcontextRolesCache{cache: c}
	}

	return nil // no cache
}

// GetExtrasConfigcontextRolesCacheKey cache key
func (c *extrasConfigcontextRolesCache) GetExtrasConfigcontextRolesCacheKey(id uint64) string {
	return extrasConfigcontextRolesCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *extrasConfigcontextRolesCache) Set(ctx context.Context, id uint64, data *model.ExtrasConfigcontextRoles, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetExtrasConfigcontextRolesCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *extrasConfigcontextRolesCache) Get(ctx context.Context, id uint64) (*model.ExtrasConfigcontextRoles, error) {
	var data *model.ExtrasConfigcontextRoles
	cacheKey := c.GetExtrasConfigcontextRolesCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *extrasConfigcontextRolesCache) MultiSet(ctx context.Context, data []*model.ExtrasConfigcontextRoles, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetExtrasConfigcontextRolesCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *extrasConfigcontextRolesCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontextRoles, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetExtrasConfigcontextRolesCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ExtrasConfigcontextRoles)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ExtrasConfigcontextRoles)
	for _, id := range ids {
		val, ok := itemMap[c.GetExtrasConfigcontextRolesCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *extrasConfigcontextRolesCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextRolesCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *extrasConfigcontextRolesCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetExtrasConfigcontextRolesCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *extrasConfigcontextRolesCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
