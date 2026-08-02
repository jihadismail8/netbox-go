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

func newVpnTunnelterminationCache() *gotest.Cache {
	record1 := &model.VpnTunneltermination{}
	record1.ID = 1
	record2 := &model.VpnTunneltermination{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVpnTunnelterminationCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_vpnTunnelterminationCache_Set(t *testing.T) {
	c := newVpnTunnelterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnTunneltermination)
	err := c.ICache.(VpnTunnelterminationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VpnTunnelterminationCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_vpnTunnelterminationCache_Get(t *testing.T) {
	c := newVpnTunnelterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnTunneltermination)
	err := c.ICache.(VpnTunnelterminationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnTunnelterminationCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VpnTunnelterminationCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_vpnTunnelterminationCache_MultiGet(t *testing.T) {
	c := newVpnTunnelterminationCache()
	defer c.Close()

	var testData []*model.VpnTunneltermination
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnTunneltermination))
	}

	err := c.ICache.(VpnTunnelterminationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnTunnelterminationCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VpnTunneltermination))
	}
}

func Test_vpnTunnelterminationCache_MultiSet(t *testing.T) {
	c := newVpnTunnelterminationCache()
	defer c.Close()

	var testData []*model.VpnTunneltermination
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnTunneltermination))
	}

	err := c.ICache.(VpnTunnelterminationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnTunnelterminationCache_Del(t *testing.T) {
	c := newVpnTunnelterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnTunneltermination)
	err := c.ICache.(VpnTunnelterminationCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnTunnelterminationCache_SetCacheWithNotFound(t *testing.T) {
	c := newVpnTunnelterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnTunneltermination)
	err := c.ICache.(VpnTunnelterminationCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VpnTunnelterminationCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVpnTunnelterminationCache(t *testing.T) {
	c := NewVpnTunnelterminationCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVpnTunnelterminationCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVpnTunnelterminationCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
