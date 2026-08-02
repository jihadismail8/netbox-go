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

func newIpamRoutetargetCache() *gotest.Cache {
	record1 := &model.IpamRoutetarget{}
	record1.ID = 1
	record2 := &model.IpamRoutetarget{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewIpamRoutetargetCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_ipamRoutetargetCache_Set(t *testing.T) {
	c := newIpamRoutetargetCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamRoutetarget)
	err := c.ICache.(IpamRoutetargetCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(IpamRoutetargetCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_ipamRoutetargetCache_Get(t *testing.T) {
	c := newIpamRoutetargetCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamRoutetarget)
	err := c.ICache.(IpamRoutetargetCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamRoutetargetCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(IpamRoutetargetCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_ipamRoutetargetCache_MultiGet(t *testing.T) {
	c := newIpamRoutetargetCache()
	defer c.Close()

	var testData []*model.IpamRoutetarget
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamRoutetarget))
	}

	err := c.ICache.(IpamRoutetargetCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamRoutetargetCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.IpamRoutetarget))
	}
}

func Test_ipamRoutetargetCache_MultiSet(t *testing.T) {
	c := newIpamRoutetargetCache()
	defer c.Close()

	var testData []*model.IpamRoutetarget
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamRoutetarget))
	}

	err := c.ICache.(IpamRoutetargetCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamRoutetargetCache_Del(t *testing.T) {
	c := newIpamRoutetargetCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamRoutetarget)
	err := c.ICache.(IpamRoutetargetCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamRoutetargetCache_SetCacheWithNotFound(t *testing.T) {
	c := newIpamRoutetargetCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamRoutetarget)
	err := c.ICache.(IpamRoutetargetCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(IpamRoutetargetCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewIpamRoutetargetCache(t *testing.T) {
	c := NewIpamRoutetargetCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewIpamRoutetargetCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewIpamRoutetargetCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
