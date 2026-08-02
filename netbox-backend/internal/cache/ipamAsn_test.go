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

func newIpamAsnCache() *gotest.Cache {
	record1 := &model.IpamAsn{}
	record1.ID = 1
	record2 := &model.IpamAsn{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewIpamAsnCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_ipamAsnCache_Set(t *testing.T) {
	c := newIpamAsnCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamAsn)
	err := c.ICache.(IpamAsnCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(IpamAsnCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_ipamAsnCache_Get(t *testing.T) {
	c := newIpamAsnCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamAsn)
	err := c.ICache.(IpamAsnCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamAsnCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(IpamAsnCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_ipamAsnCache_MultiGet(t *testing.T) {
	c := newIpamAsnCache()
	defer c.Close()

	var testData []*model.IpamAsn
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamAsn))
	}

	err := c.ICache.(IpamAsnCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamAsnCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.IpamAsn))
	}
}

func Test_ipamAsnCache_MultiSet(t *testing.T) {
	c := newIpamAsnCache()
	defer c.Close()

	var testData []*model.IpamAsn
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamAsn))
	}

	err := c.ICache.(IpamAsnCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamAsnCache_Del(t *testing.T) {
	c := newIpamAsnCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamAsn)
	err := c.ICache.(IpamAsnCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamAsnCache_SetCacheWithNotFound(t *testing.T) {
	c := newIpamAsnCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamAsn)
	err := c.ICache.(IpamAsnCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(IpamAsnCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewIpamAsnCache(t *testing.T) {
	c := NewIpamAsnCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewIpamAsnCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewIpamAsnCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
