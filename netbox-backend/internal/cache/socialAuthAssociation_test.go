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

func newSocialAuthAssociationCache() *gotest.Cache {
	record1 := &model.SocialAuthAssociation{}
	record1.ID = 1
	record2 := &model.SocialAuthAssociation{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewSocialAuthAssociationCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_socialAuthAssociationCache_Set(t *testing.T) {
	c := newSocialAuthAssociationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthAssociation)
	err := c.ICache.(SocialAuthAssociationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(SocialAuthAssociationCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_socialAuthAssociationCache_Get(t *testing.T) {
	c := newSocialAuthAssociationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthAssociation)
	err := c.ICache.(SocialAuthAssociationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SocialAuthAssociationCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(SocialAuthAssociationCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_socialAuthAssociationCache_MultiGet(t *testing.T) {
	c := newSocialAuthAssociationCache()
	defer c.Close()

	var testData []*model.SocialAuthAssociation
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SocialAuthAssociation))
	}

	err := c.ICache.(SocialAuthAssociationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SocialAuthAssociationCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.SocialAuthAssociation))
	}
}

func Test_socialAuthAssociationCache_MultiSet(t *testing.T) {
	c := newSocialAuthAssociationCache()
	defer c.Close()

	var testData []*model.SocialAuthAssociation
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SocialAuthAssociation))
	}

	err := c.ICache.(SocialAuthAssociationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_socialAuthAssociationCache_Del(t *testing.T) {
	c := newSocialAuthAssociationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthAssociation)
	err := c.ICache.(SocialAuthAssociationCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_socialAuthAssociationCache_SetCacheWithNotFound(t *testing.T) {
	c := newSocialAuthAssociationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SocialAuthAssociation)
	err := c.ICache.(SocialAuthAssociationCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(SocialAuthAssociationCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewSocialAuthAssociationCache(t *testing.T) {
	c := NewSocialAuthAssociationCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewSocialAuthAssociationCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewSocialAuthAssociationCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
