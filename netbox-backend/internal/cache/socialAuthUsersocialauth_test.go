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

func newSocialAuthUsersocialauthCache() *gotest.Cache {
	record1 := &model.SocialAuthUsersocialauth{}
	record1.ID = 1
	record2 := &model.SocialAuthUsersocialauth{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewSocialAuthUsersocialauthCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_socialAuthUsersocialauthCache_Set(t *testing.T) {
	c := newSocialAuthUsersocialauthCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthUsersocialauth)
	err := c.ICache.(SocialAuthUsersocialauthCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(SocialAuthUsersocialauthCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_socialAuthUsersocialauthCache_Get(t *testing.T) {
	c := newSocialAuthUsersocialauthCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthUsersocialauth)
	err := c.ICache.(SocialAuthUsersocialauthCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SocialAuthUsersocialauthCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(SocialAuthUsersocialauthCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_socialAuthUsersocialauthCache_MultiGet(t *testing.T) {
	c := newSocialAuthUsersocialauthCache()
	defer c.Close()

	var testData []*model.SocialAuthUsersocialauth
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SocialAuthUsersocialauth))
	}

	err := c.ICache.(SocialAuthUsersocialauthCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SocialAuthUsersocialauthCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.SocialAuthUsersocialauth))
	}
}

func Test_socialAuthUsersocialauthCache_MultiSet(t *testing.T) {
	c := newSocialAuthUsersocialauthCache()
	defer c.Close()

	var testData []*model.SocialAuthUsersocialauth
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SocialAuthUsersocialauth))
	}

	err := c.ICache.(SocialAuthUsersocialauthCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_socialAuthUsersocialauthCache_Del(t *testing.T) {
	c := newSocialAuthUsersocialauthCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthUsersocialauth)
	err := c.ICache.(SocialAuthUsersocialauthCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_socialAuthUsersocialauthCache_SetCacheWithNotFound(t *testing.T) {
	c := newSocialAuthUsersocialauthCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthUsersocialauth)
	err := c.ICache.(SocialAuthUsersocialauthCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(SocialAuthUsersocialauthCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewSocialAuthUsersocialauthCache(t *testing.T) {
	c := NewSocialAuthUsersocialauthCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewSocialAuthUsersocialauthCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewSocialAuthUsersocialauthCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
