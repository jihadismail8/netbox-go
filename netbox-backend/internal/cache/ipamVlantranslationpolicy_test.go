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

func newIpamVlantranslationpolicyCache() *gotest.Cache {
	record1 := &model.IpamVlantranslationpolicy{}
	record1.ID = 1
	record2 := &model.IpamVlantranslationpolicy{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewIpamVlantranslationpolicyCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_ipamVlantranslationpolicyCache_Set(t *testing.T) {
	c := newIpamVlantranslationpolicyCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVlantranslationpolicy)
	err := c.ICache.(IpamVlantranslationpolicyCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(IpamVlantranslationpolicyCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_ipamVlantranslationpolicyCache_Get(t *testing.T) {
	c := newIpamVlantranslationpolicyCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVlantranslationpolicy)
	err := c.ICache.(IpamVlantranslationpolicyCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamVlantranslationpolicyCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(IpamVlantranslationpolicyCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_ipamVlantranslationpolicyCache_MultiGet(t *testing.T) {
	c := newIpamVlantranslationpolicyCache()
	defer c.Close()

	var testData []*model.IpamVlantranslationpolicy
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamVlantranslationpolicy))
	}

	err := c.ICache.(IpamVlantranslationpolicyCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamVlantranslationpolicyCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.IpamVlantranslationpolicy))
	}
}

func Test_ipamVlantranslationpolicyCache_MultiSet(t *testing.T) {
	c := newIpamVlantranslationpolicyCache()
	defer c.Close()

	var testData []*model.IpamVlantranslationpolicy
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamVlantranslationpolicy))
	}

	err := c.ICache.(IpamVlantranslationpolicyCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamVlantranslationpolicyCache_Del(t *testing.T) {
	c := newIpamVlantranslationpolicyCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVlantranslationpolicy)
	err := c.ICache.(IpamVlantranslationpolicyCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamVlantranslationpolicyCache_SetCacheWithNotFound(t *testing.T) {
	c := newIpamVlantranslationpolicyCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVlantranslationpolicy)
	err := c.ICache.(IpamVlantranslationpolicyCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(IpamVlantranslationpolicyCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewIpamVlantranslationpolicyCache(t *testing.T) {
	c := NewIpamVlantranslationpolicyCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewIpamVlantranslationpolicyCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewIpamVlantranslationpolicyCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
