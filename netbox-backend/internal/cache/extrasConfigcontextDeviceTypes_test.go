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

func newExtrasConfigcontextDeviceTypesCache() *gotest.Cache {
	record1 := &model.ExtrasConfigcontextDeviceTypes{}
	record1.ID = 1
	record2 := &model.ExtrasConfigcontextDeviceTypes{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasConfigcontextDeviceTypesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasConfigcontextDeviceTypesCache_Set(t *testing.T) {
	c := newExtrasConfigcontextDeviceTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextDeviceTypes)
	err := c.ICache.(ExtrasConfigcontextDeviceTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasConfigcontextDeviceTypesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasConfigcontextDeviceTypesCache_Get(t *testing.T) {
	c := newExtrasConfigcontextDeviceTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextDeviceTypes)
	err := c.ICache.(ExtrasConfigcontextDeviceTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextDeviceTypesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasConfigcontextDeviceTypesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasConfigcontextDeviceTypesCache_MultiGet(t *testing.T) {
	c := newExtrasConfigcontextDeviceTypesCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextDeviceTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextDeviceTypes))
	}

	err := c.ICache.(ExtrasConfigcontextDeviceTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextDeviceTypesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasConfigcontextDeviceTypes))
	}
}

func Test_extrasConfigcontextDeviceTypesCache_MultiSet(t *testing.T) {
	c := newExtrasConfigcontextDeviceTypesCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextDeviceTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextDeviceTypes))
	}

	err := c.ICache.(ExtrasConfigcontextDeviceTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextDeviceTypesCache_Del(t *testing.T) {
	c := newExtrasConfigcontextDeviceTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextDeviceTypes)
	err := c.ICache.(ExtrasConfigcontextDeviceTypesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextDeviceTypesCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasConfigcontextDeviceTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextDeviceTypes)
	err := c.ICache.(ExtrasConfigcontextDeviceTypesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasConfigcontextDeviceTypesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasConfigcontextDeviceTypesCache(t *testing.T) {
	c := NewExtrasConfigcontextDeviceTypesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasConfigcontextDeviceTypesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasConfigcontextDeviceTypesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
