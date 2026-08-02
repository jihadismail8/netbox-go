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

func newExtrasScriptCache() *gotest.Cache {
	record1 := &model.ExtrasScript{}
	record1.ID = 1
	record2 := &model.ExtrasScript{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasScriptCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasScriptCache_Set(t *testing.T) {
	c := newExtrasScriptCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasScript)
	err := c.ICache.(ExtrasScriptCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasScriptCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasScriptCache_Get(t *testing.T) {
	c := newExtrasScriptCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasScript)
	err := c.ICache.(ExtrasScriptCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasScriptCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasScriptCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasScriptCache_MultiGet(t *testing.T) {
	c := newExtrasScriptCache()
	defer c.Close()

	var testData []*model.ExtrasScript
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasScript))
	}

	err := c.ICache.(ExtrasScriptCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasScriptCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasScript))
	}
}

func Test_extrasScriptCache_MultiSet(t *testing.T) {
	c := newExtrasScriptCache()
	defer c.Close()

	var testData []*model.ExtrasScript
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasScript))
	}

	err := c.ICache.(ExtrasScriptCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasScriptCache_Del(t *testing.T) {
	c := newExtrasScriptCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasScript)
	err := c.ICache.(ExtrasScriptCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasScriptCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasScriptCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasScript)
	err := c.ICache.(ExtrasScriptCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasScriptCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasScriptCache(t *testing.T) {
	c := NewExtrasScriptCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasScriptCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasScriptCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
