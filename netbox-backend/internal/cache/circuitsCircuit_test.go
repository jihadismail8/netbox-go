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

func newCircuitsCircuitCache() *gotest.Cache {
	record1 := &model.CircuitsCircuit{}
	record1.ID = 1
	record2 := &model.CircuitsCircuit{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewCircuitsCircuitCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_circuitsCircuitCache_Set(t *testing.T) {
	c := newCircuitsCircuitCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsCircuit)
	err := c.ICache.(CircuitsCircuitCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(CircuitsCircuitCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_circuitsCircuitCache_Get(t *testing.T) {
	c := newCircuitsCircuitCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsCircuit)
	err := c.ICache.(CircuitsCircuitCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsCircuitCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(CircuitsCircuitCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_circuitsCircuitCache_MultiGet(t *testing.T) {
	c := newCircuitsCircuitCache()
	defer c.Close()

	var testData []*model.CircuitsCircuit
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsCircuit))
	}

	err := c.ICache.(CircuitsCircuitCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsCircuitCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.CircuitsCircuit))
	}
}

func Test_circuitsCircuitCache_MultiSet(t *testing.T) {
	c := newCircuitsCircuitCache()
	defer c.Close()

	var testData []*model.CircuitsCircuit
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsCircuit))
	}

	err := c.ICache.(CircuitsCircuitCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsCircuitCache_Del(t *testing.T) {
	c := newCircuitsCircuitCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsCircuit)
	err := c.ICache.(CircuitsCircuitCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsCircuitCache_SetCacheWithNotFound(t *testing.T) {
	c := newCircuitsCircuitCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsCircuit)
	err := c.ICache.(CircuitsCircuitCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(CircuitsCircuitCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewCircuitsCircuitCache(t *testing.T) {
	c := NewCircuitsCircuitCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewCircuitsCircuitCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewCircuitsCircuitCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
