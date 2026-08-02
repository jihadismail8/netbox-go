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

func newExtrasConfigtemplateCache() *gotest.Cache {
	record1 := &model.ExtrasConfigtemplate{}
	record1.ID = 1
	record2 := &model.ExtrasConfigtemplate{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasConfigtemplateCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasConfigtemplateCache_Set(t *testing.T) {
	c := newExtrasConfigtemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigtemplate)
	err := c.ICache.(ExtrasConfigtemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasConfigtemplateCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasConfigtemplateCache_Get(t *testing.T) {
	c := newExtrasConfigtemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigtemplate)
	err := c.ICache.(ExtrasConfigtemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigtemplateCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasConfigtemplateCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasConfigtemplateCache_MultiGet(t *testing.T) {
	c := newExtrasConfigtemplateCache()
	defer c.Close()

	var testData []*model.ExtrasConfigtemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigtemplate))
	}

	err := c.ICache.(ExtrasConfigtemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigtemplateCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasConfigtemplate))
	}
}

func Test_extrasConfigtemplateCache_MultiSet(t *testing.T) {
	c := newExtrasConfigtemplateCache()
	defer c.Close()

	var testData []*model.ExtrasConfigtemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigtemplate))
	}

	err := c.ICache.(ExtrasConfigtemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigtemplateCache_Del(t *testing.T) {
	c := newExtrasConfigtemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigtemplate)
	err := c.ICache.(ExtrasConfigtemplateCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigtemplateCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasConfigtemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigtemplate)
	err := c.ICache.(ExtrasConfigtemplateCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasConfigtemplateCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasConfigtemplateCache(t *testing.T) {
	c := NewExtrasConfigtemplateCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasConfigtemplateCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasConfigtemplateCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
