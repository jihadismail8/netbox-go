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
	dcimInterfaceTaggedVlansCachePrefixKey = "dcimInterfaceTaggedVlans:"
	// DcimInterfaceTaggedVlansExpireTime expire time
	DcimInterfaceTaggedVlansExpireTime = 5 * time.Minute
)

var _ DcimInterfaceTaggedVlansCache = (*dcimInterfaceTaggedVlansCache)(nil)

// DcimInterfaceTaggedVlansCache cache interface
type DcimInterfaceTaggedVlansCache interface {
	Set(ctx context.Context, id uint64, data *model.DcimInterfaceTaggedVlans, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.DcimInterfaceTaggedVlans, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimInterfaceTaggedVlans, error)
	MultiSet(ctx context.Context, data []*model.DcimInterfaceTaggedVlans, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// dcimInterfaceTaggedVlansCache define a cache struct
type dcimInterfaceTaggedVlansCache struct {
	cache cache.Cache
}

// NewDcimInterfaceTaggedVlansCache new a cache
func NewDcimInterfaceTaggedVlansCache(cacheType *database.CacheType) DcimInterfaceTaggedVlansCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimInterfaceTaggedVlans{}
		})
		return &dcimInterfaceTaggedVlansCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.DcimInterfaceTaggedVlans{}
		})
		return &dcimInterfaceTaggedVlansCache{cache: c}
	}

	return nil // no cache
}

// GetDcimInterfaceTaggedVlansCacheKey cache key
func (c *dcimInterfaceTaggedVlansCache) GetDcimInterfaceTaggedVlansCacheKey(id uint64) string {
	return dcimInterfaceTaggedVlansCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *dcimInterfaceTaggedVlansCache) Set(ctx context.Context, id uint64, data *model.DcimInterfaceTaggedVlans, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetDcimInterfaceTaggedVlansCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *dcimInterfaceTaggedVlansCache) Get(ctx context.Context, id uint64) (*model.DcimInterfaceTaggedVlans, error) {
	var data *model.DcimInterfaceTaggedVlans
	cacheKey := c.GetDcimInterfaceTaggedVlansCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *dcimInterfaceTaggedVlansCache) MultiSet(ctx context.Context, data []*model.DcimInterfaceTaggedVlans, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetDcimInterfaceTaggedVlansCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *dcimInterfaceTaggedVlansCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.DcimInterfaceTaggedVlans, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetDcimInterfaceTaggedVlansCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.DcimInterfaceTaggedVlans)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.DcimInterfaceTaggedVlans)
	for _, id := range ids {
		val, ok := itemMap[c.GetDcimInterfaceTaggedVlansCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *dcimInterfaceTaggedVlansCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimInterfaceTaggedVlansCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *dcimInterfaceTaggedVlansCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetDcimInterfaceTaggedVlansCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *dcimInterfaceTaggedVlansCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
