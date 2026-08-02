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

func newUsersGroupObjectPermissionsCache() *gotest.Cache {
	record1 := &model.UsersGroupObjectPermissions{}
	record1.ID = 1
	record2 := &model.UsersGroupObjectPermissions{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewUsersGroupObjectPermissionsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_usersGroupObjectPermissionsCache_Set(t *testing.T) {
	c := newUsersGroupObjectPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersGroupObjectPermissions)
	err := c.ICache.(UsersGroupObjectPermissionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(UsersGroupObjectPermissionsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_usersGroupObjectPermissionsCache_Get(t *testing.T) {
	c := newUsersGroupObjectPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersGroupObjectPermissions)
	err := c.ICache.(UsersGroupObjectPermissionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(UsersGroupObjectPermissionsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(UsersGroupObjectPermissionsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_usersGroupObjectPermissionsCache_MultiGet(t *testing.T) {
	c := newUsersGroupObjectPermissionsCache()
	defer c.Close()

	var testData []*model.UsersGroupObjectPermissions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.UsersGroupObjectPermissions))
	}

	err := c.ICache.(UsersGroupObjectPermissionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(UsersGroupObjectPermissionsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.UsersGroupObjectPermissions))
	}
}

func Test_usersGroupObjectPermissionsCache_MultiSet(t *testing.T) {
	c := newUsersGroupObjectPermissionsCache()
	defer c.Close()

	var testData []*model.UsersGroupObjectPermissions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.UsersGroupObjectPermissions))
	}

	err := c.ICache.(UsersGroupObjectPermissionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_usersGroupObjectPermissionsCache_Del(t *testing.T) {
	c := newUsersGroupObjectPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersGroupObjectPermissions)
	err := c.ICache.(UsersGroupObjectPermissionsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_usersGroupObjectPermissionsCache_SetCacheWithNotFound(t *testing.T) {
	c := newUsersGroupObjectPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersGroupObjectPermissions)
	err := c.ICache.(UsersGroupObjectPermissionsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(UsersGroupObjectPermissionsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewUsersGroupObjectPermissionsCache(t *testing.T) {
	c := NewUsersGroupObjectPermissionsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewUsersGroupObjectPermissionsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewUsersGroupObjectPermissionsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
