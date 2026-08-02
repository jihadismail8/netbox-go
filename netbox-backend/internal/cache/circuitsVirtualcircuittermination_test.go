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

func newCircuitsVirtualcircuitterminationCache() *gotest.Cache {
	record1 := &model.CircuitsVirtualcircuittermination{}
	record1.ID = 1
	record2 := &model.CircuitsVirtualcircuittermination{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewCircuitsVirtualcircuitterminationCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_circuitsVirtualcircuitterminationCache_Set(t *testing.T) {
	c := newCircuitsVirtualcircuitterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsVirtualcircuittermination)
	err := c.ICache.(CircuitsVirtualcircuitterminationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(CircuitsVirtualcircuitterminationCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_circuitsVirtualcircuitterminationCache_Get(t *testing.T) {
	c := newCircuitsVirtualcircuitterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsVirtualcircuittermination)
	err := c.ICache.(CircuitsVirtualcircuitterminationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsVirtualcircuitterminationCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(CircuitsVirtualcircuitterminationCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_circuitsVirtualcircuitterminationCache_MultiGet(t *testing.T) {
	c := newCircuitsVirtualcircuitterminationCache()
	defer c.Close()

	var testData []*model.CircuitsVirtualcircuittermination
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsVirtualcircuittermination))
	}

	err := c.ICache.(CircuitsVirtualcircuitterminationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsVirtualcircuitterminationCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.CircuitsVirtualcircuittermination))
	}
}

func Test_circuitsVirtualcircuitterminationCache_MultiSet(t *testing.T) {
	c := newCircuitsVirtualcircuitterminationCache()
	defer c.Close()

	var testData []*model.CircuitsVirtualcircuittermination
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsVirtualcircuittermination))
	}

	err := c.ICache.(CircuitsVirtualcircuitterminationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsVirtualcircuitterminationCache_Del(t *testing.T) {
	c := newCircuitsVirtualcircuitterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsVirtualcircuittermination)
	err := c.ICache.(CircuitsVirtualcircuitterminationCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsVirtualcircuitterminationCache_SetCacheWithNotFound(t *testing.T) {
	c := newCircuitsVirtualcircuitterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsVirtualcircuittermination)
	err := c.ICache.(CircuitsVirtualcircuitterminationCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(CircuitsVirtualcircuitterminationCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewCircuitsVirtualcircuitterminationCache(t *testing.T) {
	c := NewCircuitsVirtualcircuitterminationCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewCircuitsVirtualcircuitterminationCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewCircuitsVirtualcircuitterminationCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
