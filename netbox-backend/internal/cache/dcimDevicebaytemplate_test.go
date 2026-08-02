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

func newDcimDevicebaytemplateCache() *gotest.Cache {
	record1 := &model.DcimDevicebaytemplate{}
	record1.ID = 1
	record2 := &model.DcimDevicebaytemplate{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimDevicebaytemplateCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimDevicebaytemplateCache_Set(t *testing.T) {
	c := newDcimDevicebaytemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimDevicebaytemplate)
	err := c.ICache.(DcimDevicebaytemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimDevicebaytemplateCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimDevicebaytemplateCache_Get(t *testing.T) {
	c := newDcimDevicebaytemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimDevicebaytemplate)
	err := c.ICache.(DcimDevicebaytemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimDevicebaytemplateCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimDevicebaytemplateCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimDevicebaytemplateCache_MultiGet(t *testing.T) {
	c := newDcimDevicebaytemplateCache()
	defer c.Close()

	var testData []*model.DcimDevicebaytemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimDevicebaytemplate))
	}

	err := c.ICache.(DcimDevicebaytemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimDevicebaytemplateCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimDevicebaytemplate))
	}
}

func Test_dcimDevicebaytemplateCache_MultiSet(t *testing.T) {
	c := newDcimDevicebaytemplateCache()
	defer c.Close()

	var testData []*model.DcimDevicebaytemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimDevicebaytemplate))
	}

	err := c.ICache.(DcimDevicebaytemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimDevicebaytemplateCache_Del(t *testing.T) {
	c := newDcimDevicebaytemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimDevicebaytemplate)
	err := c.ICache.(DcimDevicebaytemplateCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimDevicebaytemplateCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimDevicebaytemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimDevicebaytemplate)
	err := c.ICache.(DcimDevicebaytemplateCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimDevicebaytemplateCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimDevicebaytemplateCache(t *testing.T) {
	c := NewDcimDevicebaytemplateCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimDevicebaytemplateCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimDevicebaytemplateCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
