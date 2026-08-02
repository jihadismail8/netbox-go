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

func newExtrasCustomfieldchoicesetCache() *gotest.Cache {
	record1 := &model.ExtrasCustomfieldchoiceset{}
	record1.ID = 1
	record2 := &model.ExtrasCustomfieldchoiceset{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasCustomfieldchoicesetCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasCustomfieldchoicesetCache_Set(t *testing.T) {
	c := newExtrasCustomfieldchoicesetCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomfieldchoiceset)
	err := c.ICache.(ExtrasCustomfieldchoicesetCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasCustomfieldchoicesetCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasCustomfieldchoicesetCache_Get(t *testing.T) {
	c := newExtrasCustomfieldchoicesetCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomfieldchoiceset)
	err := c.ICache.(ExtrasCustomfieldchoicesetCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasCustomfieldchoicesetCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasCustomfieldchoicesetCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasCustomfieldchoicesetCache_MultiGet(t *testing.T) {
	c := newExtrasCustomfieldchoicesetCache()
	defer c.Close()

	var testData []*model.ExtrasCustomfieldchoiceset
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasCustomfieldchoiceset))
	}

	err := c.ICache.(ExtrasCustomfieldchoicesetCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasCustomfieldchoicesetCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasCustomfieldchoiceset))
	}
}

func Test_extrasCustomfieldchoicesetCache_MultiSet(t *testing.T) {
	c := newExtrasCustomfieldchoicesetCache()
	defer c.Close()

	var testData []*model.ExtrasCustomfieldchoiceset
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasCustomfieldchoiceset))
	}

	err := c.ICache.(ExtrasCustomfieldchoicesetCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasCustomfieldchoicesetCache_Del(t *testing.T) {
	c := newExtrasCustomfieldchoicesetCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomfieldchoiceset)
	err := c.ICache.(ExtrasCustomfieldchoicesetCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasCustomfieldchoicesetCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasCustomfieldchoicesetCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomfieldchoiceset)
	err := c.ICache.(ExtrasCustomfieldchoicesetCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasCustomfieldchoicesetCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasCustomfieldchoicesetCache(t *testing.T) {
	c := NewExtrasCustomfieldchoicesetCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasCustomfieldchoicesetCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasCustomfieldchoicesetCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
