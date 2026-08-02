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

func newVpnTunnelgroupCache() *gotest.Cache {
	record1 := &model.VpnTunnelgroup{}
	record1.ID = 1
	record2 := &model.VpnTunnelgroup{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVpnTunnelgroupCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_vpnTunnelgroupCache_Set(t *testing.T) {
	c := newVpnTunnelgroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnTunnelgroup)
	err := c.ICache.(VpnTunnelgroupCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VpnTunnelgroupCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_vpnTunnelgroupCache_Get(t *testing.T) {
	c := newVpnTunnelgroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnTunnelgroup)
	err := c.ICache.(VpnTunnelgroupCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnTunnelgroupCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VpnTunnelgroupCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_vpnTunnelgroupCache_MultiGet(t *testing.T) {
	c := newVpnTunnelgroupCache()
	defer c.Close()

	var testData []*model.VpnTunnelgroup
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnTunnelgroup))
	}

	err := c.ICache.(VpnTunnelgroupCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnTunnelgroupCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VpnTunnelgroup))
	}
}

func Test_vpnTunnelgroupCache_MultiSet(t *testing.T) {
	c := newVpnTunnelgroupCache()
	defer c.Close()

	var testData []*model.VpnTunnelgroup
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnTunnelgroup))
	}

	err := c.ICache.(VpnTunnelgroupCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnTunnelgroupCache_Del(t *testing.T) {
	c := newVpnTunnelgroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnTunnelgroup)
	err := c.ICache.(VpnTunnelgroupCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnTunnelgroupCache_SetCacheWithNotFound(t *testing.T) {
	c := newVpnTunnelgroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnTunnelgroup)
	err := c.ICache.(VpnTunnelgroupCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VpnTunnelgroupCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVpnTunnelgroupCache(t *testing.T) {
	c := NewVpnTunnelgroupCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVpnTunnelgroupCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVpnTunnelgroupCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
