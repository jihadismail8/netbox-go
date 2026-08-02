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

func newTenancyContactroleCache() *gotest.Cache {
	record1 := &model.TenancyContactrole{}
	record1.ID = 1
	record2 := &model.TenancyContactrole{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewTenancyContactroleCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_tenancyContactroleCache_Set(t *testing.T) {
	c := newTenancyContactroleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyContactrole)
	err := c.ICache.(TenancyContactroleCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(TenancyContactroleCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_tenancyContactroleCache_Get(t *testing.T) {
	c := newTenancyContactroleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyContactrole)
	err := c.ICache.(TenancyContactroleCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(TenancyContactroleCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(TenancyContactroleCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_tenancyContactroleCache_MultiGet(t *testing.T) {
	c := newTenancyContactroleCache()
	defer c.Close()

	var testData []*model.TenancyContactrole
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.TenancyContactrole))
	}

	err := c.ICache.(TenancyContactroleCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(TenancyContactroleCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.TenancyContactrole))
	}
}

func Test_tenancyContactroleCache_MultiSet(t *testing.T) {
	c := newTenancyContactroleCache()
	defer c.Close()

	var testData []*model.TenancyContactrole
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.TenancyContactrole))
	}

	err := c.ICache.(TenancyContactroleCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_tenancyContactroleCache_Del(t *testing.T) {
	c := newTenancyContactroleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyContactrole)
	err := c.ICache.(TenancyContactroleCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_tenancyContactroleCache_SetCacheWithNotFound(t *testing.T) {
	c := newTenancyContactroleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyContactrole)
	err := c.ICache.(TenancyContactroleCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(TenancyContactroleCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewTenancyContactroleCache(t *testing.T) {
	c := NewTenancyContactroleCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewTenancyContactroleCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewTenancyContactroleCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
