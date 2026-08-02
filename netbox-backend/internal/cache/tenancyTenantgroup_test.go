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

func newTenancyTenantgroupCache() *gotest.Cache {
	record1 := &model.TenancyTenantgroup{}
	record1.ID = 1
	record2 := &model.TenancyTenantgroup{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewTenancyTenantgroupCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_tenancyTenantgroupCache_Set(t *testing.T) {
	c := newTenancyTenantgroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyTenantgroup)
	err := c.ICache.(TenancyTenantgroupCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(TenancyTenantgroupCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_tenancyTenantgroupCache_Get(t *testing.T) {
	c := newTenancyTenantgroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyTenantgroup)
	err := c.ICache.(TenancyTenantgroupCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(TenancyTenantgroupCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(TenancyTenantgroupCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_tenancyTenantgroupCache_MultiGet(t *testing.T) {
	c := newTenancyTenantgroupCache()
	defer c.Close()

	var testData []*model.TenancyTenantgroup
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.TenancyTenantgroup))
	}

	err := c.ICache.(TenancyTenantgroupCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(TenancyTenantgroupCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.TenancyTenantgroup))
	}
}

func Test_tenancyTenantgroupCache_MultiSet(t *testing.T) {
	c := newTenancyTenantgroupCache()
	defer c.Close()

	var testData []*model.TenancyTenantgroup
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.TenancyTenantgroup))
	}

	err := c.ICache.(TenancyTenantgroupCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_tenancyTenantgroupCache_Del(t *testing.T) {
	c := newTenancyTenantgroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyTenantgroup)
	err := c.ICache.(TenancyTenantgroupCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_tenancyTenantgroupCache_SetCacheWithNotFound(t *testing.T) {
	c := newTenancyTenantgroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyTenantgroup)
	err := c.ICache.(TenancyTenantgroupCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(TenancyTenantgroupCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewTenancyTenantgroupCache(t *testing.T) {
	c := NewTenancyTenantgroupCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewTenancyTenantgroupCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewTenancyTenantgroupCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
