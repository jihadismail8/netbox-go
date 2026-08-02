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

func newCircuitsProvideraccountCache() *gotest.Cache {
	record1 := &model.CircuitsProvideraccount{}
	record1.ID = 1
	record2 := &model.CircuitsProvideraccount{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewCircuitsProvideraccountCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_circuitsProvideraccountCache_Set(t *testing.T) {
	c := newCircuitsProvideraccountCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProvideraccount)
	err := c.ICache.(CircuitsProvideraccountCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(CircuitsProvideraccountCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_circuitsProvideraccountCache_Get(t *testing.T) {
	c := newCircuitsProvideraccountCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProvideraccount)
	err := c.ICache.(CircuitsProvideraccountCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsProvideraccountCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(CircuitsProvideraccountCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_circuitsProvideraccountCache_MultiGet(t *testing.T) {
	c := newCircuitsProvideraccountCache()
	defer c.Close()

	var testData []*model.CircuitsProvideraccount
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsProvideraccount))
	}

	err := c.ICache.(CircuitsProvideraccountCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CircuitsProvideraccountCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.CircuitsProvideraccount))
	}
}

func Test_circuitsProvideraccountCache_MultiSet(t *testing.T) {
	c := newCircuitsProvideraccountCache()
	defer c.Close()

	var testData []*model.CircuitsProvideraccount
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CircuitsProvideraccount))
	}

	err := c.ICache.(CircuitsProvideraccountCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsProvideraccountCache_Del(t *testing.T) {
	c := newCircuitsProvideraccountCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProvideraccount)
	err := c.ICache.(CircuitsProvideraccountCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_circuitsProvideraccountCache_SetCacheWithNotFound(t *testing.T) {
	c := newCircuitsProvideraccountCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CircuitsProvideraccount)
	err := c.ICache.(CircuitsProvideraccountCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(CircuitsProvideraccountCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewCircuitsProvideraccountCache(t *testing.T) {
	c := NewCircuitsProvideraccountCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewCircuitsProvideraccountCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewCircuitsProvideraccountCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
