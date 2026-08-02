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
	dcimPoweroutlettemplateCachePrefixKey = "dcimPoweroutlettemplate:"
	// DcimPoweroutlettemplateExpireTime expire time
	DcimPoweroutlettemplateExpireTime = 5 * time.Minute
)

var _ DcimPoweroutlettemplateCache = (*dcimPoweroutlettemplateCache)(nil)

// DcimPoweroutlettemplateCache cache interface
type DcimPoweroutlettemplateCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimPoweroutlettemplate, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimPoweroutlettemplate, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimPoweroutlettemplate, error)
	MultiSet(ctx context.Context, data []*model.DcimPoweroutlettemplate, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimPoweroutlettemplateCache define a cache struct
type dcimPoweroutlettemplateCache struct {
	cache cache.Cache
}

// NewDcimPoweroutlettemplateCache new a cache
func NewDcimPoweroutlettemplateCache(cacheType *database.CacheType) DcimPoweroutlettemplateCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimPoweroutlettemplate{}
		})
		return &dcimPoweroutlettemplateCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimPoweroutlettemplate{}
		})
		return &dcimPoweroutlettemplateCache{cache: c}
	}

	return nil // no cache
}

// GetDcimPoweroutlettemplateCacheKey cache key
func (c *dcimPoweroutlettemplateCache) GetDcimPoweroutlettemplateCacheKey(id uint64) string {
	return dcimPoweroutlettemplateCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimPoweroutlettemplateCache) Set(ctx context.Context, id uint64, data *model.DcimPoweroutlettemplate, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimPoweroutlettemplateCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimPoweroutlettemplateCache) Get(ctx context.Context, id uint64) (*model.DcimPoweroutlettemplate, error) {
	var data *model.DcimPoweroutlettemplate
	cacheKey := c.GetDcimPoweroutlettemplateCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimPoweroutlettemplateCache) MultiSet(ctx context.Context, data []*model.DcimPoweroutlettemplate, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimPoweroutlettemplateCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimPoweroutlettemplateCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimPoweroutlettemplate, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimPoweroutlettemplateCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimPoweroutlettemplate)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimPoweroutlettemplate)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimPoweroutlettemplateCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimPoweroutlettemplateCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimPoweroutlettemplateCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimPoweroutlettemplateCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimPoweroutlettemplateCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimPoweroutlettemplateCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
