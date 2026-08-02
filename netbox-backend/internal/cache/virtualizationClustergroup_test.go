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

func newVirtualizationClustergroupCache() *gotest.Cache {
	record1 := &model.VirtualizationClustergroup{}
	record1.ID = 1
	record2 := &model.VirtualizationClustergroup{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVirtualizationClustergroupCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_virtualizationClustergroupCache_Set(t *testing.T) {
	c := newVirtualizationClustergroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationClustergroup)
	err := c.ICache.(VirtualizationClustergroupCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VirtualizationClustergroupCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_virtualizationClustergroupCache_Get(t *testing.T) {
	c := newVirtualizationClustergroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationClustergroup)
	err := c.ICache.(VirtualizationClustergroupCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VirtualizationClustergroupCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VirtualizationClustergroupCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_virtualizationClustergroupCache_MultiGet(t *testing.T) {
	c := newVirtualizationClustergroupCache()
	defer c.Close()

	var testData []*model.VirtualizationClustergroup
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VirtualizationClustergroup))
	}

	err := c.ICache.(VirtualizationClustergroupCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VirtualizationClustergroupCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VirtualizationClustergroup))
	}
}

func Test_virtualizationClustergroupCache_MultiSet(t *testing.T) {
	c := newVirtualizationClustergroupCache()
	defer c.Close()

	var testData []*model.VirtualizationClustergroup
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VirtualizationClustergroup))
	}

	err := c.ICache.(VirtualizationClustergroupCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_virtualizationClustergroupCache_Del(t *testing.T) {
	c := newVirtualizationClustergroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationClustergroup)
	err := c.ICache.(VirtualizationClustergroupCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_virtualizationClustergroupCache_SetCacheWithNotFound(t *testing.T) {
	c := newVirtualizationClustergroupCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationClustergroup)
	err := c.ICache.(VirtualizationClustergroupCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VirtualizationClustergroupCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVirtualizationClustergroupCache(t *testing.T) {
	c := NewVirtualizationClustergroupCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVirtualizationClustergroupCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVirtualizationClustergroupCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
