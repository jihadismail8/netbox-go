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

func newExtrasConfigcontextPlatformsCache() *gotest.Cache {
	record1 := &model.ExtrasConfigcontextPlatforms{}
	record1.ID = 1
	record2 := &model.ExtrasConfigcontextPlatforms{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasConfigcontextPlatformsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasConfigcontextPlatformsCache_Set(t *testing.T) {
	c := newExtrasConfigcontextPlatformsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextPlatforms)
	err := c.ICache.(ExtrasConfigcontextPlatformsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasConfigcontextPlatformsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasConfigcontextPlatformsCache_Get(t *testing.T) {
	c := newExtrasConfigcontextPlatformsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextPlatforms)
	err := c.ICache.(ExtrasConfigcontextPlatformsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextPlatformsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasConfigcontextPlatformsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasConfigcontextPlatformsCache_MultiGet(t *testing.T) {
	c := newExtrasConfigcontextPlatformsCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextPlatforms
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextPlatforms))
	}

	err := c.ICache.(ExtrasConfigcontextPlatformsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextPlatformsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasConfigcontextPlatforms))
	}
}

func Test_extrasConfigcontextPlatformsCache_MultiSet(t *testing.T) {
	c := newExtrasConfigcontextPlatformsCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextPlatforms
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextPlatforms))
	}

	err := c.ICache.(ExtrasConfigcontextPlatformsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextPlatformsCache_Del(t *testing.T) {
	c := newExtrasConfigcontextPlatformsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextPlatforms)
	err := c.ICache.(ExtrasConfigcontextPlatformsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextPlatformsCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasConfigcontextPlatformsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextPlatforms)
	err := c.ICache.(ExtrasConfigcontextPlatformsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasConfigcontextPlatformsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasConfigcontextPlatformsCache(t *testing.T) {
	c := NewExtrasConfigcontextPlatformsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasConfigcontextPlatformsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasConfigcontextPlatformsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
