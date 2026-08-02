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

func newUsersGroupCache() *gotest.Cache {
	record1 := &model.UsersGroup{}
	record1.ID = 1
	record2 := &model.UsersGroup{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewUsersGroupCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_usersGroupCache_Set(t *testing.T) {
	c := newUsersGroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersGroup)
	err := c.ICache.(UsersGroupCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(UsersGroupCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_usersGroupCache_Get(t *testing.T) {
	c := newUsersGroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersGroup)
	err := c.ICache.(UsersGroupCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(UsersGroupCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(UsersGroupCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_usersGroupCache_MultiGet(t *testing.T) {
	c := newUsersGroupCache()
	defer c.Close()

	var testData []*model.UsersGroup
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.UsersGroup))
	}

	err := c.ICache.(UsersGroupCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(UsersGroupCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.UsersGroup))
	}
}

func Test_usersGroupCache_MultiSet(t *testing.T) {
	c := newUsersGroupCache()
	defer c.Close()

	var testData []*model.UsersGroup
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.UsersGroup))
	}

	err := c.ICache.(UsersGroupCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_usersGroupCache_Del(t *testing.T) {
	c := newUsersGroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersGroup)
	err := c.ICache.(UsersGroupCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_usersGroupCache_SetCacheWithNotFound(t *testing.T) {
	c := newUsersGroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersGroup)
	err := c.ICache.(UsersGroupCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(UsersGroupCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewUsersGroupCache(t *testing.T) {
	c := NewUsersGroupCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewUsersGroupCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewUsersGroupCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
