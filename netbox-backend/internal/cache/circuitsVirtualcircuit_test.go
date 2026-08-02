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

func newCircuitsVirtualcircuitCache() *gotest.Cache {
	record1 := &model.CircuitsVirtualcircuit{}
	record1.ID = 1
	record2 := &model.CircuitsVirtualcircuit{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewCircuitsVirtualcircuitCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_circuitsVirtualcircuitCache_Set(t *testing.T) {
	c := newCircuitsVirtualcircuitCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsVirtualcircuit)
	err := c.ICache.(CircuitsVirtualcircuitCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(CircuitsVirtualcircuitCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_circuitsVirtualcircuitCache_Get(t *testing.T) {
	c := newCircuitsVirtualcircuitCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsVirtualcircuit)
	err := c.ICache.(CircuitsVirtualcircuitCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsVirtualcircuitCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(CircuitsVirtualcircuitCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_circuitsVirtualcircuitCache_MultiGet(t *testing.T) {
	c := newCircuitsVirtualcircuitCache()
	defer c.Close()

	var testData []*model.CircuitsVirtualcircuit
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsVirtualcircuit))
	}

	err := c.ICache.(CircuitsVirtualcircuitCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsVirtualcircuitCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.CircuitsVirtualcircuit))
	}
}

func Test_circuitsVirtualcircuitCache_MultiSet(t *testing.T) {
	c := newCircuitsVirtualcircuitCache()
	defer c.Close()

	var testData []*model.CircuitsVirtualcircuit
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsVirtualcircuit))
	}

	err := c.ICache.(CircuitsVirtualcircuitCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsVirtualcircuitCache_Del(t *testing.T) {
	c := newCircuitsVirtualcircuitCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsVirtualcircuit)
	err := c.ICache.(CircuitsVirtualcircuitCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsVirtualcircuitCache_SetCacheWithNotFound(t *testing.T) {
	c := newCircuitsVirtualcircuitCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsVirtualcircuit)
	err := c.ICache.(CircuitsVirtualcircuitCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(CircuitsVirtualcircuitCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewCircuitsVirtualcircuitCache(t *testing.T) {
	c := NewCircuitsVirtualcircuitCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewCircuitsVirtualcircuitCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewCircuitsVirtualcircuitCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
