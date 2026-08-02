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
	vpnIkepolicyProposalsCachePrefixKey = "vpnIkepolicyProposals:"
	// VpnIkepolicyProposalsExpireTime expire time
	VpnIkepolicyProposalsExpireTime = 5 * time.Minute
)

var _ VpnIkepolicyProposalsCache = (*vpnIkepolicyProposalsCache)(nil)

// VpnIkepolicyProposalsCache cache interface
type VpnIkepolicyProposalsCache interface {
	Set(ctx context.Context, id uint64, data *model.VpnIkepolicyProposals, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.VpnIkepolicyProposals, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnIkepolicyProposals, error)
	MultiSet(ctx context.Context, data []*model.VpnIkepolicyProposals, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// vpnIkepolicyProposalsCache define a cache struct
type vpnIkepolicyProposalsCache struct {
	cache cache.Cache
}

// NewVpnIkepolicyProposalsCache new a cache
func NewVpnIkepolicyProposalsCache(cacheType *database.CacheType) VpnIkepolicyProposalsCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnIkepolicyProposals{}
		})
		return &vpnIkepolicyProposalsCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.VpnIkepolicyProposals{}
		})
		return &vpnIkepolicyProposalsCache{cache: c}
	}

	return nil // no cache
}

// GetVpnIkepolicyProposalsCacheKey cache key
func (c *vpnIkepolicyProposalsCache) GetVpnIkepolicyProposalsCacheKey(id uint64) string {
	return vpnIkepolicyProposalsCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *vpnIkepolicyProposalsCache) Set(ctx context.Context, id uint64, data *model.VpnIkepolicyProposals, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetVpnIkepolicyProposalsCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *vpnIkepolicyProposalsCache) Get(ctx context.Context, id uint64) (*model.VpnIkepolicyProposals, error) {
	var data *model.VpnIkepolicyProposals
	cacheKey := c.GetVpnIkepolicyProposalsCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *vpnIkepolicyProposalsCache) MultiSet(ctx context.Context, data []*model.VpnIkepolicyProposals, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetVpnIkepolicyProposalsCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *vpnIkepolicyProposalsCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.VpnIkepolicyProposals, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetVpnIkepolicyProposalsCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.VpnIkepolicyProposals)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.VpnIkepolicyProposals)
	for _, id := range ids {
		val, ok := itemMap[c.GetVpnIkepolicyProposalsCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *vpnIkepolicyProposalsCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnIkepolicyProposalsCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *vpnIkepolicyProposalsCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetVpnIkepolicyProposalsCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *vpnIkepolicyProposalsCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
