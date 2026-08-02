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

func newTenancyContactassignmentCache() *gotest.Cache {
	record1 := &model.TenancyContactassignment{}
	record1.ID = 1
	record2 := &model.TenancyContactassignment{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewTenancyContactassignmentCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_tenancyContactassignmentCache_Set(t *testing.T) {
	c := newTenancyContactassignmentCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyContactassignment)
	err := c.ICache.(TenancyContactassignmentCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(TenancyContactassignmentCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_tenancyContactassignmentCache_Get(t *testing.T) {
	c := newTenancyContactassignmentCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyContactassignment)
	err := c.ICache.(TenancyContactassignmentCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(TenancyContactassignmentCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(TenancyContactassignmentCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_tenancyContactassignmentCache_MultiGet(t *testing.T) {
	c := newTenancyContactassignmentCache()
	defer c.Close()

	var testData []*model.TenancyContactassignment
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.TenancyContactassignment))
	}

	err := c.ICache.(TenancyContactassignmentCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(TenancyContactassignmentCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.TenancyContactassignment))
	}
}

func Test_tenancyContactassignmentCache_MultiSet(t *testing.T) {
	c := newTenancyContactassignmentCache()
	defer c.Close()

	var testData []*model.TenancyContactassignment
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.TenancyContactassignment))
	}

	err := c.ICache.(TenancyContactassignmentCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_tenancyContactassignmentCache_Del(t *testing.T) {
	c := newTenancyContactassignmentCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyContactassignment)
	err := c.ICache.(TenancyContactassignmentCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_tenancyContactassignmentCache_SetCacheWithNotFound(t *testing.T) {
	c := newTenancyContactassignmentCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.TenancyContactassignment)
	err := c.ICache.(TenancyContactassignmentCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(TenancyContactassignmentCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewTenancyContactassignmentCache(t *testing.T) {
	c := NewTenancyContactassignmentCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewTenancyContactassignmentCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewTenancyContactassignmentCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
