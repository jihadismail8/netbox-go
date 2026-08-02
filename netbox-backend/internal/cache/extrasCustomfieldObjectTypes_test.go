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

func newExtrasCustomfieldObjectTypesCache() *gotest.Cache {
	record1 := &model.ExtrasCustomfieldObjectTypes{}
	record1.ID = 1
	record2 := &model.ExtrasCustomfieldObjectTypes{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasCustomfieldObjectTypesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasCustomfieldObjectTypesCache_Set(t *testing.T) {
	c := newExtrasCustomfieldObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomfieldObjectTypes)
	err := c.ICache.(ExtrasCustomfieldObjectTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasCustomfieldObjectTypesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasCustomfieldObjectTypesCache_Get(t *testing.T) {
	c := newExtrasCustomfieldObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomfieldObjectTypes)
	err := c.ICache.(ExtrasCustomfieldObjectTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasCustomfieldObjectTypesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasCustomfieldObjectTypesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasCustomfieldObjectTypesCache_MultiGet(t *testing.T) {
	c := newExtrasCustomfieldObjectTypesCache()
	defer c.Close()

	var testData []*model.ExtrasCustomfieldObjectTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasCustomfieldObjectTypes))
	}

	err := c.ICache.(ExtrasCustomfieldObjectTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasCustomfieldObjectTypesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasCustomfieldObjectTypes))
	}
}

func Test_extrasCustomfieldObjectTypesCache_MultiSet(t *testing.T) {
	c := newExtrasCustomfieldObjectTypesCache()
	defer c.Close()

	var testData []*model.ExtrasCustomfieldObjectTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasCustomfieldObjectTypes))
	}

	err := c.ICache.(ExtrasCustomfieldObjectTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasCustomfieldObjectTypesCache_Del(t *testing.T) {
	c := newExtrasCustomfieldObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomfieldObjectTypes)
	err := c.ICache.(ExtrasCustomfieldObjectTypesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasCustomfieldObjectTypesCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasCustomfieldObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomfieldObjectTypes)
	err := c.ICache.(ExtrasCustomfieldObjectTypesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasCustomfieldObjectTypesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasCustomfieldObjectTypesCache(t *testing.T) {
	c := NewExtrasCustomfieldObjectTypesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasCustomfieldObjectTypesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasCustomfieldObjectTypesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
