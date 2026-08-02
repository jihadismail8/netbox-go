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

func newExtrasTableconfigCache() *gotest.Cache {
	record1 := &model.ExtrasTableconfig{}
	record1.ID = 1
	record2 := &model.ExtrasTableconfig{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasTableconfigCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasTableconfigCache_Set(t *testing.T) {
	c := newExtrasTableconfigCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasTableconfig)
	err := c.ICache.(ExtrasTableconfigCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasTableconfigCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasTableconfigCache_Get(t *testing.T) {
	c := newExtrasTableconfigCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasTableconfig)
	err := c.ICache.(ExtrasTableconfigCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasTableconfigCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasTableconfigCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasTableconfigCache_MultiGet(t *testing.T) {
	c := newExtrasTableconfigCache()
	defer c.Close()

	var testData []*model.ExtrasTableconfig
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasTableconfig))
	}

	err := c.ICache.(ExtrasTableconfigCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasTableconfigCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasTableconfig))
	}
}

func Test_extrasTableconfigCache_MultiSet(t *testing.T) {
	c := newExtrasTableconfigCache()
	defer c.Close()

	var testData []*model.ExtrasTableconfig
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasTableconfig))
	}

	err := c.ICache.(ExtrasTableconfigCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasTableconfigCache_Del(t *testing.T) {
	c := newExtrasTableconfigCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasTableconfig)
	err := c.ICache.(ExtrasTableconfigCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasTableconfigCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasTableconfigCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasTableconfig)
	err := c.ICache.(ExtrasTableconfigCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasTableconfigCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasTableconfigCache(t *testing.T) {
	c := NewExtrasTableconfigCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasTableconfigCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasTableconfigCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
