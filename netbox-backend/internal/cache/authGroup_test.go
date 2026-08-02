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

func newAuthGroupCache() *gotest.Cache {
	record1 := &model.AuthGroup{}
	record1.ID = 1
	record2 := &model.AuthGroup{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewAuthGroupCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_authGroupCache_Set(t *testing.T) {
	c := newAuthGroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.AuthGroup)
	err := c.ICache.(AuthGroupCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(AuthGroupCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_authGroupCache_Get(t *testing.T) {
	c := newAuthGroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.AuthGroup)
	err := c.ICache.(AuthGroupCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(AuthGroupCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(AuthGroupCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_authGroupCache_MultiGet(t *testing.T) {
	c := newAuthGroupCache()
	defer c.Close()

	var testData []*model.AuthGroup
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.AuthGroup))
	}

	err := c.ICache.(AuthGroupCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(AuthGroupCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.AuthGroup))
	}
}

func Test_authGroupCache_MultiSet(t *testing.T) {
	c := newAuthGroupCache()
	defer c.Close()

	var testData []*model.AuthGroup
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.AuthGroup))
	}

	err := c.ICache.(AuthGroupCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_authGroupCache_Del(t *testing.T) {
	c := newAuthGroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.AuthGroup)
	err := c.ICache.(AuthGroupCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_authGroupCache_SetCacheWithNotFound(t *testing.T) {
	c := newAuthGroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.AuthGroup)
	err := c.ICache.(AuthGroupCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(AuthGroupCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewAuthGroupCache(t *testing.T) {
	c := NewAuthGroupCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewAuthGroupCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewAuthGroupCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
