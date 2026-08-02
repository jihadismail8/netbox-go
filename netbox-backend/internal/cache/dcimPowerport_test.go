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

func newDcimPowerportCache() *gotest.Cache {
	record1 := &model.DcimPowerport{}
	record1.ID = 1
	record2 := &model.DcimPowerport{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimPowerportCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimPowerportCache_Set(t *testing.T) {
	c := newDcimPowerportCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPowerport)
	err := c.ICache.(DcimPowerportCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimPowerportCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimPowerportCache_Get(t *testing.T) {
	c := newDcimPowerportCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPowerport)
	err := c.ICache.(DcimPowerportCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimPowerportCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimPowerportCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimPowerportCache_MultiGet(t *testing.T) {
	c := newDcimPowerportCache()
	defer c.Close()

	var testData []*model.DcimPowerport
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimPowerport))
	}

	err := c.ICache.(DcimPowerportCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimPowerportCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimPowerport))
	}
}

func Test_dcimPowerportCache_MultiSet(t *testing.T) {
	c := newDcimPowerportCache()
	defer c.Close()

	var testData []*model.DcimPowerport
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimPowerport))
	}

	err := c.ICache.(DcimPowerportCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimPowerportCache_Del(t *testing.T) {
	c := newDcimPowerportCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPowerport)
	err := c.ICache.(DcimPowerportCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimPowerportCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimPowerportCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPowerport)
	err := c.ICache.(DcimPowerportCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimPowerportCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimPowerportCache(t *testing.T) {
	c := NewDcimPowerportCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimPowerportCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimPowerportCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
