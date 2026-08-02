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

func newAuthGroupPermissionsCache() *gotest.Cache {
	record1 := &model.AuthGroupPermissions{}
	record1.ID = 1
	record2 := &model.AuthGroupPermissions{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewAuthGroupPermissionsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_authGroupPermissionsCache_Set(t *testing.T) {
	c := newAuthGroupPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.AuthGroupPermissions)
	err := c.ICache.(AuthGroupPermissionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(AuthGroupPermissionsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_authGroupPermissionsCache_Get(t *testing.T) {
	c := newAuthGroupPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.AuthGroupPermissions)
	err := c.ICache.(AuthGroupPermissionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(AuthGroupPermissionsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(AuthGroupPermissionsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_authGroupPermissionsCache_MultiGet(t *testing.T) {
	c := newAuthGroupPermissionsCache()
	defer c.Close()

	var testData []*model.AuthGroupPermissions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.AuthGroupPermissions))
	}

	err := c.ICache.(AuthGroupPermissionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(AuthGroupPermissionsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.AuthGroupPermissions))
	}
}

func Test_authGroupPermissionsCache_MultiSet(t *testing.T) {
	c := newAuthGroupPermissionsCache()
	defer c.Close()

	var testData []*model.AuthGroupPermissions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.AuthGroupPermissions))
	}

	err := c.ICache.(AuthGroupPermissionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_authGroupPermissionsCache_Del(t *testing.T) {
	c := newAuthGroupPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.AuthGroupPermissions)
	err := c.ICache.(AuthGroupPermissionsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_authGroupPermissionsCache_SetCacheWithNotFound(t *testing.T) {
	c := newAuthGroupPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.AuthGroupPermissions)
	err := c.ICache.(AuthGroupPermissionsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(AuthGroupPermissionsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewAuthGroupPermissionsCache(t *testing.T) {
	c := NewAuthGroupPermissionsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewAuthGroupPermissionsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewAuthGroupPermissionsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
