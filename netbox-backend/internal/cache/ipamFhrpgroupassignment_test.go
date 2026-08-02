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

func newIpamFhrpgroupassignmentCache() *gotest.Cache {
	record1 := &model.IpamFhrpgroupassignment{}
	record1.ID = 1
	record2 := &model.IpamFhrpgroupassignment{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewIpamFhrpgroupassignmentCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_ipamFhrpgroupassignmentCache_Set(t *testing.T) {
	c := newIpamFhrpgroupassignmentCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamFhrpgroupassignment)
	err := c.ICache.(IpamFhrpgroupassignmentCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(IpamFhrpgroupassignmentCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_ipamFhrpgroupassignmentCache_Get(t *testing.T) {
	c := newIpamFhrpgroupassignmentCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamFhrpgroupassignment)
	err := c.ICache.(IpamFhrpgroupassignmentCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamFhrpgroupassignmentCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(IpamFhrpgroupassignmentCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_ipamFhrpgroupassignmentCache_MultiGet(t *testing.T) {
	c := newIpamFhrpgroupassignmentCache()
	defer c.Close()

	var testData []*model.IpamFhrpgroupassignment
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamFhrpgroupassignment))
	}

	err := c.ICache.(IpamFhrpgroupassignmentCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamFhrpgroupassignmentCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.IpamFhrpgroupassignment))
	}
}

func Test_ipamFhrpgroupassignmentCache_MultiSet(t *testing.T) {
	c := newIpamFhrpgroupassignmentCache()
	defer c.Close()

	var testData []*model.IpamFhrpgroupassignment
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamFhrpgroupassignment))
	}

	err := c.ICache.(IpamFhrpgroupassignmentCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamFhrpgroupassignmentCache_Del(t *testing.T) {
	c := newIpamFhrpgroupassignmentCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamFhrpgroupassignment)
	err := c.ICache.(IpamFhrpgroupassignmentCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamFhrpgroupassignmentCache_SetCacheWithNotFound(t *testing.T) {
	c := newIpamFhrpgroupassignmentCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamFhrpgroupassignment)
	err := c.ICache.(IpamFhrpgroupassignmentCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(IpamFhrpgroupassignmentCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewIpamFhrpgroupassignmentCache(t *testing.T) {
	c := NewIpamFhrpgroupassignmentCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewIpamFhrpgroupassignmentCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewIpamFhrpgroupassignmentCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
