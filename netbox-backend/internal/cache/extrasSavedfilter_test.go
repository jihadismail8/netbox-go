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

func newExtrasSavedfilterCache() *gotest.Cache {
	record1 := &model.ExtrasSavedfilter{}
	record1.ID = 1
	record2 := &model.ExtrasSavedfilter{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasSavedfilterCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasSavedfilterCache_Set(t *testing.T) {
	c := newExtrasSavedfilterCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasSavedfilter)
	err := c.ICache.(ExtrasSavedfilterCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasSavedfilterCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasSavedfilterCache_Get(t *testing.T) {
	c := newExtrasSavedfilterCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasSavedfilter)
	err := c.ICache.(ExtrasSavedfilterCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasSavedfilterCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasSavedfilterCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasSavedfilterCache_MultiGet(t *testing.T) {
	c := newExtrasSavedfilterCache()
	defer c.Close()

	var testData []*model.ExtrasSavedfilter
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasSavedfilter))
	}

	err := c.ICache.(ExtrasSavedfilterCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasSavedfilterCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasSavedfilter))
	}
}

func Test_extrasSavedfilterCache_MultiSet(t *testing.T) {
	c := newExtrasSavedfilterCache()
	defer c.Close()

	var testData []*model.ExtrasSavedfilter
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasSavedfilter))
	}

	err := c.ICache.(ExtrasSavedfilterCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasSavedfilterCache_Del(t *testing.T) {
	c := newExtrasSavedfilterCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasSavedfilter)
	err := c.ICache.(ExtrasSavedfilterCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasSavedfilterCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasSavedfilterCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasSavedfilter)
	err := c.ICache.(ExtrasSavedfilterCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasSavedfilterCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasSavedfilterCache(t *testing.T) {
	c := NewExtrasSavedfilterCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasSavedfilterCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasSavedfilterCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
