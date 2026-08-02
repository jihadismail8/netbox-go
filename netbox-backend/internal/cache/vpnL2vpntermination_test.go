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

func newVpnL2VpnterminationCache() *gotest.Cache {
	record1 := &model.VpnL2Vpntermination{}
	record1.ID = 1
	record2 := &model.VpnL2Vpntermination{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVpnL2VpnterminationCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_vpnL2VpnterminationCache_Set(t *testing.T) {
	c := newVpnL2VpnterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnL2Vpntermination)
	err := c.ICache.(VpnL2VpnterminationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VpnL2VpnterminationCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_vpnL2VpnterminationCache_Get(t *testing.T) {
	c := newVpnL2VpnterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnL2Vpntermination)
	err := c.ICache.(VpnL2VpnterminationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnL2VpnterminationCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VpnL2VpnterminationCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_vpnL2VpnterminationCache_MultiGet(t *testing.T) {
	c := newVpnL2VpnterminationCache()
	defer c.Close()

	var testData []*model.VpnL2Vpntermination
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnL2Vpntermination))
	}

	err := c.ICache.(VpnL2VpnterminationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnL2VpnterminationCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VpnL2Vpntermination))
	}
}

func Test_vpnL2VpnterminationCache_MultiSet(t *testing.T) {
	c := newVpnL2VpnterminationCache()
	defer c.Close()

	var testData []*model.VpnL2Vpntermination
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnL2Vpntermination))
	}

	err := c.ICache.(VpnL2VpnterminationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnL2VpnterminationCache_Del(t *testing.T) {
	c := newVpnL2VpnterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnL2Vpntermination)
	err := c.ICache.(VpnL2VpnterminationCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnL2VpnterminationCache_SetCacheWithNotFound(t *testing.T) {
	c := newVpnL2VpnterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnL2Vpntermination)
	err := c.ICache.(VpnL2VpnterminationCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VpnL2VpnterminationCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVpnL2VpnterminationCache(t *testing.T) {
	c := NewVpnL2VpnterminationCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVpnL2VpnterminationCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVpnL2VpnterminationCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
