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

func newTenancyContactGroupsCache() *gotest.Cache {
	record1 := &model.TenancyContactGroups{}
	record1.ID = 1
	record2 := &model.TenancyContactGroups{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewTenancyContactGroupsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_tenancyContactGroupsCache_Set(t *testing.T) {
	c := newTenancyContactGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyContactGroups)
	err := c.ICache.(TenancyContactGroupsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(TenancyContactGroupsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_tenancyContactGroupsCache_Get(t *testing.T) {
	c := newTenancyContactGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyContactGroups)
	err := c.ICache.(TenancyContactGroupsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(TenancyContactGroupsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(TenancyContactGroupsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_tenancyContactGroupsCache_MultiGet(t *testing.T) {
	c := newTenancyContactGroupsCache()
	defer c.Close()

	var testData []*model.TenancyContactGroups
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.TenancyContactGroups))
	}

	err := c.ICache.(TenancyContactGroupsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(TenancyContactGroupsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.TenancyContactGroups))
	}
}

func Test_tenancyContactGroupsCache_MultiSet(t *testing.T) {
	c := newTenancyContactGroupsCache()
	defer c.Close()

	var testData []*model.TenancyContactGroups
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.TenancyContactGroups))
	}

	err := c.ICache.(TenancyContactGroupsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_tenancyContactGroupsCache_Del(t *testing.T) {
	c := newTenancyContactGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyContactGroups)
	err := c.ICache.(TenancyContactGroupsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_tenancyContactGroupsCache_SetCacheWithNotFound(t *testing.T) {
	c := newTenancyContactGroupsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyContactGroups)
	err := c.ICache.(TenancyContactGroupsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(TenancyContactGroupsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewTenancyContactGroupsCache(t *testing.T) {
	c := NewTenancyContactGroupsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewTenancyContactGroupsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewTenancyContactGroupsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
