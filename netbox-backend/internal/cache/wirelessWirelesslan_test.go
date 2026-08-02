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

func newWirelessWirelesslanCache() *gotest.Cache {
	record1 := &model.WirelessWirelesslan{}
	record1.ID = 1
	record2 := &model.WirelessWirelesslan{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewWirelessWirelesslanCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_wirelessWirelesslanCache_Set(t *testing.T) {
	c := newWirelessWirelesslanCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.WirelessWirelesslan)
	err := c.ICache.(WirelessWirelesslanCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(WirelessWirelesslanCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_wirelessWirelesslanCache_Get(t *testing.T) {
	c := newWirelessWirelesslanCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.WirelessWirelesslan)
	err := c.ICache.(WirelessWirelesslanCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(WirelessWirelesslanCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(WirelessWirelesslanCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_wirelessWirelesslanCache_MultiGet(t *testing.T) {
	c := newWirelessWirelesslanCache()
	defer c.Close()

	var testData []*model.WirelessWirelesslan
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.WirelessWirelesslan))
	}

	err := c.ICache.(WirelessWirelesslanCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(WirelessWirelesslanCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.WirelessWirelesslan))
	}
}

func Test_wirelessWirelesslanCache_MultiSet(t *testing.T) {
	c := newWirelessWirelesslanCache()
	defer c.Close()

	var testData []*model.WirelessWirelesslan
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.WirelessWirelesslan))
	}

	err := c.ICache.(WirelessWirelesslanCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_wirelessWirelesslanCache_Del(t *testing.T) {
	c := newWirelessWirelesslanCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.WirelessWirelesslan)
	err := c.ICache.(WirelessWirelesslanCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_wirelessWirelesslanCache_SetCacheWithNotFound(t *testing.T) {
	c := newWirelessWirelesslanCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.WirelessWirelesslan)
	err := c.ICache.(WirelessWirelesslanCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(WirelessWirelesslanCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewWirelessWirelesslanCache(t *testing.T) {
	c := NewWirelessWirelesslanCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewWirelessWirelesslanCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewWirelessWirelesslanCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
