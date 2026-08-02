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
	dcimInventoryitemtemplateCachePrefixKey = "dcimInventoryitemtemplate:"
	// DcimInventoryitemtemplateExpireTime expire time
	DcimInventoryitemtemplateExpireTime = 5 * time.Minute
)

var _ DcimInventoryitemtemplateCache = (*dcimInventoryitemtemplateCache)(nil)

// DcimInventoryitemtemplateCache cache interface
type DcimInventoryitemtemplateCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimInventoryitemtemplate, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimInventoryitemtemplate, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimInventoryitemtemplate, error)
	MultiSet(ctx context.Context, data []*model.DcimInventoryitemtemplate, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimInventoryitemtemplateCache define a cache struct
type dcimInventoryitemtemplateCache struct {
	cache cache.Cache
}

// NewDcimInventoryitemtemplateCache new a cache
func NewDcimInventoryitemtemplateCache(cacheType *database.CacheType) DcimInventoryitemtemplateCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimInventoryitemtemplate{}
		})
		return &dcimInventoryitemtemplateCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimInventoryitemtemplate{}
		})
		return &dcimInventoryitemtemplateCache{cache: c}
	}

	return nil // no cache
}

// GetDcimInventoryitemtemplateCacheKey cache key
func (c *dcimInventoryitemtemplateCache) GetDcimInventoryitemtemplateCacheKey(id uint64) string {
	return dcimInventoryitemtemplateCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimInventoryitemtemplateCache) Set(ctx context.Context, id uint64, data *model.DcimInventoryitemtemplate, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimInventoryitemtemplateCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimInventoryitemtemplateCache) Get(ctx context.Context, id uint64) (*model.DcimInventoryitemtemplate, error) {
	var data *model.DcimInventoryitemtemplate
	cacheKey := c.GetDcimInventoryitemtemplateCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimInventoryitemtemplateCache) MultiSet(ctx context.Context, data []*model.DcimInventoryitemtemplate, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimInventoryitemtemplateCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimInventoryitemtemplateCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimInventoryitemtemplate, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimInventoryitemtemplateCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimInventoryitemtemplate)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimInventoryitemtemplate)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimInventoryitemtemplateCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimInventoryitemtemplateCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimInventoryitemtemplateCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimInventoryitemtemplateCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimInventoryitemtemplateCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimInventoryitemtemplateCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
