package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-dev-frame/sponge/pkg/gotest"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"netbox-go/internal/database"
	"netbox-go/internal/model"
)

func newVpnIpsecpolicyCache() *gotest.Cache {
	record1 := &model.VpnIpsecpolicy{}
	record1.ID = 1
	record2 := &model.VpnIpsecpolicy{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVpnIpsecpolicyCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_vpnIpsecpolicyCache_Set(t *testing.T) {
	c := newVpnIpsecpolicyCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIpsecpolicy)
	err := c.ICache.(VpnIpsecpolicyCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VpnIpsecpolicyCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_vpnIpsecpolicyCache_Get(t *testing.T) {
	c := newVpnIpsecpolicyCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIpsecpolicy)
	err := c.ICache.(VpnIpsecpolicyCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnIpsecpolicyCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VpnIpsecpolicyCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_vpnIpsecpolicyCache_MultiGet(t *testing.T) {
	c := newVpnIpsecpolicyCache()
	defer c.Close()

	var testData []*model.VpnIpsecpolicy
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnIpsecpolicy))
	}

	err := c.ICache.(VpnIpsecpolicyCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnIpsecpolicyCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VpnIpsecpolicy))
	}
}

func Test_vpnIpsecpolicyCache_MultiSet(t *testing.T) {
	c := newVpnIpsecpolicyCache()
	defer c.Close()

	var testData []*model.VpnIpsecpolicy
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnIpsecpolicy))
	}

	err := c.ICache.(VpnIpsecpolicyCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnIpsecpolicyCache_Del(t *testing.T) {
	c := newVpnIpsecpolicyCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIpsecpolicy)
	err := c.ICache.(VpnIpsecpolicyCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnIpsecpolicyCache_SetCacheWithNotFound(t *testing.T) {
	c := newVpnIpsecpolicyCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIpsecpolicy)
	err := c.ICache.(VpnIpsecpolicyCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VpnIpsecpolicyCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVpnIpsecpolicyCache(t *testing.T) {
	c := NewVpnIpsecpolicyCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVpnIpsecpolicyCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVpnIpsecpolicyCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
