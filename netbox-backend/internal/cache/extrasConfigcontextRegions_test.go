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

func newExtrasConfigcontextRegionsCache() *gotest.Cache {
	record1 := &model.ExtrasConfigcontextRegions{}
	record1.ID = 1
	record2 := &model.ExtrasConfigcontextRegions{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasConfigcontextRegionsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasConfigcontextRegionsCache_Set(t *testing.T) {
	c := newExtrasConfigcontextRegionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextRegions)
	err := c.ICache.(ExtrasConfigcontextRegionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasConfigcontextRegionsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasConfigcontextRegionsCache_Get(t *testing.T) {
	c := newExtrasConfigcontextRegionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextRegions)
	err := c.ICache.(ExtrasConfigcontextRegionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextRegionsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasConfigcontextRegionsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasConfigcontextRegionsCache_MultiGet(t *testing.T) {
	c := newExtrasConfigcontextRegionsCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextRegions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextRegions))
	}

	err := c.ICache.(ExtrasConfigcontextRegionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextRegionsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasConfigcontextRegions))
	}
}

func Test_extrasConfigcontextRegionsCache_MultiSet(t *testing.T) {
	c := newExtrasConfigcontextRegionsCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextRegions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextRegions))
	}

	err := c.ICache.(ExtrasConfigcontextRegionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextRegionsCache_Del(t *testing.T) {
	c := newExtrasConfigcontextRegionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextRegions)
	err := c.ICache.(ExtrasConfigcontextRegionsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextRegionsCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasConfigcontextRegionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextRegions)
	err := c.ICache.(ExtrasConfigcontextRegionsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasConfigcontextRegionsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasConfigcontextRegionsCache(t *testing.T) {
	c := NewExtrasConfigcontextRegionsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasConfigcontextRegionsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasConfigcontextRegionsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
