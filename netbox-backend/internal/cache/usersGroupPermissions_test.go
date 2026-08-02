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

func newUsersGroupPermissionsCache() *gotest.Cache {
	record1 := &model.UsersGroupPermissions{}
	record1.ID = 1
	record2 := &model.UsersGroupPermissions{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewUsersGroupPermissionsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_usersGroupPermissionsCache_Set(t *testing.T) {
	c := newUsersGroupPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersGroupPermissions)
	err := c.ICache.(UsersGroupPermissionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(UsersGroupPermissionsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_usersGroupPermissionsCache_Get(t *testing.T) {
	c := newUsersGroupPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersGroupPermissions)
	err := c.ICache.(UsersGroupPermissionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(UsersGroupPermissionsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(UsersGroupPermissionsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_usersGroupPermissionsCache_MultiGet(t *testing.T) {
	c := newUsersGroupPermissionsCache()
	defer c.Close()

	var testData []*model.UsersGroupPermissions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.UsersGroupPermissions))
	}

	err := c.ICache.(UsersGroupPermissionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(UsersGroupPermissionsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.UsersGroupPermissions))
	}
}

func Test_usersGroupPermissionsCache_MultiSet(t *testing.T) {
	c := newUsersGroupPermissionsCache()
	defer c.Close()

	var testData []*model.UsersGroupPermissions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.UsersGroupPermissions))
	}

	err := c.ICache.(UsersGroupPermissionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_usersGroupPermissionsCache_Del(t *testing.T) {
	c := newUsersGroupPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersGroupPermissions)
	err := c.ICache.(UsersGroupPermissionsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_usersGroupPermissionsCache_SetCacheWithNotFound(t *testing.T) {
	c := newUsersGroupPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersGroupPermissions)
	err := c.ICache.(UsersGroupPermissionsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(UsersGroupPermissionsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewUsersGroupPermissionsCache(t *testing.T) {
	c := NewUsersGroupPermissionsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewUsersGroupPermissionsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewUsersGroupPermissionsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
