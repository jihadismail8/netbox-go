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

func newExtrasConfigcontextSiteGroupsCache() *gotest.Cache {
	record1 := &model.ExtrasConfigcontextSiteGroups{}
	record1.ID = 1
	record2 := &model.ExtrasConfigcontextSiteGroups{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasConfigcontextSiteGroupsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasConfigcontextSiteGroupsCache_Set(t *testing.T) {
	c := newExtrasConfigcontextSiteGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextSiteGroups)
	err := c.ICache.(ExtrasConfigcontextSiteGroupsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasConfigcontextSiteGroupsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasConfigcontextSiteGroupsCache_Get(t *testing.T) {
	c := newExtrasConfigcontextSiteGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextSiteGroups)
	err := c.ICache.(ExtrasConfigcontextSiteGroupsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextSiteGroupsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasConfigcontextSiteGroupsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasConfigcontextSiteGroupsCache_MultiGet(t *testing.T) {
	c := newExtrasConfigcontextSiteGroupsCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextSiteGroups
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextSiteGroups))
	}

	err := c.ICache.(ExtrasConfigcontextSiteGroupsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextSiteGroupsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasConfigcontextSiteGroups))
	}
}

func Test_extrasConfigcontextSiteGroupsCache_MultiSet(t *testing.T) {
	c := newExtrasConfigcontextSiteGroupsCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextSiteGroups
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextSiteGroups))
	}

	err := c.ICache.(ExtrasConfigcontextSiteGroupsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextSiteGroupsCache_Del(t *testing.T) {
	c := newExtrasConfigcontextSiteGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextSiteGroups)
	err := c.ICache.(ExtrasConfigcontextSiteGroupsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextSiteGroupsCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasConfigcontextSiteGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextSiteGroups)
	err := c.ICache.(ExtrasConfigcontextSiteGroupsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasConfigcontextSiteGroupsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasConfigcontextSiteGroupsCache(t *testing.T) {
	c := NewExtrasConfigcontextSiteGroupsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasConfigcontextSiteGroupsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasConfigcontextSiteGroupsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
