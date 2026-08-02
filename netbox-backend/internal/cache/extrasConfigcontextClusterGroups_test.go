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

func newExtrasConfigcontextClusterGroupsCache() *gotest.Cache {
	record1 := &model.ExtrasConfigcontextClusterGroups{}
	record1.ID = 1
	record2 := &model.ExtrasConfigcontextClusterGroups{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasConfigcontextClusterGroupsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasConfigcontextClusterGroupsCache_Set(t *testing.T) {
	c := newExtrasConfigcontextClusterGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextClusterGroups)
	err := c.ICache.(ExtrasConfigcontextClusterGroupsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasConfigcontextClusterGroupsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasConfigcontextClusterGroupsCache_Get(t *testing.T) {
	c := newExtrasConfigcontextClusterGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextClusterGroups)
	err := c.ICache.(ExtrasConfigcontextClusterGroupsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextClusterGroupsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasConfigcontextClusterGroupsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasConfigcontextClusterGroupsCache_MultiGet(t *testing.T) {
	c := newExtrasConfigcontextClusterGroupsCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextClusterGroups
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextClusterGroups))
	}

	err := c.ICache.(ExtrasConfigcontextClusterGroupsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextClusterGroupsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasConfigcontextClusterGroups))
	}
}

func Test_extrasConfigcontextClusterGroupsCache_MultiSet(t *testing.T) {
	c := newExtrasConfigcontextClusterGroupsCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextClusterGroups
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextClusterGroups))
	}

	err := c.ICache.(ExtrasConfigcontextClusterGroupsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextClusterGroupsCache_Del(t *testing.T) {
	c := newExtrasConfigcontextClusterGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextClusterGroups)
	err := c.ICache.(ExtrasConfigcontextClusterGroupsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextClusterGroupsCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasConfigcontextClusterGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextClusterGroups)
	err := c.ICache.(ExtrasConfigcontextClusterGroupsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasConfigcontextClusterGroupsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasConfigcontextClusterGroupsCache(t *testing.T) {
	c := NewExtrasConfigcontextClusterGroupsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasConfigcontextClusterGroupsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasConfigcontextClusterGroupsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
