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

func newExtrasTagObjectTypesCache() *gotest.Cache {
	record1 := &model.ExtrasTagObjectTypes{}
	record1.ID = 1
	record2 := &model.ExtrasTagObjectTypes{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasTagObjectTypesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasTagObjectTypesCache_Set(t *testing.T) {
	c := newExtrasTagObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasTagObjectTypes)
	err := c.ICache.(ExtrasTagObjectTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasTagObjectTypesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasTagObjectTypesCache_Get(t *testing.T) {
	c := newExtrasTagObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasTagObjectTypes)
	err := c.ICache.(ExtrasTagObjectTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasTagObjectTypesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasTagObjectTypesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasTagObjectTypesCache_MultiGet(t *testing.T) {
	c := newExtrasTagObjectTypesCache()
	defer c.Close()

	var testData []*model.ExtrasTagObjectTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasTagObjectTypes))
	}

	err := c.ICache.(ExtrasTagObjectTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasTagObjectTypesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasTagObjectTypes))
	}
}

func Test_extrasTagObjectTypesCache_MultiSet(t *testing.T) {
	c := newExtrasTagObjectTypesCache()
	defer c.Close()

	var testData []*model.ExtrasTagObjectTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasTagObjectTypes))
	}

	err := c.ICache.(ExtrasTagObjectTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasTagObjectTypesCache_Del(t *testing.T) {
	c := newExtrasTagObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasTagObjectTypes)
	err := c.ICache.(ExtrasTagObjectTypesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasTagObjectTypesCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasTagObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasTagObjectTypes)
	err := c.ICache.(ExtrasTagObjectTypesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasTagObjectTypesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasTagObjectTypesCache(t *testing.T) {
	c := NewExtrasTagObjectTypesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasTagObjectTypesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasTagObjectTypesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
