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
	vpnIpsecpolicyProposalsCachePrefixKey = "vpnIpsecpolicyProposals:"
	// VpnIpsecpolicyProposalsExpireTime expire time
	VpnIpsecpolicyProposalsExpireTime = 5 * time.Minute
)

var _ VpnIpsecpolicyProposalsCache = (*vpnIpsecpolicyProposalsCache)(nil)

// VpnIpsecpolicyProposalsCache cache interface
type VpnIpsecpolicyProposalsCache interface {
	Set(ctx context.Context, id uint64, data *model.VpnIpsecpolicyProposals, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VpnIpsecpolicyProposals, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnIpsecpolicyProposals, error)
	MultiSet(ctx context.Context, data []*model.VpnIpsecpolicyProposals, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// vpnIpsecpolicyProposalsCache define a cache struct
type vpnIpsecpolicyProposalsCache struct {
	cache cache.Cache
}

// NewVpnIpsecpolicyProposalsCache new a cache
func NewVpnIpsecpolicyProposalsCache(cacheType *database.CacheType) VpnIpsecpolicyProposalsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnIpsecpolicyProposals{}
		})
		return &vpnIpsecpolicyProposalsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnIpsecpolicyProposals{}
		})
		return &vpnIpsecpolicyProposalsCache{cache: c}
	}

	return nil // no cache
}

// GetVpnIpsecpolicyProposalsCacheKey cache key
func (c *vpnIpsecpolicyProposalsCache) GetVpnIpsecpolicyProposalsCacheKey(id uint64) string {
	return vpnIpsecpolicyProposalsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *vpnIpsecpolicyProposalsCache) Set(ctx context.Context, id uint64, data *model.VpnIpsecpolicyProposals, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVpnIpsecpolicyProposalsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *vpnIpsecpolicyProposalsCache) Get(ctx context.Context, id uint64) (*model.VpnIpsecpolicyProposals, error) {
	var data *model.VpnIpsecpolicyProposals
	cacheKey := c.GetVpnIpsecpolicyProposalsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *vpnIpsecpolicyProposalsCache) MultiSet(ctx context.Context, data []*model.VpnIpsecpolicyProposals, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVpnIpsecpolicyProposalsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *vpnIpsecpolicyProposalsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnIpsecpolicyProposals, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVpnIpsecpolicyProposalsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VpnIpsecpolicyProposals)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VpnIpsecpolicyProposals)
	for _, id := range ids {
		val, ok := itemMap[c.GetVpnIpsecpolicyProposalsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *vpnIpsecpolicyProposalsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnIpsecpolicyProposalsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *vpnIpsecpolicyProposalsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnIpsecpolicyProposalsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *vpnIpsecpolicyProposalsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
