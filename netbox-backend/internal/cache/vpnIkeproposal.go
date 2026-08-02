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
	vpnIkeproposalCachePrefixKey = "vpnIkeproposal:"
	// VpnIkeproposalExpireTime expire time
	VpnIkeproposalExpireTime = 5 * time.Minute
)

var _ VpnIkeproposalCache = (*vpnIkeproposalCache)(nil)

// VpnIkeproposalCache cache interface
type VpnIkeproposalCache interface {
	Set(ctx context.Context, id uint64, data *model.VpnIkeproposal, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VpnIkeproposal, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnIkeproposal, error)
	MultiSet(ctx context.Context, data []*model.VpnIkeproposal, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// vpnIkeproposalCache define a cache struct
type vpnIkeproposalCache struct {
	cache cache.Cache
}

// NewVpnIkeproposalCache new a cache
func NewVpnIkeproposalCache(cacheType *database.CacheType) VpnIkeproposalCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnIkeproposal{}
		})
		return &vpnIkeproposalCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnIkeproposal{}
		})
		return &vpnIkeproposalCache{cache: c}
	}

	return nil // no cache
}

// GetVpnIkeproposalCacheKey cache key
func (c *vpnIkeproposalCache) GetVpnIkeproposalCacheKey(id uint64) string {
	return vpnIkeproposalCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *vpnIkeproposalCache) Set(ctx context.Context, id uint64, data *model.VpnIkeproposal, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVpnIkeproposalCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *vpnIkeproposalCache) Get(ctx context.Context, id uint64) (*model.VpnIkeproposal, error) {
	var data *model.VpnIkeproposal
	cacheKey := c.GetVpnIkeproposalCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *vpnIkeproposalCache) MultiSet(ctx context.Context, data []*model.VpnIkeproposal, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVpnIkeproposalCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *vpnIkeproposalCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnIkeproposal, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVpnIkeproposalCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VpnIkeproposal)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VpnIkeproposal)
	for _, id := range ids {
		val, ok := itemMap[c.GetVpnIkeproposalCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *vpnIkeproposalCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnIkeproposalCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *vpnIkeproposalCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnIkeproposalCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *vpnIkeproposalCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
