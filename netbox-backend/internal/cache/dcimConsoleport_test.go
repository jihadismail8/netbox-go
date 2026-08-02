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

func newDcimConsoleportCache() *gotest.Cache {
	record1 := &model.DcimConsoleport{}
	record1.ID = 1
	record2 := &model.DcimConsoleport{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimConsoleportCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimConsoleportCache_Set(t *testing.T) {
	c := newDcimConsoleportCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleport)
	err := c.ICache.(DcimConsoleportCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimConsoleportCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimConsoleportCache_Get(t *testing.T) {
	c := newDcimConsoleportCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleport)
	err := c.ICache.(DcimConsoleportCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimConsoleportCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimConsoleportCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimConsoleportCache_MultiGet(t *testing.T) {
	c := newDcimConsoleportCache()
	defer c.Close()

	var testData []*model.DcimConsoleport
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimConsoleport))
	}

	err := c.ICache.(DcimConsoleportCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimConsoleportCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimConsoleport))
	}
}

func Test_dcimConsoleportCache_MultiSet(t *testing.T) {
	c := newDcimConsoleportCache()
	defer c.Close()

	var testData []*model.DcimConsoleport
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimConsoleport))
	}

	err := c.ICache.(DcimConsoleportCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimConsoleportCache_Del(t *testing.T) {
	c := newDcimConsoleportCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleport)
	err := c.ICache.(DcimConsoleportCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimConsoleportCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimConsoleportCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleport)
	err := c.ICache.(DcimConsoleportCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimConsoleportCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimConsoleportCache(t *testing.T) {
	c := NewDcimConsoleportCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimConsoleportCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimConsoleportCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
