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

func newExtrasCustomfieldCache() *gotest.Cache {
	record1 := &model.ExtrasCustomfield{}
	record1.ID = 1
	record2 := &model.ExtrasCustomfield{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasCustomfieldCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasCustomfieldCache_Set(t *testing.T) {
	c := newExtrasCustomfieldCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomfield)
	err := c.ICache.(ExtrasCustomfieldCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasCustomfieldCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasCustomfieldCache_Get(t *testing.T) {
	c := newExtrasCustomfieldCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomfield)
	err := c.ICache.(ExtrasCustomfieldCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasCustomfieldCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasCustomfieldCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasCustomfieldCache_MultiGet(t *testing.T) {
	c := newExtrasCustomfieldCache()
	defer c.Close()

	var testData []*model.ExtrasCustomfield
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasCustomfield))
	}

	err := c.ICache.(ExtrasCustomfieldCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasCustomfieldCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasCustomfield))
	}
}

func Test_extrasCustomfieldCache_MultiSet(t *testing.T) {
	c := newExtrasCustomfieldCache()
	defer c.Close()

	var testData []*model.ExtrasCustomfield
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasCustomfield))
	}

	err := c.ICache.(ExtrasCustomfieldCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasCustomfieldCache_Del(t *testing.T) {
	c := newExtrasCustomfieldCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomfield)
	err := c.ICache.(ExtrasCustomfieldCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasCustomfieldCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasCustomfieldCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasCustomfield)
	err := c.ICache.(ExtrasCustomfieldCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasCustomfieldCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasCustomfieldCache(t *testing.T) {
	c := NewExtrasCustomfieldCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasCustomfieldCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasCustomfieldCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
