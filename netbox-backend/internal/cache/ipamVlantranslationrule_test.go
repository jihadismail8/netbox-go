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

func newIpamVlantranslationruleCache() *gotest.Cache {
	record1 := &model.IpamVlantranslationrule{}
	record1.ID = 1
	record2 := &model.IpamVlantranslationrule{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewIpamVlantranslationruleCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_ipamVlantranslationruleCache_Set(t *testing.T) {
	c := newIpamVlantranslationruleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVlantranslationrule)
	err := c.ICache.(IpamVlantranslationruleCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(IpamVlantranslationruleCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_ipamVlantranslationruleCache_Get(t *testing.T) {
	c := newIpamVlantranslationruleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVlantranslationrule)
	err := c.ICache.(IpamVlantranslationruleCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamVlantranslationruleCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(IpamVlantranslationruleCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_ipamVlantranslationruleCache_MultiGet(t *testing.T) {
	c := newIpamVlantranslationruleCache()
	defer c.Close()

	var testData []*model.IpamVlantranslationrule
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamVlantranslationrule))
	}

	err := c.ICache.(IpamVlantranslationruleCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamVlantranslationruleCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.IpamVlantranslationrule))
	}
}

func Test_ipamVlantranslationruleCache_MultiSet(t *testing.T) {
	c := newIpamVlantranslationruleCache()
	defer c.Close()

	var testData []*model.IpamVlantranslationrule
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamVlantranslationrule))
	}

	err := c.ICache.(IpamVlantranslationruleCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamVlantranslationruleCache_Del(t *testing.T) {
	c := newIpamVlantranslationruleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVlantranslationrule)
	err := c.ICache.(IpamVlantranslationruleCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamVlantranslationruleCache_SetCacheWithNotFound(t *testing.T) {
	c := newIpamVlantranslationruleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVlantranslationrule)
	err := c.ICache.(IpamVlantranslationruleCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(IpamVlantranslationruleCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewIpamVlantranslationruleCache(t *testing.T) {
	c := NewIpamVlantranslationruleCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewIpamVlantranslationruleCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewIpamVlantranslationruleCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
