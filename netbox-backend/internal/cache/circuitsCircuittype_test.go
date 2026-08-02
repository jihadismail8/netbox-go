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

func newCircuitsCircuittypeCache() *gotest.Cache {
	record1 := &model.CircuitsCircuittype{}
	record1.ID = 1
	record2 := &model.CircuitsCircuittype{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewCircuitsCircuittypeCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_circuitsCircuittypeCache_Set(t *testing.T) {
	c := newCircuitsCircuittypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsCircuittype)
	err := c.ICache.(CircuitsCircuittypeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(CircuitsCircuittypeCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_circuitsCircuittypeCache_Get(t *testing.T) {
	c := newCircuitsCircuittypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsCircuittype)
	err := c.ICache.(CircuitsCircuittypeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsCircuittypeCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(CircuitsCircuittypeCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_circuitsCircuittypeCache_MultiGet(t *testing.T) {
	c := newCircuitsCircuittypeCache()
	defer c.Close()

	var testData []*model.CircuitsCircuittype
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsCircuittype))
	}

	err := c.ICache.(CircuitsCircuittypeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsCircuittypeCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.CircuitsCircuittype))
	}
}

func Test_circuitsCircuittypeCache_MultiSet(t *testing.T) {
	c := newCircuitsCircuittypeCache()
	defer c.Close()

	var testData []*model.CircuitsCircuittype
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsCircuittype))
	}

	err := c.ICache.(CircuitsCircuittypeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsCircuittypeCache_Del(t *testing.T) {
	c := newCircuitsCircuittypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsCircuittype)
	err := c.ICache.(CircuitsCircuittypeCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsCircuittypeCache_SetCacheWithNotFound(t *testing.T) {
	c := newCircuitsCircuittypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsCircuittype)
	err := c.ICache.(CircuitsCircuittypeCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(CircuitsCircuittypeCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewCircuitsCircuittypeCache(t *testing.T) {
	c := NewCircuitsCircuittypeCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewCircuitsCircuittypeCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewCircuitsCircuittypeCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
