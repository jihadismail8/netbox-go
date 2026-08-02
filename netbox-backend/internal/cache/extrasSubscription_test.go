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

func newExtrasSubscriptionCache() *gotest.Cache {
	record1 := &model.ExtrasSubscription{}
	record1.ID = 1
	record2 := &model.ExtrasSubscription{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasSubscriptionCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasSubscriptionCache_Set(t *testing.T) {
	c := newExtrasSubscriptionCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasSubscription)
	err := c.ICache.(ExtrasSubscriptionCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasSubscriptionCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasSubscriptionCache_Get(t *testing.T) {
	c := newExtrasSubscriptionCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasSubscription)
	err := c.ICache.(ExtrasSubscriptionCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasSubscriptionCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasSubscriptionCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasSubscriptionCache_MultiGet(t *testing.T) {
	c := newExtrasSubscriptionCache()
	defer c.Close()

	var testData []*model.ExtrasSubscription
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasSubscription))
	}

	err := c.ICache.(ExtrasSubscriptionCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasSubscriptionCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasSubscription))
	}
}

func Test_extrasSubscriptionCache_MultiSet(t *testing.T) {
	c := newExtrasSubscriptionCache()
	defer c.Close()

	var testData []*model.ExtrasSubscription
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasSubscription))
	}

	err := c.ICache.(ExtrasSubscriptionCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasSubscriptionCache_Del(t *testing.T) {
	c := newExtrasSubscriptionCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasSubscription)
	err := c.ICache.(ExtrasSubscriptionCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasSubscriptionCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasSubscriptionCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasSubscription)
	err := c.ICache.(ExtrasSubscriptionCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasSubscriptionCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasSubscriptionCache(t *testing.T) {
	c := NewExtrasSubscriptionCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasSubscriptionCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasSubscriptionCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
