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

func newExtrasConfigcontextCache() *gotest.Cache {
	record1 := &model.ExtrasConfigcontext{}
	record1.ID = 1
	record2 := &model.ExtrasConfigcontext{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasConfigcontextCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasConfigcontextCache_Set(t *testing.T) {
	c := newExtrasConfigcontextCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontext)
	err := c.ICache.(ExtrasConfigcontextCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasConfigcontextCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasConfigcontextCache_Get(t *testing.T) {
	c := newExtrasConfigcontextCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontext)
	err := c.ICache.(ExtrasConfigcontextCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasConfigcontextCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasConfigcontextCache_MultiGet(t *testing.T) {
	c := newExtrasConfigcontextCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontext
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontext))
	}

	err := c.ICache.(ExtrasConfigcontextCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasConfigcontext))
	}
}

func Test_extrasConfigcontextCache_MultiSet(t *testing.T) {
	c := newExtrasConfigcontextCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontext
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontext))
	}

	err := c.ICache.(ExtrasConfigcontextCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextCache_Del(t *testing.T) {
	c := newExtrasConfigcontextCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontext)
	err := c.ICache.(ExtrasConfigcontextCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasConfigcontextCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontext)
	err := c.ICache.(ExtrasConfigcontextCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasConfigcontextCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasConfigcontextCache(t *testing.T) {
	c := NewExtrasConfigcontextCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasConfigcontextCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasConfigcontextCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
