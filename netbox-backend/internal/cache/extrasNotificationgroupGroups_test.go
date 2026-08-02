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

func newExtrasNotificationgroupGroupsCache() *gotest.Cache {
	record1 := &model.ExtrasNotificationgroupGroups{}
	record1.ID = 1
	record2 := &model.ExtrasNotificationgroupGroups{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasNotificationgroupGroupsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasNotificationgroupGroupsCache_Set(t *testing.T) {
	c := newExtrasNotificationgroupGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasNotificationgroupGroups)
	err := c.ICache.(ExtrasNotificationgroupGroupsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasNotificationgroupGroupsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasNotificationgroupGroupsCache_Get(t *testing.T) {
	c := newExtrasNotificationgroupGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasNotificationgroupGroups)
	err := c.ICache.(ExtrasNotificationgroupGroupsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasNotificationgroupGroupsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasNotificationgroupGroupsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasNotificationgroupGroupsCache_MultiGet(t *testing.T) {
	c := newExtrasNotificationgroupGroupsCache()
	defer c.Close()

	var testData []*model.ExtrasNotificationgroupGroups
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasNotificationgroupGroups))
	}

	err := c.ICache.(ExtrasNotificationgroupGroupsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasNotificationgroupGroupsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasNotificationgroupGroups))
	}
}

func Test_extrasNotificationgroupGroupsCache_MultiSet(t *testing.T) {
	c := newExtrasNotificationgroupGroupsCache()
	defer c.Close()

	var testData []*model.ExtrasNotificationgroupGroups
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasNotificationgroupGroups))
	}

	err := c.ICache.(ExtrasNotificationgroupGroupsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasNotificationgroupGroupsCache_Del(t *testing.T) {
	c := newExtrasNotificationgroupGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasNotificationgroupGroups)
	err := c.ICache.(ExtrasNotificationgroupGroupsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasNotificationgroupGroupsCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasNotificationgroupGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasNotificationgroupGroups)
	err := c.ICache.(ExtrasNotificationgroupGroupsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasNotificationgroupGroupsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasNotificationgroupGroupsCache(t *testing.T) {
	c := NewExtrasNotificationgroupGroupsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasNotificationgroupGroupsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasNotificationgroupGroupsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
