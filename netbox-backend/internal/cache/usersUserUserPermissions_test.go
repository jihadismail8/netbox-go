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

func newUsersUserUserPermissionsCache() *gotest.Cache {
	record1 := &model.UsersUserUserPermissions{}
	record1.ID = 1
	record2 := &model.UsersUserUserPermissions{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewUsersUserUserPermissionsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_usersUserUserPermissionsCache_Set(t *testing.T) {
	c := newUsersUserUserPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersUserUserPermissions)
	err := c.ICache.(UsersUserUserPermissionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(UsersUserUserPermissionsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_usersUserUserPermissionsCache_Get(t *testing.T) {
	c := newUsersUserUserPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersUserUserPermissions)
	err := c.ICache.(UsersUserUserPermissionsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(UsersUserUserPermissionsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(UsersUserUserPermissionsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_usersUserUserPermissionsCache_MultiGet(t *testing.T) {
	c := newUsersUserUserPermissionsCache()
	defer c.Close()

	var testData []*model.UsersUserUserPermissions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.UsersUserUserPermissions))
	}

	err := c.ICache.(UsersUserUserPermissionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(UsersUserUserPermissionsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.UsersUserUserPermissions))
	}
}

func Test_usersUserUserPermissionsCache_MultiSet(t *testing.T) {
	c := newUsersUserUserPermissionsCache()
	defer c.Close()

	var testData []*model.UsersUserUserPermissions
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.UsersUserUserPermissions))
	}

	err := c.ICache.(UsersUserUserPermissionsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_usersUserUserPermissionsCache_Del(t *testing.T) {
	c := newUsersUserUserPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersUserUserPermissions)
	err := c.ICache.(UsersUserUserPermissionsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_usersUserUserPermissionsCache_SetCacheWithNotFound(t *testing.T) {
	c := newUsersUserUserPermissionsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersUserUserPermissions)
	err := c.ICache.(UsersUserUserPermissionsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(UsersUserUserPermissionsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewUsersUserUserPermissionsCache(t *testing.T) {
	c := NewUsersUserUserPermissionsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewUsersUserUserPermissionsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewUsersUserUserPermissionsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
