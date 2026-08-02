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

func newIpamVlanCache() *gotest.Cache {
	record1 := &model.IpamVlan{}
	record1.ID = 1
	record2 := &model.IpamVlan{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewIpamVlanCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_ipamVlanCache_Set(t *testing.T) {
	c := newIpamVlanCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVlan)
	err := c.ICache.(IpamVlanCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(IpamVlanCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_ipamVlanCache_Get(t *testing.T) {
	c := newIpamVlanCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVlan)
	err := c.ICache.(IpamVlanCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamVlanCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(IpamVlanCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_ipamVlanCache_MultiGet(t *testing.T) {
	c := newIpamVlanCache()
	defer c.Close()

	var testData []*model.IpamVlan
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamVlan))
	}

	err := c.ICache.(IpamVlanCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamVlanCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.IpamVlan))
	}
}

func Test_ipamVlanCache_MultiSet(t *testing.T) {
	c := newIpamVlanCache()
	defer c.Close()

	var testData []*model.IpamVlan
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamVlan))
	}

	err := c.ICache.(IpamVlanCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamVlanCache_Del(t *testing.T) {
	c := newIpamVlanCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVlan)
	err := c.ICache.(IpamVlanCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamVlanCache_SetCacheWithNotFound(t *testing.T) {
	c := newIpamVlanCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVlan)
	err := c.ICache.(IpamVlanCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(IpamVlanCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewIpamVlanCache(t *testing.T) {
	c := NewIpamVlanCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewIpamVlanCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewIpamVlanCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
