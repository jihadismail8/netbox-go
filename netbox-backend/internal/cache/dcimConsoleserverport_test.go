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

func newDcimConsoleserverportCache() *gotest.Cache {
	record1 := &model.DcimConsoleserverport{}
	record1.ID = 1
	record2 := &model.DcimConsoleserverport{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimConsoleserverportCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimConsoleserverportCache_Set(t *testing.T) {
	c := newDcimConsoleserverportCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleserverport)
	err := c.ICache.(DcimConsoleserverportCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimConsoleserverportCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimConsoleserverportCache_Get(t *testing.T) {
	c := newDcimConsoleserverportCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleserverport)
	err := c.ICache.(DcimConsoleserverportCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimConsoleserverportCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimConsoleserverportCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimConsoleserverportCache_MultiGet(t *testing.T) {
	c := newDcimConsoleserverportCache()
	defer c.Close()

	var testData []*model.DcimConsoleserverport
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimConsoleserverport))
	}

	err := c.ICache.(DcimConsoleserverportCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimConsoleserverportCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimConsoleserverport))
	}
}

func Test_dcimConsoleserverportCache_MultiSet(t *testing.T) {
	c := newDcimConsoleserverportCache()
	defer c.Close()

	var testData []*model.DcimConsoleserverport
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimConsoleserverport))
	}

	err := c.ICache.(DcimConsoleserverportCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimConsoleserverportCache_Del(t *testing.T) {
	c := newDcimConsoleserverportCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleserverport)
	err := c.ICache.(DcimConsoleserverportCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimConsoleserverportCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimConsoleserverportCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleserverport)
	err := c.ICache.(DcimConsoleserverportCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimConsoleserverportCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimConsoleserverportCache(t *testing.T) {
	c := NewDcimConsoleserverportCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimConsoleserverportCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimConsoleserverportCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
