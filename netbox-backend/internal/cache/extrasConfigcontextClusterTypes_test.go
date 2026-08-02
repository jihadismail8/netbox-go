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

func newExtrasConfigcontextClusterTypesCache() *gotest.Cache {
	record1 := &model.ExtrasConfigcontextClusterTypes{}
	record1.ID = 1
	record2 := &model.ExtrasConfigcontextClusterTypes{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasConfigcontextClusterTypesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasConfigcontextClusterTypesCache_Set(t *testing.T) {
	c := newExtrasConfigcontextClusterTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextClusterTypes)
	err := c.ICache.(ExtrasConfigcontextClusterTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasConfigcontextClusterTypesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasConfigcontextClusterTypesCache_Get(t *testing.T) {
	c := newExtrasConfigcontextClusterTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextClusterTypes)
	err := c.ICache.(ExtrasConfigcontextClusterTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextClusterTypesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasConfigcontextClusterTypesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasConfigcontextClusterTypesCache_MultiGet(t *testing.T) {
	c := newExtrasConfigcontextClusterTypesCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextClusterTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextClusterTypes))
	}

	err := c.ICache.(ExtrasConfigcontextClusterTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextClusterTypesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasConfigcontextClusterTypes))
	}
}

func Test_extrasConfigcontextClusterTypesCache_MultiSet(t *testing.T) {
	c := newExtrasConfigcontextClusterTypesCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextClusterTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextClusterTypes))
	}

	err := c.ICache.(ExtrasConfigcontextClusterTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextClusterTypesCache_Del(t *testing.T) {
	c := newExtrasConfigcontextClusterTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextClusterTypes)
	err := c.ICache.(ExtrasConfigcontextClusterTypesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextClusterTypesCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasConfigcontextClusterTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextClusterTypes)
	err := c.ICache.(ExtrasConfigcontextClusterTypesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasConfigcontextClusterTypesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasConfigcontextClusterTypesCache(t *testing.T) {
	c := NewExtrasConfigcontextClusterTypesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasConfigcontextClusterTypesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasConfigcontextClusterTypesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
