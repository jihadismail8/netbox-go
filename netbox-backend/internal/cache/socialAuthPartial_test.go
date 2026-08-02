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

func newSocialAuthPartialCache() *gotest.Cache {
	record1 := &model.SocialAuthPartial{}
	record1.ID = 1
	record2 := &model.SocialAuthPartial{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewSocialAuthPartialCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_socialAuthPartialCache_Set(t *testing.T) {
	c := newSocialAuthPartialCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthPartial)
	err := c.ICache.(SocialAuthPartialCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(SocialAuthPartialCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_socialAuthPartialCache_Get(t *testing.T) {
	c := newSocialAuthPartialCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthPartial)
	err := c.ICache.(SocialAuthPartialCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SocialAuthPartialCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(SocialAuthPartialCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_socialAuthPartialCache_MultiGet(t *testing.T) {
	c := newSocialAuthPartialCache()
	defer c.Close()

	var testData []*model.SocialAuthPartial
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SocialAuthPartial))
	}

	err := c.ICache.(SocialAuthPartialCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SocialAuthPartialCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.SocialAuthPartial))
	}
}

func Test_socialAuthPartialCache_MultiSet(t *testing.T) {
	c := newSocialAuthPartialCache()
	defer c.Close()

	var testData []*model.SocialAuthPartial
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SocialAuthPartial))
	}

	err := c.ICache.(SocialAuthPartialCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_socialAuthPartialCache_Del(t *testing.T) {
	c := newSocialAuthPartialCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthPartial)
	err := c.ICache.(SocialAuthPartialCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_socialAuthPartialCache_SetCacheWithNotFound(t *testing.T) {
	c := newSocialAuthPartialCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthPartial)
	err := c.ICache.(SocialAuthPartialCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(SocialAuthPartialCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewSocialAuthPartialCache(t *testing.T) {
	c := NewSocialAuthPartialCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewSocialAuthPartialCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewSocialAuthPartialCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
