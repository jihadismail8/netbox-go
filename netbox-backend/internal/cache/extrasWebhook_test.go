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

func newExtrasWebhookCache() *gotest.Cache {
	record1 := &model.ExtrasWebhook{}
	record1.ID = 1
	record2 := &model.ExtrasWebhook{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasWebhookCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasWebhookCache_Set(t *testing.T) {
	c := newExtrasWebhookCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasWebhook)
	err := c.ICache.(ExtrasWebhookCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasWebhookCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasWebhookCache_Get(t *testing.T) {
	c := newExtrasWebhookCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasWebhook)
	err := c.ICache.(ExtrasWebhookCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasWebhookCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasWebhookCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasWebhookCache_MultiGet(t *testing.T) {
	c := newExtrasWebhookCache()
	defer c.Close()

	var testData []*model.ExtrasWebhook
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasWebhook))
	}

	err := c.ICache.(ExtrasWebhookCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasWebhookCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasWebhook))
	}
}

func Test_extrasWebhookCache_MultiSet(t *testing.T) {
	c := newExtrasWebhookCache()
	defer c.Close()

	var testData []*model.ExtrasWebhook
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasWebhook))
	}

	err := c.ICache.(ExtrasWebhookCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasWebhookCache_Del(t *testing.T) {
	c := newExtrasWebhookCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasWebhook)
	err := c.ICache.(ExtrasWebhookCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasWebhookCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasWebhookCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasWebhook)
	err := c.ICache.(ExtrasWebhookCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasWebhookCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasWebhookCache(t *testing.T) {
	c := NewExtrasWebhookCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasWebhookCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasWebhookCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
