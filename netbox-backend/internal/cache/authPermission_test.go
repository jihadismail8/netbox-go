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

func newAuthPermissionCache() *gotest.Cache {
	record1 := &model.AuthPermission{}
	record1.ID = 1
	record2 := &model.AuthPermission{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewAuthPermissionCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_authPermissionCache_Set(t *testing.T) {
	c := newAuthPermissionCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.AuthPermission)
	err := c.ICache.(AuthPermissionCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(AuthPermissionCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_authPermissionCache_Get(t *testing.T) {
	c := newAuthPermissionCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.AuthPermission)
	err := c.ICache.(AuthPermissionCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(AuthPermissionCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(AuthPermissionCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_authPermissionCache_MultiGet(t *testing.T) {
	c := newAuthPermissionCache()
	defer c.Close()

	var testData []*model.AuthPermission
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.AuthPermission))
	}

	err := c.ICache.(AuthPermissionCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(AuthPermissionCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.AuthPermission))
	}
}

func Test_authPermissionCache_MultiSet(t *testing.T) {
	c := newAuthPermissionCache()
	defer c.Close()

	var testData []*model.AuthPermission
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.AuthPermission))
	}

	err := c.ICache.(AuthPermissionCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_authPermissionCache_Del(t *testing.T) {
	c := newAuthPermissionCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.AuthPermission)
	err := c.ICache.(AuthPermissionCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_authPermissionCache_SetCacheWithNotFound(t *testing.T) {
	c := newAuthPermissionCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.AuthPermission)
	err := c.ICache.(AuthPermissionCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(AuthPermissionCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewAuthPermissionCache(t *testing.T) {
	c := NewAuthPermissionCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewAuthPermissionCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewAuthPermissionCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
