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

func newExtrasNotificationCache() *gotest.Cache {
	record1 := &model.ExtrasNotification{}
	record1.ID = 1
	record2 := &model.ExtrasNotification{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasNotificationCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasNotificationCache_Set(t *testing.T) {
	c := newExtrasNotificationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasNotification)
	err := c.ICache.(ExtrasNotificationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasNotificationCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasNotificationCache_Get(t *testing.T) {
	c := newExtrasNotificationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasNotification)
	err := c.ICache.(ExtrasNotificationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasNotificationCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasNotificationCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasNotificationCache_MultiGet(t *testing.T) {
	c := newExtrasNotificationCache()
	defer c.Close()

	var testData []*model.ExtrasNotification
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasNotification))
	}

	err := c.ICache.(ExtrasNotificationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasNotificationCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasNotification))
	}
}

func Test_extrasNotificationCache_MultiSet(t *testing.T) {
	c := newExtrasNotificationCache()
	defer c.Close()

	var testData []*model.ExtrasNotification
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasNotification))
	}

	err := c.ICache.(ExtrasNotificationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasNotificationCache_Del(t *testing.T) {
	c := newExtrasNotificationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasNotification)
	err := c.ICache.(ExtrasNotificationCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasNotificationCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasNotificationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasNotification)
	err := c.ICache.(ExtrasNotificationCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasNotificationCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasNotificationCache(t *testing.T) {
	c := NewExtrasNotificationCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasNotificationCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasNotificationCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
