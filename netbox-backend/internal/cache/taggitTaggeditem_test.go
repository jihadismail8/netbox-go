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

func newTaggitTaggeditemCache() *gotest.Cache {
	record1 := &model.TaggitTaggeditem{}
	record1.ID = 1
	record2 := &model.TaggitTaggeditem{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewTaggitTaggeditemCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_taggitTaggeditemCache_Set(t *testing.T) {
	c := newTaggitTaggeditemCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TaggitTaggeditem)
	err := c.ICache.(TaggitTaggeditemCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(TaggitTaggeditemCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_taggitTaggeditemCache_Get(t *testing.T) {
	c := newTaggitTaggeditemCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TaggitTaggeditem)
	err := c.ICache.(TaggitTaggeditemCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(TaggitTaggeditemCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(TaggitTaggeditemCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_taggitTaggeditemCache_MultiGet(t *testing.T) {
	c := newTaggitTaggeditemCache()
	defer c.Close()

	var testData []*model.TaggitTaggeditem
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.TaggitTaggeditem))
	}

	err := c.ICache.(TaggitTaggeditemCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(TaggitTaggeditemCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.TaggitTaggeditem))
	}
}

func Test_taggitTaggeditemCache_MultiSet(t *testing.T) {
	c := newTaggitTaggeditemCache()
	defer c.Close()

	var testData []*model.TaggitTaggeditem
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.TaggitTaggeditem))
	}

	err := c.ICache.(TaggitTaggeditemCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_taggitTaggeditemCache_Del(t *testing.T) {
	c := newTaggitTaggeditemCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TaggitTaggeditem)
	err := c.ICache.(TaggitTaggeditemCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_taggitTaggeditemCache_SetCacheWithNotFound(t *testing.T) {
	c := newTaggitTaggeditemCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TaggitTaggeditem)
	err := c.ICache.(TaggitTaggeditemCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(TaggitTaggeditemCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewTaggitTaggeditemCache(t *testing.T) {
	c := NewTaggitTaggeditemCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewTaggitTaggeditemCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewTaggitTaggeditemCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
