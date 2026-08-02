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

func newVpnL2VpnCache() *gotest.Cache {
	record1 := &model.VpnL2Vpn{}
	record1.ID = 1
	record2 := &model.VpnL2Vpn{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVpnL2VpnCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_vpnL2VpnCache_Set(t *testing.T) {
	c := newVpnL2VpnCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnL2Vpn)
	err := c.ICache.(VpnL2VpnCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VpnL2VpnCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_vpnL2VpnCache_Get(t *testing.T) {
	c := newVpnL2VpnCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnL2Vpn)
	err := c.ICache.(VpnL2VpnCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnL2VpnCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VpnL2VpnCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_vpnL2VpnCache_MultiGet(t *testing.T) {
	c := newVpnL2VpnCache()
	defer c.Close()

	var testData []*model.VpnL2Vpn
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnL2Vpn))
	}

	err := c.ICache.(VpnL2VpnCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnL2VpnCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VpnL2Vpn))
	}
}

func Test_vpnL2VpnCache_MultiSet(t *testing.T) {
	c := newVpnL2VpnCache()
	defer c.Close()

	var testData []*model.VpnL2Vpn
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnL2Vpn))
	}

	err := c.ICache.(VpnL2VpnCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnL2VpnCache_Del(t *testing.T) {
	c := newVpnL2VpnCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnL2Vpn)
	err := c.ICache.(VpnL2VpnCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnL2VpnCache_SetCacheWithNotFound(t *testing.T) {
	c := newVpnL2VpnCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnL2Vpn)
	err := c.ICache.(VpnL2VpnCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VpnL2VpnCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVpnL2VpnCache(t *testing.T) {
	c := NewVpnL2VpnCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVpnL2VpnCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVpnL2VpnCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
