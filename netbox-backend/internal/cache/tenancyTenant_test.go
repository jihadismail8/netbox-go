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

func newTenancyTenantCache() *gotest.Cache {
	record1 := &model.TenancyTenant{}
	record1.ID = 1
	record2 := &model.TenancyTenant{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewTenancyTenantCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_tenancyTenantCache_Set(t *testing.T) {
	c := newTenancyTenantCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyTenant)
	err := c.ICache.(TenancyTenantCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(TenancyTenantCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_tenancyTenantCache_Get(t *testing.T) {
	c := newTenancyTenantCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyTenant)
	err := c.ICache.(TenancyTenantCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(TenancyTenantCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(TenancyTenantCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_tenancyTenantCache_MultiGet(t *testing.T) {
	c := newTenancyTenantCache()
	defer c.Close()

	var testData []*model.TenancyTenant
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.TenancyTenant))
	}

	err := c.ICache.(TenancyTenantCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(TenancyTenantCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.TenancyTenant))
	}
}

func Test_tenancyTenantCache_MultiSet(t *testing.T) {
	c := newTenancyTenantCache()
	defer c.Close()

	var testData []*model.TenancyTenant
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.TenancyTenant))
	}

	err := c.ICache.(TenancyTenantCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_tenancyTenantCache_Del(t *testing.T) {
	c := newTenancyTenantCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyTenant)
	err := c.ICache.(TenancyTenantCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_tenancyTenantCache_SetCacheWithNotFound(t *testing.T) {
	c := newTenancyTenantCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyTenant)
	err := c.ICache.(TenancyTenantCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(TenancyTenantCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewTenancyTenantCache(t *testing.T) {
	c := NewTenancyTenantCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewTenancyTenantCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewTenancyTenantCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
