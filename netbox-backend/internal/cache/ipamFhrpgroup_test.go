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

func newIpamFhrpgroupCache() *gotest.Cache {
	record1 := &model.IpamFhrpgroup{}
	record1.ID = 1
	record2 := &model.IpamFhrpgroup{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewIpamFhrpgroupCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_ipamFhrpgroupCache_Set(t *testing.T) {
	c := newIpamFhrpgroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamFhrpgroup)
	err := c.ICache.(IpamFhrpgroupCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(IpamFhrpgroupCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_ipamFhrpgroupCache_Get(t *testing.T) {
	c := newIpamFhrpgroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamFhrpgroup)
	err := c.ICache.(IpamFhrpgroupCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamFhrpgroupCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(IpamFhrpgroupCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_ipamFhrpgroupCache_MultiGet(t *testing.T) {
	c := newIpamFhrpgroupCache()
	defer c.Close()

	var testData []*model.IpamFhrpgroup
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamFhrpgroup))
	}

	err := c.ICache.(IpamFhrpgroupCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamFhrpgroupCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.IpamFhrpgroup))
	}
}

func Test_ipamFhrpgroupCache_MultiSet(t *testing.T) {
	c := newIpamFhrpgroupCache()
	defer c.Close()

	var testData []*model.IpamFhrpgroup
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamFhrpgroup))
	}

	err := c.ICache.(IpamFhrpgroupCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamFhrpgroupCache_Del(t *testing.T) {
	c := newIpamFhrpgroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamFhrpgroup)
	err := c.ICache.(IpamFhrpgroupCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamFhrpgroupCache_SetCacheWithNotFound(t *testing.T) {
	c := newIpamFhrpgroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamFhrpgroup)
	err := c.ICache.(IpamFhrpgroupCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(IpamFhrpgroupCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewIpamFhrpgroupCache(t *testing.T) {
	c := NewIpamFhrpgroupCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewIpamFhrpgroupCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewIpamFhrpgroupCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
