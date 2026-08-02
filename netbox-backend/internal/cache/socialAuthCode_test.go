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

func newSocialAuthCodeCache() *gotest.Cache {
	record1 := &model.SocialAuthCode{}
	record1.ID = 1
	record2 := &model.SocialAuthCode{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewSocialAuthCodeCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_socialAuthCodeCache_Set(t *testing.T) {
	c := newSocialAuthCodeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthCode)
	err := c.ICache.(SocialAuthCodeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(SocialAuthCodeCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_socialAuthCodeCache_Get(t *testing.T) {
	c := newSocialAuthCodeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthCode)
	err := c.ICache.(SocialAuthCodeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SocialAuthCodeCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(SocialAuthCodeCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_socialAuthCodeCache_MultiGet(t *testing.T) {
	c := newSocialAuthCodeCache()
	defer c.Close()

	var testData []*model.SocialAuthCode
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SocialAuthCode))
	}

	err := c.ICache.(SocialAuthCodeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SocialAuthCodeCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.SocialAuthCode))
	}
}

func Test_socialAuthCodeCache_MultiSet(t *testing.T) {
	c := newSocialAuthCodeCache()
	defer c.Close()

	var testData []*model.SocialAuthCode
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SocialAuthCode))
	}

	err := c.ICache.(SocialAuthCodeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_socialAuthCodeCache_Del(t *testing.T) {
	c := newSocialAuthCodeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthCode)
	err := c.ICache.(SocialAuthCodeCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_socialAuthCodeCache_SetCacheWithNotFound(t *testing.T) {
	c := newSocialAuthCodeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthCode)
	err := c.ICache.(SocialAuthCodeCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(SocialAuthCodeCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewSocialAuthCodeCache(t *testing.T) {
	c := NewSocialAuthCodeCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewSocialAuthCodeCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewSocialAuthCodeCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
