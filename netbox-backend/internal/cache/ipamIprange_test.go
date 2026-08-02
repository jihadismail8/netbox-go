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

func newIpamIprangeCache() *gotest.Cache {
	record1 := &model.IpamIprange{}
	record1.ID = 1
	record2 := &model.IpamIprange{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewIpamIprangeCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_ipamIprangeCache_Set(t *testing.T) {
	c := newIpamIprangeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamIprange)
	err := c.ICache.(IpamIprangeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(IpamIprangeCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_ipamIprangeCache_Get(t *testing.T) {
	c := newIpamIprangeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamIprange)
	err := c.ICache.(IpamIprangeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamIprangeCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(IpamIprangeCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_ipamIprangeCache_MultiGet(t *testing.T) {
	c := newIpamIprangeCache()
	defer c.Close()

	var testData []*model.IpamIprange
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamIprange))
	}

	err := c.ICache.(IpamIprangeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamIprangeCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.IpamIprange))
	}
}

func Test_ipamIprangeCache_MultiSet(t *testing.T) {
	c := newIpamIprangeCache()
	defer c.Close()

	var testData []*model.IpamIprange
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamIprange))
	}

	err := c.ICache.(IpamIprangeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamIprangeCache_Del(t *testing.T) {
	c := newIpamIprangeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamIprange)
	err := c.ICache.(IpamIprangeCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamIprangeCache_SetCacheWithNotFound(t *testing.T) {
	c := newIpamIprangeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamIprange)
	err := c.ICache.(IpamIprangeCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(IpamIprangeCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewIpamIprangeCache(t *testing.T) {
	c := NewIpamIprangeCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewIpamIprangeCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewIpamIprangeCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
