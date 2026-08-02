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
	dcimPowerfeedCachePrefixKey = "dcimPowerfeed:"
	// DcimPowerfeedExpireTime expire time
	DcimPowerfeedExpireTime = 5 * time.Minute
)

var _ DcimPowerfeedCache = (*dcimPowerfeedCache)(nil)

// DcimPowerfeedCache cache interface
type DcimPowerfeedCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimPowerfeed, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimPowerfeed, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimPowerfeed, error)
	MultiSet(ctx context.Context, data []*model.DcimPowerfeed, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimPowerfeedCache define a cache struct
type dcimPowerfeedCache struct {
	cache cache.Cache
}

// NewDcimPowerfeedCache new a cache
func NewDcimPowerfeedCache(cacheType *database.CacheType) DcimPowerfeedCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimPowerfeed{}
		})
		return &dcimPowerfeedCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimPowerfeed{}
		})
		return &dcimPowerfeedCache{cache: c}
	}

	return nil // no cache
}

// GetDcimPowerfeedCacheKey cache key
func (c *dcimPowerfeedCache) GetDcimPowerfeedCacheKey(id uint64) string {
	return dcimPowerfeedCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimPowerfeedCache) Set(ctx context.Context, id uint64, data *model.DcimPowerfeed, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimPowerfeedCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimPowerfeedCache) Get(ctx context.Context, id uint64) (*model.DcimPowerfeed, error) {
	var data *model.DcimPowerfeed
	cacheKey := c.GetDcimPowerfeedCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimPowerfeedCache) MultiSet(ctx context.Context, data []*model.DcimPowerfeed, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimPowerfeedCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimPowerfeedCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimPowerfeed, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimPowerfeedCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimPowerfeed)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimPowerfeed)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimPowerfeedCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimPowerfeedCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimPowerfeedCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimPowerfeedCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimPowerfeedCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimPowerfeedCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
