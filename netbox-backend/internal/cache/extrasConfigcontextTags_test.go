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

func newExtrasConfigcontextTagsCache() *gotest.Cache {
	record1 := &model.ExtrasConfigcontextTags{}
	record1.ID = 1
	record2 := &model.ExtrasConfigcontextTags{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasConfigcontextTagsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasConfigcontextTagsCache_Set(t *testing.T) {
	c := newExtrasConfigcontextTagsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextTags)
	err := c.ICache.(ExtrasConfigcontextTagsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasConfigcontextTagsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasConfigcontextTagsCache_Get(t *testing.T) {
	c := newExtrasConfigcontextTagsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextTags)
	err := c.ICache.(ExtrasConfigcontextTagsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextTagsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasConfigcontextTagsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasConfigcontextTagsCache_MultiGet(t *testing.T) {
	c := newExtrasConfigcontextTagsCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextTags
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextTags))
	}

	err := c.ICache.(ExtrasConfigcontextTagsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextTagsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasConfigcontextTags))
	}
}

func Test_extrasConfigcontextTagsCache_MultiSet(t *testing.T) {
	c := newExtrasConfigcontextTagsCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextTags
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextTags))
	}

	err := c.ICache.(ExtrasConfigcontextTagsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextTagsCache_Del(t *testing.T) {
	c := newExtrasConfigcontextTagsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextTags)
	err := c.ICache.(ExtrasConfigcontextTagsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextTagsCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasConfigcontextTagsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextTags)
	err := c.ICache.(ExtrasConfigcontextTagsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasConfigcontextTagsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasConfigcontextTagsCache(t *testing.T) {
	c := NewExtrasConfigcontextTagsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasConfigcontextTagsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasConfigcontextTagsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
