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

func newIpamVrfImportTargetsCache() *gotest.Cache {
	record1 := &model.IpamVrfImportTargets{}
	record1.ID = 1
	record2 := &model.IpamVrfImportTargets{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewIpamVrfImportTargetsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_ipamVrfImportTargetsCache_Set(t *testing.T) {
	c := newIpamVrfImportTargetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVrfImportTargets)
	err := c.ICache.(IpamVrfImportTargetsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(IpamVrfImportTargetsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_ipamVrfImportTargetsCache_Get(t *testing.T) {
	c := newIpamVrfImportTargetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVrfImportTargets)
	err := c.ICache.(IpamVrfImportTargetsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamVrfImportTargetsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(IpamVrfImportTargetsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_ipamVrfImportTargetsCache_MultiGet(t *testing.T) {
	c := newIpamVrfImportTargetsCache()
	defer c.Close()

	var testData []*model.IpamVrfImportTargets
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamVrfImportTargets))
	}

	err := c.ICache.(IpamVrfImportTargetsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(IpamVrfImportTargetsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.IpamVrfImportTargets))
	}
}

func Test_ipamVrfImportTargetsCache_MultiSet(t *testing.T) {
	c := newIpamVrfImportTargetsCache()
	defer c.Close()

	var testData []*model.IpamVrfImportTargets
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.IpamVrfImportTargets))
	}

	err := c.ICache.(IpamVrfImportTargetsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamVrfImportTargetsCache_Del(t *testing.T) {
	c := newIpamVrfImportTargetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVrfImportTargets)
	err := c.ICache.(IpamVrfImportTargetsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ipamVrfImportTargetsCache_SetCacheWithNotFound(t *testing.T) {
	c := newIpamVrfImportTargetsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.IpamVrfImportTargets)
	err := c.ICache.(IpamVrfImportTargetsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(IpamVrfImportTargetsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewIpamVrfImportTargetsCache(t *testing.T) {
	c := NewIpamVrfImportTargetsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewIpamVrfImportTargetsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewIpamVrfImportTargetsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
