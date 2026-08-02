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

func newExtrasSavedfilterObjectTypesCache() *gotest.Cache {
	record1 := &model.ExtrasSavedfilterObjectTypes{}
	record1.ID = 1
	record2 := &model.ExtrasSavedfilterObjectTypes{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasSavedfilterObjectTypesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasSavedfilterObjectTypesCache_Set(t *testing.T) {
	c := newExtrasSavedfilterObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasSavedfilterObjectTypes)
	err := c.ICache.(ExtrasSavedfilterObjectTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasSavedfilterObjectTypesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasSavedfilterObjectTypesCache_Get(t *testing.T) {
	c := newExtrasSavedfilterObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasSavedfilterObjectTypes)
	err := c.ICache.(ExtrasSavedfilterObjectTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasSavedfilterObjectTypesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasSavedfilterObjectTypesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasSavedfilterObjectTypesCache_MultiGet(t *testing.T) {
	c := newExtrasSavedfilterObjectTypesCache()
	defer c.Close()

	var testData []*model.ExtrasSavedfilterObjectTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasSavedfilterObjectTypes))
	}

	err := c.ICache.(ExtrasSavedfilterObjectTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasSavedfilterObjectTypesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasSavedfilterObjectTypes))
	}
}

func Test_extrasSavedfilterObjectTypesCache_MultiSet(t *testing.T) {
	c := newExtrasSavedfilterObjectTypesCache()
	defer c.Close()

	var testData []*model.ExtrasSavedfilterObjectTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasSavedfilterObjectTypes))
	}

	err := c.ICache.(ExtrasSavedfilterObjectTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasSavedfilterObjectTypesCache_Del(t *testing.T) {
	c := newExtrasSavedfilterObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasSavedfilterObjectTypes)
	err := c.ICache.(ExtrasSavedfilterObjectTypesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasSavedfilterObjectTypesCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasSavedfilterObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasSavedfilterObjectTypes)
	err := c.ICache.(ExtrasSavedfilterObjectTypesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasSavedfilterObjectTypesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasSavedfilterObjectTypesCache(t *testing.T) {
	c := NewExtrasSavedfilterObjectTypesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasSavedfilterObjectTypesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasSavedfilterObjectTypesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
