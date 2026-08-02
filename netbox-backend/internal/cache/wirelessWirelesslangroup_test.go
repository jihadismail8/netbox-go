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

func newWirelessWirelesslangroupCache() *gotest.Cache {
	record1 := &model.WirelessWirelesslangroup{}
	record1.ID = 1
	record2 := &model.WirelessWirelesslangroup{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewWirelessWirelesslangroupCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_wirelessWirelesslangroupCache_Set(t *testing.T) {
	c := newWirelessWirelesslangroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.WirelessWirelesslangroup)
	err := c.ICache.(WirelessWirelesslangroupCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(WirelessWirelesslangroupCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_wirelessWirelesslangroupCache_Get(t *testing.T) {
	c := newWirelessWirelesslangroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.WirelessWirelesslangroup)
	err := c.ICache.(WirelessWirelesslangroupCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(WirelessWirelesslangroupCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(WirelessWirelesslangroupCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_wirelessWirelesslangroupCache_MultiGet(t *testing.T) {
	c := newWirelessWirelesslangroupCache()
	defer c.Close()

	var testData []*model.WirelessWirelesslangroup
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.WirelessWirelesslangroup))
	}

	err := c.ICache.(WirelessWirelesslangroupCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(WirelessWirelesslangroupCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.WirelessWirelesslangroup))
	}
}

func Test_wirelessWirelesslangroupCache_MultiSet(t *testing.T) {
	c := newWirelessWirelesslangroupCache()
	defer c.Close()

	var testData []*model.WirelessWirelesslangroup
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.WirelessWirelesslangroup))
	}

	err := c.ICache.(WirelessWirelesslangroupCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_wirelessWirelesslangroupCache_Del(t *testing.T) {
	c := newWirelessWirelesslangroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.WirelessWirelesslangroup)
	err := c.ICache.(WirelessWirelesslangroupCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_wirelessWirelesslangroupCache_SetCacheWithNotFound(t *testing.T) {
	c := newWirelessWirelesslangroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.WirelessWirelesslangroup)
	err := c.ICache.(WirelessWirelesslangroupCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(WirelessWirelesslangroupCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewWirelessWirelesslangroupCache(t *testing.T) {
	c := NewWirelessWirelesslangroupCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewWirelessWirelesslangroupCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewWirelessWirelesslangroupCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
