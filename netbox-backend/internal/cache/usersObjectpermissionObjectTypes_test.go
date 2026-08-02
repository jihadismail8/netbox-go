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

func newUsersObjectpermissionObjectTypesCache() *gotest.Cache {
	record1 := &model.UsersObjectpermissionObjectTypes{}
	record1.ID = 1
	record2 := &model.UsersObjectpermissionObjectTypes{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewUsersObjectpermissionObjectTypesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_usersObjectpermissionObjectTypesCache_Set(t *testing.T) {
	c := newUsersObjectpermissionObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersObjectpermissionObjectTypes)
	err := c.ICache.(UsersObjectpermissionObjectTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(UsersObjectpermissionObjectTypesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_usersObjectpermissionObjectTypesCache_Get(t *testing.T) {
	c := newUsersObjectpermissionObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersObjectpermissionObjectTypes)
	err := c.ICache.(UsersObjectpermissionObjectTypesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(UsersObjectpermissionObjectTypesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(UsersObjectpermissionObjectTypesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_usersObjectpermissionObjectTypesCache_MultiGet(t *testing.T) {
	c := newUsersObjectpermissionObjectTypesCache()
	defer c.Close()

	var testData []*model.UsersObjectpermissionObjectTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.UsersObjectpermissionObjectTypes))
	}

	err := c.ICache.(UsersObjectpermissionObjectTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(UsersObjectpermissionObjectTypesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.UsersObjectpermissionObjectTypes))
	}
}

func Test_usersObjectpermissionObjectTypesCache_MultiSet(t *testing.T) {
	c := newUsersObjectpermissionObjectTypesCache()
	defer c.Close()

	var testData []*model.UsersObjectpermissionObjectTypes
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.UsersObjectpermissionObjectTypes))
	}

	err := c.ICache.(UsersObjectpermissionObjectTypesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_usersObjectpermissionObjectTypesCache_Del(t *testing.T) {
	c := newUsersObjectpermissionObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersObjectpermissionObjectTypes)
	err := c.ICache.(UsersObjectpermissionObjectTypesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_usersObjectpermissionObjectTypesCache_SetCacheWithNotFound(t *testing.T) {
	c := newUsersObjectpermissionObjectTypesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersObjectpermissionObjectTypes)
	err := c.ICache.(UsersObjectpermissionObjectTypesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(UsersObjectpermissionObjectTypesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewUsersObjectpermissionObjectTypesCache(t *testing.T) {
	c := NewUsersObjectpermissionObjectTypesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewUsersObjectpermissionObjectTypesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewUsersObjectpermissionObjectTypesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
