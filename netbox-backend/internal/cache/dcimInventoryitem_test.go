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

func newDcimInventoryitemCache() *gotest.Cache {
	record1 := &model.DcimInventoryitem{}
	record1.ID = 1
	record2 := &model.DcimInventoryitem{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimInventoryitemCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimInventoryitemCache_Set(t *testing.T) {
	c := newDcimInventoryitemCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInventoryitem)
	err := c.ICache.(DcimInventoryitemCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimInventoryitemCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimInventoryitemCache_Get(t *testing.T) {
	c := newDcimInventoryitemCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInventoryitem)
	err := c.ICache.(DcimInventoryitemCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimInventoryitemCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimInventoryitemCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimInventoryitemCache_MultiGet(t *testing.T) {
	c := newDcimInventoryitemCache()
	defer c.Close()

	var testData []*model.DcimInventoryitem
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimInventoryitem))
	}

	err := c.ICache.(DcimInventoryitemCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimInventoryitemCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimInventoryitem))
	}
}

func Test_dcimInventoryitemCache_MultiSet(t *testing.T) {
	c := newDcimInventoryitemCache()
	defer c.Close()

	var testData []*model.DcimInventoryitem
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimInventoryitem))
	}

	err := c.ICache.(DcimInventoryitemCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimInventoryitemCache_Del(t *testing.T) {
	c := newDcimInventoryitemCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInventoryitem)
	err := c.ICache.(DcimInventoryitemCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimInventoryitemCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimInventoryitemCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInventoryitem)
	err := c.ICache.(DcimInventoryitemCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimInventoryitemCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimInventoryitemCache(t *testing.T) {
	c := NewDcimInventoryitemCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimInventoryitemCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimInventoryitemCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
