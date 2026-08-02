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

func newCircuitsVirtualcircuittypeCache() *gotest.Cache {
	record1 := &model.CircuitsVirtualcircuittype{}
	record1.ID = 1
	record2 := &model.CircuitsVirtualcircuittype{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewCircuitsVirtualcircuittypeCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_circuitsVirtualcircuittypeCache_Set(t *testing.T) {
	c := newCircuitsVirtualcircuittypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsVirtualcircuittype)
	err := c.ICache.(CircuitsVirtualcircuittypeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(CircuitsVirtualcircuittypeCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_circuitsVirtualcircuittypeCache_Get(t *testing.T) {
	c := newCircuitsVirtualcircuittypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsVirtualcircuittype)
	err := c.ICache.(CircuitsVirtualcircuittypeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsVirtualcircuittypeCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(CircuitsVirtualcircuittypeCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_circuitsVirtualcircuittypeCache_MultiGet(t *testing.T) {
	c := newCircuitsVirtualcircuittypeCache()
	defer c.Close()

	var testData []*model.CircuitsVirtualcircuittype
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsVirtualcircuittype))
	}

	err := c.ICache.(CircuitsVirtualcircuittypeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsVirtualcircuittypeCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.CircuitsVirtualcircuittype))
	}
}

func Test_circuitsVirtualcircuittypeCache_MultiSet(t *testing.T) {
	c := newCircuitsVirtualcircuittypeCache()
	defer c.Close()

	var testData []*model.CircuitsVirtualcircuittype
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsVirtualcircuittype))
	}

	err := c.ICache.(CircuitsVirtualcircuittypeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsVirtualcircuittypeCache_Del(t *testing.T) {
	c := newCircuitsVirtualcircuittypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsVirtualcircuittype)
	err := c.ICache.(CircuitsVirtualcircuittypeCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsVirtualcircuittypeCache_SetCacheWithNotFound(t *testing.T) {
	c := newCircuitsVirtualcircuittypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsVirtualcircuittype)
	err := c.ICache.(CircuitsVirtualcircuittypeCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(CircuitsVirtualcircuittypeCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewCircuitsVirtualcircuittypeCache(t *testing.T) {
	c := NewCircuitsVirtualcircuittypeCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewCircuitsVirtualcircuittypeCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewCircuitsVirtualcircuittypeCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
