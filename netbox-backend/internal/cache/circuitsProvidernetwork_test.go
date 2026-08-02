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

func newCircuitsProvidernetworkCache() *gotest.Cache {
	record1 := &model.CircuitsProvidernetwork{}
	record1.ID = 1
	record2 := &model.CircuitsProvidernetwork{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewCircuitsProvidernetworkCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_circuitsProvidernetworkCache_Set(t *testing.T) {
	c := newCircuitsProvidernetworkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProvidernetwork)
	err := c.ICache.(CircuitsProvidernetworkCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(CircuitsProvidernetworkCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_circuitsProvidernetworkCache_Get(t *testing.T) {
	c := newCircuitsProvidernetworkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProvidernetwork)
	err := c.ICache.(CircuitsProvidernetworkCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsProvidernetworkCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(CircuitsProvidernetworkCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_circuitsProvidernetworkCache_MultiGet(t *testing.T) {
	c := newCircuitsProvidernetworkCache()
	defer c.Close()

	var testData []*model.CircuitsProvidernetwork
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsProvidernetwork))
	}

	err := c.ICache.(CircuitsProvidernetworkCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsProvidernetworkCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.CircuitsProvidernetwork))
	}
}

func Test_circuitsProvidernetworkCache_MultiSet(t *testing.T) {
	c := newCircuitsProvidernetworkCache()
	defer c.Close()

	var testData []*model.CircuitsProvidernetwork
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsProvidernetwork))
	}

	err := c.ICache.(CircuitsProvidernetworkCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsProvidernetworkCache_Del(t *testing.T) {
	c := newCircuitsProvidernetworkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProvidernetwork)
	err := c.ICache.(CircuitsProvidernetworkCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsProvidernetworkCache_SetCacheWithNotFound(t *testing.T) {
	c := newCircuitsProvidernetworkCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProvidernetwork)
	err := c.ICache.(CircuitsProvidernetworkCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(CircuitsProvidernetworkCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewCircuitsProvidernetworkCache(t *testing.T) {
	c := NewCircuitsProvidernetworkCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewCircuitsProvidernetworkCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewCircuitsProvidernetworkCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
