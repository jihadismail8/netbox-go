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
	dcimMacaddressCachePrefixKey = "dcimMacaddress:"
	// DcimMacaddressExpireTime expire time
	DcimMacaddressExpireTime = 5 * time.Minute
)

var _ DcimMacaddressCache = (*dcimMacaddressCache)(nil)

// DcimMacaddressCache cache interface
type DcimMacaddressCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimMacaddress, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimMacaddress, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimMacaddress, error)
	MultiSet(ctx context.Context, data []*model.DcimMacaddress, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimMacaddressCache define a cache struct
type dcimMacaddressCache struct {
	cache cache.Cache
}

// NewDcimMacaddressCache new a cache
func NewDcimMacaddressCache(cacheType *database.CacheType) DcimMacaddressCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimMacaddress{}
		})
		return &dcimMacaddressCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimMacaddress{}
		})
		return &dcimMacaddressCache{cache: c}
	}

	return nil // no cache
}

// GetDcimMacaddressCacheKey cache key
func (c *dcimMacaddressCache) GetDcimMacaddressCacheKey(id uint64) string {
	return dcimMacaddressCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimMacaddressCache) Set(ctx context.Context, id uint64, data *model.DcimMacaddress, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimMacaddressCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimMacaddressCache) Get(ctx context.Context, id uint64) (*model.DcimMacaddress, error) {
	var data *model.DcimMacaddress
	cacheKey := c.GetDcimMacaddressCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimMacaddressCache) MultiSet(ctx context.Context, data []*model.DcimMacaddress, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimMacaddressCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimMacaddressCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimMacaddress, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimMacaddressCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimMacaddress)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimMacaddress)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimMacaddressCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimMacaddressCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimMacaddressCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimMacaddressCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimMacaddressCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimMacaddressCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
