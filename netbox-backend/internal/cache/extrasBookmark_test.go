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

func newExtrasBookmarkCache() *gotest.Cache {
	record1 := &model.ExtrasBookmark{}
	record1.ID = 1
	record2 := &model.ExtrasBookmark{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasBookmarkCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasBookmarkCache_Set(t *testing.T) {
	c := newExtrasBookmarkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasBookmark)
	err := c.ICache.(ExtrasBookmarkCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasBookmarkCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasBookmarkCache_Get(t *testing.T) {
	c := newExtrasBookmarkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasBookmark)
	err := c.ICache.(ExtrasBookmarkCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasBookmarkCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasBookmarkCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasBookmarkCache_MultiGet(t *testing.T) {
	c := newExtrasBookmarkCache()
	defer c.Close()

	var testData []*model.ExtrasBookmark
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasBookmark))
	}

	err := c.ICache.(ExtrasBookmarkCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasBookmarkCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasBookmark))
	}
}

func Test_extrasBookmarkCache_MultiSet(t *testing.T) {
	c := newExtrasBookmarkCache()
	defer c.Close()

	var testData []*model.ExtrasBookmark
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasBookmark))
	}

	err := c.ICache.(ExtrasBookmarkCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasBookmarkCache_Del(t *testing.T) {
	c := newExtrasBookmarkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasBookmark)
	err := c.ICache.(ExtrasBookmarkCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasBookmarkCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasBookmarkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasBookmark)
	err := c.ICache.(ExtrasBookmarkCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasBookmarkCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasBookmarkCache(t *testing.T) {
	c := NewExtrasBookmarkCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasBookmarkCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasBookmarkCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
