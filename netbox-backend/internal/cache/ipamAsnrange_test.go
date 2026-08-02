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

func newIpamAsnrangeCache() *gotest.Cache {
	record1 := &model.IpamAsnrange{}
	record1.ID = 1
	record2 := &model.IpamAsnrange{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewIpamAsnrangeCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_ipamAsnrangeCache_Set(t *testing.T) {
	c := newIpamAsnrangeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamAsnrange)
	err := c.ICache.(IpamAsnrangeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(IpamAsnrangeCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_ipamAsnrangeCache_Get(t *testing.T) {
	c := newIpamAsnrangeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamAsnrange)
	err := c.ICache.(IpamAsnrangeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamAsnrangeCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(IpamAsnrangeCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_ipamAsnrangeCache_MultiGet(t *testing.T) {
	c := newIpamAsnrangeCache()
	defer c.Close()

	var testData []*model.IpamAsnrange
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamAsnrange))
	}

	err := c.ICache.(IpamAsnrangeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamAsnrangeCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.IpamAsnrange))
	}
}

func Test_ipamAsnrangeCache_MultiSet(t *testing.T) {
	c := newIpamAsnrangeCache()
	defer c.Close()

	var testData []*model.IpamAsnrange
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamAsnrange))
	}

	err := c.ICache.(IpamAsnrangeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamAsnrangeCache_Del(t *testing.T) {
	c := newIpamAsnrangeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamAsnrange)
	err := c.ICache.(IpamAsnrangeCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamAsnrangeCache_SetCacheWithNotFound(t *testing.T) {
	c := newIpamAsnrangeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamAsnrange)
	err := c.ICache.(IpamAsnrangeCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(IpamAsnrangeCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewIpamAsnrangeCache(t *testing.T) {
	c := NewIpamAsnrangeCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewIpamAsnrangeCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewIpamAsnrangeCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
