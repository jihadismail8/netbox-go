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

func newDcimDevicebayCache() *gotest.Cache {
	record1 := &model.DcimDevicebay{}
	record1.ID = 1
	record2 := &model.DcimDevicebay{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimDevicebayCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimDevicebayCache_Set(t *testing.T) {
	c := newDcimDevicebayCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimDevicebay)
	err := c.ICache.(DcimDevicebayCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimDevicebayCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimDevicebayCache_Get(t *testing.T) {
	c := newDcimDevicebayCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimDevicebay)
	err := c.ICache.(DcimDevicebayCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimDevicebayCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimDevicebayCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimDevicebayCache_MultiGet(t *testing.T) {
	c := newDcimDevicebayCache()
	defer c.Close()

	var testData []*model.DcimDevicebay
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimDevicebay))
	}

	err := c.ICache.(DcimDevicebayCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimDevicebayCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimDevicebay))
	}
}

func Test_dcimDevicebayCache_MultiSet(t *testing.T) {
	c := newDcimDevicebayCache()
	defer c.Close()

	var testData []*model.DcimDevicebay
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimDevicebay))
	}

	err := c.ICache.(DcimDevicebayCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimDevicebayCache_Del(t *testing.T) {
	c := newDcimDevicebayCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimDevicebay)
	err := c.ICache.(DcimDevicebayCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimDevicebayCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimDevicebayCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimDevicebay)
	err := c.ICache.(DcimDevicebayCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimDevicebayCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimDevicebayCache(t *testing.T) {
	c := NewDcimDevicebayCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimDevicebayCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimDevicebayCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
