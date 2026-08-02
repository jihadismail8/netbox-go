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
	vpnIpsecproposalCachePrefixKey = "vpnIpsecproposal:"
	// VpnIpsecproposalExpireTime expire time
	VpnIpsecproposalExpireTime = 5 * time.Minute
)

var _ VpnIpsecproposalCache = (*vpnIpsecproposalCache)(nil)

// VpnIpsecproposalCache cache interface
type VpnIpsecproposalCache interface {
	Set(ctx context.Context, id uint64, data *model.VpnIpsecproposal, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VpnIpsecproposal, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnIpsecproposal, error)
	MultiSet(ctx context.Context, data []*model.VpnIpsecproposal, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// vpnIpsecproposalCache define a cache struct
type vpnIpsecproposalCache struct {
	cache cache.Cache
}

// NewVpnIpsecproposalCache new a cache
func NewVpnIpsecproposalCache(cacheType *database.CacheType) VpnIpsecproposalCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnIpsecproposal{}
		})
		return &vpnIpsecproposalCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnIpsecproposal{}
		})
		return &vpnIpsecproposalCache{cache: c}
	}

	return nil // no cache
}

// GetVpnIpsecproposalCacheKey cache key
func (c *vpnIpsecproposalCache) GetVpnIpsecproposalCacheKey(id uint64) string {
	return vpnIpsecproposalCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *vpnIpsecproposalCache) Set(ctx context.Context, id uint64, data *model.VpnIpsecproposal, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVpnIpsecproposalCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *vpnIpsecproposalCache) Get(ctx context.Context, id uint64) (*model.VpnIpsecproposal, error) {
	var data *model.VpnIpsecproposal
	cacheKey := c.GetVpnIpsecproposalCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *vpnIpsecproposalCache) MultiSet(ctx context.Context, data []*model.VpnIpsecproposal, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVpnIpsecproposalCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *vpnIpsecproposalCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnIpsecproposal, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVpnIpsecproposalCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VpnIpsecproposal)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VpnIpsecproposal)
	for _, id := range ids {
		val, ok := itemMap[c.GetVpnIpsecproposalCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *vpnIpsecproposalCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnIpsecproposalCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *vpnIpsecproposalCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnIpsecproposalCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *vpnIpsecproposalCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
