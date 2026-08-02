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

func newExtrasCustomlinkCache() *gotest.Cache {
	record1 := &model.ExtrasCustomlink{}
	record1.ID = 1
	record2 := &model.ExtrasCustomlink{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasCustomlinkCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasCustomlinkCache_Set(t *testing.T) {
	c := newExtrasCustomlinkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomlink)
	err := c.ICache.(ExtrasCustomlinkCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasCustomlinkCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasCustomlinkCache_Get(t *testing.T) {
	c := newExtrasCustomlinkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomlink)
	err := c.ICache.(ExtrasCustomlinkCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasCustomlinkCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasCustomlinkCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasCustomlinkCache_MultiGet(t *testing.T) {
	c := newExtrasCustomlinkCache()
	defer c.Close()

	var testData []*model.ExtrasCustomlink
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasCustomlink))
	}

	err := c.ICache.(ExtrasCustomlinkCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasCustomlinkCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasCustomlink))
	}
}

func Test_extrasCustomlinkCache_MultiSet(t *testing.T) {
	c := newExtrasCustomlinkCache()
	defer c.Close()

	var testData []*model.ExtrasCustomlink
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasCustomlink))
	}

	err := c.ICache.(ExtrasCustomlinkCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasCustomlinkCache_Del(t *testing.T) {
	c := newExtrasCustomlinkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomlink)
	err := c.ICache.(ExtrasCustomlinkCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasCustomlinkCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasCustomlinkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomlink)
	err := c.ICache.(ExtrasCustomlinkCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasCustomlinkCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasCustomlinkCache(t *testing.T) {
	c := NewExtrasCustomlinkCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasCustomlinkCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasCustomlinkCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
