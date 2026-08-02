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

func newVpnL2VpnExportTargetsCache() *gotest.Cache {
	record1 := &model.VpnL2VpnExportTargets{}
	record1.ID = 1
	record2 := &model.VpnL2VpnExportTargets{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVpnL2VpnExportTargetsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_vpnL2VpnExportTargetsCache_Set(t *testing.T) {
	c := newVpnL2VpnExportTargetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnL2VpnExportTargets)
	err := c.ICache.(VpnL2VpnExportTargetsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VpnL2VpnExportTargetsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_vpnL2VpnExportTargetsCache_Get(t *testing.T) {
	c := newVpnL2VpnExportTargetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnL2VpnExportTargets)
	err := c.ICache.(VpnL2VpnExportTargetsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnL2VpnExportTargetsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VpnL2VpnExportTargetsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_vpnL2VpnExportTargetsCache_MultiGet(t *testing.T) {
	c := newVpnL2VpnExportTargetsCache()
	defer c.Close()

	var testData []*model.VpnL2VpnExportTargets
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnL2VpnExportTargets))
	}

	err := c.ICache.(VpnL2VpnExportTargetsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnL2VpnExportTargetsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VpnL2VpnExportTargets))
	}
}

func Test_vpnL2VpnExportTargetsCache_MultiSet(t *testing.T) {
	c := newVpnL2VpnExportTargetsCache()
	defer c.Close()

	var testData []*model.VpnL2VpnExportTargets
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnL2VpnExportTargets))
	}

	err := c.ICache.(VpnL2VpnExportTargetsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnL2VpnExportTargetsCache_Del(t *testing.T) {
	c := newVpnL2VpnExportTargetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnL2VpnExportTargets)
	err := c.ICache.(VpnL2VpnExportTargetsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnL2VpnExportTargetsCache_SetCacheWithNotFound(t *testing.T) {
	c := newVpnL2VpnExportTargetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnL2VpnExportTargets)
	err := c.ICache.(VpnL2VpnExportTargetsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VpnL2VpnExportTargetsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVpnL2VpnExportTargetsCache(t *testing.T) {
	c := NewVpnL2VpnExportTargetsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVpnL2VpnExportTargetsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVpnL2VpnExportTargetsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
