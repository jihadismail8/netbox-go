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

func newExtrasEventruleObjectTypesCache() *gotest.Cache {
	record1 := &model.ExtrasEventruleObjectTypes{}
	record1.ID = 1
	record2 := &model.ExtrasEventruleObjectTypes{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasEventruleObjectTypesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasEventruleObjectTypesCache_Set(t *testing.T) {
	c := newExtrasEventruleObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasEventruleObjectTypes)
	err := c.ICache.(ExtrasEventruleObjectTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasEventruleObjectTypesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasEventruleObjectTypesCache_Get(t *testing.T) {
	c := newExtrasEventruleObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasEventruleObjectTypes)
	err := c.ICache.(ExtrasEventruleObjectTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasEventruleObjectTypesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasEventruleObjectTypesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasEventruleObjectTypesCache_MultiGet(t *testing.T) {
	c := newExtrasEventruleObjectTypesCache()
	defer c.Close()

	var testData []*model.ExtrasEventruleObjectTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasEventruleObjectTypes))
	}

	err := c.ICache.(ExtrasEventruleObjectTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasEventruleObjectTypesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasEventruleObjectTypes))
	}
}

func Test_extrasEventruleObjectTypesCache_MultiSet(t *testing.T) {
	c := newExtrasEventruleObjectTypesCache()
	defer c.Close()

	var testData []*model.ExtrasEventruleObjectTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasEventruleObjectTypes))
	}

	err := c.ICache.(ExtrasEventruleObjectTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasEventruleObjectTypesCache_Del(t *testing.T) {
	c := newExtrasEventruleObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasEventruleObjectTypes)
	err := c.ICache.(ExtrasEventruleObjectTypesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasEventruleObjectTypesCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasEventruleObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasEventruleObjectTypes)
	err := c.ICache.(ExtrasEventruleObjectTypesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasEventruleObjectTypesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasEventruleObjectTypesCache(t *testing.T) {
	c := NewExtrasEventruleObjectTypesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasEventruleObjectTypesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasEventruleObjectTypesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
