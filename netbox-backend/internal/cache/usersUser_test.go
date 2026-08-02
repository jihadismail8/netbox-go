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

func newUsersUserCache() *gotest.Cache {
	record1 := &model.UsersUser{}
	record1.ID = 1
	record2 := &model.UsersUser{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewUsersUserCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_usersUserCache_Set(t *testing.T) {
	c := newUsersUserCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersUser)
	err := c.ICache.(UsersUserCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(UsersUserCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_usersUserCache_Get(t *testing.T) {
	c := newUsersUserCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersUser)
	err := c.ICache.(UsersUserCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(UsersUserCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(UsersUserCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_usersUserCache_MultiGet(t *testing.T) {
	c := newUsersUserCache()
	defer c.Close()

	var testData []*model.UsersUser
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.UsersUser))
	}

	err := c.ICache.(UsersUserCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(UsersUserCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.UsersUser))
	}
}

func Test_usersUserCache_MultiSet(t *testing.T) {
	c := newUsersUserCache()
	defer c.Close()

	var testData []*model.UsersUser
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.UsersUser))
	}

	err := c.ICache.(UsersUserCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_usersUserCache_Del(t *testing.T) {
	c := newUsersUserCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersUser)
	err := c.ICache.(UsersUserCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_usersUserCache_SetCacheWithNotFound(t *testing.T) {
	c := newUsersUserCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.UsersUser)
	err := c.ICache.(UsersUserCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(UsersUserCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewUsersUserCache(t *testing.T) {
	c := NewUsersUserCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewUsersUserCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewUsersUserCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
