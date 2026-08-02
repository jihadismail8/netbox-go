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

func newExtrasConfigcontextTenantGroupsCache() *gotest.Cache {
	record1 := &model.ExtrasConfigcontextTenantGroups{}
	record1.ID = 1
	record2 := &model.ExtrasConfigcontextTenantGroups{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasConfigcontextTenantGroupsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasConfigcontextTenantGroupsCache_Set(t *testing.T) {
	c := newExtrasConfigcontextTenantGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextTenantGroups)
	err := c.ICache.(ExtrasConfigcontextTenantGroupsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasConfigcontextTenantGroupsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasConfigcontextTenantGroupsCache_Get(t *testing.T) {
	c := newExtrasConfigcontextTenantGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextTenantGroups)
	err := c.ICache.(ExtrasConfigcontextTenantGroupsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextTenantGroupsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasConfigcontextTenantGroupsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasConfigcontextTenantGroupsCache_MultiGet(t *testing.T) {
	c := newExtrasConfigcontextTenantGroupsCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextTenantGroups
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextTenantGroups))
	}

	err := c.ICache.(ExtrasConfigcontextTenantGroupsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextTenantGroupsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasConfigcontextTenantGroups))
	}
}

func Test_extrasConfigcontextTenantGroupsCache_MultiSet(t *testing.T) {
	c := newExtrasConfigcontextTenantGroupsCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextTenantGroups
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextTenantGroups))
	}

	err := c.ICache.(ExtrasConfigcontextTenantGroupsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextTenantGroupsCache_Del(t *testing.T) {
	c := newExtrasConfigcontextTenantGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextTenantGroups)
	err := c.ICache.(ExtrasConfigcontextTenantGroupsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextTenantGroupsCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasConfigcontextTenantGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextTenantGroups)
	err := c.ICache.(ExtrasConfigcontextTenantGroupsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasConfigcontextTenantGroupsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasConfigcontextTenantGroupsCache(t *testing.T) {
	c := NewExtrasConfigcontextTenantGroupsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasConfigcontextTenantGroupsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasConfigcontextTenantGroupsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
