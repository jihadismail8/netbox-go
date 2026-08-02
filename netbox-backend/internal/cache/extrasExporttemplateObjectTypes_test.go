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

func newExtrasExporttemplateObjectTypesCache() *gotest.Cache {
	record1 := &model.ExtrasExporttemplateObjectTypes{}
	record1.ID = 1
	record2 := &model.ExtrasExporttemplateObjectTypes{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasExporttemplateObjectTypesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasExporttemplateObjectTypesCache_Set(t *testing.T) {
	c := newExtrasExporttemplateObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasExporttemplateObjectTypes)
	err := c.ICache.(ExtrasExporttemplateObjectTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasExporttemplateObjectTypesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasExporttemplateObjectTypesCache_Get(t *testing.T) {
	c := newExtrasExporttemplateObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasExporttemplateObjectTypes)
	err := c.ICache.(ExtrasExporttemplateObjectTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasExporttemplateObjectTypesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasExporttemplateObjectTypesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasExporttemplateObjectTypesCache_MultiGet(t *testing.T) {
	c := newExtrasExporttemplateObjectTypesCache()
	defer c.Close()

	var testData []*model.ExtrasExporttemplateObjectTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasExporttemplateObjectTypes))
	}

	err := c.ICache.(ExtrasExporttemplateObjectTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasExporttemplateObjectTypesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasExporttemplateObjectTypes))
	}
}

func Test_extrasExporttemplateObjectTypesCache_MultiSet(t *testing.T) {
	c := newExtrasExporttemplateObjectTypesCache()
	defer c.Close()

	var testData []*model.ExtrasExporttemplateObjectTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasExporttemplateObjectTypes))
	}

	err := c.ICache.(ExtrasExporttemplateObjectTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasExporttemplateObjectTypesCache_Del(t *testing.T) {
	c := newExtrasExporttemplateObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasExporttemplateObjectTypes)
	err := c.ICache.(ExtrasExporttemplateObjectTypesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasExporttemplateObjectTypesCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasExporttemplateObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasExporttemplateObjectTypes)
	err := c.ICache.(ExtrasExporttemplateObjectTypesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasExporttemplateObjectTypesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasExporttemplateObjectTypesCache(t *testing.T) {
	c := NewExtrasExporttemplateObjectTypesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasExporttemplateObjectTypesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasExporttemplateObjectTypesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
