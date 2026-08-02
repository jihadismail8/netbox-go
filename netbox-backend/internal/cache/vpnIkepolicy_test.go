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

func newVpnIkepolicyCache() *gotest.Cache {
	record1 := &model.VpnIkepolicy{}
	record1.ID = 1
	record2 := &model.VpnIkepolicy{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVpnIkepolicyCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_vpnIkepolicyCache_Set(t *testing.T) {
	c := newVpnIkepolicyCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIkepolicy)
	err := c.ICache.(VpnIkepolicyCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VpnIkepolicyCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_vpnIkepolicyCache_Get(t *testing.T) {
	c := newVpnIkepolicyCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIkepolicy)
	err := c.ICache.(VpnIkepolicyCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnIkepolicyCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VpnIkepolicyCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_vpnIkepolicyCache_MultiGet(t *testing.T) {
	c := newVpnIkepolicyCache()
	defer c.Close()

	var testData []*model.VpnIkepolicy
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnIkepolicy))
	}

	err := c.ICache.(VpnIkepolicyCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnIkepolicyCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VpnIkepolicy))
	}
}

func Test_vpnIkepolicyCache_MultiSet(t *testing.T) {
	c := newVpnIkepolicyCache()
	defer c.Close()

	var testData []*model.VpnIkepolicy
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnIkepolicy))
	}

	err := c.ICache.(VpnIkepolicyCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnIkepolicyCache_Del(t *testing.T) {
	c := newVpnIkepolicyCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIkepolicy)
	err := c.ICache.(VpnIkepolicyCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnIkepolicyCache_SetCacheWithNotFound(t *testing.T) {
	c := newVpnIkepolicyCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIkepolicy)
	err := c.ICache.(VpnIkepolicyCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VpnIkepolicyCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVpnIkepolicyCache(t *testing.T) {
	c := NewVpnIkepolicyCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVpnIkepolicyCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVpnIkepolicyCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
