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

func newCircuitsProviderCache() *gotest.Cache {
	record1 := &model.CircuitsProvider{}
	record1.ID = 1
	record2 := &model.CircuitsProvider{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewCircuitsProviderCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_circuitsProviderCache_Set(t *testing.T) {
	c := newCircuitsProviderCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProvider)
	err := c.ICache.(CircuitsProviderCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(CircuitsProviderCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_circuitsProviderCache_Get(t *testing.T) {
	c := newCircuitsProviderCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProvider)
	err := c.ICache.(CircuitsProviderCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsProviderCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(CircuitsProviderCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_circuitsProviderCache_MultiGet(t *testing.T) {
	c := newCircuitsProviderCache()
	defer c.Close()

	var testData []*model.CircuitsProvider
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsProvider))
	}

	err := c.ICache.(CircuitsProviderCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsProviderCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.CircuitsProvider))
	}
}

func Test_circuitsProviderCache_MultiSet(t *testing.T) {
	c := newCircuitsProviderCache()
	defer c.Close()

	var testData []*model.CircuitsProvider
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsProvider))
	}

	err := c.ICache.(CircuitsProviderCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsProviderCache_Del(t *testing.T) {
	c := newCircuitsProviderCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProvider)
	err := c.ICache.(CircuitsProviderCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsProviderCache_SetCacheWithNotFound(t *testing.T) {
	c := newCircuitsProviderCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProvider)
	err := c.ICache.(CircuitsProviderCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(CircuitsProviderCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewCircuitsProviderCache(t *testing.T) {
	c := NewCircuitsProviderCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewCircuitsProviderCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewCircuitsProviderCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
