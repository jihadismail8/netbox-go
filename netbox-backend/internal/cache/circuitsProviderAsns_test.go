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

func newCircuitsProviderAsnsCache() *gotest.Cache {
	record1 := &model.CircuitsProviderAsns{}
	record1.ID = 1
	record2 := &model.CircuitsProviderAsns{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewCircuitsProviderAsnsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_circuitsProviderAsnsCache_Set(t *testing.T) {
	c := newCircuitsProviderAsnsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProviderAsns)
	err := c.ICache.(CircuitsProviderAsnsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(CircuitsProviderAsnsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_circuitsProviderAsnsCache_Get(t *testing.T) {
	c := newCircuitsProviderAsnsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProviderAsns)
	err := c.ICache.(CircuitsProviderAsnsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsProviderAsnsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(CircuitsProviderAsnsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_circuitsProviderAsnsCache_MultiGet(t *testing.T) {
	c := newCircuitsProviderAsnsCache()
	defer c.Close()

	var testData []*model.CircuitsProviderAsns
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsProviderAsns))
	}

	err := c.ICache.(CircuitsProviderAsnsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsProviderAsnsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.CircuitsProviderAsns))
	}
}

func Test_circuitsProviderAsnsCache_MultiSet(t *testing.T) {
	c := newCircuitsProviderAsnsCache()
	defer c.Close()

	var testData []*model.CircuitsProviderAsns
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsProviderAsns))
	}

	err := c.ICache.(CircuitsProviderAsnsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsProviderAsnsCache_Del(t *testing.T) {
	c := newCircuitsProviderAsnsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProviderAsns)
	err := c.ICache.(CircuitsProviderAsnsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsProviderAsnsCache_SetCacheWithNotFound(t *testing.T) {
	c := newCircuitsProviderAsnsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProviderAsns)
	err := c.ICache.(CircuitsProviderAsnsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(CircuitsProviderAsnsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewCircuitsProviderAsnsCache(t *testing.T) {
	c := NewCircuitsProviderAsnsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewCircuitsProviderAsnsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewCircuitsProviderAsnsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
