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

func newIpamVrfExportTargetsCache() *gotest.Cache {
	record1 := &model.IpamVrfExportTargets{}
	record1.ID = 1
	record2 := &model.IpamVrfExportTargets{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewIpamVrfExportTargetsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_ipamVrfExportTargetsCache_Set(t *testing.T) {
	c := newIpamVrfExportTargetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVrfExportTargets)
	err := c.ICache.(IpamVrfExportTargetsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(IpamVrfExportTargetsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_ipamVrfExportTargetsCache_Get(t *testing.T) {
	c := newIpamVrfExportTargetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVrfExportTargets)
	err := c.ICache.(IpamVrfExportTargetsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamVrfExportTargetsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(IpamVrfExportTargetsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_ipamVrfExportTargetsCache_MultiGet(t *testing.T) {
	c := newIpamVrfExportTargetsCache()
	defer c.Close()

	var testData []*model.IpamVrfExportTargets
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamVrfExportTargets))
	}

	err := c.ICache.(IpamVrfExportTargetsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamVrfExportTargetsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.IpamVrfExportTargets))
	}
}

func Test_ipamVrfExportTargetsCache_MultiSet(t *testing.T) {
	c := newIpamVrfExportTargetsCache()
	defer c.Close()

	var testData []*model.IpamVrfExportTargets
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamVrfExportTargets))
	}

	err := c.ICache.(IpamVrfExportTargetsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamVrfExportTargetsCache_Del(t *testing.T) {
	c := newIpamVrfExportTargetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVrfExportTargets)
	err := c.ICache.(IpamVrfExportTargetsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamVrfExportTargetsCache_SetCacheWithNotFound(t *testing.T) {
	c := newIpamVrfExportTargetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVrfExportTargets)
	err := c.ICache.(IpamVrfExportTargetsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(IpamVrfExportTargetsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewIpamVrfExportTargetsCache(t *testing.T) {
	c := NewIpamVrfExportTargetsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewIpamVrfExportTargetsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewIpamVrfExportTargetsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
