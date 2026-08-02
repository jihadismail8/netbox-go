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

func newIpamServicetemplateCache() *gotest.Cache {
	record1 := &model.IpamServicetemplate{}
	record1.ID = 1
	record2 := &model.IpamServicetemplate{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewIpamServicetemplateCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_ipamServicetemplateCache_Set(t *testing.T) {
	c := newIpamServicetemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamServicetemplate)
	err := c.ICache.(IpamServicetemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(IpamServicetemplateCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_ipamServicetemplateCache_Get(t *testing.T) {
	c := newIpamServicetemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamServicetemplate)
	err := c.ICache.(IpamServicetemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamServicetemplateCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(IpamServicetemplateCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_ipamServicetemplateCache_MultiGet(t *testing.T) {
	c := newIpamServicetemplateCache()
	defer c.Close()

	var testData []*model.IpamServicetemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamServicetemplate))
	}

	err := c.ICache.(IpamServicetemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamServicetemplateCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.IpamServicetemplate))
	}
}

func Test_ipamServicetemplateCache_MultiSet(t *testing.T) {
	c := newIpamServicetemplateCache()
	defer c.Close()

	var testData []*model.IpamServicetemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamServicetemplate))
	}

	err := c.ICache.(IpamServicetemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamServicetemplateCache_Del(t *testing.T) {
	c := newIpamServicetemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamServicetemplate)
	err := c.ICache.(IpamServicetemplateCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamServicetemplateCache_SetCacheWithNotFound(t *testing.T) {
	c := newIpamServicetemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamServicetemplate)
	err := c.ICache.(IpamServicetemplateCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(IpamServicetemplateCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewIpamServicetemplateCache(t *testing.T) {
	c := NewIpamServicetemplateCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewIpamServicetemplateCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewIpamServicetemplateCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
