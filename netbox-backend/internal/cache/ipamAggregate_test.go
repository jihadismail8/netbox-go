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

func newIpamAggregateCache() *gotest.Cache {
	record1 := &model.IpamAggregate{}
	record1.ID = 1
	record2 := &model.IpamAggregate{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewIpamAggregateCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_ipamAggregateCache_Set(t *testing.T) {
	c := newIpamAggregateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamAggregate)
	err := c.ICache.(IpamAggregateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(IpamAggregateCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_ipamAggregateCache_Get(t *testing.T) {
	c := newIpamAggregateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamAggregate)
	err := c.ICache.(IpamAggregateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamAggregateCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(IpamAggregateCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_ipamAggregateCache_MultiGet(t *testing.T) {
	c := newIpamAggregateCache()
	defer c.Close()

	var testData []*model.IpamAggregate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamAggregate))
	}

	err := c.ICache.(IpamAggregateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamAggregateCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.IpamAggregate))
	}
}

func Test_ipamAggregateCache_MultiSet(t *testing.T) {
	c := newIpamAggregateCache()
	defer c.Close()

	var testData []*model.IpamAggregate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamAggregate))
	}

	err := c.ICache.(IpamAggregateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamAggregateCache_Del(t *testing.T) {
	c := newIpamAggregateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamAggregate)
	err := c.ICache.(IpamAggregateCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamAggregateCache_SetCacheWithNotFound(t *testing.T) {
	c := newIpamAggregateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamAggregate)
	err := c.ICache.(IpamAggregateCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(IpamAggregateCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewIpamAggregateCache(t *testing.T) {
	c := NewIpamAggregateCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewIpamAggregateCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewIpamAggregateCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
