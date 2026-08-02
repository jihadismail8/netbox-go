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

func newIpamServiceIpaddressesCache() *gotest.Cache {
	record1 := &model.IpamServiceIpaddresses{}
	record1.ID = 1
	record2 := &model.IpamServiceIpaddresses{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewIpamServiceIpaddressesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_ipamServiceIpaddressesCache_Set(t *testing.T) {
	c := newIpamServiceIpaddressesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamServiceIpaddresses)
	err := c.ICache.(IpamServiceIpaddressesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(IpamServiceIpaddressesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_ipamServiceIpaddressesCache_Get(t *testing.T) {
	c := newIpamServiceIpaddressesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamServiceIpaddresses)
	err := c.ICache.(IpamServiceIpaddressesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamServiceIpaddressesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(IpamServiceIpaddressesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_ipamServiceIpaddressesCache_MultiGet(t *testing.T) {
	c := newIpamServiceIpaddressesCache()
	defer c.Close()

	var testData []*model.IpamServiceIpaddresses
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamServiceIpaddresses))
	}

	err := c.ICache.(IpamServiceIpaddressesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamServiceIpaddressesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.IpamServiceIpaddresses))
	}
}

func Test_ipamServiceIpaddressesCache_MultiSet(t *testing.T) {
	c := newIpamServiceIpaddressesCache()
	defer c.Close()

	var testData []*model.IpamServiceIpaddresses
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamServiceIpaddresses))
	}

	err := c.ICache.(IpamServiceIpaddressesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamServiceIpaddressesCache_Del(t *testing.T) {
	c := newIpamServiceIpaddressesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamServiceIpaddresses)
	err := c.ICache.(IpamServiceIpaddressesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamServiceIpaddressesCache_SetCacheWithNotFound(t *testing.T) {
	c := newIpamServiceIpaddressesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamServiceIpaddresses)
	err := c.ICache.(IpamServiceIpaddressesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(IpamServiceIpaddressesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewIpamServiceIpaddressesCache(t *testing.T) {
	c := NewIpamServiceIpaddressesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewIpamServiceIpaddressesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewIpamServiceIpaddressesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
