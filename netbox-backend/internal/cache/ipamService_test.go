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

func newIpamServiceCache() *gotest.Cache {
	record1 := &model.IpamService{}
	record1.ID = 1
	record2 := &model.IpamService{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewIpamServiceCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_ipamServiceCache_Set(t *testing.T) {
	c := newIpamServiceCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamService)
	err := c.ICache.(IpamServiceCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(IpamServiceCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_ipamServiceCache_Get(t *testing.T) {
	c := newIpamServiceCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamService)
	err := c.ICache.(IpamServiceCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamServiceCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(IpamServiceCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_ipamServiceCache_MultiGet(t *testing.T) {
	c := newIpamServiceCache()
	defer c.Close()

	var testData []*model.IpamService
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamService))
	}

	err := c.ICache.(IpamServiceCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamServiceCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.IpamService))
	}
}

func Test_ipamServiceCache_MultiSet(t *testing.T) {
	c := newIpamServiceCache()
	defer c.Close()

	var testData []*model.IpamService
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamService))
	}

	err := c.ICache.(IpamServiceCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamServiceCache_Del(t *testing.T) {
	c := newIpamServiceCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamService)
	err := c.ICache.(IpamServiceCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamServiceCache_SetCacheWithNotFound(t *testing.T) {
	c := newIpamServiceCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamService)
	err := c.ICache.(IpamServiceCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(IpamServiceCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewIpamServiceCache(t *testing.T) {
	c := NewIpamServiceCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewIpamServiceCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewIpamServiceCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
