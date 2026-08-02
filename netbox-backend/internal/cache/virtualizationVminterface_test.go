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

func newVirtualizationVminterfaceCache() *gotest.Cache {
	record1 := &model.VirtualizationVminterface{}
	record1.ID = 1
	record2 := &model.VirtualizationVminterface{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVirtualizationVminterfaceCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_virtualizationVminterfaceCache_Set(t *testing.T) {
	c := newVirtualizationVminterfaceCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVminterface)
	err := c.ICache.(VirtualizationVminterfaceCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VirtualizationVminterfaceCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_virtualizationVminterfaceCache_Get(t *testing.T) {
	c := newVirtualizationVminterfaceCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVminterface)
	err := c.ICache.(VirtualizationVminterfaceCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VirtualizationVminterfaceCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VirtualizationVminterfaceCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_virtualizationVminterfaceCache_MultiGet(t *testing.T) {
	c := newVirtualizationVminterfaceCache()
	defer c.Close()

	var testData []*model.VirtualizationVminterface
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VirtualizationVminterface))
	}

	err := c.ICache.(VirtualizationVminterfaceCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VirtualizationVminterfaceCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VirtualizationVminterface))
	}
}

func Test_virtualizationVminterfaceCache_MultiSet(t *testing.T) {
	c := newVirtualizationVminterfaceCache()
	defer c.Close()

	var testData []*model.VirtualizationVminterface
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VirtualizationVminterface))
	}

	err := c.ICache.(VirtualizationVminterfaceCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_virtualizationVminterfaceCache_Del(t *testing.T) {
	c := newVirtualizationVminterfaceCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVminterface)
	err := c.ICache.(VirtualizationVminterfaceCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_virtualizationVminterfaceCache_SetCacheWithNotFound(t *testing.T) {
	c := newVirtualizationVminterfaceCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVminterface)
	err := c.ICache.(VirtualizationVminterfaceCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VirtualizationVminterfaceCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVirtualizationVminterfaceCache(t *testing.T) {
	c := NewVirtualizationVminterfaceCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVirtualizationVminterfaceCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVirtualizationVminterfaceCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
