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

func newVirtualizationVminterfaceTaggedVlansCache() *gotest.Cache {
	record1 := &model.VirtualizationVminterfaceTaggedVlans{}
	record1.ID = 1
	record2 := &model.VirtualizationVminterfaceTaggedVlans{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVirtualizationVminterfaceTaggedVlansCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_virtualizationVminterfaceTaggedVlansCache_Set(t *testing.T) {
	c := newVirtualizationVminterfaceTaggedVlansCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVminterfaceTaggedVlans)
	err := c.ICache.(VirtualizationVminterfaceTaggedVlansCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VirtualizationVminterfaceTaggedVlansCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_virtualizationVminterfaceTaggedVlansCache_Get(t *testing.T) {
	c := newVirtualizationVminterfaceTaggedVlansCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVminterfaceTaggedVlans)
	err := c.ICache.(VirtualizationVminterfaceTaggedVlansCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VirtualizationVminterfaceTaggedVlansCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VirtualizationVminterfaceTaggedVlansCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_virtualizationVminterfaceTaggedVlansCache_MultiGet(t *testing.T) {
	c := newVirtualizationVminterfaceTaggedVlansCache()
	defer c.Close()

	var testData []*model.VirtualizationVminterfaceTaggedVlans
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VirtualizationVminterfaceTaggedVlans))
	}

	err := c.ICache.(VirtualizationVminterfaceTaggedVlansCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VirtualizationVminterfaceTaggedVlansCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VirtualizationVminterfaceTaggedVlans))
	}
}

func Test_virtualizationVminterfaceTaggedVlansCache_MultiSet(t *testing.T) {
	c := newVirtualizationVminterfaceTaggedVlansCache()
	defer c.Close()

	var testData []*model.VirtualizationVminterfaceTaggedVlans
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VirtualizationVminterfaceTaggedVlans))
	}

	err := c.ICache.(VirtualizationVminterfaceTaggedVlansCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_virtualizationVminterfaceTaggedVlansCache_Del(t *testing.T) {
	c := newVirtualizationVminterfaceTaggedVlansCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVminterfaceTaggedVlans)
	err := c.ICache.(VirtualizationVminterfaceTaggedVlansCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_virtualizationVminterfaceTaggedVlansCache_SetCacheWithNotFound(t *testing.T) {
	c := newVirtualizationVminterfaceTaggedVlansCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVminterfaceTaggedVlans)
	err := c.ICache.(VirtualizationVminterfaceTaggedVlansCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VirtualizationVminterfaceTaggedVlansCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVirtualizationVminterfaceTaggedVlansCache(t *testing.T) {
	c := NewVirtualizationVminterfaceTaggedVlansCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVirtualizationVminterfaceTaggedVlansCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVirtualizationVminterfaceTaggedVlansCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
